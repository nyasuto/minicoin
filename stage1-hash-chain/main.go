package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/nyasuto/minicoin/common"
)

func main() {
	// コマンドラインフラグの定義
	validateFlag := flag.Bool("validate", false, "チェーン検証のみ実行して終了")
	statsFlag := flag.Bool("stats", false, "統計情報表示のみ")
	exportFile := flag.String("export", "", "チェーンをJSON形式でエクスポート")
	importFile := flag.String("import", "", "JSON形式のチェーンをインポート")
	flag.Parse()

	// ブロックチェーンの初期化
	var bc *Blockchain
	if *importFile != "" {
		// インポート
		imported, err := importBlockchain(*importFile)
		if err != nil {
			fmt.Printf("❌ エラー: チェーンのインポートに失敗しました: %v\n", err)
			os.Exit(1)
		}
		bc = imported
		fmt.Printf("✓ チェーンを %s からインポートしました\n", *importFile)
	} else {
		bc = NewBlockchain()
	}

	// --validate フラグ: 検証のみ実行
	if *validateFlag {
		printValidationResult(bc)
		return
	}

	// --stats フラグ: 統計情報のみ表示
	if *statsFlag {
		printStats(bc)
		return
	}

	// --export フラグ: エクスポートして終了
	if *exportFile != "" {
		if err := exportBlockchain(bc, *exportFile); err != nil {
			fmt.Printf("❌ エラー: チェーンのエクスポートに失敗しました: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("✓ チェーンを %s にエクスポートしました\n", *exportFile)
		return
	}

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
			addBlockInteractive(bc, reader)
		case "2":
			displayChain(bc)
		case "3":
			printValidationResult(bc)
		case "4":
			displayBlockByIndex(bc, reader)
		case "5":
			printStats(bc)
		case "6":
			fmt.Println("\n👋 Minicoinをご利用いただきありがとうございました！")
			return
		default:
			fmt.Println("❌ 無効な選択です。1-6の数字を入力してください。")
		}
	}
}

// printHeader はヘッダーを表示します
func printHeader() {
	fmt.Println("╔════════════════════════════════════════════════════════╗")
	fmt.Println("║  Minicoin Blockchain (Stage 1: Hash Chain)            ║")
	fmt.Println("╚════════════════════════════════════════════════════════╝")
	fmt.Println()
}

// printMenu はメニューを表示します
func printMenu() {
	fmt.Println("\n====================================")
	fmt.Println("  メニュー")
	fmt.Println("====================================")
	fmt.Println("1. ブロックを追加")
	fmt.Println("2. チェーン全体を表示")
	fmt.Println("3. チェーンを検証")
	fmt.Println("4. 特定ブロックを表示")
	fmt.Println("5. 統計情報を表示")
	fmt.Println("6. 終了")
	fmt.Println("====================================")
}

// addBlockInteractive はユーザー入力からブロックを追加します
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

	err = bc.AddBlock(data)
	if err != nil {
		fmt.Printf("❌ エラー: ブロックの追加に失敗しました: %v\n", err)
		return
	}

	latestBlock := bc.GetLatestBlock()
	fmt.Printf("\n✓ ブロック #%d を追加しました\n", latestBlock.Index)
	fmt.Printf("  Data: %s\n", latestBlock.Data)
	fmt.Printf("  Hash: %s\n", latestBlock.Hash)
}

// displayChain はチェーン全体を表示します
func displayChain(bc *Blockchain) {
	fmt.Println("\n╔════════════════════════════════════════════════════════╗")
	fmt.Printf("║  ブロックチェーン (全 %d ブロック)\n", bc.GetChainLength())
	fmt.Println("╚════════════════════════════════════════════════════════╝")

	bc.mutex.RLock()
	defer bc.mutex.RUnlock()

	for _, block := range bc.Blocks {
		displayBlockDetails(block)
		fmt.Println("────────────────────────────────────────────────────────")
	}

	// チェーン全体の状態
	if bc.IsValid() {
		fmt.Println("✓ チェーン状態: 有効")
	} else {
		fmt.Println("❌ チェーン状態: 無効")
	}
}

// displayBlockDetails はブロックの詳細を表示します
func displayBlockDetails(block *Block) {
	fmt.Printf("\n📦 Block #%d", block.Index)
	if block.Index == 0 {
		fmt.Print(" (Genesis Block)")
	}
	fmt.Println()
	fmt.Println("────────────────────────────────────────────────────────")
	fmt.Printf("Timestamp:     %s\n", common.FormatTimestamp(block.Timestamp))
	fmt.Printf("Data:          %s\n", block.Data)

	// PreviousHashの表示
	if block.PreviousHash == "" {
		fmt.Printf("Previous Hash: (none)\n")
	} else {
		fmt.Printf("Previous Hash: %s\n", block.PreviousHash)
	}

	fmt.Printf("Hash:          %s\n", block.Hash)

	// バリデーション状態
	if block.Validate() {
		fmt.Println("Status:        ✓ Valid")
	} else {
		fmt.Println("Status:        ❌ Invalid")
	}
}

// displayBlockByIndex は指定されたインデックスのブロックを表示します
func displayBlockByIndex(bc *Blockchain, reader *bufio.Reader) {
	fmt.Print("\n表示するブロックのインデックスを入力してください: ")
	input, err := reader.ReadString('\n')
	if err != nil {
		fmt.Printf("❌ エラー: 入力の読み取りに失敗しました: %v\n", err)
		return
	}
	input = strings.TrimSpace(input)

	index, err := strconv.ParseInt(input, 10, 64)
	if err != nil {
		fmt.Println("❌ エラー: 有効な数値を入力してください")
		return
	}

	block, err := bc.GetBlock(index)
	if err != nil {
		fmt.Printf("❌ エラー: %v\n", err)
		return
	}

	fmt.Println("\n╔════════════════════════════════════════════════════════╗")
	fmt.Printf("║  ブロック #%d の詳細\n", index)
	fmt.Println("╚════════════════════════════════════════════════════════╝")
	displayBlockDetails(block)
}

// printValidationResult はチェーン検証結果を表示します
func printValidationResult(bc *Blockchain) {
	fmt.Println("\n🔍 チェーンの検証を実行中...")

	if bc.IsValid() {
		fmt.Println("✓ チェーンは有効です")
		fmt.Printf("  全 %d ブロックの整合性が確認されました\n", bc.GetChainLength())
	} else {
		fmt.Println("❌ チェーンが無効です")
		fmt.Println("  ブロックの改ざんまたは不整合が検出されました")
	}
}

// printStats は統計情報を表示します
func printStats(bc *Blockchain) {
	fmt.Println("\n╔════════════════════════════════════════════════════════╗")
	fmt.Println("║  ブロックチェーン統計情報")
	fmt.Println("╚════════════════════════════════════════════════════════╝")

	fmt.Printf("総ブロック数:     %d\n", bc.GetChainLength())

	latestBlock := bc.GetLatestBlock()
	if latestBlock != nil {
		fmt.Printf("最新ブロック:     #%d\n", latestBlock.Index)
		fmt.Printf("最新タイムスタンプ: %s\n", common.FormatTimestamp(latestBlock.Timestamp))
		fmt.Printf("最新ハッシュ:     %s\n", latestBlock.Hash)
	}

	genesisBlock, err := bc.GetBlock(0)
	if err == nil && genesisBlock != nil {
		fmt.Printf("ジェネシスハッシュ: %s\n", genesisBlock.Hash)
	}

	if bc.IsValid() {
		fmt.Println("チェーン状態:     ✓ 有効")
	} else {
		fmt.Println("チェーン状態:     ❌ 無効")
	}
}

// exportBlockchain はブロックチェーンをJSON形式でエクスポートします
func exportBlockchain(bc *Blockchain, filename string) error {
	bc.mutex.RLock()
	defer bc.mutex.RUnlock()

	data, err := json.MarshalIndent(bc.Blocks, "", "  ")
	if err != nil {
		return fmt.Errorf("JSON変換エラー: %w", err)
	}

	err = os.WriteFile(filename, data, 0600)
	if err != nil {
		return fmt.Errorf("ファイル書き込みエラー: %w", err)
	}

	return nil
}

// importBlockchain はJSON形式のブロックチェーンをインポートします
func importBlockchain(filename string) (*Blockchain, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("ファイル読み込みエラー: %w", err)
	}

	var blocks []*Block
	err = json.Unmarshal(data, &blocks)
	if err != nil {
		return nil, fmt.Errorf("JSON解析エラー: %w", err)
	}

	bc := &Blockchain{
		Blocks: blocks,
	}

	// インポートしたチェーンの検証
	if !bc.IsValid() {
		return nil, fmt.Errorf("インポートされたチェーンが無効です")
	}

	return bc, nil
}
