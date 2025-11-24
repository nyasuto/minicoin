// Package main implements blockchain logic for Stage 3 with transaction support.
package main

import (
	"encoding/hex"
	"fmt"
	"sync"
)

// Blockchain represents the blockchain
type Blockchain struct {
	Blocks     []*Block // ブロックのリスト
	Difficulty int      // マイニング難易度
	mutex      sync.RWMutex
}

// NewBlockchain は新しいブロックチェーンを作成します
func NewBlockchain(difficulty int, minerAddress string) *Blockchain {
	// ジェネシスブロックを作成
	genesis := NewGenesisBlock(difficulty, minerAddress)

	bc := &Blockchain{
		Blocks:     []*Block{genesis},
		Difficulty: difficulty,
	}

	return bc
}

// MineBlock はトランザクションを含むブロックをマイニングして追加します
func (bc *Blockchain) MineBlock(transactions []*Transaction) (*Block, *MiningMetrics, error) {
	bc.mutex.Lock()
	defer bc.mutex.Unlock()

	lastBlock := bc.Blocks[len(bc.Blocks)-1]

	// 新しいブロックを作成
	newBlock := NewBlock(
		lastBlock.Index+1,
		transactions,
		lastBlock.Hash,
		bc.Difficulty,
	)

	// マイニング
	metrics, err := MineBlock(newBlock)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to mine block: %w", err)
	}

	// ブロックをチェーンに追加
	bc.Blocks = append(bc.Blocks, newBlock)

	return newBlock, metrics, nil
}

// GetLatestBlock は最新のブロックを返します
func (bc *Blockchain) GetLatestBlock() *Block {
	bc.mutex.RLock()
	defer bc.mutex.RUnlock()

	if len(bc.Blocks) == 0 {
		return nil
	}

	return bc.Blocks[len(bc.Blocks)-1]
}

// GetChainLength はブロックチェーンの長さを返します
func (bc *Blockchain) GetChainLength() int {
	bc.mutex.RLock()
	defer bc.mutex.RUnlock()

	return len(bc.Blocks)
}

// IsValid はブロックチェーン全体の整合性を検証します
func (bc *Blockchain) IsValid() bool {
	bc.mutex.RLock()
	defer bc.mutex.RUnlock()

	if len(bc.Blocks) == 0 {
		return false
	}

	// ジェネシスブロックの検証
	if bc.Blocks[0].Index != 0 {
		return false
	}
	if bc.Blocks[0].PreviousHash != "" {
		return false
	}

	// 各ブロックを検証
	for i := 0; i < len(bc.Blocks); i++ {
		block := bc.Blocks[i]

		// ブロック自体の整合性
		if !block.Validate() {
			return false
		}

		// 前ブロックとのリンク検証（ジェネシス以外）
		if i > 0 {
			prevBlock := bc.Blocks[i-1]

			// インデックスの連続性
			if block.Index != prevBlock.Index+1 {
				return false
			}

			// 前ブロックのハッシュ
			if block.PreviousHash != prevBlock.Hash {
				return false
			}

			// タイムスタンプの順序
			if block.Timestamp < prevBlock.Timestamp {
				return false
			}
		}
	}

	return true
}

// FindTransaction はトランザクションIDからトランザクションを検索します
func (bc *Blockchain) FindTransaction(ID []byte) (*Transaction, error) {
	bc.mutex.RLock()
	defer bc.mutex.RUnlock()

	idStr := hex.EncodeToString(ID)

	for _, block := range bc.Blocks {
		for _, tx := range block.Transactions {
			if hex.EncodeToString(tx.ID) == idStr {
				return tx, nil
			}
		}
	}

	return nil, fmt.Errorf("transaction not found")
}

// SignTransaction はトランザクションに署名します
func (bc *Blockchain) SignTransaction(tx *Transaction, wallet *Wallet) error {
	// 前トランザクションを取得
	prevTxs := make(map[string]*Transaction)

	for _, input := range tx.Inputs {
		prevTx, err := bc.FindTransaction(input.TxID)
		if err != nil {
			return fmt.Errorf("prev transaction not found: %w", err)
		}
		prevTxs[hex.EncodeToString(prevTx.ID)] = prevTx
	}

	// 署名
	return tx.Sign(wallet, prevTxs)
}

// VerifyTransaction はトランザクションを検証します
func (bc *Blockchain) VerifyTransaction(tx *Transaction) bool {
	if tx.IsCoinbase() {
		return true
	}

	// 前トランザクションを取得
	prevTxs := make(map[string]*Transaction)

	for _, input := range tx.Inputs {
		prevTx, err := bc.FindTransaction(input.TxID)
		if err != nil {
			return false
		}
		prevTxs[hex.EncodeToString(prevTx.ID)] = prevTx
	}

	// 検証
	return tx.Verify(prevTxs)
}

// GetAllTransactions はブロックチェーン内の全トランザクションを返します
func (bc *Blockchain) GetAllTransactions() []*Transaction {
	bc.mutex.RLock()
	defer bc.mutex.RUnlock()

	var transactions []*Transaction

	for _, block := range bc.Blocks {
		transactions = append(transactions, block.Transactions...)
	}

	return transactions
}

// String はブロックチェーンの文字列表現を返します
func (bc *Blockchain) String() string {
	result := fmt.Sprintf("Blockchain (Length: %d, Difficulty: %d)\n", len(bc.Blocks), bc.Difficulty)
	result += "====================================\n"

	for _, block := range bc.Blocks {
		result += block.String()
		result += "------------------------------------\n"
	}

	if bc.IsValid() {
		result += "Status: Valid\n"
	} else {
		result += "Status: Invalid\n"
	}

	return result
}
