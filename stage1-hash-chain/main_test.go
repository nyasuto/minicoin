package main

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExportBlockchain(t *testing.T) {
	t.Run("正常なエクスポート", func(t *testing.T) {
		bc := NewBlockchain()
		_ = bc.AddBlock("Block 1")
		_ = bc.AddBlock("Block 2")

		tempFile := "test_export.json"
		defer func() { _ = os.Remove(tempFile) }()

		err := exportBlockchain(bc, tempFile)
		require.NoError(t, err)

		// ファイルが存在することを確認
		_, err = os.Stat(tempFile)
		assert.NoError(t, err)
	})

	t.Run("複数ブロックのエクスポート", func(t *testing.T) {
		bc := NewBlockchain()
		for i := 1; i <= 10; i++ {
			bc.AddBlock(fmt.Sprintf("Block %d", i))
		}

		tempFile := "test_export_large.json"
		defer func() { _ = os.Remove(tempFile) }()

		err := exportBlockchain(bc, tempFile)
		require.NoError(t, err)

		// ファイルサイズが0より大きいことを確認
		info, err := os.Stat(tempFile)
		require.NoError(t, err)
		assert.Greater(t, info.Size(), int64(0))
	})
}

func TestImportBlockchain(t *testing.T) {
	t.Run("正常なインポート", func(t *testing.T) {
		// まずエクスポート
		bc := NewBlockchain()
		bc.AddBlock("Block 1")
		bc.AddBlock("Block 2")

		tempFile := "test_import.json"
		defer func() { _ = os.Remove(tempFile) }()

		err := exportBlockchain(bc, tempFile)
		require.NoError(t, err)

		// インポート
		importedBC, err := importBlockchain(tempFile)
		require.NoError(t, err)
		require.NotNil(t, importedBC)

		// 同じ内容であることを確認
		assert.Equal(t, bc.GetChainLength(), importedBC.GetChainLength())
		assert.Equal(t, bc.GetLatestBlock().Hash, importedBC.GetLatestBlock().Hash)
	})

	t.Run("存在しないファイルのインポート", func(t *testing.T) {
		_, err := importBlockchain("non_existent.json")
		assert.Error(t, err)
	})

	t.Run("無効なJSONファイルのインポート", func(t *testing.T) {
		tempFile := "test_invalid.json"
		defer func() { _ = os.Remove(tempFile) }()

		// 無効なJSONを書き込み
		err := os.WriteFile(tempFile, []byte("{invalid json}"), 0600)
		require.NoError(t, err)

		_, err = importBlockchain(tempFile)
		assert.Error(t, err)
	})

	t.Run("無効なチェーンのインポート", func(t *testing.T) {
		bc := NewBlockchain()
		bc.AddBlock("Block 1")
		bc.AddBlock("Block 2")

		// ブロックを改ざん
		bc.Blocks[1].Hash = "tampered_hash"

		tempFile := "test_invalid_chain.json"
		defer func() { _ = os.Remove(tempFile) }()

		err := exportBlockchain(bc, tempFile)
		require.NoError(t, err)

		// インポート時にエラーが発生することを確認
		_, err = importBlockchain(tempFile)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "無効")
	})
}

func TestExportImportRoundTrip(t *testing.T) {
	t.Run("エクスポート→インポートの往復", func(t *testing.T) {
		// 元のチェーンを作成
		originalBC := NewBlockchain()
		for i := 1; i <= 5; i++ {
			_ = originalBC.AddBlock(fmt.Sprintf("Test Block %d", i))
		}

		tempFile := "test_roundtrip.json"
		defer func() { _ = os.Remove(tempFile) }()

		// エクスポート
		err := exportBlockchain(originalBC, tempFile)
		require.NoError(t, err)

		// インポート
		importedBC, err := importBlockchain(tempFile)
		require.NoError(t, err)

		// 全てのブロックが一致することを確認
		assert.Equal(t, originalBC.GetChainLength(), importedBC.GetChainLength())

		for i := 0; i < originalBC.GetChainLength(); i++ {
			originalBlock, _ := originalBC.GetBlock(int64(i))
			importedBlock, _ := importedBC.GetBlock(int64(i))

			assert.Equal(t, originalBlock.Index, importedBlock.Index)
			assert.Equal(t, originalBlock.Timestamp, importedBlock.Timestamp)
			assert.Equal(t, originalBlock.Data, importedBlock.Data)
			assert.Equal(t, originalBlock.PreviousHash, importedBlock.PreviousHash)
			assert.Equal(t, originalBlock.Hash, importedBlock.Hash)
		}

		// インポートされたチェーンが有効であることを確認
		assert.True(t, importedBC.IsValid())
	})
}

func TestDisplayFunctions(t *testing.T) {
	t.Run("displayBlockDetails - パニックしない", func(t *testing.T) {
		block := NewBlock(1, "Test Block", "abc123")

		assert.NotPanics(t, func() {
			displayBlockDetails(block)
		})
	})

	t.Run("displayBlockDetails - ジェネシスブロック", func(t *testing.T) {
		block := NewGenesisBlock()

		assert.NotPanics(t, func() {
			displayBlockDetails(block)
		})
	})
}

func TestPrintFunctions(t *testing.T) {
	bc := NewBlockchain()
	bc.AddBlock("Block 1")
	bc.AddBlock("Block 2")

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

	t.Run("printValidationResult - パニックしない", func(t *testing.T) {
		assert.NotPanics(t, func() {
			printValidationResult(bc)
		})
	})

	t.Run("printStats - パニックしない", func(t *testing.T) {
		assert.NotPanics(t, func() {
			printStats(bc)
		})
	})

	t.Run("displayChain - パニックしない", func(t *testing.T) {
		assert.NotPanics(t, func() {
			displayChain(bc)
		})
	})
}

func TestExportImportEdgeCases(t *testing.T) {
	t.Run("空のファイル名でエクスポート", func(t *testing.T) {
		bc := NewBlockchain()

		err := exportBlockchain(bc, "")
		assert.Error(t, err)
	})

	t.Run("ジェネシスブロックのみのチェーンをエクスポート/インポート", func(t *testing.T) {
		bc := NewBlockchain()

		tempFile := "test_genesis_only.json"
		defer func() { _ = os.Remove(tempFile) }()

		err := exportBlockchain(bc, tempFile)
		require.NoError(t, err)

		importedBC, err := importBlockchain(tempFile)
		require.NoError(t, err)

		assert.Equal(t, 1, importedBC.GetChainLength())
		assert.Equal(t, int64(0), importedBC.GetLatestBlock().Index)
	})

	t.Run("大量のブロックをエクスポート/インポート", func(t *testing.T) {
		bc := NewBlockchain()
		for i := 1; i <= 100; i++ {
			bc.AddBlock(fmt.Sprintf("Block %d", i))
		}

		tempFile := "test_large_chain.json"
		defer func() { _ = os.Remove(tempFile) }()

		err := exportBlockchain(bc, tempFile)
		require.NoError(t, err)

		importedBC, err := importBlockchain(tempFile)
		require.NoError(t, err)

		assert.Equal(t, 101, importedBC.GetChainLength()) // ジェネシス + 100
		assert.True(t, importedBC.IsValid())
	})
}

// ベンチマーク
func BenchmarkExportBlockchain(b *testing.B) {
	bc := NewBlockchain()
	for i := 0; i < 100; i++ {
		bc.AddBlock(fmt.Sprintf("Block %d", i))
	}

	tempFile := "bench_export.json"
	defer func() { _ = os.Remove(tempFile) }()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = exportBlockchain(bc, tempFile)
	}
}

func BenchmarkImportBlockchain(b *testing.B) {
	bc := NewBlockchain()
	for i := 0; i < 100; i++ {
		bc.AddBlock(fmt.Sprintf("Block %d", i))
	}

	tempFile := "bench_import.json"
	_ = exportBlockchain(bc, tempFile)
	defer func() { _ = os.Remove(tempFile) }()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = importBlockchain(tempFile)
	}
}
