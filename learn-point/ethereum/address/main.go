package main

import (
	"fmt"
	"math/big"

	"golang.org/x/crypto/sha3"
)

func main() {
	fmt.Println("=== 以太坊地址生成 ===")
	fmt.Println("注意：本示例使用硬编码的测试私钥，仅用于学习演示")
	fmt.Println()

	// 示例使用的硬编码私钥（与比特币示例相同，方便对比）
	privateKeyHex := "18e14a7b6a307f426a94f8114701e7c8dfbf2b030e2e148e63543c24c08b5d85"
	privateKey := fromHex(privateKeyHex)
	fmt.Printf("私钥 (HEX): %s\n", privateKeyHex)
	fmt.Println()

	// 步骤1：从私钥通过 secp256k1 椭圆曲线运算生成公钥
	publicKey := generatePublicKey(privateKey)
	fmt.Printf("公钥 (HEX, 64字节): %x\n", publicKey)
	fmt.Println()

	// 步骤2：对公钥进行 Keccak-256 哈希（注意：以太坊使用 Keccak-256，而不是 SHA3-256）
	keccakHash := keccak256(publicKey)
	fmt.Printf("Keccak-256(公钥): %x\n", keccakHash)
	fmt.Println()

	// 步骤3：取哈希结果的最后20字节作为以太坊地址（去掉前12字节）
	addressBytes := keccakHash[12:]
	fmt.Printf("地址字节 (最后20字节): %x\n", addressBytes)
	fmt.Println()

	// 步骤4：添加 0x 前缀
	address := fmt.Sprintf("0x%x", addressBytes)
	fmt.Printf("以太坊地址 (普通格式): %s\n", address)
	fmt.Println()

	// 步骤5：生成 EIP-55 校验和格式（大小写编码）
	checksumAddress := toChecksumAddress(addressBytes)
	fmt.Printf("以太坊地址 (EIP-55 校验和格式): %s\n", checksumAddress)
}

// fromHex 将十六进制字符串转换为字节数组
func fromHex(hex string) []byte {
	var result []byte
	for i := 0; i < len(hex); i += 2 {
		var b byte
		fmt.Sscanf(hex[i:i+2], "%02x", &b)
		result = append(result, b)
	}
	return result
}

// generatePublicKey 通过 secp256k1 椭圆曲线从私钥生成公钥
// 以太坊公钥格式：64字节，X坐标(32字节) + Y坐标(32字节)（去掉比特币的0x04前缀）
func generatePublicKey(privateKey []byte) []byte {
	// secp256k1 椭圆曲线参数
	p, _ := new(big.Int).SetString("FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFEFFFFFC2F", 16)
	a := big.NewInt(0)
	Gx, _ := new(big.Int).SetString("79BE667EF9DCBBAC55A06295CE870B07029BFCDB2DCE28D959F2815B16F81798", 16)
	Gy, _ := new(big.Int).SetString("483ADA7726A3C4655DA4FBFC0E1108A8FD17B448A68554199C47D08FFB10D4B8", 16)

	// 将私钥转换为大整数
	k := new(big.Int).SetBytes(privateKey)
	// 椭圆曲线标量乘法：公钥 = 私钥 * G
	x, y := scalarMult(Gx, Gy, k, a, p)

	// 构造 64 字节的公钥：X(32字节) + Y(32字节)（以太坊格式，无前缀）
	publicKey := make([]byte, 64)
	// 确保 X 和 Y 都是 32 字节，前面补零
	xBytes := x.Bytes()
	yBytes := y.Bytes()
	copy(publicKey[32-len(xBytes):32], xBytes)
	copy(publicKey[64-len(yBytes):], yBytes)
	return publicKey
}

// keccak256 计算输入的 Keccak-256 哈希
// 注意：以太坊使用 Keccak-256，而不是标准的 SHA3-256
func keccak256(data []byte) []byte {
	hash := sha3.NewLegacyKeccak256()
	hash.Write(data)
	return hash.Sum(nil)
}

// toChecksumAddress 将以太坊地址转换为 EIP-55 校验和格式
// EIP-55 通过地址哈希决定字母大小写来实现校验和
func toChecksumAddress(address []byte) string {
	// 先将地址转换为小写十六进制字符串
	addressHex := fmt.Sprintf("%x", address)
	// 对地址的十六进制字符串进行 Keccak-256 哈希
	hash := keccak256([]byte(addressHex))

	// 逐个字符处理，根据哈希值决定大小写
	result := make([]rune, 40)
	for i := 0; i < 40; i++ {
		char := rune(addressHex[i])
		// 只对字母 a-f 进行大小写处理
		if char >= 'a' && char <= 'f' {
			// 计算当前位置对应的哈希半字节（4位）
			hashByteIndex := i / 2
			hashNibble := 0
			if i%2 == 0 {
				// 偶数位置，取高4位
				hashNibble = int(hash[hashByteIndex] >> 4)
			} else {
				// 奇数位置，取低4位
				hashNibble = int(hash[hashByteIndex] & 0x0f)
			}
			// 如果哈希半字节 >= 8，则转为大写
			if hashNibble >= 8 {
				char -= 32 // 'a'-'A' = 32
			}
		}
		result[i] = char
	}

	return "0x" + string(result)
}

// scalarMult 椭圆曲线标量乘法：计算 k * (x1,y1)
func scalarMult(x1, y1, k, a, p *big.Int) (*big.Int, *big.Int) {
	x3 := new(big.Int)
	y3 := new(big.Int)
	x3.Set(x1)
	y3.Set(y1)

	// double-and-add 算法
	bits := k.BitLen()
	for i := bits - 2; i >= 0; i-- {
		x3, y3 = pointDouble(x3, y3, a, p)
		if k.Bit(i) == 1 {
			x3, y3 = pointAdd(x3, y3, x1, y1, p)
		}
	}

	return x3, y3
}

// pointAdd 椭圆曲线点加法：(x1,y1) + (x2,y2)
func pointAdd(x1, y1, x2, y2, p *big.Int) (*big.Int, *big.Int) {
	if x1.Cmp(x2) == 0 && y1.Cmp(y2) == 0 {
		return pointDouble(x1, y1, big.NewInt(0), p)
	}

	if x1.Cmp(x2) == 0 {
		return big.NewInt(0), big.NewInt(0)
	}

	// lambda = (y2 - y1) / (x2 - x1) mod p
	lambda := new(big.Int)
	num := new(big.Int).Sub(y2, y1)
	den := new(big.Int).Sub(x2, x1)
	den.ModInverse(den, p)
	lambda.Mul(num, den)
	lambda.Mod(lambda, p)

	// x3 = lambda² - x1 - x2 mod p
	x3 := new(big.Int)
	x3.Mul(lambda, lambda)
	x3.Sub(x3, x1)
	x3.Sub(x3, x2)
	x3.Mod(x3, p)

	// y3 = lambda*(x1 - x3) - y1 mod p
	y3 := new(big.Int)
	y3.Sub(x1, x3)
	y3.Mul(y3, lambda)
	y3.Sub(y3, y1)
	y3.Mod(y3, p)

	return x3, y3
}

// pointDouble 椭圆曲线点加倍：2*(x1,y1)
func pointDouble(x1, y1, a, p *big.Int) (*big.Int, *big.Int) {
	if y1.Cmp(big.NewInt(0)) == 0 {
		return big.NewInt(0), big.NewInt(0)
	}

	// lambda = (3*x1² + a) / (2*y1) mod p
	lambda := new(big.Int)
	threeX1Sq := new(big.Int).Mul(x1, x1)
	threeX1Sq.Mul(threeX1Sq, big.NewInt(3))
	threeX1Sq.Add(threeX1Sq, a)

	twoY1 := new(big.Int).Mul(y1, big.NewInt(2))
	twoY1.ModInverse(twoY1, p)

	lambda.Mul(threeX1Sq, twoY1)
	lambda.Mod(lambda, p)

	// x3 = lambda² - 2*x1 mod p
	x3 := new(big.Int)
	x3.Mul(lambda, lambda)
	x3.Sub(x3, x1)
	x3.Sub(x3, x1)
	x3.Mod(x3, p)

	// y3 = lambda*(x1 - x3) - y1 mod p
	y3 := new(big.Int)
	y3.Sub(x1, x3)
	y3.Mul(y3, lambda)
	y3.Sub(y3, y1)
	y3.Mod(y3, p)

	return x3, y3
}
