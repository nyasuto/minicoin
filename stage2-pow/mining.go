package main

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/nyasuto/minicoin/common"
)

// Block はProof of Workを含むブロックを表します
type Block struct {
	Index        int64  // ブロック番号
	Timestamp    int64  // タイムスタンプ(Unix時間)
	Data         string // ブロックに含まれるデータ
	PreviousHash string // 前のブロックのハッシュ
	Hash         string // このブロックのハッシュ
	Nonce        int64  // マイニングで使用するナンス
	Difficulty   int    // マイニング難易度
}

// MiningMetrics はマイニングのパフォーマンス情報を記録します
type MiningMetrics struct {
	AttemptsCount int64         // 試行回数
	Duration      time.Duration // マイニング時間
	HashRate      float64       // ハッシュレート(hashes/sec)
}

// NewBlock は新しいブロックを生成します（マイニングは未実施）
func NewBlock(index int64, data string, previousHash string, difficulty int) *Block {
	return &Block{
		Index:        index,
		Timestamp:    time.Now().Unix(),
		Data:         data,
		PreviousHash: previousHash,
		Nonce:        0,
		Difficulty:   difficulty,
		Hash:         "", // マイニング後に設定
	}
}

// NewGenesisBlock はジェネシスブロックを生成します
func NewGenesisBlock(difficulty int) *Block {
	block := NewBlock(0, "Genesis Block", "", difficulty)
	// ジェネシスブロックもマイニングする
	_, err := MineBlock(block, difficulty)
	if err != nil {
		// ジェネシスブロックのマイニング失敗は通常起こらないが、
		// 念のためパニックする
		panic(fmt.Sprintf("failed to mine genesis block: %v", err))
	}
	return block
}

// CalculateHashWithNonce はナンスを含むハッシュを計算します
func CalculateHashWithNonce(block *Block) string {
	record := strconv.FormatInt(block.Index, 10) +
		strconv.FormatInt(block.Timestamp, 10) +
		block.Data +
		block.PreviousHash +
		strconv.FormatInt(block.Nonce, 10) +
		strconv.Itoa(block.Difficulty)

	return common.HashString(record)
}

// CheckHashDifficulty はハッシュが指定された難易度を満たすか確認します
// 難易度は先頭のゼロの数で表現されます
func CheckHashDifficulty(hash string, difficulty int) bool {
	prefix := strings.Repeat("0", difficulty)
	return strings.HasPrefix(hash, prefix)
}

// MineBlock はブロックをマイニングします
// ハッシュが難易度条件を満たすまでナンスをインクリメントします
func MineBlock(block *Block, difficulty int) (*MiningMetrics, error) {
	if difficulty < 0 {
		return nil, fmt.Errorf("difficulty must be non-negative")
	}

	block.Difficulty = difficulty
	startTime := time.Now()
	attempts := int64(0)

	// ナンスを0から開始
	block.Nonce = 0

	for {
		// ハッシュを計算
		hash := CalculateHashWithNonce(block)
		attempts++

		// 難易度条件を満たすか確認
		if CheckHashDifficulty(hash, difficulty) {
			block.Hash = hash
			duration := time.Since(startTime)

			// メトリクスを計算
			metrics := &MiningMetrics{
				AttemptsCount: attempts,
				Duration:      duration,
			}

			if duration.Seconds() > 0 {
				metrics.HashRate = float64(attempts) / duration.Seconds()
			}

			return metrics, nil
		}

		// ナンスをインクリメント
		block.Nonce++

		// オーバーフロー防止（実際には起こりにくい）
		if block.Nonce < 0 {
			return nil, fmt.Errorf("nonce overflow - unable to find valid hash")
		}
	}
}

// ValidateProofOfWork はブロックのProof of Workを検証します
func ValidateProofOfWork(block *Block) bool {
	// ハッシュを再計算
	calculatedHash := CalculateHashWithNonce(block)

	// 保存されているハッシュと一致するか確認
	if calculatedHash != block.Hash {
		return false
	}

	// 難易度条件を満たすか確認
	return CheckHashDifficulty(block.Hash, block.Difficulty)
}

// Validate はブロックの整合性を検証します（PoWを含む）
func (b *Block) Validate() bool {
	return ValidateProofOfWork(b)
}

// String はブロックの情報を人間が読みやすい形式で返します
func (b *Block) String() string {
	return fmt.Sprintf(
		"Block #%d [%s]\n"+
			"  Timestamp: %s\n"+
			"  Data: %s\n"+
			"  Previous Hash: %s\n"+
			"  Hash: %s\n"+
			"  Nonce: %d\n"+
			"  Difficulty: %d",
		b.Index,
		common.FormatTimestamp(b.Timestamp),
		common.FormatTimestamp(b.Timestamp),
		b.Data,
		b.PreviousHash,
		b.Hash,
		b.Nonce,
		b.Difficulty,
	)
}

// GetDifficultyPrefix は難易度に応じたハッシュのプレフィックスを返します
func GetDifficultyPrefix(difficulty int) string {
	return strings.Repeat("0", difficulty)
}
