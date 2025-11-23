package main

import (
	"math"
)

// 難易度調整のパラメータ
const (
	// TargetBlockTime は目標ブロック生成時間（秒）
	TargetBlockTime = 10

	// AdjustmentInterval は難易度調整を行うブロック間隔
	AdjustmentInterval = 10

	// MaxAdjustmentFactor は最大調整倍率（急激な変化を防ぐ）
	MaxAdjustmentFactor = 2.0

	// MinDifficulty は最小難易度
	MinDifficulty = 0

	// MaxDifficulty は最大難易度
	MaxDifficulty = 10
)

// GetAverageBlockTime は直近lastNBlocks個のブロックの平均生成時間を返します（秒）
func GetAverageBlockTime(blockchain *Blockchain, lastNBlocks int) float64 {
	if len(blockchain.Blocks) <= 1 {
		return 0.0
	}

	// 計算可能なブロック数を決定
	blocksToCheck := lastNBlocks
	if len(blockchain.Blocks)-1 < blocksToCheck {
		blocksToCheck = len(blockchain.Blocks) - 1
	}

	if blocksToCheck == 0 {
		return 0.0
	}

	// 直近のブロックから過去に遡って平均時間を計算
	var totalTime int64
	for i := len(blockchain.Blocks) - 1; i >= len(blockchain.Blocks)-blocksToCheck; i-- {
		currentBlock := blockchain.Blocks[i]
		previousBlock := blockchain.Blocks[i-1]
		totalTime += currentBlock.Timestamp - previousBlock.Timestamp
	}

	return float64(totalTime) / float64(blocksToCheck)
}

// AdjustDifficulty は実際の平均時間と目標時間を比較して新しい難易度を返します
func AdjustDifficulty(currentDifficulty int, actualTime, targetTime float64) int {
	if actualTime == 0.0 || targetTime == 0.0 {
		return currentDifficulty
	}

	// 調整比率を計算
	ratio := actualTime / targetTime

	// 急激な変化を防ぐ
	if ratio > MaxAdjustmentFactor {
		ratio = MaxAdjustmentFactor
	} else if ratio < 1.0/MaxAdjustmentFactor {
		ratio = 1.0 / MaxAdjustmentFactor
	}

	// 難易度を調整
	// 実際の時間が目標より長い → 難易度を下げる（マイニングを簡単に）
	// 実際の時間が目標より短い → 難易度を上げる（マイニングを難しく）
	var newDifficulty int
	if ratio > 1.0 {
		// 時間がかかりすぎている → 難易度を下げる
		adjustment := int(math.Ceil(math.Log2(ratio)))
		newDifficulty = currentDifficulty - adjustment
	} else {
		// 時間が短すぎる → 難易度を上げる
		adjustment := int(math.Ceil(math.Log2(1.0 / ratio)))
		newDifficulty = currentDifficulty + adjustment
	}

	// 難易度の範囲を制限
	if newDifficulty < MinDifficulty {
		newDifficulty = MinDifficulty
	} else if newDifficulty > MaxDifficulty {
		newDifficulty = MaxDifficulty
	}

	return newDifficulty
}

// CalculateDifficulty はブロックチェーン全体から次の難易度を計算します
func CalculateDifficulty(blockchain *Blockchain, targetTime int) int {
	// ブロックが少ない場合は現在の難易度を維持
	if len(blockchain.Blocks) < AdjustmentInterval {
		return blockchain.Difficulty
	}

	// 調整間隔でのみ難易度を更新
	if len(blockchain.Blocks)%AdjustmentInterval != 0 {
		return blockchain.Difficulty
	}

	// 直近のブロックの平均生成時間を取得
	avgTime := GetAverageBlockTime(blockchain, AdjustmentInterval)

	// 難易度を調整
	return AdjustDifficulty(blockchain.Difficulty, avgTime, float64(targetTime))
}

// ShouldAdjustDifficulty は難易度調整が必要かどうかを判定します
func ShouldAdjustDifficulty(blockchain *Blockchain) bool {
	return len(blockchain.Blocks) >= AdjustmentInterval &&
		len(blockchain.Blocks)%AdjustmentInterval == 0
}

// GetDifficultyStats は難易度に関する統計情報を返します
type DifficultyStats struct {
	CurrentDifficulty int     // 現在の難易度
	AverageBlockTime  float64 // 平均ブロック生成時間
	TargetBlockTime   int     // 目標ブロック生成時間
	NextAdjustment    int     // 次の調整までのブロック数
}

// GetDifficultyStats は難易度統計を取得します
func GetDifficultyStatsFromChain(blockchain *Blockchain) *DifficultyStats {
	stats := &DifficultyStats{
		CurrentDifficulty: blockchain.Difficulty,
		TargetBlockTime:   TargetBlockTime,
	}

	// 平均ブロック生成時間を計算
	if len(blockchain.Blocks) > 1 {
		stats.AverageBlockTime = GetAverageBlockTime(blockchain, AdjustmentInterval)
	}

	// 次の調整までのブロック数
	if len(blockchain.Blocks) < AdjustmentInterval {
		stats.NextAdjustment = AdjustmentInterval - len(blockchain.Blocks)
	} else {
		stats.NextAdjustment = AdjustmentInterval - (len(blockchain.Blocks) % AdjustmentInterval)
	}

	return stats
}
