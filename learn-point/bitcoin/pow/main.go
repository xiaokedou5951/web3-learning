package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/big"
	"strconv"
	"strings"
	"time"
)

// ============================================================================
// 比特币工作量证明（Proof of Work, PoW）学习实现
// ============================================================================
// 核心概念:
// 1. 区块包含: 版本号、前一区块哈希、Merkle根、时间戳、难度目标、Nonce
// 2. 挖矿目标: 找到一个Nonce值,使得 区块哈希 < 目标值(Target)
// 3. 哈希计算: SHA256(SHA256(区块头))
// 4. 难度表示: 前导零越多,难度越大,需要尝试的Nonce越多
// 5. 比特币难度: 每2016个区块调整一次,目标出块时间10分钟
// ============================================================================

// Block 表示比特币的一个区块
type Block struct {
	// Version 区块版本号,标识区块遵循的验证规则
	Version int32

	// PrevBlockHash 前一个区块的哈希值,形成区块链
	PrevBlockHash string

	// MerkleRoot 区块内所有交易的Merkle树根哈希
	MerkleRoot string

	// Timestamp 区块创建的时间戳(Unix时间)
	Timestamp int64

	// Difficulty 难度:要求哈希值前导零的十六进制位数
	// 例如 Difficulty=4 表示哈希必须以 "0000" 开头
	Difficulty int

	// Nonce 工作量证明的随机数,矿工通过不断改变它来寻找有效哈希
	Nonce uint32

	// Hash 当前区块的哈希值(由区块头计算得出)
	Hash string
}

// PowResult 工作量证明的计算结果
type PowResult struct {
	// Nonce 找到的有效Nonce值
	Nonce uint32

	// Hash 对应的区块哈希
	Hash string

	// Attempts 总共尝试了多少次
	Attempts uint64

	// ElapsedTime 挖矿耗时
	ElapsedTime time.Duration
}

// BlockHeaderToBytes 将区块头字段转换为字节切片用于哈希计算
// 比特币区块头结构(80字节):
//   - 版本号: 4字节
//   - 前一区块哈希: 32字节
//   - Merkle根: 32字节
//   - 时间戳: 4字节
//   - 难度目标(Bits): 4字节
//   - Nonce: 4字节
func BlockHeaderToBytes(block *Block) []byte {
	// 组合所有区块头字段为一个字节切片
	// 格式: version|prev_hash|merkle_root|timestamp|difficulty|nonce
	header := fmt.Sprintf("%d%s%s%d%d%d",
		block.Version,
		block.PrevBlockHash,
		block.MerkleRoot,
		block.Timestamp,
		block.Difficulty,
		block.Nonce,
	)
	return []byte(header)
}

// DoubleSHA256 双重SHA256哈希: SHA256(SHA256(data))
// 比特币使用双重哈希来防止长度扩展攻击
func DoubleSHA256(data []byte) []byte {
	// 第一次SHA256哈希
	hash1 := sha256.Sum256(data)
	// 第二次SHA256哈希
	hash2 := sha256.Sum256(hash1[:])
	return hash2[:]
}

// CalculateHash 计算区块的哈希值
// 对区块头进行双重SHA256运算
func CalculateHash(block *Block) string {
	// 将区块头转换为字节
	headerBytes := BlockHeaderToBytes(block)
	// 计算双重SHA256
	hashBytes := DoubleSHA256(headerBytes)
	// 转换为十六进制字符串
	return hex.EncodeToString(hashBytes)
}

// GetTargetPrefix 根据难度获取目标前缀
// 难度=1: 哈希需要以 "0" 开头 (1个前导零)
// 难度=2: 哈希需要以 "00" 开头 (2个前导零)
// 难度=4: 哈希需要以 "0000" 开头 (4个前导零)
// 每增加1个难度,尝试次数大约增加16倍(因为十六进制每位有16种可能)
func GetTargetPrefix(difficulty int) string {
	return strings.Repeat("0", difficulty)
}

// HashToBigInt 将十六进制哈希字符串转换为大整数
func HashToBigInt(hash string) *big.Int {
	hashBytes, _ := hex.DecodeString(hash)
	result := new(big.Int)
	result.SetBytes(hashBytes)
	return result
}

// ProofOfWork 工作量证明算法
// 通过不断改变Nonce值,计算区块哈希,直到哈希值满足难度要求
func ProofOfWork(block *Block) *PowResult {
	// 获取目标前缀(例如难度4时为"0000")
	targetPrefix := GetTargetPrefix(block.Difficulty)

	fmt.Printf("  难度(Difficulty): %d\n", block.Difficulty)
	fmt.Printf("  目标前缀: 哈希必须以 \"%s\" 开头\n", targetPrefix)
	fmt.Printf("  说明: 每增加1个难度,工作量约增加16倍\n\n")

	// 记录开始时间和尝试次数
	startTime := time.Now()
	var attempts uint64 = 0

	// 挖矿循环:不断尝试不同的Nonce值
	// uint32最大值约为42.9亿
	for nonce := uint32(0); nonce < 4294967295; nonce++ {
		// 设置当前Nonce
		block.Nonce = nonce

		// 计算区块哈希
		hash := CalculateHash(block)
		attempts++

		// 每10万次尝试打印一次进度
		if attempts%100000 == 0 {
			fmt.Printf("  已尝试 %d 万次,当前Nonce: %d, 当前哈希: %s...\n",
				attempts/10000, nonce, hash[:16])
		}

		// 检查哈希是否满足难度要求(前导零数量)
		if strings.HasPrefix(hash, targetPrefix) {
			elapsed := time.Since(startTime)
			leadingZeros := CountLeadingZeros(hash)

			fmt.Printf("\n  ✓ 找到有效Nonce!\n")
			fmt.Printf("  Nonce: %d\n", nonce)
			fmt.Printf("  区块哈希: %s\n", hash)
			fmt.Printf("  前导零数量: %d (要求: %d)\n", leadingZeros, block.Difficulty)
			fmt.Printf("  总尝试次数: %d\n", attempts)
			fmt.Printf("  耗时: %v\n", elapsed)
			fmt.Printf("  每秒尝试: %.0f 次\n", float64(attempts)/elapsed.Seconds())

			return &PowResult{
				Nonce:       nonce,
				Hash:        hash,
				Attempts:    attempts,
				ElapsedTime: elapsed,
			}
		}
	}

	// 如果遍历完所有Nonce都没找到,返回失败
	return nil
}

// ValidateProofOfWork 验证工作量证明是否有效
// 任何人都可以验证:只需计算一次哈希,检查是否满足难度要求
func ValidateProofOfWork(block *Block) bool {
	// 计算区块哈希
	hash := CalculateHash(block)
	// 获取目标前缀
	targetPrefix := GetTargetPrefix(block.Difficulty)
	// 检查哈希是否以目标前缀开头
	return strings.HasPrefix(hash, targetPrefix)
}

// CountLeadingZeros 计算十六进制哈希字符串的前导零数量
// 前导零越多,说明哈希值越小,难度越大
func CountLeadingZeros(hash string) int {
	count := 0
	for _, ch := range hash {
		if ch == '0' {
			count++
		} else {
			break
		}
	}
	return count
}

// NewGenesisBlock 创建创世区块(区块链的第一个区块)
// 创世区块没有前一个区块,所以PrevBlockHash为空
func NewGenesisBlock() *Block {
	return &Block{
		Version:       1,
		PrevBlockHash: strings.Repeat("0", 64), // 创世区块的前一哈希全为0
		MerkleRoot:    strings.Repeat("0", 64), // 简化处理,无交易时Merkle根全为0
		Timestamp:     1231006505,              // 比特币创世区块时间戳: 2009-01-03
		Difficulty:    4,                       // 使用较低难度便于演示
		Nonce:         0,
	}
}

// NewBlock 创建新区块
func NewBlock(version int32, prevHash, merkleRoot string, timestamp int64, difficulty int) *Block {
	return &Block{
		Version:       version,
		PrevBlockHash: prevHash,
		MerkleRoot:    merkleRoot,
		Timestamp:     timestamp,
		Difficulty:    difficulty,
		Nonce:         0,
	}
}

// ============================================================================
// 主程序: 演示比特币工作量证明的完整流程
// ============================================================================
func main() {
	fmt.Println("============================================================")
	fmt.Println("  比特币工作量证明(Proof of Work)学习演示")
	fmt.Println("============================================================")
	fmt.Println()

	// ------------------------------------------------------------------------
	// 第一部分: 理解哈希函数的特性
	// ------------------------------------------------------------------------
	fmt.Println("【第一部分: 哈希函数特性演示】")
	fmt.Println("------------------------------------------------------------")
	demonstrateHashProperties()
	fmt.Println()

	// ------------------------------------------------------------------------
	// 第二部分: 难度概念解析
	// ------------------------------------------------------------------------
	fmt.Println("【第二部分: 难度概念解析】")
	fmt.Println("------------------------------------------------------------")
	demonstrateDifficulty()
	fmt.Println()

	// ------------------------------------------------------------------------
	// 第三部分: 执行工作量证明(低难度)
	// ------------------------------------------------------------------------
	fmt.Println("【第三部分: 执行工作量证明(难度=2,快速演示)】")
	fmt.Println("------------------------------------------------------------")

	// 创建测试区块(使用低难度以便快速演示)
	testBlock := NewBlock(
		1,
		strings.Repeat("0", 64),
		"test_merkle_root_for_demo",
		time.Now().Unix(),
		2, // 难度2,哈希需要以"00"开头,通常只需尝试几百次
	)

	fmt.Printf("开始挖矿...\n\n")
	result := ProofOfWork(testBlock)

	if result != nil {
		testBlock.Hash = result.Hash
		fmt.Printf("\n挖矿成功!\n")
		fmt.Printf("  区块哈希: %s\n", testBlock.Hash)
	} else {
		fmt.Println("挖矿失败: 未找到有效Nonce")
	}

	fmt.Println()

	// ------------------------------------------------------------------------
	// 第四部分: 执行工作量证明(中等难度)
	// ------------------------------------------------------------------------
	fmt.Println("【第四部分: 执行工作量证明(难度=4,中等难度)】")
	fmt.Println("------------------------------------------------------------")

	mediumBlock := NewBlock(
		1,
		strings.Repeat("0", 64),
		"medium_difficulty_demo",
		time.Now().Unix(),
		4, // 难度4,哈希需要以"0000"开头,通常需要尝试数万次
	)

	fmt.Printf("开始挖矿...\n\n")
	result2 := ProofOfWork(mediumBlock)

	if result2 != nil {
		mediumBlock.Hash = result2.Hash
		fmt.Printf("\n挖矿成功!\n")
	}
	fmt.Println()

	// ------------------------------------------------------------------------
	// 第五部分: 验证工作量证明
	// ------------------------------------------------------------------------
	fmt.Println("【第五部分: 验证工作量证明】")
	fmt.Println("------------------------------------------------------------")

	// 验证区块的PoW是否有效
	valid := ValidateProofOfWork(mediumBlock)
	fmt.Printf("区块验证结果: %v\n", valid)
	fmt.Printf("验证原理: 只需计算一次哈希,检查前导零数量是否满足难度要求\n")
	fmt.Printf("这正是比特币的特性: 挖矿困难,验证简单!\n")
	fmt.Println()

	// ------------------------------------------------------------------------
	// 第六部分: 不同难度的对比
	// ------------------------------------------------------------------------
	fmt.Println("【第六部分: 不同难度的理论对比】")
	fmt.Println("------------------------------------------------------------")

	fmt.Println("难度与前导零的关系:")
	fmt.Println("  难度 | 前导零 | 目标前缀   | 期望尝试次数 | 说明")
	fmt.Println("  -----|--------|-----------|-------------|------------------")

	for diff := 1; diff <= 8; diff++ {
		prefix := GetTargetPrefix(diff)
		// 每个十六进制位有16种可能,期望尝试次数 = 16^difficulty
		expected := big.NewInt(16)
		expected.Exp(expected, big.NewInt(int64(diff)), nil)
		fmt.Printf("  %5d | %6d | %-10s | %12s | 约%s次尝试\n",
			diff, diff, "\""+prefix+"\"", expected.String(), expected.String())
	}

	fmt.Println()
	fmt.Println("比特币当前难度(2024年): 约83万亿(8.3×10^13)")
	fmt.Println("这意味着比特币区块哈希需要约17-18个前导零")
	fmt.Println()

	// ------------------------------------------------------------------------
	// 第七部分: 区块链链接演示
	// ------------------------------------------------------------------------
	fmt.Println("【第七部分: 区块链链接演示(两个区块)】")
	fmt.Println("------------------------------------------------------------")

	// 创世区块
	fmt.Println("挖创世区块(难度=2)...")
	genesisBlock := NewBlock(1, strings.Repeat("0", 64), "genesis_block",
		time.Now().Unix(), 2)
	genesisResult := ProofOfWork(genesisBlock)
	if genesisResult != nil {
		genesisBlock.Hash = genesisResult.Hash
	}
	fmt.Println()

	// 第二个区块(使用创世区块的哈希作为PrevBlockHash)
	fmt.Println("挖第二个区块(难度=2,链接到创世区块)...")
	secondBlock := NewBlock(1, genesisBlock.Hash, "second_block_transactions",
		time.Now().Unix(), 2)
	secondResult := ProofOfWork(secondBlock)
	if secondResult != nil {
		secondBlock.Hash = secondResult.Hash
	}
	fmt.Println()

	// 显示区块链结构
	fmt.Println("区块链结构:")
	fmt.Printf("  创世区块:\n")
	fmt.Printf("    哈希: %s\n", genesisBlock.Hash[:16]+"...")
	fmt.Printf("    前一区块: %s\n", genesisBlock.PrevBlockHash[:16]+"...")
	fmt.Printf("    Nonce: %d\n", genesisBlock.Nonce)
	fmt.Println()
	fmt.Printf("  第二个区块:\n")
	fmt.Printf("    哈希: %s\n", secondBlock.Hash[:16]+"...")
	fmt.Printf("    前一区块: %s\n", secondBlock.PrevBlockHash[:16]+"...")
	fmt.Printf("    Nonce: %d\n", secondBlock.Nonce)
	fmt.Println()
	fmt.Println("注意: 第二个区块的PrevBlockHash等于创世区块的Hash")
	fmt.Println("这就是'区块链'中'链'的含义!")
	fmt.Println()

	// ------------------------------------------------------------------------
	// 总结
	// ------------------------------------------------------------------------
	fmt.Println("============================================================")
	fmt.Println("  学习要点总结:")
	fmt.Println("============================================================")
	fmt.Println("  1. PoW核心: 找到Nonce使区块哈希满足难度要求(前导零)")
	fmt.Println("  2. 哈希函数: SHA256(SHA256(区块头)),不可逆,雪崩效应")
	fmt.Println("  3. 难度调整: 前导零越多,难度越大,期望尝试次数 = 16^难度")
	fmt.Println("  4. 验证简单: 挖矿困难,但验证只需一次哈希计算")
	fmt.Println("  5. 区块链: 每个区块包含前一区块的哈希,形成不可篡改的链")
	fmt.Println("  6. 安全性: 修改区块需要重新计算PoW,成本极高")
	fmt.Println("============================================================")
}

// ============================================================================
// 辅助演示函数
// ============================================================================

// demonstrateHashProperties 演示哈希函数的特性
func demonstrateHashProperties() {
	// 演示1: 相同输入产生相同哈希
	input1 := []byte("Hello, Bitcoin!")
	hash1 := hex.EncodeToString(DoubleSHA256(input1))
	hash1Again := hex.EncodeToString(DoubleSHA256(input1))
	fmt.Printf("相同输入产生相同哈希:\n")
	fmt.Printf("  输入: \"Hello, Bitcoin!\"\n")
	fmt.Printf("  哈希1: %s\n", hash1)
	fmt.Printf("  哈希2: %s\n", hash1Again)
	fmt.Printf("  结果: %v\n\n", hash1 == hash1Again)

	// 演示2: 微小改变导致完全不同的哈希(雪崩效应)
	input2 := []byte("Hello, Bitcoin!") // 原输入
	input3 := []byte("Hello, bitcoin!") // 只改变B->b
	hash2 := hex.EncodeToString(DoubleSHA256(input2))
	hash3 := hex.EncodeToString(DoubleSHA256(input3))
	fmt.Printf("雪崩效应(一个字符改变):\n")
	fmt.Printf("  输入1: \"Hello, Bitcoin!\"\n")
	fmt.Printf("  输入2: \"Hello, bitcoin!\"\n")
	fmt.Printf("  哈希1: %s\n", hash2)
	fmt.Printf("  哈希2: %s\n", hash3)
	fmt.Printf("  完全不同: %v\n\n", hash2 != hash3)

	// 演示3: 哈希的随机性
	fmt.Printf("哈希值看起来是随机的:\n")
	for i := 0; i < 5; i++ {
		data := []byte("test" + strconv.Itoa(i))
		hash := hex.EncodeToString(DoubleSHA256(data))
		fmt.Printf("  input=\"test%d\" -> hash=%s\n", i, hash[:16]+"...")
	}
}

// demonstrateDifficulty 演示难度概念
func demonstrateDifficulty() {
	fmt.Println("难度是通过要求哈希值的前导零数量来控制的:")
	fmt.Println()

	// 演示不同输入的前导零数量
	inputs := []string{
		"block_data_1",
		"block_data_2",
		"block_data_3",
		"block_data_4",
		"block_data_5",
	}

	fmt.Println("随机数据的哈希前导零分布:")
	for _, input := range inputs {
		hash := hex.EncodeToString(DoubleSHA256([]byte(input)))
		zeros := CountLeadingZeros(hash)
		fmt.Printf("  输入: %-15s -> 哈希: %s -> 前导零: %d\n",
			input, hash[:16]+"...", zeros)
	}

	fmt.Println()
	fmt.Println("说明:")
	fmt.Println("  - 大约1/16的哈希以1个0开头 (难度1)")
	fmt.Println("  - 大约1/256的哈希以2个0开头 (难度2)")
	fmt.Println("  - 大约1/4096的哈希以3个0开头 (难度3)")
	fmt.Println("  - 以此类推,每增加1个难度,概率降低16倍")
}
