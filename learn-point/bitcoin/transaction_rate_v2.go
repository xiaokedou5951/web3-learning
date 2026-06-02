package main

import "fmt"

func main() {
	// 比特币网络平均出块时间约为 10 分钟
	minutesPerBlock := 10
	secondsPerBlock := minutesPerBlock * 60
	
	// 比特币区块大小限制为 1MB（1,048,576 字节）
	blockSizeBytes := 1024 * 1024 // 1MB
	
	// 平均每笔交易大小约为 250 字节（参考实际数据）
	avgTxSizeBytes := 250
	
	// 计算每个区块理论上能容纳的交易数量
	txsPerBlock := blockSizeBytes / avgTxSizeBytes
	
	// 计算每秒成交量
	transactionsPerSecond := float64(txsPerBlock) / float64(secondsPerBlock)
	
	fmt.Println("=== 比特币每秒成交量分析（基于区块大小）===")
	fmt.Printf("平均出块时间: %d 分钟 (%d 秒)\n", minutesPerBlock, secondsPerBlock)
	fmt.Printf("区块大小限制: %d 字节 (1MB)\n", blockSizeBytes)
	fmt.Printf("平均每笔交易大小: %d 字节\n", avgTxSizeBytes)
	fmt.Printf("理论每区块交易数量: %d 笔\n", txsPerBlock)
	fmt.Printf("每秒成交量: %.2f 笔/秒\n", transactionsPerSecond)
	fmt.Printf("每分钟成交量: %.0f 笔/分钟\n", transactionsPerSecond*60)
	fmt.Printf("每小时成交量: %.0f 笔/小时\n", transactionsPerSecond*60*60)
	fmt.Printf("每天成交量: %.0f 笔/天\n", transactionsPerSecond*60*60*24)
}
