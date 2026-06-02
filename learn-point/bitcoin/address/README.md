# 比特币地址生成

本项目使用 Go 语言实现了完整的比特币 P2PKH (Pay-to-Public-Key-Hash) 地址生成流程。

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
cd /Users/mac/learn/web3/2026/06/web3-learning/learn-point/bitcoin/address

# 下载依赖
go mod tidy

# 运行程序
go run main.go
```

## 比特币地址生成流程

### 1. 私钥生成
- 私钥是 256 位（32 字节）的随机数
- 本示例使用硬编码的测试私钥用于演示

### 2. 公钥生成
- 使用 secp256k1 椭圆曲线算法
- 通过私钥与基点 G 进行标量乘法运算得到公钥
- 公钥格式：0x04 + X坐标(32字节) + Y坐标(32字节)

### 3. 公钥压缩
- 根据 Y 坐标的奇偶性，使用前缀 0x02（偶数）或 0x03（奇数）
- 压缩公钥格式：前缀 + X坐标(32字节)

### 4. 公钥哈希
- SHA256(压缩公钥)
- RIPEMD160(SHA256结果)
- 得到 20 字节的公钥哈希

### 5. 添加版本字节
- 主网地址：0x00
- 测试网地址：0x6f

### 6. 计算校验和
- SHA256(版本字节 + 公钥哈希)
- SHA256(第一次SHA256结果)
- 取前 4 字节作为校验和

### 7. Base58 编码
- 使用 Bitcoin Base58 字母表：`123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz`
- 编码：版本字节 + 公钥哈希 + 校验和

## 核心函数说明

- [Base58Encode](file:///Users/mac/learn/web3/2026/06/web3-learning/learn-point/bitcoin/address/main.go#L13-L39) - Base58 编码函数
- [generatePublicKey](file:///Users/mac/learn/web3/2026/06/web3-learning/learn-point/bitcoin/address/main.go#L84-L99) - 生成公钥
- [compressPublicKey](file:///Users/mac/learn/web3/2026/06/web3-learning/learn-point/bitcoin/address/main.go#L101-L116) - 压缩公钥
- [scalarMult](file:///Users/mac/learn/web3/2026/06/web3-learning/learn-point/bitcoin/address/main.go#L118-L133) - 椭圆曲线标量乘法
- [pointAdd](file:///Users/mac/learn/web3/2026/06/web3-learning/learn-point/bitcoin/address/main.go#L135-L164) - 椭圆曲线点加法
- [pointDouble](file:///Users/mac/learn/web3/2026/06/web3-learning/learn-point/bitcoin/address/main.go#L166-L195) - 椭圆曲线点加倍

## 注意事项

⚠️ **重要提示：**
- 本代码仅用于学习演示目的
- 硬编码的私钥仅用于测试，不要在真实环境中使用
- 真实的比特币地址生成需要使用密码学安全的随机数生成器
- 在生产环境中，建议使用经过审计的成熟库

## 依赖

- `golang.org/x/crypto` - 提供 RIPEMD160 哈希算法
