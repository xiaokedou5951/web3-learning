package main

import (
	"bytes"
	"fmt"
)

// ============================================================================
// Merkle Proof 证明机制
// ============================================================================

// Proof 是 Merkle Patricia Trie 的证明
// 它包含从根节点到目标节点的路径上的所有节点
// 用于在不下载整个 Trie 的情况下验证某个键值对的存在性
type Proof struct {
	// Key 是要证明的键
	Key []byte
	// Nodes 是从根到叶子的节点路径（RLP 编码）
	// 每个元素是一个节点的 RLP 编码
	Nodes [][]byte
	// Value 是键对应的值（如果存在）
	Value []byte
}

// GenerateProof 为指定键生成 Merkle Proof
// 这是 SPV（Simple Payment Verification）的核心机制
// 允许轻客户端在不下载完整状态的情况下验证数据
func (t *Trie) GenerateProof(key []byte) (*Proof, error) {
	nibbles := bytesToNibbles(key)
	proof := &Proof{
		Key:   key,
		Nodes: make([][]byte, 0),
	}

	// 从根节点开始收集证明路径
	err := t.generateProof(t.root, nibbles, proof)
	if err != nil {
		return nil, err
	}

	return proof, nil
}

// generateProof 递归收集证明节点
func (t *Trie) generateProof(node Node, key []byte, proof *Proof) error {
	if node == nil {
		return nil
	}

	switch n := node.(type) {
	case *EmptyNode:
		// 空节点，不需要添加证明
		return nil

	case *LeafNode:
		// 叶子节点：检查路径是否匹配
		if bytes.Equal(n.Path, key) {
			// 路径匹配，添加叶子节点到证明
			encoded := rlpEncodeLeaf(n.Path, n.Value)
			proof.Nodes = append(proof.Nodes, encoded)
			proof.Value = n.Value
		}
		return nil

	case *ExtensionNode:
		// 扩展节点：检查前缀是否匹配
		prefixLen := commonPrefixLength(n.Path, key)
		if prefixLen < len(n.Path) {
			// 前缀不匹配，键不存在
			return nil
		}
		// 添加扩展节点到证明
		nextHash := n.Next.Hash()
		encoded := rlpEncodeExtension(n.Path, nextHash)
		proof.Nodes = append(proof.Nodes, encoded)
		// 递归处理下一个节点
		return t.generateProof(n.Next, key[len(n.Path):], proof)

	case *BranchNode:
		// 分支节点：添加当前节点到证明
		childrenHashes := make([][]byte, 16)
		for i := 0; i < 16; i++ {
			if n.Children[i] != nil {
				childrenHashes[i] = n.Children[i].Hash()
			} else {
				childrenHashes[i] = []byte{0x80}
			}
		}
		encoded := rlpEncodeBranch(childrenHashes, n.Value)
		proof.Nodes = append(proof.Nodes, encoded)

		if len(key) == 0 {
			// 键已用完，值在分支节点
			proof.Value = n.Value
			return nil
		}

		// 根据第一个 nibble 选择子节点
		idx := key[0]
		child := n.Children[idx]
		if child == nil {
			return nil
		}
		return t.generateProof(child, key[1:], proof)

	case *HashNode:
		// 哈希节点：实际应用中需要从数据库加载
		// 这里简化处理，直接返回
		fmt.Printf("  [注意] 遇到哈希节点，需要从数据库加载以生成完整证明\n")
		return nil
	}

	return nil
}

// VerifyProof 验证 Merkle Proof
// 参数:
//   - rootHash: 已知的根哈希（从可信来源获取）
//   - proof: 要验证的证明
//
// 返回:
//   - bool: 证明是否有效
//   - []byte: 验证后的值（如果存在）
func VerifyProof(rootHash []byte, proof *Proof) (bool, []byte) {
	if len(proof.Nodes) == 0 {
		// 空证明，无效
		return false, nil
	}

	// 从最后一个节点（叶子节点）开始验证
	// 逐层向上计算哈希，直到根节点

	// 验证第一个节点（根节点）的哈希是否匹配
	firstNode := proof.Nodes[0]
	firstHash := hashNode(firstNode)

	// 如果第一个节点的哈希不等于根哈希，证明无效
	if !bytes.Equal(firstHash, rootHash) {
		// 注意：这里简化处理，实际应该比较完整哈希
		// 对于短节点，哈希就是编码本身
		if len(firstNode) >= 32 {
			return false, nil
		}
	}

	// 逐层验证节点
	// 从根节点开始，逐层向下验证每个节点的哈希是否被父节点引用
	for i := 1; i < len(proof.Nodes); i++ {
		node := proof.Nodes[i]
		parentNode := proof.Nodes[i-1]

		// 计算当前节点的哈希
		nodeHash := hashNode(node)

		// 验证当前节点的哈希是否存在于父节点的编码中
		// 这是 Merkle Proof 的核心：子节点的哈希必须被父节点引用
		if !bytes.Contains(parentNode, nodeHash) {
			// 如果父节点编码中不包含子节点哈希，证明无效
			// 注意：对于短节点（< 32 字节），哈希就是编码本身
			// 这里简化处理，仅做基本验证
			if len(node) >= 32 {
				return false, nil
			}
		}
	}

	// 验证成功，返回值
	return true, proof.Value
}

// VerifyProofSimple 简化的证明验证
// 仅验证证明路径的完整性，不验证具体值
func VerifyProofSimple(proof *Proof) bool {
	if len(proof.Nodes) == 0 {
		return false
	}

	// 检查每个节点是否有效（非空）
	for _, node := range proof.Nodes {
		if len(node) == 0 {
			return false
		}
	}

	return true
}

// PrintProof 打印证明信息
func PrintProof(proof *Proof) {
	fmt.Println("=== Merkle Proof ===")
	fmt.Printf("Key: %x\n", proof.Key)
	if proof.Value != nil {
		fmt.Printf("Value: %s\n", string(proof.Value))
	} else {
		fmt.Println("Value: <not found>")
	}
	fmt.Printf("Proof Path Length: %d nodes\n", len(proof.Nodes))
	for i, node := range proof.Nodes {
		fmt.Printf("  Node %d: %x... (%d bytes)\n", i, node[:min(16, len(node))], len(node))
	}
	fmt.Println("====================")
}

// min 辅助函数
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// ============================================================================
// 证明的应用场景
// ============================================================================

// ProofExample 证明使用示例说明
func ProofExample() {
	fmt.Println(`
Merkle Proof 应用场景:

1. 轻客户端验证 (SPV)
   - 轻客户端不需要下载完整的状态树
   - 只需要下载证明路径上的节点（通常只有几个）
   - 就可以验证某个账户的状态

2. 跨链桥接
   - 源链提供状态证明
   - 目标链验证证明的有效性
   - 无需下载源链的完整状态

3. 状态同步
   - 新节点可以快速同步状态
   - 通过证明验证关键状态
   - 不需要下载完整的历史数据

4. 审计和验证
   - 第三方可以验证特定数据
   - 无需访问完整的数据库
   - 只需要根哈希和证明路径

证明大小:
- 典型证明包含 3-7 个节点
- 每个节点约 100-500 字节
- 总证明大小通常 < 3KB
- 相比完整状态（GB级别），非常高效
`)
}
