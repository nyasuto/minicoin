// Package main implements Proof of Work mining for Stage 3.
package main

import (
	"fmt"
	"strings"
	"time"
)

// MiningMetrics はマイニングのパフォーマンス指標を保持します
type MiningMetrics struct {
	Attempts  int64         // 試行回数
	Duration  time.Duration // マイニング時間
	HashRate  float64       // ハッシュレート (hashes/second)
	Nonce     int64         // 見つかったナンス
	Hash      string        // 見つかったハッシュ
	Difficult int           // 難易度
}

// CheckHashDifficulty はハッシュが指定の難易度を満たすか確認します
func CheckHashDifficulty(hash string, difficulty int) bool {
	if difficulty == 0 {
		return true
	}

	prefix := strings.Repeat("0", difficulty)
	return strings.HasPrefix(hash, prefix)
}

// MineBlock はブロックをマイニングします
func MineBlock(block *Block) (*MiningMetrics, error) {
	if block.Difficulty < 0 {
		return nil, fmt.Errorf("difficulty must be non-negative")
	}

	startTime := time.Now()
	attempts := int64(0)

	// マイニング: 難易度を満たすハッシュを見つける
	for {
		hash := block.CalculateHashWithNonce()
		attempts++

		if CheckHashDifficulty(hash, block.Difficulty) {
			// 見つかった!
			block.Hash = hash
			duration := time.Since(startTime)

			metrics := &MiningMetrics{
				Attempts:  attempts,
				Duration:  duration,
				HashRate:  float64(attempts) / duration.Seconds(),
				Nonce:     block.Nonce,
				Hash:      hash,
				Difficult: block.Difficulty,
			}

			return metrics, nil
		}

		// ナンスをインクリメント
		block.Nonce++
	}
}

// ValidateProofOfWork はブロックのProof of Workを検証します
func ValidateProofOfWork(block *Block) bool {
	// ハッシュが難易度を満たすか確認
	return CheckHashDifficulty(block.Hash, block.Difficulty)
}
