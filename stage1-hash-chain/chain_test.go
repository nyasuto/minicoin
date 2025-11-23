package main

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewBlockchain(t *testing.T) {
	t.Run("新規ブロックチェーン生成", func(t *testing.T) {
		bc := NewBlockchain()

		require.NotNil(t, bc)
		assert.Equal(t, 1, len(bc.Blocks))
		assert.Equal(t, int64(0), bc.Blocks[0].Index)
		assert.Equal(t, "Genesis Block", bc.Blocks[0].Data)
	})

	t.Run("ジェネシスブロックが含まれる", func(t *testing.T) {
		bc := NewBlockchain()

		genesis := bc.Blocks[0]
		assert.Equal(t, int64(0), genesis.Index)
		assert.Equal(t, "", genesis.PreviousHash)
		assert.NotEmpty(t, genesis.Hash)
	})

	t.Run("生成直後のチェーンは有効", func(t *testing.T) {
		bc := NewBlockchain()

		assert.True(t, bc.IsValid())
	})
}

func TestAddBlock(t *testing.T) {
	t.Run("ブロックを正常に追加", func(t *testing.T) {
		bc := NewBlockchain()

		err := bc.AddBlock("Block 1")
		require.NoError(t, err)

		assert.Equal(t, 2, len(bc.Blocks))
		assert.Equal(t, "Block 1", bc.Blocks[1].Data)
		assert.Equal(t, int64(1), bc.Blocks[1].Index)
	})

	t.Run("複数のブロックを追加", func(t *testing.T) {
		bc := NewBlockchain()

		for i := 1; i <= 5; i++ {
			err := bc.AddBlock(fmt.Sprintf("Block %d", i))
			require.NoError(t, err)
		}

		assert.Equal(t, 6, len(bc.Blocks))
		assert.Equal(t, "Block 5", bc.Blocks[5].Data)
	})

	t.Run("追加されたブロックのPreviousHashが正しい", func(t *testing.T) {
		bc := NewBlockchain()

		_ = bc.AddBlock("Block 1")
		_ = bc.AddBlock("Block 2")

		block1 := bc.Blocks[1]
		block2 := bc.Blocks[2]

		assert.Equal(t, bc.Blocks[0].Hash, block1.PreviousHash)
		assert.Equal(t, block1.Hash, block2.PreviousHash)
	})

	t.Run("追加後もチェーンが有効", func(t *testing.T) {
		bc := NewBlockchain()

		for i := 1; i <= 10; i++ {
			_ = bc.AddBlock(fmt.Sprintf("Block %d", i))
			assert.True(t, bc.IsValid(), "Block %d追加後もチェーンは有効であるべき", i)
		}
	})
}

func TestGetLatestBlock(t *testing.T) {
	t.Run("ジェネシスブロックのみの場合", func(t *testing.T) {
		bc := NewBlockchain()

		latest := bc.GetLatestBlock()

		require.NotNil(t, latest)
		assert.Equal(t, int64(0), latest.Index)
		assert.Equal(t, "Genesis Block", latest.Data)
	})

	t.Run("ブロック追加後の最新ブロック", func(t *testing.T) {
		bc := NewBlockchain()

		_ = bc.AddBlock("Block 1")
		_ = bc.AddBlock("Block 2")
		_ = bc.AddBlock("Block 3")

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

func TestGetBlock(t *testing.T) {
	t.Run("有効なインデックスでブロック取得", func(t *testing.T) {
		bc := NewBlockchain()
		_ = bc.AddBlock("Block 1")
		_ = bc.AddBlock("Block 2")

		block, err := bc.GetBlock(1)

		require.NoError(t, err)
		require.NotNil(t, block)
		assert.Equal(t, int64(1), block.Index)
		assert.Equal(t, "Block 1", block.Data)
	})

	t.Run("範囲外のインデックス（負の数）", func(t *testing.T) {
		bc := NewBlockchain()

		block, err := bc.GetBlock(-1)

		assert.Error(t, err)
		assert.Nil(t, block)
		assert.Contains(t, err.Error(), "index out of range")
	})

	t.Run("範囲外のインデックス（大きすぎる）", func(t *testing.T) {
		bc := NewBlockchain()

		block, err := bc.GetBlock(100)

		assert.Error(t, err)
		assert.Nil(t, block)
	})

	t.Run("ジェネシスブロック取得", func(t *testing.T) {
		bc := NewBlockchain()

		block, err := bc.GetBlock(0)

		require.NoError(t, err)
		require.NotNil(t, block)
		assert.Equal(t, int64(0), block.Index)
		assert.Equal(t, "Genesis Block", block.Data)
	})
}

func TestGetChainLength(t *testing.T) {
	t.Run("ジェネシスブロックのみ", func(t *testing.T) {
		bc := NewBlockchain()

		length := bc.GetChainLength()

		assert.Equal(t, 1, length)
	})

	t.Run("ブロック追加後の長さ", func(t *testing.T) {
		bc := NewBlockchain()

		for i := 1; i <= 5; i++ {
			_ = bc.AddBlock(fmt.Sprintf("Block %d", i))
		}

		length := bc.GetChainLength()

		assert.Equal(t, 6, length)
	})

	t.Run("空のブロックチェーン", func(t *testing.T) {
		bc := &Blockchain{Blocks: []*Block{}}

		length := bc.GetChainLength()

		assert.Equal(t, 0, length)
	})
}

func TestIsValid(t *testing.T) {
	t.Run("正常なチェーンは有効", func(t *testing.T) {
		bc := NewBlockchain()

		for i := 1; i <= 5; i++ {
			_ = bc.AddBlock(fmt.Sprintf("Block %d", i))
		}

		assert.True(t, bc.IsValid())
	})

	t.Run("空のチェーンは無効", func(t *testing.T) {
		bc := &Blockchain{Blocks: []*Block{}}

		assert.False(t, bc.IsValid())
	})

	t.Run("ハッシュが改ざんされたブロックを検出", func(t *testing.T) {
		bc := NewBlockchain()
		_ = bc.AddBlock("Block 1")
		_ = bc.AddBlock("Block 2")

		// ブロック1のハッシュを改ざん
		bc.Blocks[1].Hash = "tampered_hash"

		assert.False(t, bc.IsValid())
	})

	t.Run("データが改ざんされたブロックを検出", func(t *testing.T) {
		bc := NewBlockchain()
		_ = bc.AddBlock("Block 1")
		_ = bc.AddBlock("Block 2")

		// ブロック1のデータを改ざん（ハッシュはそのまま）
		bc.Blocks[1].Data = "Tampered Data"

		assert.False(t, bc.IsValid())
	})

	t.Run("PreviousHashの不一致を検出", func(t *testing.T) {
		bc := NewBlockchain()
		_ = bc.AddBlock("Block 1")
		_ = bc.AddBlock("Block 2")

		// ブロック2のPreviousHashを改ざん
		bc.Blocks[2].PreviousHash = "wrong_hash"

		assert.False(t, bc.IsValid())
	})

	t.Run("インデックスの不連続を検出", func(t *testing.T) {
		bc := NewBlockchain()
		_ = bc.AddBlock("Block 1")
		_ = bc.AddBlock("Block 2")

		// ブロック2のインデックスを改ざん
		bc.Blocks[2].Index = 999

		assert.False(t, bc.IsValid())
	})

	t.Run("タイムスタンプの逆転を検出", func(t *testing.T) {
		bc := NewBlockchain()
		_ = bc.AddBlock("Block 1")
		_ = bc.AddBlock("Block 2")

		// ブロック2のタイムスタンプを過去に変更
		bc.Blocks[2].Timestamp = bc.Blocks[1].Timestamp - 1000

		assert.False(t, bc.IsValid())
	})

	t.Run("タイムスタンプが同じでも有効", func(t *testing.T) {
		bc := NewBlockchain()
		_ = bc.AddBlock("Block 1")

		// ブロック1と同じタイムスタンプでブロック2を作成
		timestamp := bc.Blocks[1].Timestamp
		_ = bc.AddBlock("Block 2")
		bc.Blocks[2].Timestamp = timestamp
		bc.Blocks[2].Hash = bc.Blocks[2].CalculateHash()

		assert.True(t, bc.IsValid())
	})

	t.Run("ジェネシスブロックのインデックスが0でない場合", func(t *testing.T) {
		bc := NewBlockchain()
		bc.Blocks[0].Index = 1

		assert.False(t, bc.IsValid())
	})

	t.Run("ジェネシスブロックのPreviousHashが空でない場合", func(t *testing.T) {
		bc := NewBlockchain()
		bc.Blocks[0].PreviousHash = "not_empty"

		assert.False(t, bc.IsValid())
	})
}

func TestPrintChain(t *testing.T) {
	t.Run("PrintChainが正常に実行される", func(t *testing.T) {
		bc := NewBlockchain()
		_ = bc.AddBlock("Block 1")
		_ = bc.AddBlock("Block 2")

		// パニックしないことを確認
		assert.NotPanics(t, func() {
			bc.PrintChain()
		})
	})
}

func TestConcurrentAccess(t *testing.T) {
	t.Run("並行ブロック追加", func(t *testing.T) {
		bc := NewBlockchain()
		var wg sync.WaitGroup

		// 10個のゴルーチンから同時にブロックを追加
		numGoroutines := 10
		blocksPerGoroutine := 10

		wg.Add(numGoroutines)
		for i := 0; i < numGoroutines; i++ {
			go func(id int) {
				defer wg.Done()
				for j := 0; j < blocksPerGoroutine; j++ {
					_ = bc.AddBlock(fmt.Sprintf("Goroutine %d - Block %d", id, j))
				}
			}(i)
		}

		wg.Wait()

		// 期待される総ブロック数（ジェネシス + 追加分）
		expectedLength := 1 + (numGoroutines * blocksPerGoroutine)
		assert.Equal(t, expectedLength, bc.GetChainLength())
	})

	t.Run("並行読み取りと書き込み", func(t *testing.T) {
		bc := NewBlockchain()
		var wg sync.WaitGroup

		// 書き込みゴルーチン
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < 50; i++ {
				_ = bc.AddBlock(fmt.Sprintf("Block %d", i))
				time.Sleep(1 * time.Millisecond)
			}
		}()

		// 読み取りゴルーチン（複数）
		for i := 0; i < 5; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for j := 0; j < 100; j++ {
					bc.GetLatestBlock()
					bc.GetChainLength()
					bc.IsValid()
					time.Sleep(1 * time.Millisecond)
				}
			}()
		}

		wg.Wait()

		// 最終的にチェーンが有効であることを確認
		assert.True(t, bc.IsValid())
		assert.Equal(t, 51, bc.GetChainLength()) // ジェネシス + 50ブロック
	})
}

// ベンチマーク
func BenchmarkAddBlock(b *testing.B) {
	bc := NewBlockchain()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = bc.AddBlock(fmt.Sprintf("Block %d", i))
	}
}

func BenchmarkIsValid(b *testing.B) {
	bc := NewBlockchain()
	for i := 0; i < 100; i++ {
		_ = bc.AddBlock(fmt.Sprintf("Block %d", i))
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bc.IsValid()
	}
}

func BenchmarkGetLatestBlock(b *testing.B) {
	bc := NewBlockchain()
	for i := 0; i < 100; i++ {
		_ = bc.AddBlock(fmt.Sprintf("Block %d", i))
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bc.GetLatestBlock()
	}
}
