package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewBlockchain(t *testing.T) {
	t.Run("新規ブロックチェーン生成（難易度2）", func(t *testing.T) {
		bc := NewBlockchain(2)

		require.NotNil(t, bc)
		assert.Equal(t, 1, len(bc.Blocks))
		assert.Equal(t, 2, bc.Difficulty)
		assert.Equal(t, int64(0), bc.Blocks[0].Index)
		assert.Equal(t, "Genesis Block", bc.Blocks[0].Data)
	})

	t.Run("異なる難易度でのブロックチェーン生成", func(t *testing.T) {
		difficulties := []int{0, 1, 2, 3}

		for _, diff := range difficulties {
			bc := NewBlockchain(diff)

			assert.Equal(t, diff, bc.Difficulty)
			assert.Equal(t, diff, bc.Blocks[0].Difficulty)
			assert.True(t, ValidateProofOfWork(bc.Blocks[0]))
		}
	})

	t.Run("生成直後のチェーンは有効", func(t *testing.T) {
		bc := NewBlockchain(2)

		assert.True(t, bc.IsValid())
	})
}

func TestAddBlock(t *testing.T) {
	t.Run("ブロックを正常に追加（難易度1）", func(t *testing.T) {
		bc := NewBlockchain(1)

		metrics, err := bc.AddBlock("Block 1")

		require.NoError(t, err)
		require.NotNil(t, metrics)
		assert.Equal(t, 2, len(bc.Blocks))
		assert.Equal(t, "Block 1", bc.Blocks[1].Data)
		assert.Equal(t, int64(1), bc.Blocks[1].Index)
		assert.Greater(t, metrics.AttemptsCount, int64(0))
	})

	t.Run("複数のブロックを追加", func(t *testing.T) {
		bc := NewBlockchain(1)

		for i := 1; i <= 3; i++ {
			metrics, err := bc.AddBlock("Block " + string(rune(i+'0')))

			require.NoError(t, err)
			require.NotNil(t, metrics)
		}

		assert.Equal(t, 4, len(bc.Blocks)) // ジェネシス + 3
	})

	t.Run("追加されたブロックのPreviousHashが正しい", func(t *testing.T) {
		bc := NewBlockchain(1)

		bc.AddBlock("Block 1")
		bc.AddBlock("Block 2")

		block1 := bc.Blocks[1]
		block2 := bc.Blocks[2]

		assert.Equal(t, bc.Blocks[0].Hash, block1.PreviousHash)
		assert.Equal(t, block1.Hash, block2.PreviousHash)
	})

	t.Run("追加後もチェーンが有効", func(t *testing.T) {
		bc := NewBlockchain(1)

		for i := 1; i <= 5; i++ {
			bc.AddBlock("Block " + string(rune(i+'0')))
			assert.True(t, bc.IsValid(), "Block %d追加後もチェーンは有効であるべき", i)
		}
	})

	t.Run("マイニングメトリクスが正しく返される", func(t *testing.T) {
		bc := NewBlockchain(2)

		metrics, err := bc.AddBlock("Test Block")

		require.NoError(t, err)
		require.NotNil(t, metrics)
		assert.Greater(t, metrics.AttemptsCount, int64(0))
		assert.Greater(t, metrics.Duration.Nanoseconds(), int64(0))
		assert.GreaterOrEqual(t, metrics.HashRate, 0.0)
	})
}

func TestGetLatestBlock(t *testing.T) {
	t.Run("ジェネシスブロックのみの場合", func(t *testing.T) {
		bc := NewBlockchain(1)

		latest := bc.GetLatestBlock()

		require.NotNil(t, latest)
		assert.Equal(t, int64(0), latest.Index)
		assert.Equal(t, "Genesis Block", latest.Data)
	})

	t.Run("ブロック追加後の最新ブロック", func(t *testing.T) {
		bc := NewBlockchain(1)

		bc.AddBlock("Block 1")
		bc.AddBlock("Block 2")
		bc.AddBlock("Block 3")

		latest := bc.GetLatestBlock()

		require.NotNil(t, latest)
		assert.Equal(t, int64(3), latest.Index)
		assert.Equal(t, "Block 3", latest.Data)
	})

	t.Run("空のブロックチェーン", func(t *testing.T) {
		bc := &Blockchain{Blocks: []*Block{}}

		latest := bc.GetLatestBlock()

		assert.Nil(t, latest)
	})
}

func TestGetChainLength(t *testing.T) {
	t.Run("ジェネシスブロックのみ", func(t *testing.T) {
		bc := NewBlockchain(1)

		length := bc.GetChainLength()

		assert.Equal(t, 1, length)
	})

	t.Run("ブロック追加後の長さ", func(t *testing.T) {
		bc := NewBlockchain(1)

		for i := 1; i <= 5; i++ {
			bc.AddBlock("Block " + string(rune(i+'0')))
		}

		length := bc.GetChainLength()

		assert.Equal(t, 6, length) // ジェネシス + 5
	})

	t.Run("空のブロックチェーン", func(t *testing.T) {
		bc := &Blockchain{Blocks: []*Block{}}

		length := bc.GetChainLength()

		assert.Equal(t, 0, length)
	})
}

func TestIsValid(t *testing.T) {
	t.Run("正常なチェーンは有効", func(t *testing.T) {
		bc := NewBlockchain(1)

		for i := 1; i <= 5; i++ {
			bc.AddBlock("Block " + string(rune(i+'0')))
		}

		assert.True(t, bc.IsValid())
	})

	t.Run("空のチェーンは無効", func(t *testing.T) {
		bc := &Blockchain{Blocks: []*Block{}}

		assert.False(t, bc.IsValid())
	})

	t.Run("ハッシュが改ざんされたブロックを検出", func(t *testing.T) {
		bc := NewBlockchain(1)
		bc.AddBlock("Block 1")
		bc.AddBlock("Block 2")

		// ブロック1のハッシュを改ざん
		bc.Blocks[1].Hash = "tampered_hash"

		assert.False(t, bc.IsValid())
	})

	t.Run("データが改ざんされたブロックを検出", func(t *testing.T) {
		bc := NewBlockchain(1)
		bc.AddBlock("Block 1")
		bc.AddBlock("Block 2")

		// ブロック1のデータを改ざん
		bc.Blocks[1].Data = "Tampered Data"

		assert.False(t, bc.IsValid())
	})

	t.Run("PreviousHashの不一致を検出", func(t *testing.T) {
		bc := NewBlockchain(1)
		bc.AddBlock("Block 1")
		bc.AddBlock("Block 2")

		// ブロック2のPreviousHashを改ざん
		bc.Blocks[2].PreviousHash = "wrong_hash"

		assert.False(t, bc.IsValid())
	})

	t.Run("インデックスの不連続を検出", func(t *testing.T) {
		bc := NewBlockchain(1)
		bc.AddBlock("Block 1")
		bc.AddBlock("Block 2")

		// ブロック2のインデックスを改ざん
		bc.Blocks[2].Index = 999

		assert.False(t, bc.IsValid())
	})

	t.Run("タイムスタンプの逆転を検出", func(t *testing.T) {
		bc := NewBlockchain(1)
		bc.AddBlock("Block 1")
		bc.AddBlock("Block 2")

		// ブロック2のタイムスタンプを過去に変更
		bc.Blocks[2].Timestamp = bc.Blocks[1].Timestamp - 1000

		assert.False(t, bc.IsValid())
	})

	t.Run("ジェネシスブロックのインデックスが0でない場合", func(t *testing.T) {
		bc := NewBlockchain(1)
		bc.Blocks[0].Index = 1

		assert.False(t, bc.IsValid())
	})

	t.Run("ジェネシスブロックのPreviousHashが空でない場合", func(t *testing.T) {
		bc := NewBlockchain(1)
		bc.Blocks[0].PreviousHash = "not_empty"

		assert.False(t, bc.IsValid())
	})

	t.Run("PoW検証が失敗した場合", func(t *testing.T) {
		bc := NewBlockchain(2)
		bc.AddBlock("Block 1")

		// ブロック1のNonceを改ざん
		bc.Blocks[1].Nonce = 999999

		assert.False(t, bc.IsValid())
	})
}

func TestDisplayFunctions(t *testing.T) {
	bc := NewBlockchain(1)
	bc.AddBlock("Block 1")

	t.Run("printHeader - パニックしない", func(t *testing.T) {
		assert.NotPanics(t, func() {
			printHeader()
		})
	})

	t.Run("printMenu - パニックしない", func(t *testing.T) {
		assert.NotPanics(t, func() {
			printMenu()
		})
	})

	t.Run("displayChain - パニックしない", func(t *testing.T) {
		assert.NotPanics(t, func() {
			displayChain(bc)
		})
	})

	t.Run("validateChain - パニックしない", func(t *testing.T) {
		assert.NotPanics(t, func() {
			validateChain(bc)
		})
	})
}

func TestDifficultySettings(t *testing.T) {
	t.Run("異なる難易度でブロックチェーンを初期化", func(t *testing.T) {
		bc0 := NewBlockchain(0)
		bc1 := NewBlockchain(1)
		bc2 := NewBlockchain(2)

		assert.Equal(t, 0, bc0.Difficulty)
		assert.Equal(t, 1, bc1.Difficulty)
		assert.Equal(t, 2, bc2.Difficulty)
	})

	t.Run("追加されたブロックがブロックチェーンの難易度を継承", func(t *testing.T) {
		bc := NewBlockchain(2)

		bc.AddBlock("Block 1")

		assert.Equal(t, 2, bc.Blocks[1].Difficulty)
	})

	t.Run("難易度変更後に追加されたブロックが新しい難易度を使用", func(t *testing.T) {
		bc := NewBlockchain(1)
		bc.AddBlock("Block 1")

		// 難易度を変更
		bc.Difficulty = 2
		bc.AddBlock("Block 2")

		assert.Equal(t, 1, bc.Blocks[1].Difficulty)
		assert.Equal(t, 2, bc.Blocks[2].Difficulty)
	})
}

// ベンチマーク
func BenchmarkAddBlock(b *testing.B) {
	bc := NewBlockchain(1)
	// ベンチマーク中に難易度が上がらないように調整機能を無効化
	bc.TargetBlockTime = 0

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bc.AddBlock("Benchmark Block")
		// メモリ爆発を防ぐために定期的にチェーンをリセット
		if len(bc.Blocks) > 1000 {
			// 最新のブロックを保持して新しいチェーンのベースにする
			lastBlock := bc.Blocks[len(bc.Blocks)-1]
			bc.Blocks = []*Block{lastBlock}
		}
	}
}

func BenchmarkIsValid(b *testing.B) {
	bc := NewBlockchain(1)
	for i := 0; i < 10; i++ {
		bc.AddBlock("Block")
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bc.IsValid()
	}
}
