# Merkle Patricia Trie (MPT) - 教学实现

本项目使用 Go 语言实现了以太坊的核心数据结构 **Merkle Patricia Trie**，用于教学目的，帮助理解世界状态树的构建、查询、更新和证明原理。

## 目录结构

```
Merkle-Patricia-Trie/
├── node.go      # 节点类型定义与编码
├── trie.go      # Trie 的核心操作（构建、查询、更新、删除）
├── proof.go     # Merkle Proof 生成与验证
├── main.go      # 完整示例演示
├── go.mod       # Go 模块定义
└── README.md    # 说明文档
```

## 快速开始

### 运行环境

- Go 1.21+
- 依赖：`golang.org/x/crypto`（用于 Keccak256 哈希）

### 安装依赖

```bash
go mod tidy
```

### 运行示例

```bash
go run .
```

## Merkle Patricia Trie 原理

### 什么是 MPT？

MPT 是以太坊用于存储世界状态的核心数据结构，它结合了三种数据结构的优势：

1. **Merkle Tree** - 提供密码学证明，验证数据完整性
2. **Patricia Trie** - 压缩公共前缀，节省存储空间
3. **Radix Trie** - 高效的路径查找

### 节点类型

#### 1. Empty Node（空节点）

表示 Trie 为空，当 Trie 中没有任何键值对时使用。

```go
type EmptyNode struct{}
```

#### 2. Leaf Node（叶子节点）

存储最终的键值对，包含两个字段：

- `Path`：剩余路径（nibbles 编码）
- `Value`：存储的值

```go
type LeafNode struct {
    Path  []byte
    Value []byte
}
```

#### 3. Extension Node（扩展节点）

压缩共享路径，当多个键共享相同前缀时使用，避免创建不必要的分支节点。

- `Path`：共享路径片段
- `Next`：指向下一个节点

```go
type ExtensionNode struct {
    Path []byte
    Next Node
}
```

#### 4. Branch Node（分支节点）

16 叉树结构，当路径出现分叉时使用。

- `Children`：16 个子节点（对应 0-f 共 16 个 nibble）
- `Value`：可选值（如果键在此结束）

```go
type BranchNode struct {
    Children [16]Node
    Value    []byte
}
```

#### 5. Hash Node（哈希节点）

存储子节点的哈希引用，当子节点的 RLP 编码 >= 32 字节时，只存储哈希而不存储完整节点，这是 MPT 节省空间的关键机制。

```go
type HashNode struct {
    HashValue []byte
}
```

### 路径编码

#### Nibble 编码

将每个字节拆分为两个 4-bit 值（nibble）：

```
0x12 34 -> [1, 2, 3, 4]
```

这是 MPT 的基础，所有操作都基于 nibble 进行。

#### HP 编码（Hex-Prefix Encoding）

用于标识节点类型和存储路径：

| 前缀   | 含义           | 路径长度 |
|--------|----------------|----------|
| `0x00` | 扩展节点       | 偶数     |
| `0x01` | 扩展节点       | 奇数     |
| `0x20` | 叶子节点       | 偶数     |
| `0x3`  | 叶子节点       | 奇数     |

编码规则：
```
flag = 2 * isLeaf + (len(path) % 2)
```

### 哈希机制

MPT 使用 Keccak256 哈希算法（注意：不是标准的 SHA3-256，而是其前身）：

- **节点 RLP 编码 < 32 字节**：直接内联（不计算哈希）
- **节点 RLP 编码 >= 32 字节**：存储 `Keccak256(RLP(node))`

这种机制确保了 Trie 的紧凑性，同时提供密码学安全性。

## 核心操作

### 1. 构建（Insert）

插入键值对的过程：

1. 将键转换为 nibble 序列
2. 从根节点开始，逐层向下查找
3. 根据当前节点类型处理：
   - **空节点**：直接创建叶子节点
   - **叶子节点**：计算公共前缀，可能需要创建分支节点
   - **扩展节点**：检查前缀匹配，可能需要分裂
   - **分支节点**：根据第一个 nibble 选择子节点继续

```go
trie.Insert([]byte{0x12, 0x34}, []byte("Alice"))
```

### 2. 查询（Get）

查询过程相对简单：

1. 将键转换为 nibble 序列
2. 从根节点开始，根据 nibbles 逐层向下
3. 到达叶子节点后，检查路径是否完全匹配

```go
value, found := trie.Get([]byte{0x12, 0x34})
```

### 3. 更新（Update）

更新操作与插入相同：
- 找到匹配的键则更新值
- 未找到则插入新键值对

```go
trie.Insert([]byte{0x12, 0x34}, []byte("Alice Updated"))
```

### 4. 删除（Delete）

删除操作需要处理节点合并：

1. 找到要删除的叶子节点
2. 删除后检查是否需要合并：
   - 分支节点只剩一个子节点 -> 合并为扩展或叶子节点
   - 扩展节点的下一个节点为空 -> 删除扩展节点

```go
trie.Delete([]byte{0x12, 0x34})
```

### 5. Merkle Proof（证明）

生成证明：

```go
proof, err := trie.GenerateProof([]byte{0x12, 0x34})
```

验证证明：

```go
valid, value := VerifyProof(rootHash, proof)
```

**证明原理**：
- 提供从根节点到目标节点的路径上所有节点
- 验证者只需计算根哈希并与已知根哈希比较
- 如果匹配，证明数据存在且未被篡改

## 以太坊世界状态树

### 结构

```
世界状态树 (State Trie)
├── 键: Keccak256(地址)
└── 值: RLP 编码的账户信息
    ├── nonce: 交易计数器
    ├── balance: 账户余额
    ├── storageRoot: 存储树根哈希
    └── codeHash: 合约代码哈希
```

### 区块中的使用

每个以太坊区块头部包含：

- `stateRoot`：世界状态树的根哈希
- `transactionsRoot`：交易树的根哈希
- `receiptsRoot`：收据树的根哈希

```
Block Header
├── stateRoot: 0x...  ← 世界状态树根哈希
├── transactionsRoot: 0x...
├── receiptsRoot: 0x...
└── ...
```

### 状态变化

当执行交易时：

1. 查询发送者账户状态
2. 更新 nonce 和 balance
3. 可能更新接收者账户状态
4. 计算新的状态根哈希

```
交易前状态:
  addr1: {nonce: 1, balance: 1000}
  addr2: {nonce: 0, balance: 500}
  stateRoot: 0xabc...

交易：addr1 转 100 ETH 给 addr2

交易后状态:
  addr1: {nonce: 2, balance: 900}
  addr2: {nonce: 0, balance: 600}
  stateRoot: 0xdef...  ← 新的状态根
```

## 应用场景

### 1. 轻客户端验证（SPV）

轻客户端不需要下载完整的状态树，只需：

1. 获取区块头部的 `stateRoot`
2. 请求完整节点提供 Merkle Proof
3. 本地验证 Proof 的有效性
4. 确认账户状态

### 2. 跨链桥接

- 源链提供状态证明
- 目标链验证证明的有效性
- 无需下载源链的完整状态

### 3. 状态同步

- 新节点可以快速同步状态
- 通过证明验证关键状态
- 不需要下载完整的历史数据

### 4. 审计和验证

- 第三方可以验证特定数据
- 无需访问完整的数据库
- 只需要根哈希和证明路径

## 代码示例

### 基本操作

```go
// 创建 Trie
trie := NewTrie()

// 插入
trie.Insert([]byte("key1"), []byte("value1"))
trie.Insert([]byte("key2"), []byte("value2"))

// 查询
if value, found := trie.Get([]byte("key1")); found {
    fmt.Printf("Found: %s\n", value)
}

// 更新
trie.Insert([]byte("key1"), []byte("new value"))

// 删除
trie.Delete([]byte("key2"))

// 获取根哈希
rootHash := trie.RootHash()
```

### 生成和验证证明

```go
// 生成证明
proof, _ := trie.GenerateProof([]byte("key1"))

// 验证证明
valid, value := VerifyProof(rootHash, proof)
if valid {
    fmt.Printf("Verified value: %s\n", value)
}
```

### 遍历 Trie

```go
trie.ForEach(func(key, value []byte) bool {
    fmt.Printf("Key: %x -> Value: %s\n", key, value)
    return true // 返回 false 可停止遍历
})
```

## 证明大小估算

典型证明包含 3-7 个节点，每个节点约 100-500 字节：

| 深度 | 证明大小 |
|------|----------|
| 3    | ~450 B   |
| 5    | ~750 B   |
| 7    | ~1.1 KB  |

相比完整状态（GB 级别），非常高效。

## 学习建议

1. **理解 nibble 编码**
   - 这是 MPT 的基础
   - 每个字节拆分为两个 4-bit 值

2. **掌握节点类型转换**
   - 叶子 <-> 分支
   - 扩展 <-> 分支
   - 理解何时需要分裂和合并

3. **理解哈希机制**
   - 为什么 < 32 字节要内联？
   - 如何节省存储空间？

4. **实践 Merkle 证明**
   - 理解 SPV 原理
   - 如何实现轻客户端验证

5. **参考 go-ethereum 源码**
   - `core/state/statedb.go`
   - `trie/trie.go`
   - `trie/proof.go`

## 注意事项

1. **RLP 编码**：本项目使用简化的 RLP 编码用于教学，实际以太坊使用完整的 RLP 规范

2. **Keccak256**：以太坊使用 Keccak256（SHA-3 的前身），不是标准的 SHA3-256

3. **数据库存储**：实际实现中，节点会持久化到 LevelDB/Badger 等数据库

4. **不可变性**：MPT 是不可变的，每次更新都会创建新的节点，旧节点保持不变

## 参考资源

- [Ethereum Yellow Paper](https://ethereum.org/en/developers/docs/whitepaper/)
- [Merkle Patricia Trie - Ethereum Wiki](https://ethereum.org/en/developers/docs/data-structures-and-encoding/patricia-merkle-trie/)
- [go-ethereum Trie 实现](https://github.com/ethereum/go-ethereum/tree/master/trie)
