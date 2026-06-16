package main

import (
	"bytes"
	"fmt"
	"reflect"

	sha3 "golang.org/x/crypto/sha3"
)

// ============================================================================
// MPT 节点类型定义
// ============================================================================

// Node 是 MPT 中所有节点的接口
// 每个节点都可以计算自己的哈希值，用于构建 Merkle Tree
type Node interface {
	// Hash 计算节点的哈希值
	// 如果节点内容较短（RLP 编码后 < 32 字节），直接返回编码本身
	// 否则返回 SHA3(RLP(node))
	Hash() []byte
}

// EmptyNode 表示空节点
type EmptyNode struct{}

func (n *EmptyNode) Hash() []byte {
	// 空节点的哈希值是固定的：SHA3(RLP(""))
	// 这是以太坊规范中定义的空根哈希
	return emptyRootHash
}

// emptyRootHash 是空 Trie 的根哈希
// 等于 SHA3(RLP("")) = 0x56e81f171bcc55a6ff8345e692c0f86e5b48e01b996cadc001622fb5e363b421
var emptyRootHash = sha3Keccak256([]byte{0x80}) // 0x80 是空字符串的 RLP 编码

// LeafNode 叶子节点：存储键值对
// Path: 压缩路径（nibbles 编码），表示从父节点到当前节点的剩余键
// Value: 存储的实际值（如账户信息）
type LeafNode struct {
	Path  []byte // 压缩路径（nibbles）
	Value []byte // 存储的值
}

func (n *LeafNode) Hash() []byte {
	// 叶子节点的哈希 = SHA3(RLP([path, value]))
	encoded := rlpEncodeLeaf(n.Path, n.Value)
	return hashNode(encoded)
}

// ExtensionNode 扩展节点：压缩共享路径
// Path: 共享的路径片段（nibbles）
// Next: 指向下一个节点的哈希或内联节点
type ExtensionNode struct {
	Path []byte // 共享路径（nibbles）
	Next Node   // 下一个节点（通常是 Branch 或 Leaf）
}

func (n *ExtensionNode) Hash() []byte {
	// 扩展节点的哈希 = SHA3(RLP([path, next_hash]))
	nextHash := n.Next.Hash()
	encoded := rlpEncodeExtension(n.Path, nextHash)
	return hashNode(encoded)
}

// BranchNode 分支节点：16 叉树结构
// Children: 16 个子节点（对应 16 个 nibble: 0-f）
// Value: 可选的值（如果键在此结束）
type BranchNode struct {
	Children [16]Node // 16 个子节点
	Value    []byte   // 可选值
}

func (n *BranchNode) Hash() []byte {
	// 分支节点的哈希 = SHA3(RLP([child0, child1, ..., child15, value]))
	childrenHashes := make([][]byte, 16)
	for i := 0; i < 16; i++ {
		if n.Children[i] != nil {
			childrenHashes[i] = n.Children[i].Hash()
		} else {
			childrenHashes[i] = []byte{0x80} // 空节点的 RLP 编码
		}
	}
	encoded := rlpEncodeBranch(childrenHashes, n.Value)
	return hashNode(encoded)
}

// HashNode 存储子节点的哈希引用
// 当子节点的 RLP 编码 >= 32 字节时，存储其哈希而不是完整节点
// 这是 MPT 节省空间的关键机制
type HashNode struct {
	HashValue []byte // 子节点的 32 字节哈希
}

func (n *HashNode) Hash() []byte {
	return n.HashValue
}

// ============================================================================
// RLP 编码辅助函数（简化版本，用于教学）
// ============================================================================

// rlpEncodeLeaf 编码叶子节点
// 格式: [path, value]
// path 使用 HP 编码（Hex-Prefix）添加前缀标识叶子节点
func rlpEncodeLeaf(path, value []byte) []byte {
	// HP 编码：叶子节点前缀为 0x20
	// 如果路径长度为奇数，前缀为 0x3
	hpPath := hexPrefixEncode(path, true)
	return rlpEncodeList([][]byte{hpPath, value})
}

// rlpEncodeExtension 编码扩展节点
// 格式: [path, next_hash]
func rlpEncodeExtension(path, nextHash []byte) []byte {
	// HP 编码：扩展节点前缀为 0x00
	hpPath := hexPrefixEncode(path, false)
	return rlpEncodeList([][]byte{hpPath, nextHash})
}

// rlpEncodeBranch 编码分支节点
// 格式: [child0, child1, ..., child15, value]
func rlpEncodeBranch(children [][]byte, value []byte) []byte {
	items := make([][]byte, 17)
	for i := 0; i < 16; i++ {
		items[i] = children[i]
	}
	if value == nil {
		items[16] = []byte{0x80} // 空值
	} else {
		items[16] = value
	}
	return rlpEncodeList(items)
}

// hexPrefixEncode 实现 HP 编码（Hex-Prefix Encoding）
// HP 编码用于将 nibble 路径转换为字节，同时标识节点类型
//
// 编码规则：
// - 第一个 nibble 标识节点类型和路径长度奇偶性
// - 叶子节点偶数长度: 0x00 + path
// - 叶子节点奇数长度: 0x1 + path[0]
// - 扩展节点偶数长度: 0x00 + path
// - 扩展节点奇数长度: 0x0 + path[0]
//
// 实际上，以太坊使用以下格式：
// - flag = 2 * isLeaf + (len(path) % 2)
// - 如果 flag >= 2: 前缀 = flag
// - 如果 flag < 2: 前缀 = 0x00 + flag
func hexPrefixEncode(path []byte, isLeaf bool) []byte {
	// 计算 flag
	flag := 0
	if isLeaf {
		flag += 2
	}
	if len(path)%2 == 1 {
		flag += 1
	}

	// 构建编码后的路径
	result := make([]byte, 0, (len(path)+2)/2+1)

	if flag >= 2 {
		// 奇数长度：第一个字节包含 flag 和第一个 nibble
		if len(path) > 0 {
			result = append(result, byte(flag<<4)|path[0])
			path = path[1:]
		} else {
			result = append(result, byte(flag<<4))
		}
	} else {
		// 偶数长度：第一个字节是 flag，第二个字节开始是路径
		result = append(result, byte(flag<<4))
	}

	// 将剩余的 nibbles 两两组合成字节
	for i := 0; i < len(path); i += 2 {
		if i+1 < len(path) {
			result = append(result, path[i]<<4|path[i+1])
		} else {
			result = append(result, path[i]<<4)
		}
	}

	return result
}

// rlpEncodeList 简化的 RLP 列表编码
// 实际以太坊使用完整的 RLP 编码，这里简化用于教学
func rlpEncodeList(items [][]byte) []byte {
	// 计算总长度
	totalLen := 0
	for _, item := range items {
		totalLen += len(item)
	}

	// 构建列表编码
	// RLP 列表格式: [prefix][length][items...]
	// 如果总长度 < 55: prefix = 0xc0 + length
	// 否则: prefix = 0xf7 + len(length), 然后跟长度
	result := make([]byte, 0, totalLen+10)

	if totalLen < 55 {
		result = append(result, byte(0xc0+totalLen))
	} else {
		// 简化：假设长度不会超过 255
		result = append(result, 0xf8, byte(totalLen))
	}

	for _, item := range items {
		result = append(result, item...)
	}

	return result
}

// hashNode 计算节点哈希
// 规则：
// - 如果 RLP 编码长度 < 32 字节，直接返回编码（内联节点）
// - 否则返回 SHA3(RLP(node))
func hashNode(encoded []byte) []byte {
	if len(encoded) < 32 {
		// 短节点直接内联，不计算哈希
		return encoded
	}
	// 长节点计算 Keccak256 哈希
	return sha3Keccak256(encoded)
}

// sha3Keccak256 计算 Keccak256 哈希
// 以太坊使用 Keccak256（不是标准的 SHA3-256）
// Keccak256 是 SHA-3 标准的前身，两者有细微差别
// 使用 golang.org/x/crypto/sha3 提供的 Keccak256
func sha3Keccak256(data []byte) []byte {
	h := sha3.NewLegacyKeccak256()
	h.Write(data)
	return h.Sum(nil)
}

// ============================================================================
// 辅助函数
// ============================================================================

// bytesToNibbles 将字节切片转换为 nibble 切片
// 每个字节转换为两个 nibble（高4位和低4位）
// 例如: 0x1a -> [0x1, 0xa]
func bytesToNibbles(data []byte) []byte {
	nibbles := make([]byte, len(data)*2)
	for i, b := range data {
		nibbles[i*2] = b >> 4     // 高4位
		nibbles[i*2+1] = b & 0x0f // 低4位
	}
	return nibbles
}

// nibblesToBytes 将 nibble 切片转换回字节切片
// 必须是偶数长度
func nibblesToBytes(nibbles []byte) []byte {
	if len(nibbles)%2 != 0 {
		panic("nibbles length must be even")
	}
	result := make([]byte, len(nibbles)/2)
	for i := 0; i < len(nibbles); i += 2 {
		result[i/2] = nibbles[i]<<4 | nibbles[i+1]
	}
	return result
}

// commonPrefixLength 计算两个 nibble 切片的公共前缀长度
// 用于确定路径的共享部分
func commonPrefixLength(a, b []byte) int {
	maxLen := len(a)
	if len(b) < maxLen {
		maxLen = len(b)
	}
	for i := 0; i < maxLen; i++ {
		if a[i] != b[i] {
			return i
		}
	}
	return maxLen
}

// PrintNode 打印节点信息（用于调试）
func PrintNode(node Node, indent string) {
	if node == nil {
		fmt.Printf("%s<nil>\n", indent)
		return
	}

	switch n := node.(type) {
	case *EmptyNode:
		fmt.Printf("%sEmptyNode\n", indent)
	case *LeafNode:
		fmt.Printf("%sLeafNode{path: %x, value: %s}\n", indent, n.Path, string(n.Value))
	case *ExtensionNode:
		fmt.Printf("%sExtensionNode{path: %x}\n", indent, n.Path)
		PrintNode(n.Next, indent+"  ")
	case *BranchNode:
		fmt.Printf("%sBranchNode{\n", indent)
		for i := 0; i < 16; i++ {
			if n.Children[i] != nil {
				fmt.Printf("%s  [%x]:\n", indent, i)
				PrintNode(n.Children[i], indent+"    ")
			}
		}
		if n.Value != nil {
			fmt.Printf("%s  value: %s\n", indent, string(n.Value))
		}
		fmt.Printf("%s}\n", indent)
	case *HashNode:
		fmt.Printf("%sHashNode{%x}\n", indent, n.HashValue[:8])
	default:
		fmt.Printf("%sUnknown: %v\n", indent, reflect.TypeOf(node))
	}
}

// CompareNodes 比较两个节点是否相等
func CompareNodes(a, b Node) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return bytes.Equal(a.Hash(), b.Hash())
}
