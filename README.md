# Web3 学习笔记

本项目是 Web3 区块链技术的系统性学习记录，使用 Go 语言实现核心算法与机制。

## 目录结构

```
web3-learning/
├── learn-point/
│   ├── bitcoin/          # 比特币相关学习
│   │   ├── address/      # 比特币地址生成（P2PKH）
│   │   ├── pow/          # 工作量证明（Proof of Work）
│   │   ├── reward.go     # 区块奖励与减半机制
│   │   └── transaction_rate.go  # 交易费率相关
│   └── ethereum/         # 以太坊相关学习
│       └── address/      # 以太坊地址生成
└── LICENSE
```

## 学习内容纲要

### 一、比特币（Bitcoin）

#### 1. 密码学基础
- [地址生成](learn-point/bitcoin/address/) - 完整的 P2PKH 地址生成流程
  - secp256k1 椭圆曲线算法
  - 公钥生成与压缩
  - SHA256 + RIPEMD160 双重哈希
  - Base58Check 编码与校验和

#### 2. 共识机制
- [工作量证明（PoW）](learn-point/bitcoin/pow/) - 比特币核心共识算法
  - SHA256 双重哈希原理
  - 难度调整机制
  - 挖矿与验证的不对称性
  - 区块链链接原理

#### 3. 经济学模型
- [区块奖励](learn-point/bitcoin/reward.go) - 比特币发行机制
  - 初始奖励 50 BTC
  - 每 210,000 区块减半
  - 2100 万上限计算
  - 当前区块奖励查询

#### 4. 网络与交易
- 交易费率相关（待完善）

### 二、以太坊（Ethereum）

#### 1. 地址系统
- [地址生成](learn-point/ethereum/address/) - 以太坊账户地址生成
  - secp256k1 公钥生成
  - Keccak-256 哈希算法
  - 20 字节地址提取
  - EIP-55 校验和格式

## 运行方式

每个子模块都是独立的 Go 项目，可直接运行：

```bash
# 进入具体模块目录
cd learn-point/bitcoin/address

# 下载依赖并运行
go mod tidy
go run main.go
```

## 技术栈

- **编程语言**: Go
- **核心库**: `golang.org/x/crypto`（RIPEMD160、Keccak-256）
- **学习重点**: 密码学原语、共识算法、协议原理

## 学习路径建议

1. **密码学基础** → 先学习地址生成，理解椭圆曲线与哈希函数
2. **共识机制** → 学习 PoW，理解比特币如何达成去中心化共识
3. **经济模型** → 理解代币发行与通胀机制
4. **对比学习** → 对比比特币与以太坊的异同

## 参考资源

- [比特币白皮书](https://bitcoin.org/bitcoin.pdf)
- [以太坊白皮书](https://ethereum.org/en/whitepaper/)
- [Bitcoin Wiki](https://en.bitcoin.it/)
- [Mastering Bitcoin](https://github.com/bitcoinbook/bitcoinbook)
