package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

// doubleSHA256 执行双重 SHA256 哈希运算
// 比特币中使用双重哈希是为了增加安全性，防止长度扩展攻击
func doubleSHA256(data []byte) []byte {
	// 第一次 SHA256 哈希
	firstHash := sha256.Sum256(data)
	// 第二次 SHA256 哈希（对第一次的结果再次哈希）
	secondHash := sha256.Sum256(firstHash[:])
	return secondHash[:]
}

// hashTransaction 对交易数据进行哈希计算
// 在比特币中，交易 ID (txid) 就是交易数据的双重 SHA256 哈希值
func hashTransaction(txData string) []byte {
	// 将十六进制字符串转换为字节数组
	txBytes, _ := hex.DecodeString(txData)
	// 返回双重 SHA256 哈希结果
	return doubleSHA256(txBytes)
}

// calculateMerkleRoot 计算 Merkle Root
// txHashes: 交易哈希列表（字节数组切片）
// 返回: Merkle Root 的字节数组
func calculateMerkleRoot(txHashes [][]byte) []byte {
	// 边界条件：如果没有交易，返回空哈希
	if len(txHashes) == 0 {
		return make([]byte, 32)
	}

	// 边界条件：如果只有一个交易，它的哈希就是 Merkle Root
	if len(txHashes) == 1 {
		return txHashes[0]
	}

	// 如果交易数量是奇数，复制最后一个交易哈希添加到末尾
	// 这是比特币 Merkle 树的特殊规则：奇数节点时复制最后一个节点
	if len(txHashes)%2 == 1 {
		lastHash := txHashes[len(txHashes)-1]
		txHashes = append(txHashes, lastHash)
	}

	// 创建下一层的哈希列表
	// 每一层的节点数量是上一层的一半
	nextLevel := make([][]byte, 0, len(txHashes)/2)

	// 两两配对计算哈希
	for i := 0; i < len(txHashes); i += 2 {
		// 将两个相邻节点的哈希拼接
		combined := append(txHashes[i], txHashes[i+1]...)
		// 对拼接后的数据进行双重 SHA256 哈希，得到父节点
		parentHash := doubleSHA256(combined)
		nextLevel = append(nextLevel, parentHash)
	}

	// 递归计算上一层的 Merkle Root
	return calculateMerkleRoot(nextLevel)
}

// bytesToHex 将字节数组转换为十六进制字符串（便于阅读）
func bytesToHex(data []byte) string {
	return hex.EncodeToString(data)
}

func main() {
	// 示例：模拟比特币区块中的交易
	// 这些是简化的交易数据（实际比特币交易数据更复杂）
	transactions := []string{
		"01000000010000000000000000000000000000000000000000000000000000000000000000ffffffff0704ffff001d0104ffffffff0100f2052a0100000043410496b538e853519c726a2c91e61ec11600ae1390813a627c66fb8be7947be63c52da7589379515d4e0a604f8141781e62294721166bf621e73a82cbf2342c858eeac00000000",                                                                                                                                                                                                                                                          // Coinbase 交易
		"01000000016774a696604f0d335d9b176717a4d68609b4b72c799643111e08810a47da37da000000008b483045022100d94f2cd165ff8ae88a7e84af84174b6b36254e000d27d8bf2450447817e1b18d0220630f06f89d0776ed0b3f01ab7e911b6e9e3f3448e6cc709d6174218c630c76950141046c4300f753ed182d74c3c2ea86472f6c77561b416d53a2647798a6dfc900763881992ca50335d2d26e1d86a71dd1cc80d479a064ef13ef9f65de66987e1d935e5ffffffff0280841e00000000001976a91465f445a2760980b15350cfd90c19e6fe71e71c6988ac40f99e00000000001976a9141e981842adb53c1cf1a8f99259a411b26e76852688ac00000000", // 普通交易 1
		"0100000001ad7c8929ea5222c8e0e7ee3cc82f0a9c98d63b58a0e563e6df5024fb5d6f8934000000008a4730440220580b4040000b290893dda774490b6c7c34434043862e16752f6852b96d060850022075a5e0d04c4370ecbcf4cc5416178c207985b639f96f697904162e1a959f9b32014104f17d749b6c76be3da0497cfc1d18c3d68732eef75ed9a74d833a15d4b653b19c51e3c925b67e273c12c032f46e19b0b1b6c3c4a5573d969331d95b85718a0124ffffffff02804a5d05000000001976a914960656099e311f85e1b8b2c36b5786159d8ef78588ac40d20901000000001976a9146804605c96b05c57bf521c5c959c17e2e44517e388ac00000000",    // 普通交易 2
	}

	fmt.Println("=== 比特币 Merkle Root 计算示例 ===\n")

	// 步骤 1: 对所有交易进行哈希计算，得到叶子节点
	fmt.Println("步骤 1: 计算所有交易的哈希值（叶子节点）")
	txHashes := make([][]byte, len(transactions))
	for i, tx := range transactions {
		txHashes[i] = hashTransaction(tx)
		fmt.Printf("  交易 %d 的哈希 (txid): %s\n", i+1, bytesToHex(txHashes[i]))
	}

	// 步骤 2: 计算 Merkle Root
	fmt.Println("\n步骤 2: 构建 Merkle 树并计算 Merkle Root")
	merkleRoot := calculateMerkleRoot(txHashes)

	// 输出最终结果
	fmt.Printf("\n=== Merkle Root ===\n")
	fmt.Printf("Merkle Root: %s\n", bytesToHex(merkleRoot))
	fmt.Printf("Merkle Root (反转): %s\n", reverseHex(bytesToHex(merkleRoot)))

	fmt.Println("\n=== Merkle 树结构示意图 ===")
	fmt.Println("        [Merkle Root]")
	fmt.Println("          /        \\")
	fmt.Println("     [H0-1]        [H2-3]")
	fmt.Println("     /    \\        /    \\")
	fmt.Println("   [H0]  [H1]    [H2]  [H3]")
	fmt.Println("\n注: [H0]-[H3] 为叶子节点（交易哈希），当交易数量为奇数时，最后一个节点会被复制（标记为 *）")
}

// reverseHex 反转十六进制字符串的字节顺序
// 比特币内部存储使用小端字节序，所以需要反转显示
func reverseHex(hexStr string) string {
	bytes := []byte(hexStr)
	result := make([]byte, len(bytes))
	// 以两个字符（一个字节）为单位进行反转
	for i := 0; i < len(bytes); i += 2 {
		result[len(bytes)-2-i] = bytes[i]
		result[len(bytes)-1-i] = bytes[i+1]
	}
	return string(result)
}
