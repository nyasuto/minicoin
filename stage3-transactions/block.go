// Package main implements the Block structure for Stage 3 with transaction support.
package main

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"time"

	"github.com/nyasuto/minicoin/common"
)

// Block represents a block in the blockchain with transactions
type Block struct {
	Index        int64          // ブロック番号
	Timestamp    int64          // タイムスタンプ
	Transactions []*Transaction // トランザクションリスト
	PreviousHash string         // 前ブロックのハッシュ
	Hash         string         // このブロックのハッシュ
	Nonce        int64          // PoWのナンス
	Difficulty   int            // マイニング難易度
}

// NewBlock は新しいブロックを作成します
func NewBlock(index int64, transactions []*Transaction, previousHash string, difficulty int) *Block {
	block := &Block{
		Index:        index,
		Timestamp:    time.Now().Unix(),
		Transactions: transactions,
		PreviousHash: previousHash,
		Nonce:        0,
		Difficulty:   difficulty,
	}

	// ハッシュは後でマイニング時に計算
	block.Hash = ""

	return block
}

// NewGenesisBlock はジェネシスブロックを作成します
func NewGenesisBlock(difficulty int, minerAddress string) *Block {
	// コインベーストランザクションを作成
	coinbaseTx := NewCoinbaseTx(minerAddress, "Genesis Block")

	block := &Block{
		Index:        0,
		Timestamp:    time.Now().Unix(),
		Transactions: []*Transaction{coinbaseTx},
		PreviousHash: "",
		Nonce:        0,
		Difficulty:   difficulty,
	}

	// ジェネシスブロックをマイニング
	_, err := MineBlock(block)
	if err != nil {
		// ジェネシスブロック作成は失敗してはいけない
		panic(fmt.Sprintf("failed to mine genesis block: %v", err))
	}

	return block
}

// CalculateHashWithNonce はナンスを含めたブロックのハッシュを計算します
func (b *Block) CalculateHashWithNonce() string {
	data := b.prepareData()
	hash := common.Hash(data)
	return common.BytesToHex(hash)
}

// prepareData はハッシュ計算用のデータを準備します
func (b *Block) prepareData() []byte {
	var buffer bytes.Buffer

	// ブロックデータをシリアライズ
	encoder := gob.NewEncoder(&buffer)

	// トランザクションのハッシュリストを作成
	txHashes := make([][]byte, 0, len(b.Transactions))
	for _, tx := range b.Transactions {
		txHashes = append(txHashes, tx.ID)
	}

	// ハッシュ計算用のデータ構造
	type hashData struct {
		Index        int64
		Timestamp    int64
		TxHashes     [][]byte
		PreviousHash string
		Nonce        int64
		Difficulty   int
	}

	data := hashData{
		Index:        b.Index,
		Timestamp:    b.Timestamp,
		TxHashes:     txHashes,
		PreviousHash: b.PreviousHash,
		Nonce:        b.Nonce,
		Difficulty:   b.Difficulty,
	}

	err := encoder.Encode(data)
	if err != nil {
		return []byte{}
	}

	return buffer.Bytes()
}

// HashTransactions はブロック内の全トランザクションのマークルルートを計算します
func (b *Block) HashTransactions() []byte {
	txHashes := make([][]byte, 0, len(b.Transactions))

	for _, tx := range b.Transactions {
		txHashes = append(txHashes, tx.ID)
	}

	return common.MerkleRoot(txHashes)
}

// Validate はブロックの整合性を検証します
func (b *Block) Validate() bool {
	// ハッシュの再計算
	calculatedHash := b.CalculateHashWithNonce()

	// ハッシュが一致するか
	if calculatedHash != b.Hash {
		return false
	}

	// PoW検証
	if !ValidateProofOfWork(b) {
		return false
	}

	return true
}

// String はブロックの文字列表現を返します
func (b *Block) String() string {
	result := fmt.Sprintf("Block #%d\n", b.Index)
	result += fmt.Sprintf("Timestamp: %s\n", time.Unix(b.Timestamp, 0).Format("2006-01-02 15:04:05"))
	result += fmt.Sprintf("Transactions: %d\n", len(b.Transactions))
	result += fmt.Sprintf("Previous Hash: %s\n", b.PreviousHash)
	result += fmt.Sprintf("Hash: %s\n", b.Hash)
	result += fmt.Sprintf("Nonce: %d\n", b.Nonce)
	result += fmt.Sprintf("Difficulty: %d\n", b.Difficulty)

	return result
}
