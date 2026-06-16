# 比特币 Merkle Tree (默克尔树)

## 什么是 Merkle Tree?

Merkle Tree 是一种二叉树结构，用于高效且安全地验证大量数据的完整性。比特币使用 Merkle Tree 将区块中的所有交易压缩成一个 32 字节的 Merkle Root。

## Merkle Tree 的结构

```
              [Merkle Root]
              /         \
         [Hash AB]    [Hash CD]
         /      \      /      \
      [Hash A] [Hash B] [Hash C] [Hash D]
```

- **叶子节点 (Hash A, Hash B...)**: 对原始交易执行双重 SHA256 后得到的交易哈希 (txid)
- **中间节点 (Hash AB, Hash CD...)**: 对子节点拼接后的双重 SHA256 哈希
- **Merkle Root**: 树的顶端节点，代表整个区块所有交易的唯一指纹

## 为什么比特币使用 Merkle Tree?

1. **快速验证**: 只需下载 Merkle Root 和验证路径即可确认某笔交易是否存在

### 快速验证示例（Merkle Proof）

假设你想验证交易 B 是否存在于区块中，**不需要下载所有交易**，只需要：
- Merkle Root（从区块头获取，32 字节）
- Hash B（你要验证的交易哈希）
- Hash A（B 的兄弟节点）
- Hash CD（B 所在子树对面的兄弟节点）

验证过程：
```
步骤 1: 计算 Hash AB = SHA256(SHA256(Hash A + Hash B))
步骤 2: 计算 Root' = SHA256(SHA256(Hash AB + Hash CD))
步骤 3: 比较 Root' 是否等于区块头中的 Merkle Root
```

如果相等，说明交易 B 确实存在于该区块中。

**为什么高效？** 对于包含 1024 笔交易的区块，完整下载需要 1024 个哈希，而 Merkle Proof 只需 `log2(1024) = 10` 个哈希值即可完成验证。

2. **数据完整性**: Merkle Root 的任何微小变化都意味着数据被篡改
3. **SPV 钱包**: 轻节点只需存储区块头（包含 Merkle Root），无需下载完整区块
4. **节省空间**: 32 字节的 Merkle Root 可以代表成千上万笔交易

## 核心概念

### 交易数据 vs 交易哈希

Merkle 树的**叶子节点是交易哈希**，而非原始交易数据：

```
原始交易数据 --[doubleSHA256]--> 交易哈希 (txid) --[作为叶子节点构建 Merkle 树]--> Merkle Root
```

### 双重 SHA256

比特币对所有数据都使用双重 SHA256 哈希：
```
Hash = SHA256(SHA256(data))
```

为什么要双重哈希？防止长度扩展攻击 (Length Extension Attack)。

### 奇数节点处理

当某层节点数量为奇数时，**复制最后一个节点**配对：

```
      [Merkle Root]
         /     \
     [H0-1]   [H2-2]   ← H2 与自身的副本配对
     /    \     /   \
   [H0]  [H1] [H2] [H2*]
```

### 小端字节序

比特币内部存储哈希值时使用**小端字节序**，所以显示时需要反转：
```
正常显示: 982051fd1e4ba744bbbe680e1fee14677ba1a3c3540bf7b1cdb606e857233e0e
小端存储: 0e3e2357e806b1cdb60b0f54c3a1ab766714ee1f0e68bebb44a74b1efd512098
```

## 算法步骤

1. 对所有交易执行双重 SHA256，得到叶子节点哈希
2. 如果节点数为奇数，复制最后一个节点
3. 两两配对：`ParentHash = SHA256(SHA256(LeftHash + RightHash))`
4. 重复步骤 2-3，直到只剩一个节点，即为 Merkle Root

## 代码示例

运行 `go run merkle.go` 查看完整的 Merkle Root 计算过程。
