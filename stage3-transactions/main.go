// Package main implements CLI for Stage 3 blockchain with UTXO model.
package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

const walletFile = "wallet.dat"

func main() {
	printHeader()

	// ウォレットの読み込みまたは作成
	wallet, err := loadOrCreateWallet()
	if err != nil {
		fmt.Printf("❌ Failed to load wallet: %v\n", err)
		return
	}

	fmt.Printf("📱 Your Address: %s\n\n", wallet.GetAddress())

	// ブロックチェーン初期化
	bc := NewBlockchain(2, wallet.GetAddress())
	utxoSet := NewUTXOSet(bc)

	scanner := bufio.NewScanner(os.Stdin)

	for {
		printMenu()
		fmt.Print("選択: ")

		if !scanner.Scan() {
			break
		}

		choice := strings.TrimSpace(scanner.Text())

		switch choice {
		case "1":
			displayBalance(wallet, utxoSet)
		case "2":
			createWallet()
		case "3":
			displayChain(bc)
		case "4":
			displayTransactions(bc)
		case "5":
			mineBlock(bc, utxoSet, wallet)
		case "6":
			displayUTXOs(wallet, utxoSet)
		case "7":
			validateChain(bc)
		case "8":
			fmt.Println("\n👋 Goodbye!")
			return
		default:
			fmt.Println("❌ Invalid choice. Please try again.")
		}
	}
}

func printHeader() {
	fmt.Println("╔════════════════════════════════════════════════════════╗")
	fmt.Println("║  Minicoin Blockchain (Stage 3: Transactions + UTXO)   ║")
	fmt.Println("╚════════════════════════════════════════════════════════╝")
	fmt.Println()
}

func printMenu() {
	fmt.Println("\n====================================")
	fmt.Println("  メニュー")
	fmt.Println("====================================")
	fmt.Println("1. 残高確認")
	fmt.Println("2. 新しいウォレット作成")
	fmt.Println("3. ブロックチェーン表示")
	fmt.Println("4. トランザクション履歴")
	fmt.Println("5. ブロックをマイニング")
	fmt.Println("6. UTXOセット表示")
	fmt.Println("7. チェーン検証")
	fmt.Println("8. 終了")
	fmt.Println("====================================")
}

func loadOrCreateWallet() (*Wallet, error) {
	// ウォレットファイルが存在するか確認
	if _, err := os.Stat(walletFile); err == nil {
		fmt.Println("📂 Loading existing wallet...")
		wallet, err := LoadWalletFromFile(walletFile)
		if err != nil {
			fmt.Println("⚠️  Failed to load wallet, creating new one...")
			return createAndSaveWallet()
		}
		fmt.Println("✅ Wallet loaded successfully!")
		return wallet, nil
	}

	// 新しいウォレットを作成
	fmt.Println("🆕 Creating new wallet...")
	return createAndSaveWallet()
}

func createAndSaveWallet() (*Wallet, error) {
	wallet, err := NewWallet()
	if err != nil {
		return nil, err
	}

	err = wallet.SaveToFile(walletFile)
	if err != nil {
		fmt.Printf("⚠️  Warning: Could not save wallet: %v\n", err)
	} else {
		fmt.Println("✅ Wallet created and saved!")
	}

	return wallet, nil
}

func displayBalance(wallet *Wallet, utxoSet *UTXOSet) {
	balance := utxoSet.GetBalance(wallet.GetAddress())

	fmt.Println("\n💰 Current Balance")
	fmt.Println("────────────────────────────────────────────────────────")
	fmt.Printf("Address: %s\n", wallet.GetAddress())
	fmt.Printf("Balance: %d coins\n", balance)
	fmt.Println("────────────────────────────────────────────────────────")
}

func createWallet() {
	wallet, err := NewWallet()
	if err != nil {
		fmt.Printf("❌ Failed to create wallet: %v\n", err)
		return
	}

	filename := fmt.Sprintf("wallet_%s.dat", wallet.GetAddress()[:8])
	err = wallet.SaveToFile(filename)
	if err != nil {
		fmt.Printf("❌ Failed to save wallet: %v\n", err)
		return
	}

	fmt.Println("\n✅ New Wallet Created!")
	fmt.Println("────────────────────────────────────────────────────────")
	fmt.Printf("Address: %s\n", wallet.GetAddress())
	fmt.Printf("Saved to: %s\n", filename)
	fmt.Println("────────────────────────────────────────────────────────")
}

func displayChain(bc *Blockchain) {
	fmt.Printf("\n╔════════════════════════════════════════════════════════╗\n")
	fmt.Printf("║  ブロックチェーン (全 %d ブロック, 難易度: %d)\n", bc.GetChainLength(), bc.Difficulty)
	fmt.Printf("╚════════════════════════════════════════════════════════╝\n\n")

	for _, block := range bc.Blocks {
		fmt.Printf("📦 Block #%d\n", block.Index)
		fmt.Println("────────────────────────────────────────────────────────")
		fmt.Printf("Timestamp:     %s\n", formatTimestamp(block.Timestamp))
		fmt.Printf("Transactions:  %d\n", len(block.Transactions))
		fmt.Printf("Previous Hash: %s\n", truncateHash(block.PreviousHash))
		fmt.Printf("Hash:          %s\n", truncateHash(block.Hash))
		fmt.Printf("Nonce:         %d\n", block.Nonce)
		fmt.Printf("Difficulty:    %d\n", block.Difficulty)
		fmt.Println()
	}
}

func displayTransactions(bc *Blockchain) {
	fmt.Println("\n📝 Transaction History")
	fmt.Println("════════════════════════════════════════════════════════")

	for _, block := range bc.Blocks {
		fmt.Printf("\nBlock #%d:\n", block.Index)
		for i, tx := range block.Transactions {
			fmt.Printf("  [%d] TxID: %s\n", i, truncateHash(string(tx.ID)))
			fmt.Printf("      Inputs: %d, Outputs: %d\n", len(tx.Inputs), len(tx.Outputs))
			if tx.IsCoinbase() {
				fmt.Println("      Type: Coinbase (Mining Reward)")
			}
		}
	}

	fmt.Println("════════════════════════════════════════════════════════")
}

func mineBlock(bc *Blockchain, utxoSet *UTXOSet, wallet *Wallet) {
	fmt.Println("\n⛏️  Mining new block...")

	// コインベーストランザクションを作成
	coinbaseTx := NewCoinbaseTx(wallet.GetAddress(), fmt.Sprintf("Block %d reward", bc.GetChainLength()))

	// ブロックをマイニング
	block, metrics, err := bc.MineBlock([]*Transaction{coinbaseTx})
	if err != nil {
		fmt.Printf("❌ Mining failed: %v\n", err)
		return
	}

	// UTXO更新
	err = utxoSet.Update(block)
	if err != nil {
		fmt.Printf("⚠️  Warning: UTXO update failed: %v\n", err)
	}

	fmt.Println("\n✅ Block mined successfully!")
	fmt.Println("────────────────────────────────────────────────────────")
	fmt.Printf("Block #%d\n", block.Index)
	fmt.Printf("Hash:       %s\n", truncateHash(block.Hash))
	fmt.Printf("Nonce:      %d\n", metrics.Nonce)
	fmt.Printf("Attempts:   %d\n", metrics.Attempts)
	fmt.Printf("Duration:   %s\n", metrics.Duration)
	fmt.Printf("Hash Rate:  %.2f H/s\n", metrics.HashRate)
	fmt.Println("────────────────────────────────────────────────────────")
}

func displayUTXOs(wallet *Wallet, utxoSet *UTXOSet) {
	utxos := utxoSet.FindUTXO(wallet.GetAddress())

	fmt.Println("\n💎 Your UTXOs")
	fmt.Println("════════════════════════════════════════════════════════")

	if len(utxos) == 0 {
		fmt.Println("  No UTXOs found.")
	} else {
		total := 0
		for i, utxo := range utxos {
			fmt.Printf("[%d] TxID: %s\n", i+1, truncateHash(string(utxo.TxID)))
			fmt.Printf("    Index: %d, Value: %d coins\n", utxo.OutIndex, utxo.Output.Value)
			total += utxo.Output.Value
		}
		fmt.Println("────────────────────────────────────────────────────────")
		fmt.Printf("Total: %d coins\n", total)
	}

	fmt.Println("════════════════════════════════════════════════════════")
}

func validateChain(bc *Blockchain) {
	fmt.Println("\n🔍 Validating blockchain...")

	if bc.IsValid() {
		fmt.Println("✅ Blockchain is valid!")
		fmt.Printf("   All %d blocks verified successfully.\n", bc.GetChainLength())
	} else {
		fmt.Println("❌ Blockchain is INVALID!")
		fmt.Println("   Chain integrity compromised.")
	}
}

// Helper functions

func formatTimestamp(timestamp int64) string {
	return fmt.Sprintf("%d", timestamp)
}

func truncateHash(hash string) string {
	if len(hash) > 16 {
		return hash[:16] + "..."
	}
	return hash
}

// GetBlock returns a block by index (needed for compatibility)
func (bc *Blockchain) GetBlock(index int64) (*Block, error) {
	bc.mutex.RLock()
	defer bc.mutex.RUnlock()

	if index < 0 || index >= int64(len(bc.Blocks)) {
		return nil, fmt.Errorf("index out of range")
	}

	return bc.Blocks[index], nil
}
