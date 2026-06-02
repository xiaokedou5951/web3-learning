# 以太坊地址生成

本项目使用 Go 语言实现了完整的以太坊地址生成流程，包括 EIP-55 校验和格式。

## 目录结构

```
address/
├── go.mod      # Go 模块配置
├── main.go     # 主程序代码
└── README.md   # 说明文档（本文件）
```

## 运行方式

```bash
# 进入项目目录
cd /Users/mac/learn/web3/2026/06/web3-learning/learn-point/ethereum/address

# 下载依赖
go mod tidy

# 运行程序
go run main.go
```

## 以太坊地址生成流程

### 1. 私钥生成
- 私钥是 256 位（32 字节）的随机数
- 本示例使用硬编码的测试私钥用于演示

### 2. 公钥生成
- 使用 secp256k1 椭圆曲线算法（与比特币相同）
- 通过私钥与基点 G 进行标量乘法运算得到公钥
- 公钥格式：X坐标(32字节) + Y坐标(32字节)，共64字节（注意：没有比特币的 0x04 前缀）

### 3. 公钥哈希
- 使用 Keccak-256 算法对公钥进行哈希（注意：以太坊使用 Keccak-256，不是标准 SHA3-256）
- 得到 32 字节的哈希值

### 4. 地址生成
- 取 Keccak-256 哈希结果的最后 20 字节作为地址
- 添加 "0x" 前缀

### 5. EIP-55 校验和（可选）
- 对地址的小写十六进制字符串进行 Keccak-256 哈希
- 根据哈希值的每一位决定地址字母的大小写
- 哈希值 >= 8 时字母大写，否则小写

## 与比特币地址的区别

| 特性 | 比特币 | 以太坊 |
|------|--------|--------|
| 公钥格式 | 0x04 + X + Y（65字节）或压缩格式 | X + Y（64字节，无压缩） |
| 哈希算法 | SHA256 + RIPEMD160 | Keccak-256 |
| 地址长度 | 20 字节（Base58编码） | 20 字节（Hex编码） |
| 地址前缀 | 0x00（主网） | 0x |
| 校验方式 | Base58Check（4字节校验和） | EIP-55（大小写校验） |

## 核心函数说明

- [keccak256](file:///Users/mac/learn/web3/2026/06/web3-learning/learn-point/ethereum/address/main.go#L83-L87) - Keccak-256 哈希函数
- [generatePublicKey](file:///Users/mac/learn/web3/2026/06/web3-learning/learn-point/ethereum/address/main.go#L59-L79) - 生成以太坊格式公钥
- [toChecksumAddress](file:///Users/mac/learn/web3/2026/06/web3-learning/learn-point/ethereum/address/main.go#L91-L114) - EIP-55 校验和地址生成

## 注意事项

⚠️ **重要提示：**
- 本代码仅用于学习演示目的
- 硬编码的私钥仅用于测试，不要在真实环境中使用
- 真实的以太坊地址生成需要使用密码学安全的随机数生成器
- 在生产环境中，建议使用经过审计的成熟库（如 go-ethereum）
- Keccak-256 ≠ SHA3-256，这是以太坊使用的特殊版本

## 依赖

- `golang.org/x/crypto` - 提供 Keccak-256 哈希算法
