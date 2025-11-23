package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetAverageBlockTime(t *testing.T) {
	t.Run("ブロックが1つの場合", func(t *testing.T) {
		bc := NewBlockchain(1)

		avgTime := GetAverageBlockTime(bc, 10)

		assert.Equal(t, 0.0, avgTime)
	})

	t.Run("ブロックが2つの場合", func(t *testing.T) {
		bc := NewBlockchain(1)
		bc.Blocks[0].Timestamp = 0

		// 手動でタイムスタンプを設定
		block1 := &Block{Index: 1, Timestamp: 10, Data: "Block 1", PreviousHash: bc.Blocks[0].Hash, Difficulty: 1}
		_, _ = MineBlock(block1, 1)
		bc.Blocks = append(bc.Blocks, block1)

		avgTime := GetAverageBlockTime(bc, 10)

		// 10秒のはず
		assert.Equal(t, 10.0, avgTime)
	})

	t.Run("複数ブロックの平均時間", func(t *testing.T) {
		bc := NewBlockchain(1)

		// タイムスタンプを手動で設定してテスト
		// ブロック0: 0秒
		// ブロック1: 10秒
		// ブロック2: 20秒
		// ブロック3: 30秒
		bc.Blocks[0].Timestamp = 0

		block1 := &Block{Index: 1, Timestamp: 10, Data: "Block 1", PreviousHash: bc.Blocks[0].Hash, Difficulty: 1}
		_, _ = MineBlock(block1, 1)
		bc.Blocks = append(bc.Blocks, block1)

		block2 := &Block{Index: 2, Timestamp: 20, Data: "Block 2", PreviousHash: block1.Hash, Difficulty: 1}
		_, _ = MineBlock(block2, 1)
		bc.Blocks = append(bc.Blocks, block2)

		block3 := &Block{Index: 3, Timestamp: 30, Data: "Block 3", PreviousHash: block2.Hash, Difficulty: 1}
		_, _ = MineBlock(block3, 1)
		bc.Blocks = append(bc.Blocks, block3)

		// 直近3ブロックの平均: (10-0 + 20-10 + 30-20) / 3 = 30 / 3 = 10秒
		avgTime := GetAverageBlockTime(bc, 3)

		assert.Equal(t, 10.0, avgTime)
	})

	t.Run("lastNBlocksより少ないブロック数", func(t *testing.T) {
		bc := NewBlockchain(1)
		bc.Blocks[0].Timestamp = 0

		block1 := &Block{Index: 1, Timestamp: 10, Data: "Block 1", PreviousHash: bc.Blocks[0].Hash, Difficulty: 1}
		_, _ = MineBlock(block1, 1)
		bc.Blocks = append(bc.Blocks, block1)

		// 10ブロック要求するが、2ブロックしかない
		// (10 - 0) / 1 = 10秒
		avgTime := GetAverageBlockTime(bc, 10)

		assert.Equal(t, 10.0, avgTime)
	})
}

func TestAdjustDifficulty(t *testing.T) {
	t.Run("実際の時間が目標時間と同じ", func(t *testing.T) {
		newDiff := AdjustDifficulty(2, 10.0, 10.0)

		// 変化なし
		assert.Equal(t, 2, newDiff)
	})

	t.Run("実際の時間が目標時間の2倍（遅い）", func(t *testing.T) {
		newDiff := AdjustDifficulty(2, 20.0, 10.0)

		// 難易度を下げる（簡単にする）
		assert.Less(t, newDiff, 2)
	})

	t.Run("実際の時間が目標時間の半分（速い）", func(t *testing.T) {
		newDiff := AdjustDifficulty(2, 5.0, 10.0)

		// 難易度を上げる（難しくする）
		assert.Greater(t, newDiff, 2)
	})

	t.Run("実際の時間が極端に長い（MaxAdjustmentFactorで制限）", func(t *testing.T) {
		// 100倍遅いが、MaxAdjustmentFactor=2で制限される
		newDiff := AdjustDifficulty(5, 1000.0, 10.0)

		// 最大でも1段階しか下がらない
		assert.GreaterOrEqual(t, newDiff, 4)
	})

	t.Run("実際の時間が極端に短い（MaxAdjustmentFactorで制限）", func(t *testing.T) {
		// 100倍速いが、MaxAdjustmentFactor=2で制限される
		newDiff := AdjustDifficulty(5, 0.1, 10.0)

		// 最大でも1段階しか上がらない
		assert.LessOrEqual(t, newDiff, 6)
	})

	t.Run("難易度の最小値制限", func(t *testing.T) {
		// 非常に遅い時間で難易度0から調整
		newDiff := AdjustDifficulty(0, 100.0, 10.0)

		// MinDifficulty = 0以下にはならない
		assert.GreaterOrEqual(t, newDiff, MinDifficulty)
	})

	t.Run("難易度の最大値制限", func(t *testing.T) {
		// 非常に速い時間で難易度10から調整
		newDiff := AdjustDifficulty(10, 0.1, 10.0)

		// MaxDifficulty = 10以上にはならない
		assert.LessOrEqual(t, newDiff, MaxDifficulty)
	})

	t.Run("actualTimeが0の場合", func(t *testing.T) {
		newDiff := AdjustDifficulty(2, 0.0, 10.0)

		// 変化なし
		assert.Equal(t, 2, newDiff)
	})

	t.Run("targetTimeが0の場合", func(t *testing.T) {
		newDiff := AdjustDifficulty(2, 10.0, 0.0)

		// 変化なし
		assert.Equal(t, 2, newDiff)
	})
}

func TestCalculateDifficulty(t *testing.T) {
	t.Run("ブロック数が調整間隔未満", func(t *testing.T) {
		bc := NewBlockchain(2)

		// AdjustmentInterval = 10なので、9ブロック追加
		for i := 1; i < AdjustmentInterval; i++ {
			bc.AddBlock("Block " + string(rune(i+'0')))
		}

		newDiff := CalculateDifficulty(bc, TargetBlockTime)

		// 現在の難易度を維持
		assert.Equal(t, bc.Difficulty, newDiff)
	})

	t.Run("調整間隔でない場合", func(t *testing.T) {
		bc := NewBlockchain(2)

		// 11ブロック（調整間隔=10の倍数ではない）
		for i := 1; i <= 11; i++ {
			bc.AddBlock("Block " + string(rune(i+'0')))
		}

		newDiff := CalculateDifficulty(bc, TargetBlockTime)

		// 現在の難易度を維持
		assert.Equal(t, bc.Difficulty, newDiff)
	})

	t.Run("調整間隔での難易度計算", func(t *testing.T) {
		bc := NewBlockchain(2)

		// タイムスタンプを手動で設定
		bc.Blocks[0].Timestamp = 0

		// 10ブロック追加（各ブロック間20秒 = 目標の2倍遅い）
		for i := 1; i < AdjustmentInterval; i++ {
			block := &Block{
				Index:        int64(i),
				Timestamp:    int64(i * 20), // 20秒間隔
				Data:         "Block " + string(rune(i+'0')),
				PreviousHash: bc.Blocks[i-1].Hash,
				Difficulty:   2,
			}
			_, _ = MineBlock(block, 2)
			bc.Blocks = append(bc.Blocks, block)
		}

		newDiff := CalculateDifficulty(bc, TargetBlockTime)

		// 平均20秒、目標10秒なので、難易度を下げるべき
		assert.Less(t, newDiff, bc.Difficulty)
	})
}

func TestShouldAdjustDifficulty(t *testing.T) {
	t.Run("ブロック数が調整間隔未満", func(t *testing.T) {
		bc := NewBlockchain(2)

		// 5ブロック（調整間隔=10未満）
		for i := 1; i < 5; i++ {
			bc.AddBlock("Block " + string(rune(i+'0')))
		}

		shouldAdjust := ShouldAdjustDifficulty(bc)

		assert.False(t, shouldAdjust)
	})

	t.Run("ブロック数が調整間隔ちょうど", func(t *testing.T) {
		bc := NewBlockchain(2)

		// 9ブロック追加（合計10ブロック）
		for i := 1; i < AdjustmentInterval; i++ {
			bc.AddBlock("Block " + string(rune(i+'0')))
		}

		shouldAdjust := ShouldAdjustDifficulty(bc)

		assert.True(t, shouldAdjust)
	})

	t.Run("ブロック数が調整間隔の倍数", func(t *testing.T) {
		bc := NewBlockchain(2)

		// 19ブロック追加（合計20ブロック）
		for i := 1; i < AdjustmentInterval*2; i++ {
			bc.AddBlock("Block " + string(rune(i+'0')))
		}

		shouldAdjust := ShouldAdjustDifficulty(bc)

		assert.True(t, shouldAdjust)
	})

	t.Run("ブロック数が調整間隔の倍数でない", func(t *testing.T) {
		bc := NewBlockchain(2)

		// 10ブロック追加（合計11ブロック）
		for i := 1; i <= AdjustmentInterval; i++ {
			bc.AddBlock("Block " + string(rune(i+'0')))
		}

		shouldAdjust := ShouldAdjustDifficulty(bc)

		assert.False(t, shouldAdjust)
	})
}

func TestGetDifficultyStats(t *testing.T) {
	t.Run("統計情報の取得", func(t *testing.T) {
		bc := NewBlockchain(2)
		bc.Blocks[0].Timestamp = 0

		// 5ブロック追加（手動でタイムスタンプ設定）
		for i := 1; i <= 5; i++ {
			block := &Block{
				Index:        int64(i),
				Timestamp:    int64(i * 10), // 10秒間隔
				Data:         "Block " + string(rune(i+'0')),
				PreviousHash: bc.Blocks[i-1].Hash,
				Difficulty:   2,
			}
			_, _ = MineBlock(block, 2)
			bc.Blocks = append(bc.Blocks, block)
		}

		stats := GetDifficultyStatsFromChain(bc)

		require.NotNil(t, stats)
		assert.Equal(t, 2, stats.CurrentDifficulty)
		assert.Equal(t, TargetBlockTime, stats.TargetBlockTime)
		assert.Equal(t, 10.0, stats.AverageBlockTime)
		// 6ブロック存在、次の調整は10ブロック時なので、あと4ブロック
		assert.Equal(t, 4, stats.NextAdjustment)
	})

	t.Run("ジェネシスブロックのみの場合", func(t *testing.T) {
		bc := NewBlockchain(2)

		stats := GetDifficultyStatsFromChain(bc)

		require.NotNil(t, stats)
		assert.Equal(t, 2, stats.CurrentDifficulty)
		assert.Equal(t, 0.0, stats.AverageBlockTime)
		// 次の調整まであと9ブロック
		assert.Equal(t, 9, stats.NextAdjustment)
	})

	t.Run("調整間隔直前", func(t *testing.T) {
		bc := NewBlockchain(2)

		// 8ブロック追加（合計9ブロック）
		for i := 1; i < AdjustmentInterval-1; i++ {
			bc.AddBlock("Block " + string(rune(i+'0')))
		}

		stats := GetDifficultyStatsFromChain(bc)

		// 次の調整まであと1ブロック
		assert.Equal(t, 1, stats.NextAdjustment)
	})
}
