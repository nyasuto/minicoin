package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/nyasuto/minicoin/common"
)

// Blockchain はPoWマイニング対応のブロックチェーン
type Blockchain struct {
	Blocks          []*Block
	Difficulty      int // 現在の難易度
	TargetBlockTime int // 目標ブロック生成時間（秒）
	mutex           sync.RWMutex
}

// NewBlockchain は新しいブロックチェーンを生成します
func NewBlockchain(difficulty int) *Blockchain {
	return &Blockchain{
		Blocks:          []*Block{NewGenesisBlock(difficulty)},
		Difficulty:      difficulty,
		TargetBlockTime: TargetBlockTime, // difficulty.goの定数を使用
	}
}

// AddBlock はチェーンに新しいブロックを追加します（マイニング実行）
func (bc *Blockchain) AddBlock(data string) (*MiningMetrics, error) {
	bc.mutex.Lock()
	defer bc.mutex.Unlock()

	previousBlock := bc.Blocks[len(bc.Blocks)-1]

	newBlock := NewBlock(
		previousBlock.Index+1,
		data,
		previousBlock.Hash,
		bc.Difficulty,
	)

	// マイニング実行
	metrics, err := MineBlock(newBlock, bc.Difficulty)
	if err != nil {
		return nil, err
	}

	bc.Blocks = append(bc.Blocks, newBlock)

	// 難易度の自動調整
	if ShouldAdjustDifficulty(bc) {
		oldDifficulty := bc.Difficulty
		bc.Difficulty = CalculateDifficulty(bc, bc.TargetBlockTime)
		if oldDifficulty != bc.Difficulty {
			// 難易度が変更された場合のログ（オプション）
			_ = oldDifficulty // 将来のログ用に残す
		}
	}

	return metrics, nil
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

// GetChainLength はチェーンの長さを返します
func (bc *Blockchain) GetChainLength() int {
	bc.mutex.RLock()
	defer bc.mutex.RUnlock()

	return len(bc.Blocks)
}

// IsValid はチェーン全体の整合性を検証します（PoW検証を含む）
func (bc *Blockchain) IsValid() bool {
	bc.mutex.RLock()
	defer bc.mutex.RUnlock()

	if len(bc.Blocks) == 0 {
		return false
	}

	// ジェネシスブロックの検証
	genesis := bc.Blocks[0]
	if genesis.Index != 0 || genesis.PreviousHash != "" {
		return false
	}
	if !ValidateProofOfWork(genesis) {
		return false
	}

	// 各ブロックを検証
	for i := 1; i < len(bc.Blocks); i++ {
		currentBlock := bc.Blocks[i]
		previousBlock := bc.Blocks[i-1]

		// PoW検証
		if !ValidateProofOfWork(currentBlock) {
			return false
		}

		// PreviousHashの一致確認
		if currentBlock.PreviousHash != previousBlock.Hash {
			return false
		}

		// インデックスの連続性確認
		if currentBlock.Index != previousBlock.Index+1 {
			return false
		}

		// タイムスタンプの単調増加確認
		if currentBlock.Timestamp < previousBlock.Timestamp {
			return false
		}
	}

	return true
}

func main() {
	// コマンドラインフラグの定義
	difficultyFlag := flag.Int("difficulty", 2, "デフォルトのマイニング難易度")
	flag.Parse()

	// ブロックチェーンの初期化
	bc := NewBlockchain(*difficultyFlag)

	// 対話型CLI
	runInteractiveCLI(bc)
}

// runInteractiveCLI は対話型CLIを実行します
func runInteractiveCLI(bc *Blockchain) {
	reader := bufio.NewReader(os.Stdin)

	printHeader()

	for {
		printMenu()
		fmt.Print("選択してください: ")

		input, err := reader.ReadString('\n')
		if err != nil {
			fmt.Printf("❌ エラー: 入力の読み取りに失敗しました: %v\n", err)
			return
		}
		input = strings.TrimSpace(input)

		switch input {
		case "1":
			miningDemo(reader)
		case "2":
			addBlockInteractive(bc, reader)
		case "3":
			displayChain(bc)
		case "4":
			validateChain(bc)
		case "5":
			performanceComparison(reader)
		case "6":
			changeDifficulty(bc, reader)
		case "7":
			displayDifficultyStats(bc)
		case "8":
			fmt.Println("\n👋 Minicoinをご利用いただきありがとうございました！")
			return
		default:
			fmt.Println("❌ 無効な選択です。1-8の数字を入力してください。")
		}
	}
}

// printHeader はヘッダーを表示します
func printHeader() {
	fmt.Println("╔════════════════════════════════════════════════════════╗")
	fmt.Println("║  Minicoin Blockchain (Stage 2: Proof of Work)         ║")
	fmt.Println("╚════════════════════════════════════════════════════════╝")
	fmt.Println()
}

// printMenu はメニューを表示します
func printMenu() {
	fmt.Println("\n====================================")
	fmt.Println("  メニュー")
	fmt.Println("====================================")
	fmt.Println("1. マイニングデモを実行")
	fmt.Println("2. ブロックをマイニングして追加")
	fmt.Println("3. チェーン全体を表示")
	fmt.Println("4. チェーンを検証")
	fmt.Println("5. パフォーマンス比較")
	fmt.Println("6. 難易度を変更")
	fmt.Println("7. 難易度統計を表示")
	fmt.Println("8. 終了")
	fmt.Println("====================================")
}

// miningDemo はマイニングデモを実行します
func miningDemo(reader *bufio.Reader) {
	fmt.Print("\n難易度を選択してください (0-5): ")
	input, err := reader.ReadString('\n')
	if err != nil {
		fmt.Printf("❌ エラー: 入力の読み取りに失敗しました: %v\n", err)
		return
	}
	input = strings.TrimSpace(input)

	difficulty, err := strconv.Atoi(input)
	if err != nil || difficulty < 0 || difficulty > 5 {
		fmt.Println("❌ 難易度は0-5の範囲で指定してください")
		return
	}

	fmt.Print("ブロックデータを入力してください: ")
	data, err := reader.ReadString('\n')
	if err != nil {
		fmt.Printf("❌ エラー: 入力の読み取りに失敗しました: %v\n", err)
		return
	}
	data = strings.TrimSpace(data)

	if data == "" {
		data = "Demo Block"
	}

	fmt.Printf("\n⛏️  難易度 %d でマイニング開始...\n", difficulty)
	fmt.Printf("   目標: 先頭に %d 個の 0 が並ぶハッシュを見つける\n\n", difficulty)

	block := NewBlock(1, data, "0000000000000000000000000000000000000000000000000000000000000000", difficulty)

	metrics, err := MineBlock(block, difficulty)
	if err != nil {
		fmt.Printf("❌ マイニングエラー: %v\n", err)
		return
	}

	fmt.Println("\n✅ マイニング成功！")
	fmt.Println("────────────────────────────────────────────────────────")
	fmt.Printf("⏱️  所要時間:     %v\n", metrics.Duration)
	fmt.Printf("🔢 試行回数:     %d 回\n", metrics.AttemptsCount)
	fmt.Printf("⚡ ハッシュレート: %.2f hashes/sec\n", metrics.HashRate)
	fmt.Printf("🎲 Nonce:        %d\n", block.Nonce)
	fmt.Printf("🔐 Hash:         %s\n", block.Hash)
	fmt.Printf("✓  Difficulty:   %s%s\n", GetDifficultyPrefix(difficulty), strings.Repeat("x", 64-difficulty))
	fmt.Println("────────────────────────────────────────────────────────")

	// ハッシュが難易度条件を満たすか確認
	if CheckHashDifficulty(block.Hash, difficulty) {
		fmt.Println("✓ ハッシュは難易度条件を満たしています")
	}
}

// addBlockInteractive はユーザー入力からブロックをマイニングして追加します
func addBlockInteractive(bc *Blockchain, reader *bufio.Reader) {
	fmt.Print("\nブロックに含めるデータを入力してください: ")
	data, err := reader.ReadString('\n')
	if err != nil {
		fmt.Printf("❌ エラー: 入力の読み取りに失敗しました: %v\n", err)
		return
	}
	data = strings.TrimSpace(data)

	if data == "" {
		fmt.Println("❌ データが空です。ブロックは追加されませんでした。")
		return
	}

	// 難易度変更を検出するため、現在の難易度を保存
	oldDifficulty := bc.Difficulty

	fmt.Printf("\n⛏️  難易度 %d でマイニング中...\n", bc.Difficulty)

	metrics, err := bc.AddBlock(data)
	if err != nil {
		fmt.Printf("❌ エラー: ブロックの追加に失敗しました: %v\n", err)
		return
	}

	latestBlock := bc.GetLatestBlock()
	fmt.Println("\n✅ ブロックをマイニングしてチェーンに追加しました！")
	fmt.Println("────────────────────────────────────────────────────────")
	fmt.Printf("📦 Block #%d\n", latestBlock.Index)
	fmt.Printf("   Data:         %s\n", latestBlock.Data)
	fmt.Printf("   Hash:         %s\n", latestBlock.Hash)
	fmt.Printf("   Nonce:        %d\n", latestBlock.Nonce)
	fmt.Printf("   Difficulty:   %d\n", latestBlock.Difficulty)
	fmt.Printf("   ⏱️  所要時間:     %v\n", metrics.Duration)
	fmt.Printf("   🔢 試行回数:     %d 回\n", metrics.AttemptsCount)
	fmt.Printf("   ⚡ ハッシュレート: %.2f hashes/sec\n", metrics.HashRate)
	fmt.Println("────────────────────────────────────────────────────────")

	// 難易度が変更された場合に通知
	if bc.Difficulty != oldDifficulty {
		fmt.Println()
		fmt.Println("🔧 難易度調整が発生しました！")
		fmt.Println("────────────────────────────────────────────────────────")
		fmt.Printf("   %d → %d", oldDifficulty, bc.Difficulty)
		if bc.Difficulty > oldDifficulty {
			fmt.Println(" (難易度上昇 ⬆️)")
			fmt.Println("   平均ブロック時間が目標より短かったため、難易度が上がりました")
		} else {
			fmt.Println(" (難易度低下 ⬇️)")
			fmt.Println("   平均ブロック時間が目標より長かったため、難易度が下がりました")
		}
		fmt.Println("────────────────────────────────────────────────────────")
	}
}

// displayChain はチェーン全体を表示します
func displayChain(bc *Blockchain) {
	fmt.Println("\n╔════════════════════════════════════════════════════════╗")
	fmt.Printf("║  ブロックチェーン (全 %d ブロック, 難易度: %d)\n", bc.GetChainLength(), bc.Difficulty)
	fmt.Println("╚════════════════════════════════════════════════════════╝")

	for _, block := range bc.Blocks {
		fmt.Println()
		if block.Index == 0 {
			fmt.Printf("📦 Block #%d (Genesis Block)\n", block.Index)
		} else {
			fmt.Printf("📦 Block #%d\n", block.Index)
		}
		fmt.Println("────────────────────────────────────────────────────────")
		fmt.Printf("Timestamp:     %s\n", common.FormatTimestamp(block.Timestamp))
		fmt.Printf("Data:          %s\n", block.Data)
		if block.PreviousHash == "" {
			fmt.Printf("Previous Hash: (none)\n")
		} else {
			fmt.Printf("Previous Hash: %s\n", block.PreviousHash)
		}
		fmt.Printf("Hash:          %s\n", block.Hash)
		fmt.Printf("Nonce:         %d\n", block.Nonce)
		fmt.Printf("Difficulty:    %d\n", block.Difficulty)

		if ValidateProofOfWork(block) {
			fmt.Println("Status:        ✓ Valid PoW")
		} else {
			fmt.Println("Status:        ❌ Invalid PoW")
		}
	}

	fmt.Println("\n────────────────────────────────────────────────────────")
	if bc.IsValid() {
		fmt.Println("✓ チェーン状態: 有効")
	} else {
		fmt.Println("❌ チェーン状態: 無効")
	}
}

// validateChain はチェーンを検証します
func validateChain(bc *Blockchain) {
	fmt.Println("\n🔍 チェーンの検証を実行中...")

	if bc.IsValid() {
		fmt.Println("✓ チェーンは有効です")
		fmt.Printf("  全 %d ブロックのPoW検証に成功しました\n", bc.GetChainLength())
	} else {
		fmt.Println("❌ チェーンが無効です")
		fmt.Println("  ブロックの改ざんまたはPoW検証の失敗が検出されました")
	}
}

// performanceComparison は異なる難易度でのパフォーマンス比較を実行します
func performanceComparison(reader *bufio.Reader) {
	fmt.Println("\n⚡ パフォーマンス比較")
	fmt.Println("異なる難易度でのマイニング性能を比較します")
	fmt.Println()

	difficulties := []int{0, 1, 2, 3, 4}
	data := "Performance Test Block"

	fmt.Println("難易度 | 所要時間    | 試行回数  | ハッシュレート")
	fmt.Println("──────────────────────────────────────────────")

	for _, diff := range difficulties {
		block := NewBlock(1, data, "0000000000000000000000000000000000000000000000000000000000000000", diff)

		startTime := time.Now()
		metrics, err := MineBlock(block, diff)
		if err != nil {
			fmt.Printf("  %d    | エラー\n", diff)
			continue
		}

		fmt.Printf("  %d    | %10v | %8d | %10.2f h/s\n",
			diff,
			metrics.Duration,
			metrics.AttemptsCount,
			metrics.HashRate,
		)

		// 難易度4以上は時間がかかるので、デモではスキップ
		if diff >= 3 && time.Since(startTime) > 5*time.Second {
			fmt.Println("\n⚠️  難易度が高いため、これ以上の比較を中止します")
			break
		}
	}

	fmt.Println("──────────────────────────────────────────────")
	fmt.Println("難易度が1増えるごとに、平均で約16倍の時間がかかります")
}

// changeDifficulty はブロックチェーンの難易度を変更します
func changeDifficulty(bc *Blockchain, reader *bufio.Reader) {
	fmt.Printf("\n現在の難易度: %d\n", bc.Difficulty)
	fmt.Print("新しい難易度を入力してください (0-5): ")

	input, err := reader.ReadString('\n')
	if err != nil {
		fmt.Printf("❌ エラー: 入力の読み取りに失敗しました: %v\n", err)
		return
	}
	input = strings.TrimSpace(input)

	difficulty, err := strconv.Atoi(input)
	if err != nil || difficulty < 0 || difficulty > 5 {
		fmt.Println("❌ 難易度は0-5の範囲で指定してください")
		return
	}

	bc.Difficulty = difficulty
	fmt.Printf("✓ 難易度を %d に変更しました\n", difficulty)
}

// displayDifficultyStats は難易度統計情報を表示します
func displayDifficultyStats(bc *Blockchain) {
	stats := GetDifficultyStatsFromChain(bc)

	fmt.Println("\n╔════════════════════════════════════════════════════════╗")
	fmt.Println("║  難易度調整統計                                        ║")
	fmt.Println("╚════════════════════════════════════════════════════════╝")
	fmt.Println()
	fmt.Println("📊 現在の難易度情報")
	fmt.Println("────────────────────────────────────────────────────────")
	fmt.Printf("現在の難易度:       %d\n", stats.CurrentDifficulty)
	fmt.Printf("目標ブロック時間:   %d 秒\n", stats.TargetBlockTime)
	if stats.AverageBlockTime > 0 {
		fmt.Printf("平均ブロック時間:   %.2f 秒\n", stats.AverageBlockTime)

		// パフォーマンス評価
		ratio := stats.AverageBlockTime / float64(stats.TargetBlockTime)
		if ratio > 1.2 {
			fmt.Printf("状態:               ⚠️  遅い (目標の %.1f倍)\n", ratio)
		} else if ratio < 0.8 {
			fmt.Printf("状態:               ⚡ 速い (目標の %.1f倍)\n", ratio)
		} else {
			fmt.Printf("状態:               ✓ 適正 (目標の %.1f倍)\n", ratio)
		}
	} else {
		fmt.Println("平均ブロック時間:   (データ不足)")
		fmt.Println("状態:               (計算不可)")
	}
	fmt.Println()
	fmt.Println("📈 調整情報")
	fmt.Println("────────────────────────────────────────────────────────")
	fmt.Printf("調整間隔:           %d ブロックごと\n", AdjustmentInterval)
	fmt.Printf("次回調整まで:       %d ブロック\n", stats.NextAdjustment)
	fmt.Printf("チェーンの長さ:     %d ブロック\n", bc.GetChainLength())
	fmt.Println("────────────────────────────────────────────────────────")
	fmt.Println()
	fmt.Println("💡 ヒント:")
	fmt.Println("  - 難易度は自動調整されます")
	fmt.Println("  - 調整は10ブロックごとに行われます")
	fmt.Println("  - 平均時間が目標より長い場合、難易度は下がります")
	fmt.Println("  - 平均時間が目標より短い場合、難易度は上がります")
}
