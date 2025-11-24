// Package main implements UTXO (Unspent Transaction Output) model.
// This provides balance calculation and transaction validation capabilities.
package main

import (
	"encoding/hex"
	"fmt"
	"sync"
)

// UTXO represents an unspent transaction output
type UTXO struct {
	TxID     []byte   // トランザクションID
	OutIndex int      // 出力のインデックス
	Output   TxOutput // 出力データ
}

// UTXOSet はUTXO集合を管理します
type UTXOSet struct {
	UTXOs map[string][]UTXO // address -> UTXOs
	mutex sync.RWMutex
}

// NewUTXOSet はブロックチェーンからUTXO集合を生成します
func NewUTXOSet(blockchain *Blockchain) *UTXOSet {
	us := &UTXOSet{
		UTXOs: make(map[string][]UTXO),
	}

	if err := us.Reindex(blockchain); err != nil {
		// 初期化時のエラーは通常発生しないが、念のため空のセットを返す
		return &UTXOSet{UTXOs: make(map[string][]UTXO)}
	}

	return us
}

// FindSpendableOutputs は指定金額を満たす使用可能な出力を検索します
// 戻り値: (実際の合計額, トランザクションID -> 出力インデックスのマップ)
func (us *UTXOSet) FindSpendableOutputs(address string, amount int) (int, map[string][]int) {
	us.mutex.RLock()
	defer us.mutex.RUnlock()

	unspentOutputs := make(map[string][]int)
	accumulated := 0

	utxos := us.UTXOs[address]
	for _, utxo := range utxos {
		txID := hex.EncodeToString(utxo.TxID)
		unspentOutputs[txID] = append(unspentOutputs[txID], utxo.OutIndex)
		accumulated += utxo.Output.Value

		if accumulated >= amount {
			break
		}
	}

	return accumulated, unspentOutputs
}

// FindUTXO は指定アドレスのすべてのUTXOを取得します
func (us *UTXOSet) FindUTXO(address string) []UTXO {
	us.mutex.RLock()
	defer us.mutex.RUnlock()

	utxos := make([]UTXO, len(us.UTXOs[address]))
	copy(utxos, us.UTXOs[address])
	return utxos
}

// GetBalance は指定アドレスの残高を計算します
func (us *UTXOSet) GetBalance(address string) int {
	us.mutex.RLock()
	defer us.mutex.RUnlock()

	balance := 0
	for _, utxo := range us.UTXOs[address] {
		balance += utxo.Output.Value
	}

	return balance
}

// Update はブロック追加時にUTXOセットを更新します
func (us *UTXOSet) Update(block *Block) error {
	us.mutex.Lock()
	defer us.mutex.Unlock()

	// まず、使用された出力（inputs）を削除
	for _, tx := range block.Transactions {
		if !tx.IsCoinbase() {
			for _, input := range tx.Inputs {
				txID := hex.EncodeToString(input.TxID)

				// すべてのアドレスのUTXOから該当する出力を削除
				for address := range us.UTXOs {
					newUTXOs := []UTXO{}
					for _, utxo := range us.UTXOs[address] {
						if hex.EncodeToString(utxo.TxID) != txID || utxo.OutIndex != input.OutIndex {
							newUTXOs = append(newUTXOs, utxo)
						}
					}
					us.UTXOs[address] = newUTXOs
				}
			}
		}

		// 新しい出力（outputs）を追加
		for outIdx, output := range tx.Outputs {
			address := hex.EncodeToString(output.PubKeyHash)
			utxo := UTXO{
				TxID:     tx.ID,
				OutIndex: outIdx,
				Output:   output,
			}
			us.UTXOs[address] = append(us.UTXOs[address], utxo)
		}
	}

	return nil
}

// Reindex はブロックチェーン全体からUTXOセットを再構築します
func (us *UTXOSet) Reindex(blockchain *Blockchain) error {
	us.mutex.Lock()
	defer us.mutex.Unlock()

	// UTXOセットをクリア
	us.UTXOs = make(map[string][]UTXO)

	// 使用済み出力を追跡
	spentTXOs := make(map[string]map[int]bool)

	// ブロックチェーンを逆順に走査（最新ブロックから）
	for i := len(blockchain.Blocks) - 1; i >= 0; i-- {
		block := blockchain.Blocks[i]

		for _, tx := range block.Transactions {
			txID := hex.EncodeToString(tx.ID)

			// 出力を処理
			for outIdx, output := range tx.Outputs {
				// すでに使用済みの出力はスキップ
				if spentTXOs[txID] != nil && spentTXOs[txID][outIdx] {
					continue
				}

				// UTXOとして登録
				address := hex.EncodeToString(output.PubKeyHash)
				utxo := UTXO{
					TxID:     tx.ID,
					OutIndex: outIdx,
					Output:   output,
				}
				us.UTXOs[address] = append(us.UTXOs[address], utxo)
			}

			// 入力を処理（コインベース以外）
			if !tx.IsCoinbase() {
				for _, input := range tx.Inputs {
					inTxID := hex.EncodeToString(input.TxID)
					if spentTXOs[inTxID] == nil {
						spentTXOs[inTxID] = make(map[int]bool)
					}
					spentTXOs[inTxID][input.OutIndex] = true
				}
			}
		}
	}

	return nil
}

// String はUTXOセットの文字列表現を返します
func (us *UTXOSet) String() string {
	us.mutex.RLock()
	defer us.mutex.RUnlock()

	result := "UTXO Set:\n"
	for address, utxos := range us.UTXOs {
		result += fmt.Sprintf("  Address %s: %d UTXOs\n", address[:16]+"...", len(utxos))
		for _, utxo := range utxos {
			result += fmt.Sprintf("    - TxID: %s, Index: %d, Value: %d\n",
				hex.EncodeToString(utxo.TxID)[:16]+"...",
				utxo.OutIndex,
				utxo.Output.Value)
		}
	}

	return result
}
