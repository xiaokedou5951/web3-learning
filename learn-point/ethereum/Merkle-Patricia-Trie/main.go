package main

import (
	"fmt"
)

// ============================================================================
// 主程序 - Merkle Patricia Trie 教学演示
// ============================================================================

func main() {
	fmt.Println(`
================================================================================
以太坊世界状态树 - Merkle Patricia Trie 教学实现
================================================================================

MPT (Merkle Patricia Trie) 是以太坊的核心数据结构，用于存储世界状态。
它结合了三种数据结构的优势：

1. Merkle Tree  - 提供密码学证明，验证数据完整性
2. Patricia Trie - 压缩公共前缀，节省存储空间
3. Radix Trie   - 高效的路径查找

--------------------------------------------------------------------------------
节点类型
--------------------------------------------------------------------------------

1. Empty Node (空节点)
   - 表示 Trie 为空

2. Leaf Node (叶子节点)
   - 存储最终的键值对
   - Path: 剩余路径（nibbles 编码）
   - Value: 存储的值

3. Extension Node (扩展节点)
   - 压缩共享路径
   - Path: 共享路径片段
   - Next: 指向下一个节点

4. Branch Node (分支节点)
   - 16 叉树结构
   - Children: 16 个子节点（对应 0-f）
   - Value: 可选值

5. Hash Node (哈希节点)
   - 存储子节点的哈希引用
   - 当子节点 RLP 编码 >= 32 字节时使用

--------------------------------------------------------------------------------
核心操作
--------------------------------------------------------------------------------

1. 构建 (Insert)
   - 从根节点开始，根据键的 nibbles 逐层插入
   - 遇到路径冲突时，创建分支节点

2. 查询 (Get)
   - 从根节点开始，根据键的 nibbles 逐层查找
   - 时间复杂度: O(log n)

3. 更新 (Update)
   - 与插入相同，找到匹配路径后更新值
   - 保持 MPT 的不可变特性

4. 删除 (Delete)
   - 删除后需要合并节点以保持压缩特性

5. 证明 (Proof)
   - 生成从根到叶子的路径证明
   - 验证数据存在性而无需完整 Trie

--------------------------------------------------------------------------------
HP 编码 (Hex-Prefix Encoding)
--------------------------------------------------------------------------------

MPT 使用 HP 编码来标识节点类型和存储路径：

- 0x00... 前缀标识扩展节点（偶数长度）
- 0x1...  前缀标识扩展节点（奇数长度）
- 0x20... 前缀标识叶子节点（偶数长度）
- 0x3...  前缀标识叶子节点（奇数长度）

规则:
- flag = 2 * isLeaf + (len(path) % 2)
- 如果 flag >= 2: 第一个字节 = flag << 4 | path[0]
- 如果 flag < 2: 第一个字节 = 0x00, 第二个字节 = flag << 4 | path[0]

--------------------------------------------------------------------------------
哈希机制
--------------------------------------------------------------------------------

- 节点 RLP 编码 < 32 字节: 直接内联（不计算哈希）
- 节点 RLP 编码 >= 32 字节: 存储 Keccak256 哈希

这确保了 Trie 的紧凑性，同时提供密码学安全性。

================================================================================
运行示例
================================================================================
`)

	// ========================================================================
	// 示例 1: 基本操作
	// ========================================================================
	fmt.Println("--- 示例 1: 基本插入和查询 ---\n")

	// 创建一个新的 Trie
	trie := NewTrie()

	// 插入键值对
	fmt.Println("插入键值对:")

	trie.Insert([]byte{0x12, 0x34}, []byte("Alice"))
	fmt.Println("  Key: 0x1234 -> Value: Alice")

	trie.Insert([]byte{0x12, 0x35}, []byte("Bob"))
	fmt.Println("  Key: 0x1235 -> Value: Bob")

	trie.Insert([]byte{0x56, 0x78}, []byte("Charlie"))
	fmt.Println("  Key: 0x5678 -> Value: Charlie")

	// 查询
	fmt.Println("\n查询操作:")

	if value, found := trie.Get([]byte{0x12, 0x34}); found {
		fmt.Printf("  Get 0x1234: %s (找到)\n", string(value))
	} else {
		fmt.Println("  Get 0x1234: 未找到")
	}

	if value, found := trie.Get([]byte{0x12, 0x35}); found {
		fmt.Printf("  Get 0x1235: %s (找到)\n", string(value))
	} else {
		fmt.Println("  Get 0x1235: 未找到")
	}

	if value, found := trie.Get([]byte{0x99, 0x99}); found {
		fmt.Printf("  Get 0x9999: %s (找到)\n", string(value))
	} else {
		fmt.Println("  Get 0x9999: 未找到 (正确，未插入)")
	}

	// 打印 Trie 结构
	fmt.Println("\nTrie 结构:")
	trie.Print()

	// 更新值
	fmt.Println("\n更新键值对:")
	trie.Insert([]byte{0x12, 0x34}, []byte("Alice Updated"))
	fmt.Println("  更新 0x1234 -> Alice Updated")

	if value, found := trie.Get([]byte{0x12, 0x34}); found {
		fmt.Printf("  Get 0x1234: %s (找到)\n", string(value))
	}

	// ========================================================================
	// 示例 2: 删除操作
	// ========================================================================
	fmt.Println("\n--- 示例 2: 删除操作 ---\n")

	trie2 := NewTrie()
	trie2.Insert([]byte{0x12, 0x34}, []byte("A"))
	trie2.Insert([]byte{0x12, 0x35}, []byte("B"))
	trie2.Insert([]byte{0x12, 0x36}, []byte("C"))

	fmt.Println("删除前:")
	trie2.Print()

	fmt.Println("删除 0x1235:")
	trie2.Delete([]byte{0x12, 0x35})

	fmt.Println("删除后:")
	trie2.Print()

	// ========================================================================
	// 示例 3: Merkle Proof
	// ========================================================================
	fmt.Println("\n--- 示例 3: Merkle Proof ---\n")

	trie3 := NewTrie()
	trie3.Insert([]byte{0x12, 0x34}, []byte("Value1"))
	trie3.Insert([]byte{0x56, 0x78}, []byte("Value2"))
	trie3.Insert([]byte{0x9A, 0xBC}, []byte("Value3"))

	rootHash := trie3.RootHash()
	fmt.Printf("根哈希: %x\n\n", rootHash)

	// 生成证明
	fmt.Println("为 Key 0x1234 生成 Merkle Proof:")
	proof, err := trie3.GenerateProof([]byte{0x12, 0x34})
	if err != nil {
		fmt.Printf("  生成证明失败: %v\n", err)
	} else {
		PrintProof(proof)

		// 验证证明
		valid, value := VerifyProof(rootHash, proof)
		fmt.Printf("证明验证结果: %v\n", valid)
		if valid {
			fmt.Printf("验证后的值: %s\n", string(value))
		}
	}

	// ========================================================================
	// 示例 4: 遍历
	// ========================================================================
	fmt.Println("\n--- 示例 4: 遍历 Trie ---\n")

	fmt.Println("遍历所有键值对:")
	trie.ForEach(func(key, value []byte) bool {
		fmt.Printf("  Key: %x -> Value: %s\n", key, string(value))
		return true // 返回 false 可停止遍历
	})

	// ========================================================================
	// 示例 5: 世界状态树模拟
	// ========================================================================
	fmt.Println("\n--- 示例 5: 以太坊世界状态树模拟 ---\n")

	// 模拟账户状态
	type Account struct {
		Nonce    uint64
		Balance  uint64
		CodeHash []byte
	}

	// 创建状态树
	stateTrie := NewTrie()

	// 添加账户（键为地址的简化表示）
	accounts := map[string]*Account{
		"addr1": {Nonce: 1, Balance: 1000, CodeHash: []byte("code1")},
		"addr2": {Nonce: 0, Balance: 500, CodeHash: []byte("code2")},
		"addr3": {Nonce: 5, Balance: 2000, CodeHash: []byte("code3")},
	}

	fmt.Println("插入账户状态:")
	for addr, acc := range accounts {
		// 实际以太坊中，账户信息使用 RLP 编码
		accountData := []byte(fmt.Sprintf("{nonce:%d, balance:%d, codeHash:%x}",
			acc.Nonce, acc.Balance, acc.CodeHash))
		stateTrie.Insert([]byte(addr), accountData)
		fmt.Printf("  %s -> %s\n", addr, string(accountData))
	}

	fmt.Printf("\n状态根哈希: %x\n", stateTrie.RootHash())

	// 查询账户状态
	fmt.Println("\n查询账户状态:")
	if value, found := stateTrie.Get([]byte("addr1")); found {
		fmt.Printf("  addr1: %s\n", string(value))
	}

	// 为账户状态生成证明
	fmt.Println("\n为 addr1 生成状态证明:")
	stateProof, _ := stateTrie.GenerateProof([]byte("addr1"))
	PrintProof(stateProof)

	// 验证状态证明
	valid, val := VerifyProof(stateTrie.RootHash(), stateProof)
	fmt.Printf("状态证明验证: %v, 值: %s\n", valid, string(val))

	fmt.Println(`
================================================================================
学习要点总结
================================================================================

1. MPT 是以太坊状态管理的核心数据结构

2. 三种主要节点类型:
   - Leaf: 存储键值对
   - Extension: 压缩共享路径
   - Branch: 16 叉分支

3. HP 编码用于标识节点类型:
   - 0x[0-1]... 前缀标识扩展节点
   - 0x[2-3]... 前缀标识叶子节点

4. 哈希机制:
   - 节点 RLP 编码 < 32 字节：直接内联
   - 节点 RLP 编码 >= 32 字节：存储 Keccak256 哈希

5. Merkle 证明:
   - 只需提供从根到叶子的路径节点
   - 轻客户端可验证数据存在性
   - 证明大小通常 < 3KB

6. 世界状态树:
   - 键：地址的 Keccak256 哈希
   - 值：RLP 编码的账户信息
   - 每个区块都有独立的状态树
   - 通过状态根哈希链接

================================================================================
学习建议
================================================================================

1. 理解 nibble 编码
   - 这是 MPT 的基础
   - 每个字节拆分为两个 4-bit 值

2. 掌握节点类型转换
   - 叶子 <-> 分支
   - 扩展 <-> 分支
   - 理解何时需要分裂和合并

3. 理解哈希机制
   - 为什么 < 32 字节要内联？
   - 如何节省存储空间？

4. 实践 Merkle 证明
   - 理解 SPV 原理
   - 如何实现轻客户端验证

5. 参考 go-ethereum 源码
   - core/state/statedb.go
   - trie/trie.go
   - trie/proof.go

================================================================================
`)
}
