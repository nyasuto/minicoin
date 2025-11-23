package main

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"encoding/gob"
	"fmt"
	"math/big"
	"os"
)

// Wallets は複数のウォレットを管理します
type Wallets struct {
	Wallets map[string]*Wallet // address -> Wallet
}

// NewWallets は新しいウォレットコレクションを作成します
func NewWallets() *Wallets {
	return &Wallets{
		Wallets: make(map[string]*Wallet),
	}
}

// CreateWallet は新しいウォレットを作成し、アドレスを返します
func (ws *Wallets) CreateWallet() (string, error) {
	wallet, err := NewWallet()
	if err != nil {
		return "", fmt.Errorf("failed to create wallet: %w", err)
	}

	address := wallet.GetAddress()
	ws.Wallets[address] = wallet

	return address, nil
}

// GetWallet は指定されたアドレスのウォレットを取得します
func (ws *Wallets) GetWallet(address string) (*Wallet, error) {
	wallet, exists := ws.Wallets[address]
	if !exists {
		return nil, fmt.Errorf("wallet not found: %s", address)
	}
	return wallet, nil
}

// GetAddresses は全てのウォレットアドレスを返します
func (ws *Wallets) GetAddresses() []string {
	addresses := make([]string, 0, len(ws.Wallets))
	for address := range ws.Wallets {
		addresses = append(addresses, address)
	}
	return addresses
}

// walletsData はウォレット保存用の構造体
type walletsData struct {
	Wallets map[string]*walletData
}

// SaveToFile は全てのウォレットをファイルに保存します
func (ws *Wallets) SaveToFile(filename string) error {
	// ウォレットデータを変換
	data := walletsData{
		Wallets: make(map[string]*walletData),
	}

	for address, wallet := range ws.Wallets {
		data.Wallets[address] = &walletData{
			PrivateKeyD: wallet.PrivateKey.D.Bytes(),
			PrivateKeyX: wallet.PrivateKey.X.Bytes(),
			PrivateKeyY: wallet.PrivateKey.Y.Bytes(),
			Address:     wallet.Address,
		}
	}

	// エンコード
	var buffer bytes.Buffer
	encoder := gob.NewEncoder(&buffer)
	err := encoder.Encode(data)
	if err != nil {
		return fmt.Errorf("failed to encode wallets: %w", err)
	}

	// ファイルに書き込み
	err = os.WriteFile(filename, buffer.Bytes(), 0600)
	if err != nil {
		return fmt.Errorf("failed to write wallets file: %w", err)
	}

	return nil
}

// LoadWalletsFromFile はウォレットをファイルから読み込みます
func LoadWalletsFromFile(filename string) (*Wallets, error) {
	// ファイルが存在しない場合は空のコレクションを返す
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		return NewWallets(), nil
	}

	// ファイルを読み込み
	fileData, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read wallets file: %w", err)
	}

	// デコード
	var data walletsData
	decoder := gob.NewDecoder(bytes.NewReader(fileData))
	err = decoder.Decode(&data)
	if err != nil {
		return nil, fmt.Errorf("failed to decode wallets: %w", err)
	}

	// ウォレットを復元
	wallets := NewWallets()
	for address, wData := range data.Wallets {
		wallet, err := restoreWallet(wData)
		if err != nil {
			return nil, fmt.Errorf("failed to restore wallet %s: %w", address, err)
		}
		wallets.Wallets[address] = wallet
	}

	return wallets, nil
}

// restoreWallet はwalletDataからWalletを復元します
func restoreWallet(data *walletData) (*Wallet, error) {
	// 秘密鍵を復元
	curve := elliptic.P256()
	privateKey := &ecdsa.PrivateKey{
		PublicKey: ecdsa.PublicKey{
			Curve: curve,
			X:     new(big.Int).SetBytes(data.PrivateKeyX),
			Y:     new(big.Int).SetBytes(data.PrivateKeyY),
		},
		D: new(big.Int).SetBytes(data.PrivateKeyD),
	}

	wallet := &Wallet{
		PrivateKey: privateKey,
		PublicKey:  &privateKey.PublicKey,
		Address:    data.Address,
	}

	return wallet, nil
}
