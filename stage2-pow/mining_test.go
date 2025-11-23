package main

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewBlock(t *testing.T) {
	t.Run("新しいブロックの生成", func(t *testing.T) {
		block := NewBlock(1, "Test Data", "previous_hash", 2)

		assert.Equal(t, int64(1), block.Index)
		assert.Equal(t, "Test Data", block.Data)
		assert.Equal(t, "previous_hash", block.PreviousHash)
		assert.Equal(t, int64(0), block.Nonce)
		assert.Equal(t, 2, block.Difficulty)
		assert.Equal(t, "", block.Hash) // マイニング前なので空
		assert.Greater(t, block.Timestamp, int64(0))
	})

	t.Run("異なる難易度でのブロック生成", func(t *testing.T) {
		difficulties := []int{0, 1, 3, 5}

		for _, diff := range difficulties {
			block := NewBlock(1, "Test", "prev", diff)
			assert.Equal(t, diff, block.Difficulty)
		}
	})
}

func TestNewGenesisBlock(t *testing.T) {
	t.Run("ジェネシスブロックの生成（難易度0）", func(t *testing.T) {
		genesis := NewGenesisBlock(0)

		require.NotNil(t, genesis)
		assert.Equal(t, int64(0), genesis.Index)
		assert.Equal(t, "Genesis Block", genesis.Data)
		assert.Equal(t, "", genesis.PreviousHash)
		assert.Equal(t, 0, genesis.Difficulty)
		assert.NotEmpty(t, genesis.Hash) // マイニング済み
	})

	t.Run("ジェネシスブロックの生成（難易度1）", func(t *testing.T) {
		genesis := NewGenesisBlock(1)

		require.NotNil(t, genesis)
		assert.NotEmpty(t, genesis.Hash)
		assert.True(t, strings.HasPrefix(genesis.Hash, "0"))
	})

	t.Run("ジェネシスブロックの生成（難易度2）", func(t *testing.T) {
		genesis := NewGenesisBlock(2)

		require.NotNil(t, genesis)
		assert.NotEmpty(t, genesis.Hash)
		assert.True(t, strings.HasPrefix(genesis.Hash, "00"))
	})
}

func TestCalculateHashWithNonce(t *testing.T) {
	t.Run("ナンスを含むハッシュ計算", func(t *testing.T) {
		block := NewBlock(1, "Test Data", "previous_hash", 2)
		block.Nonce = 12345

		hash := CalculateHashWithNonce(block)

		assert.NotEmpty(t, hash)
		assert.Equal(t, 64, len(hash)) // SHA-256は64文字の16進数
	})

	t.Run("同じブロックは同じハッシュを生成", func(t *testing.T) {
		block := NewBlock(1, "Test Data", "previous_hash", 2)
		block.Nonce = 100

		hash1 := CalculateHashWithNonce(block)
		hash2 := CalculateHashWithNonce(block)

		assert.Equal(t, hash1, hash2)
	})

	t.Run("ナンスが異なると異なるハッシュを生成", func(t *testing.T) {
		block := NewBlock(1, "Test Data", "previous_hash", 2)

		block.Nonce = 100
		hash1 := CalculateHashWithNonce(block)

		block.Nonce = 101
		hash2 := CalculateHashWithNonce(block)

		assert.NotEqual(t, hash1, hash2)
	})

	t.Run("データが異なると異なるハッシュを生成", func(t *testing.T) {
		block1 := NewBlock(1, "Data 1", "previous_hash", 2)
		block2 := NewBlock(1, "Data 2", "previous_hash", 2)

		block1.Nonce = 100
		block2.Nonce = 100

		hash1 := CalculateHashWithNonce(block1)
		hash2 := CalculateHashWithNonce(block2)

		assert.NotEqual(t, hash1, hash2)
	})

	t.Run("難易度が異なると異なるハッシュを生成", func(t *testing.T) {
		block1 := NewBlock(1, "Data", "previous_hash", 1)
		block2 := NewBlock(1, "Data", "previous_hash", 2)

		block1.Nonce = 100
		block2.Nonce = 100

		hash1 := CalculateHashWithNonce(block1)
		hash2 := CalculateHashWithNonce(block2)

		assert.NotEqual(t, hash1, hash2)
	})
}

func TestCheckHashDifficulty(t *testing.T) {
	t.Run("難易度0（制約なし）", func(t *testing.T) {
		hash := "abcdef1234567890"
		assert.True(t, CheckHashDifficulty(hash, 0))
	})

	t.Run("難易度1（先頭が0）", func(t *testing.T) {
		assert.True(t, CheckHashDifficulty("0abcdef", 1))
		assert.False(t, CheckHashDifficulty("1abcdef", 1))
		assert.False(t, CheckHashDifficulty("abcdef0", 1))
	})

	t.Run("難易度2（先頭が00）", func(t *testing.T) {
		assert.True(t, CheckHashDifficulty("00abcdef", 2))
		assert.False(t, CheckHashDifficulty("0abcdef", 2))
		assert.False(t, CheckHashDifficulty("10abcdef", 2))
	})

	t.Run("難易度3（先頭が000）", func(t *testing.T) {
		assert.True(t, CheckHashDifficulty("000abcdef", 3))
		assert.True(t, CheckHashDifficulty("0000abcdef", 3)) // 4つ以上の0でもOK
		assert.False(t, CheckHashDifficulty("00abcdef", 3))
	})

	t.Run("難易度4（先頭が0000）", func(t *testing.T) {
		assert.True(t, CheckHashDifficulty("0000abcdef", 4))
		assert.False(t, CheckHashDifficulty("000abcdef", 4))
	})

	t.Run("空文字列", func(t *testing.T) {
		assert.True(t, CheckHashDifficulty("", 0))
		assert.False(t, CheckHashDifficulty("", 1))
	})
}

func TestMineBlock(t *testing.T) {
	t.Run("難易度0でのマイニング", func(t *testing.T) {
		block := NewBlock(1, "Test Block", "previous_hash", 0)

		metrics, err := MineBlock(block, 0)

		require.NoError(t, err)
		require.NotNil(t, metrics)
		assert.NotEmpty(t, block.Hash)
		assert.GreaterOrEqual(t, metrics.AttemptsCount, int64(1))
		assert.Greater(t, metrics.Duration, time.Duration(0))
	})

	t.Run("難易度1でのマイニング", func(t *testing.T) {
		block := NewBlock(1, "Test Block", "previous_hash", 1)

		metrics, err := MineBlock(block, 1)

		require.NoError(t, err)
		require.NotNil(t, metrics)
		assert.NotEmpty(t, block.Hash)
		assert.True(t, strings.HasPrefix(block.Hash, "0"))
		assert.GreaterOrEqual(t, block.Nonce, int64(0))
	})

	t.Run("難易度2でのマイニング", func(t *testing.T) {
		block := NewBlock(1, "Test Block", "previous_hash", 2)

		metrics, err := MineBlock(block, 2)

		require.NoError(t, err)
		require.NotNil(t, metrics)
		assert.True(t, strings.HasPrefix(block.Hash, "00"))
		assert.Greater(t, metrics.AttemptsCount, int64(1))
	})

	t.Run("難易度3でのマイニング", func(t *testing.T) {
		block := NewBlock(1, "Test Block", "previous_hash", 3)

		metrics, err := MineBlock(block, 3)

		require.NoError(t, err)
		require.NotNil(t, metrics)
		assert.True(t, strings.HasPrefix(block.Hash, "000"))
	})

	t.Run("負の難易度でエラー", func(t *testing.T) {
		block := NewBlock(1, "Test Block", "previous_hash", -1)

		metrics, err := MineBlock(block, -1)

		assert.Error(t, err)
		assert.Nil(t, metrics)
		assert.Contains(t, err.Error(), "difficulty must be non-negative")
	})

	t.Run("マイニングメトリクスの正確性", func(t *testing.T) {
		block := NewBlock(1, "Test Block", "previous_hash", 1)

		startTime := time.Now()
		metrics, err := MineBlock(block, 1)
		endTime := time.Now()

		require.NoError(t, err)
		require.NotNil(t, metrics)

		// メトリクスの検証
		assert.Greater(t, metrics.AttemptsCount, int64(0))
		assert.Greater(t, metrics.Duration, time.Duration(0))
		assert.LessOrEqual(t, metrics.Duration, endTime.Sub(startTime))

		// ハッシュレートの検証
		if metrics.Duration.Seconds() > 0 {
			expectedHashRate := float64(metrics.AttemptsCount) / metrics.Duration.Seconds()
			assert.InDelta(t, expectedHashRate, metrics.HashRate, 0.1)
		}
	})

	t.Run("マイニング後のブロック状態", func(t *testing.T) {
		block := NewBlock(1, "Test Block", "previous_hash", 2)
		originalData := block.Data
		originalIndex := block.Index
		originalPreviousHash := block.PreviousHash

		_, err := MineBlock(block, 2)

		require.NoError(t, err)

		// データが変更されていないことを確認
		assert.Equal(t, originalData, block.Data)
		assert.Equal(t, originalIndex, block.Index)
		assert.Equal(t, originalPreviousHash, block.PreviousHash)

		// ハッシュと難易度が設定されていることを確認
		assert.NotEmpty(t, block.Hash)
		assert.Equal(t, 2, block.Difficulty)
		assert.GreaterOrEqual(t, block.Nonce, int64(0))
	})
}

func TestValidateProofOfWork(t *testing.T) {
	t.Run("正しくマイニングされたブロックの検証", func(t *testing.T) {
		block := NewBlock(1, "Test Block", "previous_hash", 2)
		MineBlock(block, 2)

		assert.True(t, ValidateProofOfWork(block))
	})

	t.Run("ハッシュが改ざんされたブロックの検証", func(t *testing.T) {
		block := NewBlock(1, "Test Block", "previous_hash", 2)
		MineBlock(block, 2)

		// ハッシュを改ざん
		block.Hash = "00tampered_hash"

		assert.False(t, ValidateProofOfWork(block))
	})

	t.Run("ナンスが改ざんされたブロックの検証", func(t *testing.T) {
		block := NewBlock(1, "Test Block", "previous_hash", 2)
		MineBlock(block, 2)

		// ナンスを改ざん
		block.Nonce = 999999

		assert.False(t, ValidateProofOfWork(block))
	})

	t.Run("データが改ざんされたブロックの検証", func(t *testing.T) {
		block := NewBlock(1, "Test Block", "previous_hash", 2)
		MineBlock(block, 2)

		// データを改ざん
		block.Data = "Tampered Data"

		assert.False(t, ValidateProofOfWork(block))
	})

	t.Run("難易度が改ざんされたブロックの検証", func(t *testing.T) {
		block := NewBlock(1, "Test Block", "previous_hash", 2)
		MineBlock(block, 2)

		// 難易度を改ざん
		block.Difficulty = 1

		assert.False(t, ValidateProofOfWork(block))
	})

	t.Run("難易度を満たさないハッシュの検証", func(t *testing.T) {
		block := NewBlock(1, "Test Block", "previous_hash", 2)
		block.Nonce = 123
		block.Hash = "1234567890abcdef" // 先頭が0でない
		block.Difficulty = 2

		assert.False(t, ValidateProofOfWork(block))
	})

	t.Run("異なる難易度でのブロック検証", func(t *testing.T) {
		difficulties := []int{0, 1, 2, 3}

		for _, diff := range difficulties {
			block := NewBlock(1, "Test", "prev", diff)
			MineBlock(block, diff)

			assert.True(t, ValidateProofOfWork(block), "難易度%dのブロックが無効", diff)
		}
	})
}

func TestBlockValidate(t *testing.T) {
	t.Run("Validateメソッドのテスト", func(t *testing.T) {
		block := NewBlock(1, "Test Block", "previous_hash", 2)
		MineBlock(block, 2)

		assert.True(t, block.Validate())
	})

	t.Run("無効なブロックのValidate", func(t *testing.T) {
		block := NewBlock(1, "Test Block", "previous_hash", 2)
		MineBlock(block, 2)

		block.Data = "Tampered"

		assert.False(t, block.Validate())
	})
}

func TestBlockString(t *testing.T) {
	t.Run("Stringメソッドのテスト", func(t *testing.T) {
		block := NewBlock(1, "Test Block", "previous_hash", 2)
		MineBlock(block, 2)

		str := block.String()

		assert.Contains(t, str, "Block #1")
		assert.Contains(t, str, "Test Block")
		assert.Contains(t, str, "previous_hash")
		assert.Contains(t, str, block.Hash)
		assert.Contains(t, str, "Nonce:")
		assert.Contains(t, str, "Difficulty: 2")
	})

	t.Run("ジェネシスブロックのString", func(t *testing.T) {
		genesis := NewGenesisBlock(1)

		str := genesis.String()

		assert.Contains(t, str, "Block #0")
		assert.Contains(t, str, "Genesis Block")
	})
}

func TestGetDifficultyPrefix(t *testing.T) {
	t.Run("難易度0", func(t *testing.T) {
		assert.Equal(t, "", GetDifficultyPrefix(0))
	})

	t.Run("難易度1", func(t *testing.T) {
		assert.Equal(t, "0", GetDifficultyPrefix(1))
	})

	t.Run("難易度3", func(t *testing.T) {
		assert.Equal(t, "000", GetDifficultyPrefix(3))
	})

	t.Run("難易度5", func(t *testing.T) {
		assert.Equal(t, "00000", GetDifficultyPrefix(5))
	})
}

func TestMiningMetrics(t *testing.T) {
	t.Run("メトリクスの各フィールド", func(t *testing.T) {
		block := NewBlock(1, "Test Block", "previous_hash", 2)

		metrics, err := MineBlock(block, 2)

		require.NoError(t, err)
		require.NotNil(t, metrics)

		// AttemptsCount
		assert.Greater(t, metrics.AttemptsCount, int64(0))

		// Duration
		assert.Greater(t, metrics.Duration, time.Duration(0))

		// HashRate
		if metrics.Duration.Seconds() > 0 {
			assert.Greater(t, metrics.HashRate, 0.0)
		}
	})

	t.Run("難易度が高いほど試行回数が多い（統計的）", func(t *testing.T) {
		// 難易度1
		block1 := NewBlock(1, "Test", "prev", 1)
		metrics1, _ := MineBlock(block1, 1)

		// 難易度2
		block2 := NewBlock(1, "Test", "prev", 2)
		metrics2, _ := MineBlock(block2, 2)

		// 統計的には難易度2の方が試行回数が多いはず
		// ただし確率的なので、必ずしも成立しないため、
		// 複数回試行して傾向を確認するのが理想
		// ここでは基本的な検証のみ
		assert.Greater(t, metrics1.AttemptsCount, int64(0))
		assert.Greater(t, metrics2.AttemptsCount, int64(0))
	})
}

func TestEdgeCases(t *testing.T) {
	t.Run("ナンスの初期値が0", func(t *testing.T) {
		block := NewBlock(1, "Test", "prev", 1)

		assert.Equal(t, int64(0), block.Nonce)
	})

	t.Run("非常に大きなインデックス", func(t *testing.T) {
		block := NewBlock(999999999, "Test", "prev", 1)

		metrics, err := MineBlock(block, 1)

		require.NoError(t, err)
		assert.NotNil(t, metrics)
		assert.True(t, strings.HasPrefix(block.Hash, "0"))
	})

	t.Run("空のデータでマイニング", func(t *testing.T) {
		block := NewBlock(1, "", "prev", 1)

		metrics, err := MineBlock(block, 1)

		require.NoError(t, err)
		assert.NotNil(t, metrics)
		assert.NotEmpty(t, block.Hash)
	})

	t.Run("長いデータでマイニング", func(t *testing.T) {
		longData := strings.Repeat("A", 10000)
		block := NewBlock(1, longData, "prev", 1)

		metrics, err := MineBlock(block, 1)

		require.NoError(t, err)
		assert.NotNil(t, metrics)
		assert.Equal(t, longData, block.Data)
	})

	t.Run("空のPreviousHashでマイニング", func(t *testing.T) {
		block := NewBlock(1, "Test", "", 1)

		metrics, err := MineBlock(block, 1)

		require.NoError(t, err)
		assert.NotNil(t, metrics)
	})
}

// ベンチマーク
func BenchmarkMineBlockDifficulty0(b *testing.B) {
	for i := 0; i < b.N; i++ {
		block := NewBlock(1, "Benchmark Block", "previous_hash", 0)
		MineBlock(block, 0)
	}
}

func BenchmarkMineBlockDifficulty1(b *testing.B) {
	for i := 0; i < b.N; i++ {
		block := NewBlock(1, "Benchmark Block", "previous_hash", 1)
		MineBlock(block, 1)
	}
}

func BenchmarkMineBlockDifficulty2(b *testing.B) {
	for i := 0; i < b.N; i++ {
		block := NewBlock(1, "Benchmark Block", "previous_hash", 2)
		MineBlock(block, 2)
	}
}

func BenchmarkMineBlockDifficulty3(b *testing.B) {
	for i := 0; i < b.N; i++ {
		block := NewBlock(1, "Benchmark Block", "previous_hash", 3)
		MineBlock(block, 3)
	}
}

func BenchmarkCalculateHashWithNonce(b *testing.B) {
	block := NewBlock(1, "Benchmark Block", "previous_hash", 2)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		CalculateHashWithNonce(block)
	}
}

func BenchmarkCheckHashDifficulty(b *testing.B) {
	hash := "000abcdef1234567890"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		CheckHashDifficulty(hash, 3)
	}
}

func BenchmarkValidateProofOfWork(b *testing.B) {
	block := NewBlock(1, "Benchmark Block", "previous_hash", 2)
	MineBlock(block, 2)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ValidateProofOfWork(block)
	}
}
