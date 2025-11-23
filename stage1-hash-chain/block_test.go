package main

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewBlock(t *testing.T) {
	t.Run("正常なブロック生成", func(t *testing.T) {
		index := int64(1)
		data := "Test Block Data"
		previousHash := "abc123"

		block := NewBlock(index, data, previousHash)

		require.NotNil(t, block)
		assert.Equal(t, index, block.Index)
		assert.Equal(t, data, block.Data)
		assert.Equal(t, previousHash, block.PreviousHash)
		assert.NotEmpty(t, block.Hash)
		assert.NotZero(t, block.Timestamp)
	})

	t.Run("タイムスタンプが現在時刻に近い", func(t *testing.T) {
		before := time.Now().Unix()
		block := NewBlock(1, "data", "hash")
		after := time.Now().Unix()

		assert.GreaterOrEqual(t, block.Timestamp, before)
		assert.LessOrEqual(t, block.Timestamp, after)
	})

	t.Run("ハッシュが自動計算される", func(t *testing.T) {
		block := NewBlock(1, "data", "previousHash")
		expectedHash := block.CalculateHash()

		assert.Equal(t, expectedHash, block.Hash)
	})
}

func TestCalculateHash(t *testing.T) {
	t.Run("同じ内容のブロックは同じハッシュを生成", func(t *testing.T) {
		block1 := &Block{
			Index:        1,
			Timestamp:    1234567890,
			Data:         "Test Data",
			PreviousHash: "abc123",
		}
		block2 := &Block{
			Index:        1,
			Timestamp:    1234567890,
			Data:         "Test Data",
			PreviousHash: "abc123",
		}

		hash1 := block1.CalculateHash()
		hash2 := block2.CalculateHash()

		assert.Equal(t, hash1, hash2)
	})

	t.Run("異なる内容のブロックは異なるハッシュを生成", func(t *testing.T) {
		block1 := &Block{
			Index:        1,
			Timestamp:    1234567890,
			Data:         "Data 1",
			PreviousHash: "abc123",
		}
		block2 := &Block{
			Index:        1,
			Timestamp:    1234567890,
			Data:         "Data 2",
			PreviousHash: "abc123",
		}

		hash1 := block1.CalculateHash()
		hash2 := block2.CalculateHash()

		assert.NotEqual(t, hash1, hash2)
	})

	t.Run("ハッシュがSHA-256形式（64文字の16進数）", func(t *testing.T) {
		block := NewBlock(1, "data", "hash")

		assert.Len(t, block.Hash, 64)
		assert.Regexp(t, "^[a-f0-9]{64}$", block.Hash)
	})

	t.Run("Index変更でハッシュが変わる", func(t *testing.T) {
		block := &Block{
			Index:        1,
			Timestamp:    1234567890,
			Data:         "Data",
			PreviousHash: "abc123",
		}
		hash1 := block.CalculateHash()

		block.Index = 2
		hash2 := block.CalculateHash()

		assert.NotEqual(t, hash1, hash2)
	})

	t.Run("Timestamp変更でハッシュが変わる", func(t *testing.T) {
		block := &Block{
			Index:        1,
			Timestamp:    1234567890,
			Data:         "Data",
			PreviousHash: "abc123",
		}
		hash1 := block.CalculateHash()

		block.Timestamp = 9876543210
		hash2 := block.CalculateHash()

		assert.NotEqual(t, hash1, hash2)
	})

	t.Run("PreviousHash変更でハッシュが変わる", func(t *testing.T) {
		block := &Block{
			Index:        1,
			Timestamp:    1234567890,
			Data:         "Data",
			PreviousHash: "abc123",
		}
		hash1 := block.CalculateHash()

		block.PreviousHash = "xyz789"
		hash2 := block.CalculateHash()

		assert.NotEqual(t, hash1, hash2)
	})
}

func TestValidate(t *testing.T) {
	t.Run("正常なブロックはバリデーション成功", func(t *testing.T) {
		block := NewBlock(1, "Test Data", "abc123")

		assert.True(t, block.Validate())
	})

	t.Run("ハッシュが改ざんされたブロックはバリデーション失敗", func(t *testing.T) {
		block := NewBlock(1, "Test Data", "abc123")
		block.Hash = "tampered_hash"

		assert.False(t, block.Validate())
	})

	t.Run("データが改ざんされたブロックはバリデーション失敗", func(t *testing.T) {
		block := NewBlock(1, "Original Data", "abc123")
		// ハッシュはそのままでデータだけ変更
		block.Data = "Tampered Data"

		assert.False(t, block.Validate())
	})

	t.Run("Indexが改ざんされたブロックはバリデーション失敗", func(t *testing.T) {
		block := NewBlock(1, "Data", "abc123")
		block.Index = 999

		assert.False(t, block.Validate())
	})
}

func TestString(t *testing.T) {
	t.Run("String()が必要な情報を含む", func(t *testing.T) {
		block := NewBlock(1, "Test Data", "abc123")
		str := block.String()

		assert.Contains(t, str, "Block #1")
		assert.Contains(t, str, "Test Data")
		assert.Contains(t, str, "abc123")
		assert.Contains(t, str, block.Hash)
	})

	t.Run("String()が空文字列を返さない", func(t *testing.T) {
		block := NewBlock(0, "", "")
		str := block.String()

		assert.NotEmpty(t, str)
	})
}

func TestNewGenesisBlock(t *testing.T) {
	t.Run("ジェネシスブロック生成", func(t *testing.T) {
		genesis := NewGenesisBlock()

		require.NotNil(t, genesis)
		assert.Equal(t, int64(0), genesis.Index)
		assert.Equal(t, "Genesis Block", genesis.Data)
		assert.Equal(t, "", genesis.PreviousHash)
		assert.NotEmpty(t, genesis.Hash)
	})

	t.Run("ジェネシスブロックはバリデーション成功", func(t *testing.T) {
		genesis := NewGenesisBlock()

		assert.True(t, genesis.Validate())
	})

	t.Run("複数回生成しても同じIndex", func(t *testing.T) {
		genesis1 := NewGenesisBlock()
		genesis2 := NewGenesisBlock()

		assert.Equal(t, genesis1.Index, genesis2.Index)
		assert.Equal(t, genesis1.Data, genesis2.Data)
		assert.Equal(t, genesis1.PreviousHash, genesis2.PreviousHash)
	})

	t.Run("タイムスタンプのみ異なる", func(t *testing.T) {
		genesis1 := NewGenesisBlock()
		time.Sleep(1100 * time.Millisecond) // 確実に秒が変わるまで待機
		genesis2 := NewGenesisBlock()

		// タイムスタンプが異なるため、ハッシュも異なる
		assert.NotEqual(t, genesis1.Timestamp, genesis2.Timestamp)
		assert.NotEqual(t, genesis1.Hash, genesis2.Hash)
	})
}

// ベンチマーク
func BenchmarkNewBlock(b *testing.B) {
	for i := 0; i < b.N; i++ {
		NewBlock(int64(i), "Benchmark Data", "previousHash")
	}
}

func BenchmarkCalculateHash(b *testing.B) {
	block := &Block{
		Index:        1,
		Timestamp:    time.Now().Unix(),
		Data:         "Benchmark Data",
		PreviousHash: "abc123",
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		block.CalculateHash()
	}
}

func BenchmarkValidate(b *testing.B) {
	block := NewBlock(1, "Benchmark Data", "abc123")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		block.Validate()
	}
}
