package main

import (
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewUTXOSet(t *testing.T) {
	t.Run("ジェネシスブロックからUTXOセット生成", func(t *testing.T) {
		wallet, err := NewWallet()
		require.NoError(t, err)

		bc := NewBlockchain(1, wallet.GetAddress())
		utxoSet := NewUTXOSet(bc)

		require.NotNil(t, utxoSet)
		assert.NotNil(t, utxoSet.UTXOs)

		// ジェネシスブロックのコインベーストランザクションがUTXOに含まれる
		balance := utxoSet.GetBalance(wallet.GetAddress())
		assert.Equal(t, 50, balance) // コインベース報酬
	})
}

func TestGetBalance(t *testing.T) {
	t.Run("単一UTXOの残高計算", func(t *testing.T) {
		wallet, err := NewWallet()
		require.NoError(t, err)

		bc := NewBlockchain(1, wallet.GetAddress())
		utxoSet := NewUTXOSet(bc)

		balance := utxoSet.GetBalance(wallet.GetAddress())
		assert.Equal(t, 50, balance)
	})

	t.Run("存在しないアドレスの残高は0", func(t *testing.T) {
		wallet, err := NewWallet()
		require.NoError(t, err)

		bc := NewBlockchain(1, wallet.GetAddress())
		utxoSet := NewUTXOSet(bc)

		balance := utxoSet.GetBalance("nonexistent")
		assert.Equal(t, 0, balance)
	})
}

func TestFindSpendableOutputs(t *testing.T) {
	t.Run("十分な残高がある場合", func(t *testing.T) {
		wallet, err := NewWallet()
		require.NoError(t, err)

		bc := NewBlockchain(1, wallet.GetAddress())
		utxoSet := NewUTXOSet(bc)

		accumulated, outputs := utxoSet.FindSpendableOutputs(wallet.GetAddress(), 30)

		assert.Equal(t, 50, accumulated) // コインベース報酬全額
		assert.NotEmpty(t, outputs)
	})

	t.Run("残高不足の場合", func(t *testing.T) {
		wallet, err := NewWallet()
		require.NoError(t, err)

		bc := NewBlockchain(1, wallet.GetAddress())
		utxoSet := NewUTXOSet(bc)

		accumulated, outputs := utxoSet.FindSpendableOutputs(wallet.GetAddress(), 100)

		assert.Equal(t, 50, accumulated) // 不足
		assert.NotEmpty(t, outputs)
	})

	t.Run("存在しないアドレス", func(t *testing.T) {
		wallet, err := NewWallet()
		require.NoError(t, err)

		bc := NewBlockchain(1, wallet.GetAddress())
		utxoSet := NewUTXOSet(bc)

		accumulated, outputs := utxoSet.FindSpendableOutputs("nonexistent", 10)

		assert.Equal(t, 0, accumulated)
		assert.Empty(t, outputs)
	})
}

func TestFindUTXO(t *testing.T) {
	t.Run("UTXOの取得", func(t *testing.T) {
		wallet, err := NewWallet()
		require.NoError(t, err)

		bc := NewBlockchain(1, wallet.GetAddress())
		utxoSet := NewUTXOSet(bc)

		utxos := utxoSet.FindUTXO(wallet.GetAddress())

		assert.Len(t, utxos, 1) // ジェネシスブロックのコインベース
		assert.Equal(t, 50, utxos[0].Output.Value)
	})
}

func TestUpdate(t *testing.T) {
	t.Run("ブロック追加時のUTXO更新", func(t *testing.T) {
		wallet1, err := NewWallet()
		require.NoError(t, err)
		wallet2, err := NewWallet()
		require.NoError(t, err)

		bc := NewBlockchain(1, wallet1.GetAddress())
		utxoSet := NewUTXOSet(bc)

		// 初期残高確認
		balance1 := utxoSet.GetBalance(wallet1.GetAddress())
		assert.Equal(t, 50, balance1)

		// 新しいブロック（コインベーストランザクション）を作成
		coinbaseTx := NewCoinbaseTx(wallet2.GetAddress(), "Block 1")
		block, _, err := bc.MineBlock([]*Transaction{coinbaseTx})
		require.NoError(t, err)

		// UTXOセット更新
		err = utxoSet.Update(block)
		require.NoError(t, err)

		// wallet2の残高が増加
		balance2 := utxoSet.GetBalance(wallet2.GetAddress())
		assert.Equal(t, 50, balance2)

		// wallet1の残高は変わらない
		balance1 = utxoSet.GetBalance(wallet1.GetAddress())
		assert.Equal(t, 50, balance1)
	})
}

func TestReindex(t *testing.T) {
	t.Run("ブロックチェーンからUTXO再構築", func(t *testing.T) {
		wallet1, err := NewWallet()
		require.NoError(t, err)
		wallet2, err := NewWallet()
		require.NoError(t, err)

		bc := NewBlockchain(1, wallet1.GetAddress())

		// ブロックを追加
		coinbaseTx := NewCoinbaseTx(wallet2.GetAddress(), "Block 1")
		_, _, err = bc.MineBlock([]*Transaction{coinbaseTx})
		require.NoError(t, err)

		// UTXOセットを再構築
		utxoSet := NewUTXOSet(bc)

		// 両方のウォレットに残高がある
		balance1 := utxoSet.GetBalance(wallet1.GetAddress())
		balance2 := utxoSet.GetBalance(wallet2.GetAddress())

		assert.Equal(t, 50, balance1)
		assert.Equal(t, 50, balance2)
	})
}

func TestUTXOSetConcurrency(t *testing.T) {
	t.Run("並行アクセスでパニックしない", func(t *testing.T) {
		wallet, err := NewWallet()
		require.NoError(t, err)

		bc := NewBlockchain(1, wallet.GetAddress())
		utxoSet := NewUTXOSet(bc)

		// 並行して残高確認
		done := make(chan bool)
		for i := 0; i < 10; i++ {
			go func() {
				_ = utxoSet.GetBalance(wallet.GetAddress())
				_ = utxoSet.FindUTXO(wallet.GetAddress())
				done <- true
			}()
		}

		for i := 0; i < 10; i++ {
			<-done
		}
	})
}

func TestUTXOSetString(t *testing.T) {
	t.Run("String()がパニックしない", func(t *testing.T) {
		wallet, err := NewWallet()
		require.NoError(t, err)

		bc := NewBlockchain(1, wallet.GetAddress())
		utxoSet := NewUTXOSet(bc)

		assert.NotPanics(t, func() {
			_ = utxoSet.String()
		})

		str := utxoSet.String()
		assert.Contains(t, str, "UTXO Set")
	})
}

func TestUTXOWithMultipleTransactions(t *testing.T) {
	t.Run("複数トランザクションの統合テスト", func(t *testing.T) {
		wallet1, err := NewWallet()
		require.NoError(t, err)
		wallet2, err := NewWallet()
		require.NoError(t, err)
		wallet3, err := NewWallet()
		require.NoError(t, err)

		bc := NewBlockchain(1, wallet1.GetAddress())
		utxoSet := NewUTXOSet(bc)

		// ブロック1: wallet2とwallet3にコインベース
		coinbase2 := NewCoinbaseTx(wallet2.GetAddress(), "Block 1 - wallet2")
		coinbase3 := NewCoinbaseTx(wallet3.GetAddress(), "Block 1 - wallet3")
		block1, _, err := bc.MineBlock([]*Transaction{coinbase2, coinbase3})
		require.NoError(t, err)

		err = utxoSet.Update(block1)
		require.NoError(t, err)

		// 各ウォレットの残高確認
		balance1 := utxoSet.GetBalance(wallet1.GetAddress())
		balance2 := utxoSet.GetBalance(wallet2.GetAddress())
		balance3 := utxoSet.GetBalance(wallet3.GetAddress())

		assert.Equal(t, 50, balance1) // ジェネシス
		assert.Equal(t, 50, balance2) // coinbase2
		assert.Equal(t, 50, balance3) // coinbase3

		// UTXO数の確認
		utxos1 := utxoSet.FindUTXO(wallet1.GetAddress())
		utxos2 := utxoSet.FindUTXO(wallet2.GetAddress())
		utxos3 := utxoSet.FindUTXO(wallet3.GetAddress())

		assert.Len(t, utxos1, 1)
		assert.Len(t, utxos2, 1)
		assert.Len(t, utxos3, 1)
	})
}

func TestUTXOSpentOutputRemoval(t *testing.T) {
	t.Run("使用済み出力の削除", func(t *testing.T) {
		wallet1, err := NewWallet()
		require.NoError(t, err)
		wallet2, err := NewWallet()
		require.NoError(t, err)

		bc := NewBlockchain(1, wallet1.GetAddress())
		utxoSet := NewUTXOSet(bc)

		// wallet1から送金するトランザクションを作成（簡略版）
		// 実際の送金ロジックは後で実装
		coinbaseTx := bc.Blocks[0].Transactions[0]

		// 入力: wallet1のコインベース出力を使用
		txIn := TxInput{
			TxID:      coinbaseTx.ID,
			OutIndex:  0,
			Signature: nil,
			PubKey:    []byte(wallet1.GetAddress()),
		}

		// 出力: wallet2に送金
		wallet2PubKeyHash, _ := hex.DecodeString(wallet2.GetAddress())
		txOut := TxOutput{
			Value:      30,
			PubKeyHash: wallet2PubKeyHash,
		}

		// おつり: wallet1に返す
		wallet1PubKeyHash, _ := hex.DecodeString(wallet1.GetAddress())
		changeOut := TxOutput{
			Value:      20,
			PubKeyHash: wallet1PubKeyHash,
		}

		tx := &Transaction{
			Inputs:  []TxInput{txIn},
			Outputs: []TxOutput{txOut, changeOut},
		}
		tx.ID = tx.Hash()

		// ブロックに追加
		block, _, err := bc.MineBlock([]*Transaction{tx})
		require.NoError(t, err)

		// UTXO更新
		err = utxoSet.Update(block)
		require.NoError(t, err)

		// wallet1の残高: 20 (おつり)
		// wallet2の残高: 30
		balance1 := utxoSet.GetBalance(wallet1.GetAddress())
		balance2 := utxoSet.GetBalance(wallet2.GetAddress())

		assert.Equal(t, 20, balance1)
		assert.Equal(t, 30, balance2)
	})
}
