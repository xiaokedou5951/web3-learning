package main

import "fmt"

func main() {
	// 比特币网络平均出块时间约为 10 分钟
	minutesPerBlock := 10
	secondsPerBlock := minutesPerBlock * 60
	
	// 平均每个区块的交易数量约为 4200 笔（参考实际数据）
	txsPerBlock := 4200
	
	// 计算每秒成交量
	transactionsPerSecond := float64(txsPerBlock) / float64(secondsPerBlock)
	
	fmt.Println("=== 比特币每秒成交量分析 ===")
	fmt.Printf("平均出块时间: %d 分钟 (%d 秒)\n", minutesPerBlock, secondsPerBlock)
	fmt.Printf("平均每区块交易数量: %d 笔\n", txsPerBlock)
	fmt.Printf("每秒成交量: %.2f 笔/秒\n", transactionsPerSecond)
	fmt.Printf("每分钟成交量: %.0f 笔/分钟\n", transactionsPerSecond*60)
	fmt.Printf("每小时成交量: %.0f 笔/小时\n", transactionsPerSecond*60*60)
	fmt.Printf("每天成交量: %.0f 笔/天\n", transactionsPerSecond*60*60*24)
}
