// Package main implements a basic hash chain blockchain.
// This includes block structure and hash calculation.
package main

import (
	"fmt"
	"strconv"
	"time"

	"github.com/nyasuto/minicoin/common"
)

// Block はブロックチェーンの基本単位となるブロックを表します
type Block struct {
	Index        int64  // ブロック番号（0から始まる連番）
	Timestamp    int64  // ブロック生成時のUnixタイムスタンプ
	Data         string // ブロックに含まれるデータ
	PreviousHash string // 前のブロックのハッシュ値（16進数文字列）
	Hash         string // このブロックのハッシュ値（16進数文字列）
}

// NewBlock は新しいブロックを生成します
// index: ブロック番号
// data: ブロックに含めるデータ
// previousHash: 前のブロックのハッシュ値
func NewBlock(index int64, data string, previousHash string) *Block {
	block := &Block{
		Index:        index,
		Timestamp:    time.Now().Unix(),
		Data:         data,
		PreviousHash: previousHash,
	}
	// ブロックのハッシュを計算して設定
	block.Hash = block.CalculateHash()
	return block
}

// CalculateHash はブロックのSHA-256ハッシュを計算します
// Index + Timestamp + Data + PreviousHash を結合してハッシュ化
func (b *Block) CalculateHash() string {
	// ブロックの内容を文字列として結合
	record := strconv.FormatInt(b.Index, 10) +
		strconv.FormatInt(b.Timestamp, 10) +
		b.Data +
		b.PreviousHash

	// SHA-256ハッシュを計算して16進数文字列として返す
	return common.HashString(record)
}

// Validate はブロックの整合性を検証します
// 保存されているハッシュ値と再計算したハッシュ値が一致するかチェック
func (b *Block) Validate() bool {
	return b.Hash == b.CalculateHash()
}

// String はブロックの情報を人間が読みやすい形式で返します
func (b *Block) String() string {
	return fmt.Sprintf(
		"Block #%d [%s]\n"+
			"  Timestamp: %s\n"+
			"  Data: %s\n"+
			"  Previous Hash: %s\n"+
			"  Hash: %s",
		b.Index,
		common.FormatTimestamp(b.Timestamp),
		common.FormatTimestamp(b.Timestamp),
		b.Data,
		b.PreviousHash,
		b.Hash,
	)
}

// NewGenesisBlock はブロックチェーンの最初のブロック（ジェネシスブロック）を生成します
// ジェネシスブロックは前のブロックが存在しないため、PreviousHashは空文字列
func NewGenesisBlock() *Block {
	return NewBlock(0, "Genesis Block", "")
}
