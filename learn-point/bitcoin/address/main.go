package main

import (
	"crypto/sha256"
	"fmt"
	"math/big"

	"golang.org/x/crypto/ripemd160"
)

// Bitcoin Base58 编码字母表，移除了易混淆字符（0, O, l, I）
var b58Alphabet = []byte("123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz")

// Base58Encode 将字节数组编码为 Bitcoin Base58 格式
// Base58 相比 Base64 移除了易混淆字符，更适合人类读取
func Base58Encode(input []byte) []byte {
	var result []byte

	// 将输入字节转换为大整数
	x := new(big.Int).SetBytes(input)

	base := big.NewInt(58)
	zero := big.NewInt(0)
	mod := &big.Int{}

	// 不断除以 58 取余数，得到逆序的编码
	for x.Cmp(zero) != 0 {
		x.DivMod(x, base, mod)
		result = append(result, b58Alphabet[mod.Int64()])
	}

	// 处理前导零，每个前导零对应一个 '1'
	for _, b := range input {
		if b != 0x00 {
			break
		}
		result = append(result, b58Alphabet[0])
	}

	// 反转结果，得到正确的 Base58 编码
	for i, j := 0, len(result)-1; i < j; i, j = i+1, j-1 {
		result[i], result[j] = result[j], result[i]
	}

	return result
}

func main() {
	fmt.Println("=== 比特币地址生成 ===")
	fmt.Println("注意：本示例使用硬编码的测试私钥，仅用于学习演示")
	fmt.Println()

	// 示例使用的硬编码私钥
	privateKeyHex := "18e14a7b6a307f426a94f8114701e7c8dfbf2b030e2e148e63543c24c08b5d85"
	privateKey := fromHex(privateKeyHex)
	fmt.Printf("私钥 (HEX): %s\n", privateKeyHex)
	fmt.Println()

	// 步骤1：从私钥通过椭圆曲线运算生成公钥
	publicKey := generatePublicKey(privateKey)
	// 步骤2：对公钥进行压缩处理
	compressedPublicKey := compressPublicKey(publicKey)
	fmt.Printf("压缩公钥 (HEX): %x\n", compressedPublicKey)
	fmt.Println()

	// 步骤3：对压缩公钥进行双重哈希：先 SHA256，后 RIPEMD160
	sha256Hash := sha256.Sum256(compressedPublicKey)

	ripemd160Hasher := ripemd160.New()
	ripemd160Hasher.Write(sha256Hash[:])
	publicKeyHash := ripemd160Hasher.Sum(nil)
	fmt.Printf("公钥哈希 (RIPEMD160(SHA256(压缩公钥))): %x\n", publicKeyHash)
	fmt.Println()

	// 步骤4：添加比特币主网版本字节（0x00表示主网P2PKH地址）
	version := byte(0x00)
	versionedPayload := append([]byte{version}, publicKeyHash...)

	// 步骤5：计算双重双重SHA256的校验和：SHA256(SHA256(version + publicKeyHash)) 取前4字节
	firstSHA := sha256.Sum256(versionedPayload)
	secondSHA := sha256.Sum256(firstSHA[:])
	checksum := secondSHA[:4]

	// 步骤6：组合版本字节+公钥哈希+4字节校验和，得到完整的编码载荷
	fullPayload := append(versionedPayload, checksum...)
	// 步骤7：Base58编码生成最终的比特币地址
	address := string(Base58Encode(fullPayload))

	fmt.Printf("比特币地址 (P2PKH): %s\n", address)
}

// fromHex 将十六进制字符串转换为字节数组
// 例如："a1b2" → []byte{0xa1, 0xb2}
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
// 公钥 = 私钥 * G（G是secp256k1的椭圆曲线基点）
func generatePublicKey(privateKey []byte) []byte {
	// secp256k1 椭圆曲线参数
	// p: 有限域的素数模数
	p, _ := new(big.Int).SetString("FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFEFFFFFC2F", 16)
	// a: 椭圆曲线方程 y² = x³ + ax + b 中的 a 参数
	a := big.NewInt(0)
	// Gx, Gy: secp256k1 的基点 G 的坐标
	Gx, _ := new(big.Int).SetString("79BE667EF9DCBBAC55A06295CE870B07029BFCDB2DCE28D959F2815B16F81798", 16)
	Gy, _ := new(big.Int).SetString("483ADA7726A3C4655DA4FBFC0E1108A8FD17B448A68554199C47D08FFB10D4B8", 16)

	// 将私钥字节数组转换为大整数
	k := new(big.Int).SetBytes(privateKey)
	// 进行椭圆曲线标量乘法：公钥 = 私钥 * G
	x, y := scalarMult(Gx, Gy, k, a, p)

	// 构造 65 字节的未压缩公钥：0x04 + X(32字节) + Y(32字节)
	publicKey := make([]byte, 65)
	publicKey[0] = 0x04 // 0x04表示未压缩的公钥格式前缀
	// 复制 X 坐标到公钥字节数组，确保是32字节
	copy(publicKey[1:33], x.Bytes())
	// 复制 Y 坐标到公钥字节数组，确保是32字节
	copy(publicKey[33:], y.Bytes())
	return publicKey
}

// compressPublicKey 将未压缩的公钥压缩为压缩公钥格式
// 压缩公钥格式：前缀 + X坐标(32字节)
// 前缀：0x02 表示Y坐标为偶数，0x03 表示Y坐标为奇数
func compressPublicKey(publicKey []byte) []byte {
	// 从65字节的未压缩公钥中提取X坐标（字节1-32）
	x := new(big.Int).SetBytes(publicKey[1:33])
	// 从65字节的未压缩公钥中提取Y坐标（字节33-64）
	y := new(big.Int).SetBytes(publicKey[33:])

	var prefix byte
	// 根据Y坐标的最后一位判断奇偶性
	if y.Bit(0) == 0 {
		prefix = 0x02 // Y为偶数用 0x02 前缀
	} else {
		prefix = 0x03 // Y为奇数用 0x03 前缀
	}

	// 构造压缩公钥：前缀(1字节) + X坐标(32字节)
	compressed := make([]byte, 33)
	compressed[0] = prefix
	copy(compressed[1:], x.Bytes())
	return compressed
}

// scalarMult 椭圆曲线标量乘法：计算 k * (x1,y1)
// 使用 double-and-add 算法实现，高效计算大整数的标量乘法
func scalarMult(x1, y1, k, a, p *big.Int) (*big.Int, *big.Int) {
	x3 := new(big.Int)
	y3 := new(big.Int)
	x3.Set(x1)
	y3.Set(y1)

	// 从k的次高位开始遍历到最低位（double-and-add算法）
	bits := k.BitLen()
	for i := bits - 2; i >= 0; i-- {
		// 先加倍（点加倍运算）
		x3, y3 = pointDouble(x3, y3, a, p)
		// 如果当前位是1，则加上基点（点加法运算）
		if k.Bit(i) == 1 {
			x3, y3 = pointAdd(x3, y3, x1, y1, p)
		}
	}

	return x3, y3
}

// pointAdd 椭圆曲线点加法：(x1,y1) + (x2,y2)
// 返回椭圆曲线上的和点 (x3,y3)
func pointAdd(x1, y1, x2, y2, p *big.Int) (*big.Int, *big.Int) {
	// 如果两个点相同，则执行点加倍运算
	if x1.Cmp(x2) == 0 && y1.Cmp(y2) == 0 {
		return pointDouble(x1, y1, big.NewInt(0), p)
	}

	// 如果两个点是垂直直线（x相同但y相反），则返回无穷远点
	if x1.Cmp(x2) == 0 {
		return big.NewInt(0), big.NewInt(0)
	}

	// 计算斜率 lambda = (y2 - y1) / (x2 - x1) mod p
	lambda := new(big.Int)
	num := new(big.Int).Sub(y2, y1)
	den := new(big.Int).Sub(x2, x1)
	den.ModInverse(den, p) // 计算模p下的逆元
	lambda.Mul(num, den)
	lambda.Mod(lambda, p)

	// 计算 x3 = lambda² - x1 - x2 mod p
	x3 := new(big.Int)
	x3.Mul(lambda, lambda)
	x3.Sub(x3, x1)
	x3.Sub(x3, x2)
	x3.Mod(x3, p)

	// 计算 y3 = lambda*(x1 - x3) - y1 mod p
	y3 := new(big.Int)
	y3.Sub(x1, x3)
	y3.Mul(y3, lambda)
	y3.Sub(y3, y1)
	y3.Mod(y3, p)

	return x3, y3
}

// pointDouble 椭圆曲线点加倍：2*(x1,y1)
// 返回椭圆曲线上的加倍点 (x3,y3)
func pointDouble(x1, y1, a, p *big.Int) (*big.Int, *big.Int) {
	// 如果是无穷远点或y=0，则返回无穷远点
	if y1.Cmp(big.NewInt(0)) == 0 {
		return big.NewInt(0), big.NewInt(0)
	}

	// 计算切线斜率 lambda = (3*x1² + a) / (2*y1) mod p
	lambda := new(big.Int)
	threeX1Sq := new(big.Int).Mul(x1, x1)
	threeX1Sq.Mul(threeX1Sq, big.NewInt(3))
	threeX1Sq.Add(threeX1Sq, a)

	twoY1 := new(big.Int).Mul(y1, big.NewInt(2))
	twoY1.ModInverse(twoY1, p) // 计算模p下的逆元

	lambda.Mul(threeX1Sq, twoY1)
	lambda.Mod(lambda, p)

	// 计算 x3 = lambda² - 2*x1 mod p
	x3 := new(big.Int)
	x3.Mul(lambda, lambda)
	x3.Sub(x3, x1)
	x3.Sub(x3, x1)
	x3.Mod(x3, p)

	// 计算 y3 = lambda*(x1 - x3) - y1 mod p
	y3 := new(big.Int)
	y3.Sub(x1, x3)
	y3.Mul(y3, lambda)
	y3.Sub(y3, y1)
	y3.Mod(y3, p)

	return x3, y3
}
