# 零知识证明 (Zero-Knowledge Proof) - Go 语言实现

## 什么是零知识证明？

零知识证明（ZKP）是一种密码学协议，允许**证明者（Prover）**向**验证者（Verifier）**证明某个声明为真，而无需透露任何额外信息。

### 三个核心性质

| 性质 | 英文 | 说明 |
|------|------|------|
| 完备性 | Completeness | 如果声明为真，诚实的证明者能说服诚实的验证者 |
| 可靠性 | Soundness | 如果声明为假，作弊的证明者无法说服验证者 |
| 零知识性 | Zero-Knowledge | 验证者除了"声明为真"外，学不到任何其他信息 |

### 生活中的例子

> **洞穴比喻**：想象一个环形洞穴，只有一个入口，内部有一个魔法门，需要密码才能打开。Alice 知道密码，她想向 Bob 证明自己知道密码，但不想告诉 Bob 密码是什么。
>
> 1. Bob 站在入口外
> 2. Alice 走进洞穴，随机选择左路或右路
> 3. Bob 走到分叉口，随机喊出"从左路出来"或"从右路出来"
> 4. 如果 Alice 知道密码，她总能从指定路线出来；如果不知道，只有 50% 概率成功
> 5. 重复多次，Bob 确信 Alice 知道密码，但仍然不知道密码是什么

---

## 本实现：Schnorr 协议

### 数学基础

基于**离散对数难题（Discrete Logarithm Problem, DLP）**：
- 给定 `g`, `p`, `y = g^x mod p`，求 `x` 是计算困难的
- 但验证 `y = g^x mod p` 很容易

### 协议流程

#### 交互式版本

```
证明者                                    验证者
  |                                        |
  |  1. 选择随机 r，计算 t = g^r mod p     |
  |--------------------------------------->|
  |              t (承诺)                   |
  |                                        |
  |  2. 生成随机挑战 c                     |
  |<---------------------------------------|
  |              c (挑战)                   |
  |                                        |
  |  3. 计算 s = r + c*x mod (p-1)        |
  |--------------------------------------->|
  |              s (响应)                   |
  |                                        |
  |  4. 验证: g^s ≡ t * y^c (mod p)?     |
```

#### 非交互式版本（本实现）

使用 **Fiat-Shamir 启发式方法**，用哈希函数代替验证者的随机挑战：

```
c = SHA256(g || y || t) mod (p-1)
```

这使得协议变成非交互式，证明者可以一次性生成完整证明。

---

## 代码结构

```
Zero-Knowledge-Proof/
├── main.go        # 核心实现
└── README.md      # 说明文档
```

### 主要类型

| 类型 | 说明 |
|------|------|
| `SystemParams` | 系统参数 (p, g) |
| `PrivateKey` | 私钥 x（秘密） |
| `PublicKey` | 公钥 y = g^x mod p |
| `Proof` | 零知识证明 (T, C, S) |

### 核心函数

| 函数 | 说明 |
|------|------|
| `GenerateSystemParams()` | 生成系统参数 |
| `GenerateKeyPair()` | 生成公私钥对 |
| `GenerateProof()` | 生成零知识证明 |
| `VerifyProof()` | 验证证明 |
| `GenerateFakeProof()` | 生成假证明（用于测试） |

---

## 运行方式

```bash
# 进入目录
cd Zero-Knowledge-Proof

# 运行
go run main.go
```

### 预期输出

```
==============================================
  零知识证明 (Zero-Knowledge Proof) 演示
  基于 Schnorr 协议 + Fiat-Shamir 启发式
==============================================

--- 步骤 1: 生成系统参数 ---
[系统参数] 素数 p (16进制): fffffffffffffffffffffffffffffffffffffffffffffffffffffffefffffc2f
[系统参数] 生成元 g: 7

--- 步骤 2: 生成公私钥对 ---
[密钥生成] 秘密值 x (不公开): b20d2bf37cddc3d5c57df070389371c44ac67ab5d951acd742870ee42270900b
[密钥生成] 公钥 y = g^x mod p: f6ab218297a898eb238a81f7af0f0f2e5114c4f86082d825a2471d022beb2f44

--- 步骤 3: 证明者生成零知识证明 ---
...
[验证成功] g^s ≡ t * y^c (mod p)，证明有效！✓

--- 验证者验证伪造的证明 ---
[验证失败] g^s ≠ t * y^c (mod p)，证明无效！
✓  伪造的证明被正确拒绝！
```

---

## 验证等式推导

为什么 `g^s ≡ t * y^c (mod p)` 能证明证明者知道 `x`？

```
已知:
  s = r + c*x mod (p-1)
  t = g^r mod p
  y = g^x mod p

推导:
  g^s = g^(r + c*x)
      = g^r * g^(c*x)           [指数法则]
      = g^r * (g^x)^c
      = t * y^c                 [代入定义]

因此:
  g^s ≡ t * y^c (mod p)
```

如果证明者不知道 `x`，就无法构造出满足此等式的 `s`。

---

## 实际应用

零知识证明在 Web3 和区块链中有广泛应用：

| 应用 | 说明 | 项目示例 |
|------|------|----------|
| 隐私交易 | 隐藏交易金额和参与者 | Zcash, Tornado Cash |
| Layer 2 扩容 | 压缩链上验证数据 | zkSync, StarkNet |
| 身份验证 | 证明身份而不泄露信息 | World ID |
| 合规证明 | 证明满足条件而不暴露数据 | zkKYC |

---

## 注意事项

1. **演示用途**: 本实现使用较小的素数用于学习，实际应用需要 2048 位以上的素数
2. **安全参数**: 生产环境应使用经过审计的密码学库
3. **扩展学习**: 可进一步研究 zk-SNARKs、zk-STARKs 等更复杂的零知识证明系统

---

## 参考资源

- [Zero Knowledge Proofs - An Illustrated Primer](https://blog.cryptographyengineering.com/2014/11/27/zero-knowledge-proofs-illustrated-primer/)
- [Schnorr Identification Protocol](https://en.wikipedia.org/wiki/Schnorr_identification_protocol)
- [Fiat-Shamir Heuristic](https://en.wikipedia.org/wiki/Fiat%E2%80%93Shamir_heuristic)
