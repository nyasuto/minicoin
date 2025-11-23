package main

import (
	"errors"
	"fmt"
	"sync"
)

// Blockchain はブロックチェーン全体を管理する構造体
type Blockchain struct {
	Blocks []*Block     // ブロックのスライス（ジェネシスブロックから順に格納）
	mutex  sync.RWMutex // 並行アクセス制御用のRWMutex
}

// NewBlockchain は新しいブロックチェーンを生成します
// ジェネシスブロックが自動的に追加されます
func NewBlockchain() *Blockchain {
	return &Blockchain{
		Blocks: []*Block{NewGenesisBlock()},
	}
}

// AddBlock はチェーンに新しいブロックを追加します
// data: ブロックに含めるデータ
// エラーが発生した場合はエラーを返します
func (bc *Blockchain) AddBlock(data string) error {
	bc.mutex.Lock()
	defer bc.mutex.Unlock()

	// 最新ブロックを取得
	previousBlock := bc.Blocks[len(bc.Blocks)-1]

	// 新しいブロックを生成
	newBlock := NewBlock(
		previousBlock.Index+1,
		data,
		previousBlock.Hash,
	)

	// チェーンに追加
	bc.Blocks = append(bc.Blocks, newBlock)

	return nil
}

// GetLatestBlock はチェーンの最新ブロックを返します
func (bc *Blockchain) GetLatestBlock() *Block {
	bc.mutex.RLock()
	defer bc.mutex.RUnlock()

	if len(bc.Blocks) == 0 {
		return nil
	}

	return bc.Blocks[len(bc.Blocks)-1]
}

// GetBlock は指定されたインデックスのブロックを返します
// index: 取得したいブロックのインデックス
// 範囲外のインデックスが指定された場合はエラーを返します
func (bc *Blockchain) GetBlock(index int64) (*Block, error) {
	bc.mutex.RLock()
	defer bc.mutex.RUnlock()

	if index < 0 || index >= int64(len(bc.Blocks)) {
		return nil, errors.New("index out of range")
	}

	return bc.Blocks[index], nil
}

// GetChainLength はチェーンの長さ（ブロック数）を返します
func (bc *Blockchain) GetChainLength() int {
	bc.mutex.RLock()
	defer bc.mutex.RUnlock()

	return len(bc.Blocks)
}

// IsValid はチェーン全体の整合性を検証します
// 以下の項目を検証:
// 1. 各ブロックのハッシュが正しく計算されているか
// 2. PreviousHashが実際に前のブロックのハッシュと一致するか
// 3. インデックスが連続しているか
// 4. タイムスタンプが単調増加しているか（等しいのは許容）
func (bc *Blockchain) IsValid() bool {
	bc.mutex.RLock()
	defer bc.mutex.RUnlock()

	// 空のチェーンは無効
	if len(bc.Blocks) == 0 {
		return false
	}

	// ジェネシスブロックの検証
	genesis := bc.Blocks[0]
	if genesis.Index != 0 {
		return false
	}
	if genesis.PreviousHash != "" {
		return false
	}
	if !genesis.Validate() {
		return false
	}

	// 各ブロックを検証
	for i := 1; i < len(bc.Blocks); i++ {
		currentBlock := bc.Blocks[i]
		previousBlock := bc.Blocks[i-1]

		// 1. ブロックのハッシュが正しく計算されているか
		if !currentBlock.Validate() {
			return false
		}

		// 2. PreviousHashが実際に前のブロックのハッシュと一致するか
		if currentBlock.PreviousHash != previousBlock.Hash {
			return false
		}

		// 3. インデックスが連続しているか
		if currentBlock.Index != previousBlock.Index+1 {
			return false
		}

		// 4. タイムスタンプが単調増加しているか（等しいのは許容）
		if currentBlock.Timestamp < previousBlock.Timestamp {
			return false
		}
	}

	return true
}

// PrintChain はチェーン全体を表示します（デバッグ用）
func (bc *Blockchain) PrintChain() {
	bc.mutex.RLock()
	defer bc.mutex.RUnlock()

	fmt.Println("====================================")
	fmt.Printf("Blockchain (Length: %d)\n", len(bc.Blocks))
	fmt.Println("====================================")

	for _, block := range bc.Blocks {
		fmt.Println(block.String())
		fmt.Println("------------------------------------")
	}
}
