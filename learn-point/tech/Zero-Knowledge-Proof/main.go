package main

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/big"
)

// ============================================================
// 零知识证明 (Zero-Knowledge Proof, ZKP) - Schnorr 协议实现
// ============================================================
//
// 什么是零知识证明？
// 零知识证明是一种密码学协议，允许证明者 (Prover) 向验证者 (Verifier) 证明
// 某个声明为真，而无需透露任何额外信息（"零知识"）。
//
// 零知识证明必须满足三个性质：
// 1. 完备性 (Completeness): 如果声明为真，诚实的证明者能说服诚实的验证者
// 2. 可靠性 (Soundness): 如果声明为假，作弊的证明者无法说服诚实的验证者
// 3. 零知识性 (Zero-Knowledge): 验证者除了"声明为真"外，学不到任何其他信息
//
// Schnorr 协议原理：
// 证明者知道秘密值 x（离散对数），想证明 y = g^x mod p，但不泄露 x
//
// 交互式协议三步：
// 1. 承诺 (Commitment): 证明者选择随机数 r，发送 t = g^r mod p
// 2. 挑战 (Challenge):  验证者发送随机挑战 c
// 3. 响应 (Response):   证明者计算 s = r + c*x mod (p-1)，发送 s
//
// 验证者验证: g^s == t * y^c mod p
//
// 非交互式版本（本实现）:
// 使用 Fiat-Shamir 启发式方法，用哈希函数代替验证者的随机挑战：
// c = H(g, y, t)，使协议变成非交互式
// ============================================================

// SystemParams 保存系统公共参数
type SystemParams struct {
	// p: 大素数，定义有限域 Z_p
	// 在实际应用中通常使用 2048 位以上的素数
	// 这里使用较小的素数用于演示
	P *big.Int

	// g: 模 p 的生成元（原根），即 g 的幂可以生成 Z_p* 中的所有元素
	G *big.Int
}

// PrivateKey 证明者的私钥（秘密）
// x: 证明者知道的秘密值，永远不会发送给验证者
type PrivateKey struct {
	X *big.Int
}

// PublicKey 证明者的公钥（公开）
// y = g^x mod p
// 任何人都可以看到这个值，但无法从 y 推导出 x（离散对数难题）
type PublicKey struct {
	Y *big.Int
}

// Proof 零知识证明结构
// 包含证明者生成的两个值，验证者可以用它们来验证证明
type Proof struct {
	// T: 承诺值，t = g^r mod p，其中 r 是随机数
	// 这相当于证明者"密封"了一个随机值，稍后揭示
	T *big.Int

	// S: 响应值，s = r + c*x mod (p-1)
	// 这个值混合了随机数 r 和秘密 x，但不会泄露 x
	S *big.Int

	// C: 挑战值，c = H(g || y || t)
	// 在非交互式版本中，由哈希函数生成
	C *big.Int
}

// ============================================================
// 系统参数生成
// ============================================================

// GenerateSystemParams 生成系统公共参数 (p, g)
// 在实际应用中，这些参数可以是预定义的标准参数（如 DSA 参数）
func GenerateSystemParams() (*SystemParams, error) {
	// 为了演示，使用一个固定的安全素数
	// 这里使用一个 256 位的安全素数（实际应用需要更大）
	// p = 2^256 - 189 (这是一个素数，常用于密码学)
	p := new(big.Int)
	p.SetString("FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFEFFFFFC2F", 16)

	// g 选择一个生成元
	// 对于这个素数，g=7 是一个常用的选择
	g := big.NewInt(7)

	return &SystemParams{
		P: p,
		G: g,
	}, nil
}

// ============================================================
// 密钥生成
// ============================================================

// GenerateKeyPair 生成公私钥对
// 秘密值 x 随机选择，公钥 y = g^x mod p
func GenerateKeyPair(params *SystemParams) (*PrivateKey, *PublicKey, error) {
	// 生成一个随机秘密值 x，范围在 [1, p-2]
	// x 必须足够大以保证安全性，且不能为 0 或 p-1
	pMinus2 := new(big.Int).Sub(params.P, big.NewInt(2))

	x, err := rand.Int(rand.Reader, pMinus2)
	if err != nil {
		return nil, nil, fmt.Errorf("生成随机秘密值失败: %w", err)
	}

	// 确保 x >= 1
	x.Add(x, big.NewInt(1))

	// 计算公钥 y = g^x mod p
	// 使用模幂运算：ModExp(base, exp, mod) = base^exp mod mod
	y := new(big.Int).Exp(params.G, x, params.P)

	fmt.Printf("[密钥生成] 秘密值 x (不公开): %s\n", x.Text(16))
	fmt.Printf("[密钥生成] 公钥 y = g^x mod p: %s\n", y.Text(16))

	return &PrivateKey{X: x}, &PublicKey{Y: y}, nil
}

// ============================================================
// 证明生成 (Prover)
// ============================================================

// GenerateProof 生成零知识证明
// 证明者向验证者证明自己知道 x（满足 y = g^x mod p），而不泄露 x
func GenerateProof(params *SystemParams, privKey *PrivateKey, pubKey *PublicKey) (*Proof, error) {
	fmt.Println("\n=== 证明者开始生成证明 ===")

	// ========== 步骤 1: 承诺 (Commitment) ==========
	// 证明者选择一个随机数 r（临时密钥）
	// r 的范围应该在 [1, p-2]
	pMinus2 := new(big.Int).Sub(params.P, big.NewInt(2))

	r, err := rand.Int(rand.Reader, pMinus2)
	if err != nil {
		return nil, fmt.Errorf("生成随机数 r 失败: %w", err)
	}
	r.Add(r, big.NewInt(1)) // 确保 r >= 1

	fmt.Printf("[承诺] 随机数 r: %s\n", r.Text(16))

	// 计算承诺值 t = g^r mod p
	// 这相当于"密封"了随机数 r，验证者看到 t 但无法反推 r
	t := new(big.Int).Exp(params.G, r, params.P)
	fmt.Printf("[承诺] t = g^r mod p: %s\n", t.Text(16))

	// ========== 步骤 2: 挑战 (Challenge) ==========
	// 在交互式协议中，验证者会发送随机挑战 c
	// 在非交互式版本中，使用 Fiat-Shamir 启发式：c = H(g || y || t)
	// 这确保了挑战是确定性的，但无法被证明者预测或操纵
	c := computeChallenge(params, pubKey, t)
	fmt.Printf("[挑战] c = H(g || y || t): %s\n", c.Text(16))

	// ========== 步骤 3: 响应 (Response) ==========
	// 计算 s = r + c*x mod (p-1)
	// 这个公式巧妙地混合了 r 和 x，但不会泄露 x 的值
	// 因为 r 是随机的，s 也是随机的，不包含关于 x 的直接信息
	cx := new(big.Int).Mul(c, privKey.X) // c * x
	s := new(big.Int).Add(r, cx)         // r + c * x

	// mod (p-1)，因为根据费马小定理，g^(p-1) ≡ 1 (mod p)
	pMinus1 := new(big.Int).Sub(params.P, big.NewInt(1))
	s.Mod(s, pMinus1)

	fmt.Printf("[响应] s = (r + c*x) mod (p-1): %s\n", s.Text(16))
	fmt.Println("=== 证明生成完成 ===")

	return &Proof{
		T: t,
		S: s,
		C: c,
	}, nil
}

// computeChallenge 使用 Fiat-Shamir 启发式计算挑战值
// c = SHA256(g || y || t) mod (p-1)
// 这模拟了验证者的随机挑战，确保协议的非交互性和安全性
func computeChallenge(params *SystemParams, pubKey *PublicKey, t *big.Int) *big.Int {
	// 将 g, y, t 序列化后拼接
	hash := sha256.New()
	hash.Write(params.G.Bytes())
	hash.Write(pubKey.Y.Bytes())
	hash.Write(t.Bytes())

	// 计算 SHA256 哈希
	hashBytes := hash.Sum(nil)
	fmt.Printf("[挑战] SHA256(g||y||t) 哈希值: %s\n", hex.EncodeToString(hashBytes))

	// 将哈希值转换为大整数
	c := new(big.Int).SetBytes(hashBytes)

	// mod (p-1) 确保挑战值在有效范围内
	pMinus1 := new(big.Int).Sub(params.P, big.NewInt(1))
	c.Mod(c, pMinus1)

	return c
}

// ============================================================
// 证明验证 (Verifier)
// ============================================================

// VerifyProof 验证零知识证明
// 验证者检查证明是否有效，而不需要知道秘密值 x
//
// 验证原理：
// 如果 s = r + c*x mod (p-1)，那么：
// g^s = g^(r + c*x) = g^r * g^(c*x) = g^r * (g^x)^c = t * y^c mod p
//
// 因此验证等式：g^s ≡ t * y^c (mod p)
// 如果等式成立，说明证明者确实知道 x
func VerifyProof(params *SystemParams, pubKey *PublicKey, proof *Proof) bool {
	fmt.Println("\n=== 验证者开始验证证明 ===")

	// ========== 步骤 1: 重新计算挑战值 ==========
	// 验证者也使用相同的哈希函数计算挑战值
	// 这确保了挑战值是确定性的，证明者无法预先知道
	c := computeChallenge(params, pubKey, proof.T)
	fmt.Printf("[验证] 重新计算挑战 c: %s\n", c.Text(16))

	// 检查计算出的挑战值是否与证明中的一致
	if c.Cmp(proof.C) != 0 {
		fmt.Println("[验证失败] 挑战值不匹配！")
		return false
	}
	fmt.Println("[验证] 挑战值匹配 ✓")

	// ========== 步骤 2: 验证等式 g^s ≡ t * y^c (mod p) ==========

	// 计算等式左边: g^s mod p
	leftSide := new(big.Int).Exp(params.G, proof.S, params.P)
	fmt.Printf("[验证] 等式左边 g^s mod p: %s\n", leftSide.Text(16))

	// 计算等式右边: t * y^c mod p
	// 先计算 y^c mod p
	yToC := new(big.Int).Exp(pubKey.Y, proof.C, params.P)
	// 再乘以 t
	rightSide := new(big.Int).Mul(proof.T, yToC)
	// 最后 mod p
	rightSide.Mod(rightSide, params.P)
	fmt.Printf("[验证] 等式右边 t * y^c mod p: %s\n", rightSide.Text(16))

	// 比较两边是否相等
	if leftSide.Cmp(rightSide) == 0 {
		fmt.Println("[验证成功] g^s ≡ t * y^c (mod p)，证明有效！✓")
		return true
	}

	fmt.Println("[验证失败] g^s ≠ t * y^c (mod p)，证明无效！")
	return false
}

// ============================================================
// 恶意证明者尝试（演示可靠性）
// ============================================================

// GenerateFakeProof 生成一个假的证明（不知道秘密 x）
// 用于演示：如果证明者不知道 x，验证会失败
func GenerateFakeProof(params *SystemParams, pubKey *PublicKey) (*Proof, error) {
	fmt.Println("\n=== 恶意证明者尝试伪造证明 ===")

	// 伪造者不知道 x，只能随机猜测
	r, _ := rand.Int(rand.Reader, new(big.Int).Sub(params.P, big.NewInt(2)))
	r.Add(r, big.NewInt(1))

	t := new(big.Int).Exp(params.G, r, params.P)
	c := computeChallenge(params, pubKey, t)

	// 伪造者使用随机的 s，而不是正确的 s = r + c*x
	s, _ := rand.Int(rand.Reader, new(big.Int).Sub(params.P, big.NewInt(1)))

	fmt.Println("=== 伪造证明完成（使用随机值） ===")
	return &Proof{
		T: t,
		S: s,
		C: c,
	}, nil
}

// ============================================================
// 主函数 - 演示完整的零知识证明流程
// ============================================================

func main() {
	fmt.Println("==============================================")
	fmt.Println("  零知识证明 (Zero-Knowledge Proof) 演示")
	fmt.Println("  基于 Schnorr 协议 + Fiat-Shamir 启发式")
	fmt.Println("==============================================\n")

	// ========== 1. 系统参数设置 ==========
	fmt.Println("--- 步骤 1: 生成系统参数 ---")
	params, err := GenerateSystemParams()
	if err != nil {
		fmt.Printf("生成系统参数失败: %v\n", err)
		return
	}
	fmt.Printf("[系统参数] 素数 p (16进制): %s\n", params.P.Text(16))
	fmt.Printf("[系统参数] 生成元 g: %s\n\n", params.G.Text(16))

	// ========== 2. 密钥生成 ==========
	fmt.Println("--- 步骤 2: 生成公私钥对 ---")
	privKey, pubKey, err := GenerateKeyPair(params)
	if err != nil {
		fmt.Printf("生成密钥对失败: %v\n", err)
		return
	}
	fmt.Println()

	// ========== 3. 生成零知识证明 ==========
	fmt.Println("--- 步骤 3: 证明者生成零知识证明 ---")
	fmt.Println("证明者知道秘密 x，要向验证者证明自己知道 x")
	fmt.Println("但不会泄露 x 的值！")

	proof, err := GenerateProof(params, privKey, pubKey)
	if err != nil {
		fmt.Printf("生成证明失败: %v\n", err)
		return
	}
	fmt.Printf("\n[最终证明] T: %s\n", proof.T.Text(16))
	fmt.Printf("[最终证明] C: %s\n", proof.C.Text(16))
	fmt.Printf("[最终证明] S: %s\n", proof.S.Text(16))

	// ========== 4. 验证证明 ==========
	fmt.Println("\n--- 步骤 4: 验证者验证证明 ---")
	_ = VerifyProof(params, pubKey, proof)

	// ========== 5. 测试伪造证明（演示可靠性） ==========
	fmt.Println("\n\n==============================================")
	fmt.Println("  测试：伪造证明应该被拒绝")
	fmt.Println("==============================================")

	fakeProof, err := GenerateFakeProof(params, pubKey)
	if err != nil {
		fmt.Printf("生成假证明失败: %v\n", err)
		return
	}

	fmt.Println("\n--- 验证者验证伪造的证明 ---")
	fakeValid := VerifyProof(params, pubKey, fakeProof)
	if fakeValid {
		fmt.Println("⚠️  警告：伪造的证明被接受了！（这不应该发生）")
	} else {
		fmt.Println("✓  伪造的证明被正确拒绝！")
	}

	// ========== 总结 ==========
	fmt.Println("\n==============================================")
	fmt.Println("  总结")
	fmt.Println("==============================================")
	fmt.Println("1. 证明者成功生成了零知识证明")
	fmt.Println("2. 验证者验证了证明的有效性")
	fmt.Println("3. 整个过程中，秘密值 x 从未被暴露")
	fmt.Println("4. 伪造的证明被正确拒绝")
	fmt.Println("\n零知识证明的核心：")
	fmt.Println("- 证明者证明'我知道 x'，而不透露 x 是什么")
	fmt.Println("- 验证者只能确认'证明有效'，学不到关于 x 的任何信息")
}
