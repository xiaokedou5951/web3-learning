# 以太坊 Gas 费用机制详解（EIP-1559 后）

本文档系统梳理以太坊（Ethereum）在 **EIP-1559（伦敦硬分叉，2021 年 8 月）** 之后的 Gas 费用机制。重点讲解：

1. 一笔交易中所有核心指标的具体定义与计算方式
2. 各指标之间的公式推演关系
3. 实际支付给矿工（Miner / Validator）的 **Priority Fee**（小费）计算公式与直觉

---

## 一、核心概念与术语速查表

| 指标 | 英文/缩写 | 单位 | 含义 |
|------|-----------|------|------|
| Gas 上限 | `gasLimit` | Gas | 用户愿意为这笔交易最多花费的计算工作量 |
| Gas 实际消耗 | `gasUsed` | Gas | 交易实际执行消耗的计算工作量（`gasUsed ≤ gasLimit`） |
| 基础费用 | `baseFeePerGas` | Gwei/Gas | 由协议根据区块拥堵程度动态调整，**必须支付**且会被**销毁** |
| 最高费用 | `maxFeePerGas` | Gwei/Gas | 用户愿意支付的最高单价（含 base fee + tip） |
| 最高小费 | `maxPriorityFeePerGas` | Gwei/Gas | 用户愿意支付给矿工的最高小费（奖励激励） |
| 实际小费 / Priority Fee | `priorityFeePerGas` | Gwei/Gas | 实际支付给矿工的小费单价（下文推导） |
| 有效 Gas 单价 | `effectiveGasPrice` | Gwei/Gas | 实际每 Gas 总花费单价 = `baseFee + priorityFee` |
| 交易手续费总额 | `transactionFee` | ETH | = `gasUsed × effectiveGasPrice` |
| 销毁金额 | `burntFee` | ETH | = `gasUsed × baseFeePerGas`（销毁，从流通中移除） |
| 矿工奖励 | `minerReward` | ETH | = `gasUsed × priorityFeePerGas`（实际支付给矿工） |

> 💡 区分 **Legacy 交易** 与 **EIP-1559 交易**：
> - **Legacy（旧格式）**：只有 `gasPrice` 一个参数（`gasPrice = baseFee + priorityFee`），无法分开指定。
> - **EIP-1559（Type 2）**：有 `maxFeePerGas` 与 `maxPriorityFeePerGas` 两个参数，用户对"最多花多少"和"给矿工多少 tip"都有明确控制。
>
> 下文以 EIP-1559 交易为默认讲解对象。

---

## 二、所有核心指标的计算方式与公式推演

### 2.1 `gasUsed`——交易实际消耗的 Gas

以太坊上的每一次操作（如加法、哈希、存储写入、合约调用）都有固定的 Gas 成本，由 **Yellow Paper** 定义。

**核心公式：**

```
gasUsed = Σ opGas[i]   对交易中执行的所有操作 i 求和
```

**典型常数：**

| 操作 | Gas 消耗 | 说明 |
|------|---------|------|
| 转账普通 ETH（非合约） | **21,000 Gas** | 普通转账固定消耗 |
| 往合约写入一个新 slot | 20,000 Gas | SSTORE 冷存储写入 |
| 往合约改写一个已有 slot | 5,000 Gas | SSTORE 热改写 |
| Keccak-256 哈希一次 | 30 + 6 × word | 基础 30 + 每字 6 |
| 调用另一个合约 | 最低 100 Gas | CALL 基础成本 |
| 交易数据的零字节（`0x00`） | 4 Gas / byte | `calldata` 每零字节 |
| 交易数据的非零字节 | 16 Gas / byte | `calldata` 每非零字节 |

**直观理解：**

- 一笔最简单的 ETH 转账 = **21,000 Gas**
- 一笔带数据的转账或简单 ERC-20 `approve` ≈ **46,000 – 65,000 Gas**
- 一次复杂 DeFi 交互（如 Uniswap V3 swap） ≈ **120,000 – 350,000 Gas**

**约束：**

```
gasUsed ≤ gasLimit           （否则交易 OOG 回滚）
Σ gasUsed 同一区块 ≤ blockGasLimit
```

---

### 2.2 `baseFeePerGas`——基础费用的动态调整

EIP-1559 之后，每个区块都有一个由协议**自动计算**的 `baseFeePerGas`。

**目标区块大小**：`TARGET_GAS = 15,000,000`（目标）

**区块容量上限**：`MAX_BLOCK_GAS = 30,000,000`（最大，即 2× 目标）

**核心调整公式**（EIP-1559 第 17 条）：

```
baseFee_new = baseFee_parent + baseFee_parent × (gasUsed_parent − TARGET_GAS) / TARGET_GAS / 8
```

等价写法：

```
baseFee_new = baseFee_parent × ⎛ 1 + (Δ / TARGET_GAS) / 8 ⎞

其中 Δ = gasUsed_parent − TARGET_GAS
```

**三种状态：**

| 状态 | `gasUsed_parent` | `baseFee_new` 变化 |
|------|------------------|-------------------|
| 正好 | `= 15M` | 不变（`baseFee_new = baseFee_parent`） |
| 超满 | `> 15M` | 上涨（最多 +12.5% / 区块） |
| 不满 | `< 15M` | 下跌（最少 −12.5% / 区块） |

**最大波动幅度**：单区块 ±12.5%

```
max_delta = baseFee_parent × 1/8 = 12.5%
```

**直觉：**

- 当区块"超堵"（用满 30M，即 2× 目标）：
  ```
  Δ = 15M
  baseFee_new = baseFee × (1 + 15M/15M / 8) = baseFee × 1.125
  ```
  每个满区块上涨 12.5%，连续 6 个满区块 ≈ 翻 1 倍。

- 当区块"空"（只用了 0）：
  ```
  Δ = −15M
  baseFee_new = baseFee × (1 − 15M/15M / 8) = baseFee × 0.875
  ```
  每个空区块下跌 12.5%。

**为什么 base fee 会被销毁？**

`baseFee` 被销毁（不是给矿工）是 EIP-1559 的核心设计，目的是：
1. 让矿工没有动机**操控 Gas 价格**（矿工拿不到 base fee）
2. 使 ETH 具有**通缩 / 低通胀**属性（部分 ETH 永久销毁）

---

### 2.3 `effectiveGasPrice`——用户实际每 Gas 支付价格

**核心公式**：

```
effectiveGasPrice = baseFeePerGas + priorityFeePerGas
```

其中 `priorityFeePerGas` 的计算见第三节。

**与用户设置参数的关系**：

- 用户设置 `maxFeePerGas`（最多每 Gas 愿意支付多少）
- 用户设置 `maxPriorityFeePerGas`（最多每 Gas 愿意给矿工多少小费）
- 最终 `effectiveGasPrice` **不超过** `maxFeePerGas`

---

### 2.4 `transactionFee`——整笔交易总手续费

**核心公式**：

```
transactionFee = gasUsed × effectiveGasPrice
```

**展开**：

```
transactionFee = gasUsed × (baseFeePerGas + priorityFeePerGas)
               = gasUsed × baseFeePerGas  +  gasUsed × priorityFeePerGas
               = burntFee                +  minerReward
```

**单位换算**：

- `1 Gwei = 10^9 Wei`
- `1 ETH  = 10^18 Wei = 10^9 Gwei`
- 以 ETH 计价的手续费：

```
fee_ETH = gasUsed × effectiveGasPrice_Gwei / 10^9
```

---

### 2.5 `burntFee`——被销毁的金额

```
burntFee = gasUsed × baseFeePerGas
```

这部分 ETH 被永久从流通中移除（发送到一个无法取出的地址 `0x0...0` 或由协议逻辑销毁）。

---

### 2.6 `minerReward`——矿工/验证者实际收到的奖励

```
minerReward = gasUsed × priorityFeePerGas
```

这就是"实际支付给矿工的费用"，下一节详细推导 `priorityFeePerGas`。

---

## 三、Priority Fee（小费）——矿工实际收入的计算

### 3.1 用户设置的两个参数

用户发送 EIP-1559 交易时，设置了：

```
maxFeePerGas            （单位：Gwei/Gas）—— 用户愿意支付的最高单价
maxPriorityFeePerGas    （单位：Gwei/Gas）—— 用户愿意支付给矿工的最高小费
```

### 3.2 矿工实际拿到的小费：`priorityFeePerGas`

**核心公式**：

```
priorityFeePerGas = min( maxPriorityFeePerGas, maxFeePerGas − baseFeePerGas )
```

**等价写法**（更直观）：

```
priorityFeePerGas = maxPriorityFeePerGas
                    ,  当 maxPriorityFeePerGas ≤ maxFeePerGas − baseFeePerGas

priorityFeePerGas = maxFeePerGas − baseFeePerGas
                    ,  当 maxPriorityFeePerGas > maxFeePerGas − baseFeePerGas
```

**为什么是这样？**

矿工从这笔交易里总共得到的手续费 = `gasUsed × effectiveGasPrice`，但其中 `gasUsed × baseFeePerGas` 会被协议销毁，所以：

```
minerReward = gasUsed × effectiveGasPrice − gasUsed × baseFeePerGas
            = gasUsed × (effectiveGasPrice − baseFeePerGas)
```

因此矿工**每 Gas 的净收入**就是 `effectiveGasPrice − baseFeePerGas`，我们称它为 `priorityFeePerGas`。

同时，用户约束必须被满足：

```
effectiveGasPrice ≤ maxFeePerGas                 —— 单价不超过用户上限
effectiveGasPrice − baseFeePerGas ≤ maxPriorityFeePerGas   —— 小费不超过用户 tip 上限
```

联立两个约束：

```
effectiveGasPrice ≤ maxFeePerGas
effectiveGasPrice ≤ baseFeePerGas + maxPriorityFeePerGas
```

所以矿工能拿到的最大小费：

```
priorityFeePerGas_max = effectiveGasPrice_max − baseFeePerGas
                      = min(maxFeePerGas, baseFeePerGas + maxPriorityFeePerGas) − baseFeePerGas
                      = min(maxFeePerGas − baseFeePerGas, maxPriorityFeePerGas)
```

这就推导出了前面的核心公式：

```
priorityFeePerGas = min( maxPriorityFeePerGas, maxFeePerGas − baseFeePerGas )
```

### 3.3 结论：矿工实际收入 = gasUsed × priorityFeePerGas

```
minerReward = gasUsed × min( maxPriorityFeePerGas, maxFeePerGas − baseFeePerGas )
```

### 3.4 三种典型情形

**情形 A：maxFeePerGas 充足，小费按用户 tip 上限支付**

```
maxFeePerGas = 50 Gwei
maxPriorityFeePerGas = 2 Gwei
baseFeePerGas = 30 Gwei

priorityFeePerGas = min(2, 50 − 30) = min(2, 20) = 2 Gwei
effectiveGasPrice   = 30 + 2 = 32 Gwei   （≤ 50，满足约束）
```

用户按自己"想给矿工 2 Gwei"的意图支付。

**情形 B：baseFee 飙升，maxFee 刚够覆盖 baseFee，小费被压缩**

```
maxFeePerGas = 50 Gwei
maxPriorityFeePerGas = 2 Gwei
baseFeePerGas = 49 Gwei   （网络拥堵暴涨）

priorityFeePerGas = min(2, 50 − 49) = min(2, 1) = 1 Gwei
effectiveGasPrice   = 49 + 1 = 50 Gwei   （= maxFeePerGas，触顶）
```

用户想给 2 Gwei 小费，但总预算只剩 1 Gwei 空间，小费降到 1 Gwei。

**情形 C：baseFee 超过 maxFee——交易不会被打包**

```
maxFeePerGas = 40 Gwei
baseFeePerGas = 50 Gwei

maxFeePerGas − baseFeePerGas = −10 Gwei   （负数）
priorityFeePerGas = min(maxPriorityFeePerGas, −10) < 0
```

此时 `effectiveGasPrice ≥ baseFeePerGas > maxFeePerGas`，不满足用户约束，**矿工不会打包该交易**，交易留在 mempool 里直到 base fee 回落。

> 💡 这就是为什么当网络极度拥堵时，你会看到一笔交易"卡住了几个区块"——它不是失败了，而是在等 base fee 降到你设置的 `maxFeePerGas` 以下。

### 3.5 为什么协议用 `min` 而不是 `max`？

这是 EIP-1559 机制的核心思想：**矿工在不违反用户预算的前提下，取两者中更保守的那个作为小费**。直觉：

- `maxPriorityFeePerGas` 是用户对"给矿工多少 tip"的上限
- `maxFeePerGas − baseFeePerGas` 是总预算扣掉 base fee 后剩下的空间
- 取 **min** 保证两者都不被突破

---

## 四、完整计算链：一个数字实例

假设：

```
gasLimit              = 100,000 Gas
gasUsed               = 60,000 Gas
maxFeePerGas          = 40 Gwei
maxPriorityFeePerGas  = 3 Gwei
baseFeePerGas         = 25 Gwei
```

**Step 1 — 计算 priority fee 单价**

```
priorityFeePerGas = min(maxPriorityFeePerGas, maxFeePerGas − baseFeePerGas)
                  = min(3, 40 − 25)
                  = min(3, 15)
                  = 3 Gwei
```

**Step 2 — 计算实际每 Gas 支付价格**

```
effectiveGasPrice = baseFeePerGas + priorityFeePerGas
                  = 25 + 3
                  = 28 Gwei          （≤ 40 Gwei，满足 maxFee 约束）
```

**Step 3 — 计算各部分费用**

```
transactionFee = gasUsed × effectiveGasPrice
               = 60,000 × 28
               = 1,680,000 Gwei
               = 0.00168 ETH

burntFee       = gasUsed × baseFeePerGas
               = 60,000 × 25
               = 1,500,000 Gwei       （永久销毁）
               = 0.00150 ETH

minerReward    = gasUsed × priorityFeePerGas
               = 60,000 × 3
               = 180,000 Gwei          （给矿工）
               = 0.00018 ETH
```

**Step 4 — 校验**

```
burntFee + minerReward = 1,500,000 + 180,000 = 1,680,000 Gwei = transactionFee   ✓
```

---

## 五、公式速查总表

| 指标 | 公式 |
|------|------|
| 实际小费单价 | `priorityFeePerGas = min(maxPriorityFeePerGas, maxFeePerGas − baseFeePerGas)` |
| 实际单价 | `effectiveGasPrice = baseFeePerGas + priorityFeePerGas` |
| 交易总手续费 | `transactionFee = gasUsed × effectiveGasPrice` |
| 销毁金额 | `burntFee = gasUsed × baseFeePerGas` |
| 矿工奖励（给矿工的实际费用） | `minerReward = gasUsed × priorityFeePerGas` |
| 总金额（转账金额 + 手续费） | `totalCost = value + transactionFee` |
| 下一区块 base fee | `baseFee_new = baseFee_parent × (1 + (gasUsed_parent − 15M) / 15M / 8)` |

---

## 六、Legacy 交易与 EIP-1559 交易对比

### Legacy 交易（旧格式）

用户只设置一个 `gasPrice`：

```
transactionFee = gasUsed × gasPrice
minerReward    = gasUsed × (gasPrice − baseFeePerGas)   （剩余部分归矿工）
burntFee       = gasUsed × baseFeePerGas
```

矿工从 Legacy 交易拿到的"小费" = `gasPrice − baseFeePerGas`，用户无法单独控制。

### EIP-1559 交易（新格式，Type 2）

```
priorityFeePerGas = min(maxPriorityFeePerGas, maxFeePerGas − baseFeePerGas)
minerReward       = gasUsed × priorityFeePerGas
```

用户分别对"最多花多少"和"给矿工多少 tip"做预算，更精细。

---

## 七、常见误区

| 误区 | 正确理解 |
|------|---------|
| "矿工拿到 `maxPriorityFeePerGas`" | 只有在 `maxFeePerGas − baseFeePerGas ≥ maxPriorityFeePerGas` 时才是；否则小费会被压缩 |
| "矿工拿到 `maxFeePerGas`" | 矿工只拿到 `min(maxPriorityFeePerGas, maxFee − baseFee)` × gasUsed，其余销毁或根本没花出去 |
| "base fee 是矿工设置的" | base fee 由协议根据上一区块使用率公式计算，矿工无权修改 |
| "交易被打包后，用户付的就是 `maxFeePerGas`" | 实际单价是 `baseFee + priorityFee`，`maxFeePerGas` 只是上限 |
| "gasLimit 越高越贵" | 最终费用按 `gasUsed` 算，不是 `gasLimit`；但 `gasLimit` 过低会导致 OOG 回滚并损失已用 gas |

---

## 八、小结：一句话总结 Priority Fee

> **矿工实际收入 = `gasUsed × min(maxPriorityFeePerGas, maxFeePerGas − baseFeePerGas)`**

直觉：**用户在"给矿工多少 tip"与"扣完 base fee 还剩多少预算"之间，取较小的那个付给矿工**。这就是 EIP-1559 费用模型最核心的一行逻辑。
