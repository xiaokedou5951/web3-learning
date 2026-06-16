package main

import (
	"bytes"
	"fmt"
)

// ============================================================================
// Merkle Patricia Trie 核心实现
// ============================================================================

// Trie 是 Merkle Patricia Trie 的主结构
// 它维护一个根节点，并提供插入、查询、删除等操作
type Trie struct {
	root Node // 根节点
}

// NewTrie 创建一个新的空 Trie
func NewTrie() *Trie {
	return &Trie{
		root: &EmptyNode{},
	}
}

// ============================================================================
// 查询操作 (Get)
// ============================================================================

// Get 查询指定键的值
// 这是 MPT 最基本的操作，从根节点开始，根据键的 nibbles 逐层向下查找
func (t *Trie) Get(key []byte) ([]byte, bool) {
	// 将字节键转换为 nibble 序列
	// 例如: 0x1234 -> [1, 2, 3, 4]
	nibbles := bytesToNibbles(key)

	// 从根节点开始递归查找
	value, found := t.get(t.root, nibbles)
	return value, found
}

// get 递归查找节点
// node: 当前节点
// key: 剩余的 nibble 键
func (t *Trie) get(node Node, key []byte) ([]byte, bool) {
	// 空节点，未找到
	if _, ok := node.(*EmptyNode); ok {
		return nil, false
	}

	switch n := node.(type) {
	case *LeafNode:
		// 叶子节点：比较路径是否完全匹配
		if bytes.Equal(n.Path, key) {
			return n.Value, true
		}
		return nil, false

	case *ExtensionNode:
		// 扩展节点：检查路径前缀是否匹配
		prefixLen := commonPrefixLength(n.Path, key)
		if prefixLen < len(n.Path) {
			// 前缀不匹配，说明键不存在
			return nil, false
		}
		// 前缀匹配，继续在下一个节点中查找剩余键
		return t.get(n.Next, key[len(n.Path):])

	case *BranchNode:
		// 分支节点：根据第一个 nibble 选择子节点
		if len(key) == 0 {
			// 键已用完，返回分支节点存储的值
			if n.Value != nil {
				return n.Value, true
			}
			return nil, false
		}

		// 取第一个 nibble 作为索引
		idx := key[0]
		child := n.Children[idx]
		if child == nil {
			return nil, false
		}
		// 递归查找剩余键
		return t.get(child, key[1:])

	case *HashNode:
		// 哈希节点：实际应用中需要从数据库加载完整节点
		// 这里简化处理，返回未找到
		// 实际以太坊实现中，这里会调用 t.db.Get(n.Hash) 获取节点
		fmt.Printf("  [注意] 遇到哈希节点，需要从数据库加载: %x...\n", n.HashValue[:8])
		return nil, false
	}

	return nil, false
}

// ============================================================================
// 更新操作 (Insert/Update)
// ============================================================================

// Insert 插入或更新键值对
// 这是 MPT 最复杂的操作，需要处理各种节点分裂和合并情况
func (t *Trie) Insert(key, value []byte) {
	// 将字节键转换为 nibble 序列
	nibbles := bytesToNibbles(key)

	// 递归插入，返回新的根节点
	t.root = t.insert(t.root, nibbles, value)
}

// insert 递归插入键值对
// node: 当前节点
// key: 剩余的 nibble 键
// value: 要插入的值
// 返回: 新的节点（可能是新创建的，也可能是修改后的）
func (t *Trie) insert(node Node, key, value []byte) Node {
	// 空节点或 nil：直接创建叶子节点
	if node == nil {
		return &LeafNode{
			Path:  key,
			Value: value,
		}
	}
	if _, ok := node.(*EmptyNode); ok {
		return &LeafNode{
			Path:  key,
			Value: value,
		}
	}

	switch n := node.(type) {
	case *LeafNode:
		// 叶子节点：需要处理路径冲突
		return t.insertIntoLeaf(n, key, value)

	case *ExtensionNode:
		// 扩展节点：需要处理路径分裂
		return t.insertIntoExtension(n, key, value)

	case *BranchNode:
		// 分支节点：根据第一个 nibble 递归插入
		return t.insertIntoBranch(n, key, value)

	case *HashNode:
		// 哈希节点：实际应用中需要先加载完整节点
		// 这里简化处理，创建新的叶子节点
		fmt.Printf("  [注意] 遇到哈希节点，需要加载后更新: %x...\n", n.HashValue[:8])
		return &LeafNode{Path: key, Value: value}
	}

	return node
}

// insertIntoLeaf 处理叶子节点的插入
// 情况1: 键完全匹配 -> 更新值
// 情况2: 键部分匹配 -> 创建分支节点
// 情况3: 键完全不匹配 -> 创建分支节点 + 两个叶子
func (t *Trie) insertIntoLeaf(leaf *LeafNode, key, value []byte) Node {
	// 计算公共前缀长度
	commonLen := commonPrefixLength(leaf.Path, key)

	// 情况1: 键完全匹配，直接更新值
	if commonLen == len(leaf.Path) && commonLen == len(key) {
		return &LeafNode{
			Path:  leaf.Path,
			Value: value, // 更新值
		}
	}

	// 情况2和3: 需要创建分支节点
	// 创建新的分支节点
	branch := &BranchNode{}

	// 如果存在公共前缀，需要创建扩展节点
	var extensionPath []byte
	if commonLen > 0 {
		extensionPath = leaf.Path[:commonLen]
	}

	// 处理原有叶子节点的剩余路径
	leafRemaining := leaf.Path[commonLen:]
	if len(leafRemaining) == 0 {
		// 原叶子节点的路径用完了，值放在分支节点
		branch.Value = leaf.Value
	} else {
		// 原叶子节点还有剩余路径，创建新的叶子节点
		// 第一个 nibble 作为分支索引
		idx := leafRemaining[0]
		branch.Children[idx] = &LeafNode{
			Path:  leafRemaining[1:],
			Value: leaf.Value,
		}
	}

	// 处理新键的剩余路径
	keyRemaining := key[commonLen:]
	if len(keyRemaining) == 0 {
		// 新键用完了，值放在分支节点
		branch.Value = value
	} else {
		// 新键还有剩余路径，创建新的叶子节点
		idx := keyRemaining[0]
		branch.Children[idx] = &LeafNode{
			Path:  keyRemaining[1:],
			Value: value,
		}
	}

	// 如果有公共前缀，用扩展节点包装分支节点
	if len(extensionPath) > 0 {
		return &ExtensionNode{
			Path: extensionPath,
			Next: branch,
		}
	}

	return branch
}

// insertIntoExtension 处理扩展节点的插入
// 情况1: 键包含整个扩展路径 -> 递归插入到下一个节点
// 情况2: 键只匹配部分扩展路径 -> 分裂扩展节点
func (t *Trie) insertIntoExtension(ext *ExtensionNode, key, value []byte) Node {
	// 计算公共前缀长度
	commonLen := commonPrefixLength(ext.Path, key)

	// 情况1: 键包含整个扩展路径
	if commonLen == len(ext.Path) {
		// 递归插入到下一个节点
		newNext := t.insert(ext.Next, key[commonLen:], value)
		return &ExtensionNode{
			Path: ext.Path,
			Next: newNext,
		}
	}

	// 情况2: 键只匹配部分扩展路径，需要分裂
	// 创建新的分支节点
	branch := &BranchNode{}

	// 处理扩展节点的剩余路径
	extRemaining := ext.Path[commonLen:]
	if len(extRemaining) == 1 {
		// 扩展路径只剩1个nibble，直接作为分支索引
		idx := extRemaining[0]
		branch.Children[idx] = ext.Next
	} else {
		// 扩展路径还有多个nibble，创建新的扩展节点
		idx := extRemaining[0]
		branch.Children[idx] = &ExtensionNode{
			Path: extRemaining[1:],
			Next: ext.Next,
		}
	}

	// 处理新键的剩余路径
	keyRemaining := key[commonLen:]
	if len(keyRemaining) == 0 {
		// 新键用完了，值放在分支节点
		branch.Value = value
	} else {
		// 新键还有剩余路径
		idx := keyRemaining[0]
		branch.Children[idx] = t.insert(nil, keyRemaining[1:], value)
	}

	// 如果有公共前缀，用扩展节点包装
	if commonLen > 0 {
		return &ExtensionNode{
			Path: ext.Path[:commonLen],
			Next: branch,
		}
	}

	return branch
}

// insertIntoBranch 处理分支节点的插入
func (t *Trie) insertIntoBranch(branch *BranchNode, key, value []byte) Node {
	// 创建新的分支节点（避免修改原节点）
	newBranch := &BranchNode{
		Value: branch.Value,
	}
	copy(newBranch.Children[:], branch.Children[:])

	if len(key) == 0 {
		// 键已用完，更新分支节点的值
		newBranch.Value = value
		return newBranch
	}

	// 根据第一个 nibble 选择子节点
	idx := key[0]
	child := branch.Children[idx]

	// 递归插入到子节点
	if child == nil {
		// 子节点为空，创建新的叶子节点
		newBranch.Children[idx] = &LeafNode{
			Path:  key[1:],
			Value: value,
		}
	} else {
		// 子节点存在，递归插入
		newBranch.Children[idx] = t.insert(child, key[1:], value)
	}

	return newBranch
}

// ============================================================================
// 删除操作 (Delete)
// ============================================================================

// Delete 删除指定键
// 删除后需要合并节点以保持 MPT 的压缩特性
func (t *Trie) Delete(key []byte) bool {
	nibbles := bytesToNibbles(key)
	newRoot, deleted := t.delete(t.root, nibbles)
	if deleted {
		t.root = newRoot
	}
	return deleted
}

// delete 递归删除键
func (t *Trie) delete(node Node, key []byte) (Node, bool) {
	if _, ok := node.(*EmptyNode); ok {
		return node, false
	}

	switch n := node.(type) {
	case *LeafNode:
		if bytes.Equal(n.Path, key) {
			// 找到要删除的叶子节点
			return &EmptyNode{}, true
		}
		return node, false

	case *ExtensionNode:
		prefixLen := commonPrefixLength(n.Path, key)
		if prefixLen < len(n.Path) {
			return node, false
		}
		// 递归删除
		newNext, deleted := t.delete(n.Next, key[len(n.Path):])
		if !deleted {
			return node, false
		}
		// 检查是否需要合并
		return t.mergeExtension(n.Path, newNext), true

	case *BranchNode:
		if len(key) == 0 {
			// 删除分支节点的值
			if n.Value == nil {
				return node, false
			}
			newBranch := &BranchNode{}
			copy(newBranch.Children[:], n.Children[:])
			newBranch.Value = nil
			return t.mergeBranch(newBranch), true
		}

		idx := key[0]
		child := n.Children[idx]
		if child == nil {
			return node, false
		}

		newChild, deleted := t.delete(child, key[1:])
		if !deleted {
			return node, false
		}

		newBranch := &BranchNode{
			Value: n.Value,
		}
		copy(newBranch.Children[:], n.Children[:])
		newBranch.Children[idx] = newChild
		return t.mergeBranch(newBranch), true
	}

	return node, false
}

// mergeExtension 合并扩展节点
func (t *Trie) mergeExtension(path []byte, next Node) Node {
	if _, ok := next.(*EmptyNode); ok {
		return &EmptyNode{}
	}

	// 如果下一个节点也是扩展节点，合并路径
	if ext, ok := next.(*ExtensionNode); ok {
		combinedPath := append(path, ext.Path...)
		return &ExtensionNode{
			Path: combinedPath,
			Next: ext.Next,
		}
	}

	// 如果下一个节点是叶子节点，合并路径
	if leaf, ok := next.(*LeafNode); ok {
		combinedPath := append(path, leaf.Path...)
		return &LeafNode{
			Path: combinedPath,
			Value: leaf.Value,
		}
	}

	return &ExtensionNode{
		Path: path,
		Next: next,
	}
}

// mergeBranch 合并分支节点
// 如果分支节点只剩一个子节点，需要合并
func (t *Trie) mergeBranch(branch *BranchNode) Node {
	// 统计非空子节点数量
	childCount := 0
	var lastChild Node
	var lastIdx int

	for i := 0; i < 16; i++ {
		if branch.Children[i] != nil {
			if _, ok := branch.Children[i].(*EmptyNode); !ok {
				childCount++
				lastChild = branch.Children[i]
				lastIdx = i
			}
		}
	}

	// 如果没有子节点且有值，保留为分支节点
	if childCount == 0 {
		if branch.Value != nil {
			return branch
		}
		return &EmptyNode{}
	}

	// 如果只有一个子节点且没有值，需要合并
	if childCount == 1 && branch.Value == nil {
		// 根据子节点类型合并
		switch child := lastChild.(type) {
		case *LeafNode:
			// 合并为新的叶子节点
			newPath := append([]byte{byte(lastIdx)}, child.Path...)
			return &LeafNode{
				Path:  newPath,
				Value: child.Value,
			}
		case *ExtensionNode:
			// 合并为新的扩展节点
			newPath := append([]byte{byte(lastIdx)}, child.Path...)
			return &ExtensionNode{
				Path: newPath,
				Next: child.Next,
			}
		default:
			// 其他情况，创建扩展节点
			return &ExtensionNode{
				Path: []byte{byte(lastIdx)},
				Next: child,
			}
		}
	}

	return branch
}

// ============================================================================
// 遍历操作
// ============================================================================

// ForEach 遍历所有键值对
func (t *Trie) ForEach(fn func(key, value []byte) bool) {
	t.forEach(t.root, nil, fn)
}

func (t *Trie) forEach(node Node, prefix []byte, fn func(key, value []byte) bool) bool {
	if node == nil {
		return true
	}

	switch n := node.(type) {
	case *EmptyNode:
		return true

	case *LeafNode:
		key := append(prefix, n.Path...)
		return fn(nibblesToBytes(key), n.Value)

	case *ExtensionNode:
		newPrefix := append(prefix, n.Path...)
		return t.forEach(n.Next, newPrefix, fn)

	case *BranchNode:
		if n.Value != nil {
			if !fn(nibblesToBytes(prefix), n.Value) {
				return false
			}
		}
		for i := 0; i < 16; i++ {
			if n.Children[i] != nil {
				newPrefix := append(prefix, byte(i))
				if !t.forEach(n.Children[i], newPrefix, fn) {
					return false
				}
			}
		}
	}

	return true
}

// RootHash 获取根节点哈希
func (t *Trie) RootHash() []byte {
	return t.root.Hash()
}

// Print 打印 Trie 结构
func (t *Trie) Print() {
	fmt.Println("=== Merkle Patricia Trie ===")
	PrintNode(t.root, "")
	fmt.Printf("Root Hash: %x\n", t.RootHash())
	fmt.Println("===========================")
}
