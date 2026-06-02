package main

import (
	"fmt"
	"time"
)

func main() {
	// 理论值（无限等比数列极限）
	theoreticalBTC := 21000000.0

	// 使用聪作为单位计算，避免浮点数精度问题
	totalSatoshi := uint64(0)
	rewardSatoshi := uint64(50 * 100000000) // 50 BTC = 50亿聪
	blocksPerHalving := uint64(210000)
	reduceCount := 0

	for rewardSatoshi > 0 {
		totalSatoshi += rewardSatoshi * blocksPerHalving
		rewardSatoshi /= 2
		reduceCount++
	}

	totalBTC := float64(totalSatoshi) / 100000000.0
	diffBTC := theoreticalBTC - totalBTC
	diffSatoshi := uint64(diffBTC * 100000000.0)

	fmt.Println("=== 比特币发行总量分析 ===")
	fmt.Printf("理论极限总量: %.8f BTC\n", theoreticalBTC)
	fmt.Printf("实际总量: %.8f BTC\n", totalBTC)
	fmt.Printf("差值: %.8f BTC (%d 聪)\n", diffBTC, diffSatoshi)
	fmt.Printf("实际总量（聪）: %d\n", totalSatoshi)
	fmt.Printf("减半次数: %d\n", reduceCount)
	fmt.Printf("预计挖完时间: %d年\n", 2008+reduceCount*4)

	// 计算当前时间的区块奖励
	now := time.Now()
	genesisYear := 2009
	yearsSinceGenesis := now.Year() - genesisYear
	currentHalvingCount := yearsSinceGenesis / 4

	// 计算当前奖励
	currentRewardSatoshi := uint64(50 * 100000000)
	for i := 0; i < currentHalvingCount; i++ {
		currentRewardSatoshi /= 2
		if currentRewardSatoshi == 0 {
			break
		}
	}
	currentRewardBTC := float64(currentRewardSatoshi) / 100000000.0

	fmt.Println("\n=== 当前区块奖励 ===")
	fmt.Printf("当前时间: %s\n", now.Format("2006-01-02"))
	fmt.Printf("已减半次数: %d\n", currentHalvingCount)
	fmt.Printf("当前区块奖励: %.8f BTC\n", currentRewardBTC)
}
