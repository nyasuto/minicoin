package main

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"encoding/gob"
	"fmt"
	"math/big"
	"os"

	"github.com/nyasuto/minicoin/common"
)

// Wallet はユーザーのウォレットを表します
type Wallet struct {
	PrivateKey *ecdsa.PrivateKey
	PublicKey  *ecdsa.PublicKey
	Address    string
}

// NewWallet は新しいウォレットを生成します
func NewWallet() (*Wallet, error) {
	// 鍵ペアを生成
	privateKey, err := common.GenerateKeyPair()
	if err != nil {
		return nil, fmt.Errorf("failed to generate key pair: %w", err)
	}

	// 公開鍵を取得
	publicKey := &privateKey.PublicKey

	// アドレスを生成
	address := common.PublicKeyToAddress(publicKey)

	wallet := &Wallet{
		PrivateKey: privateKey,
		PublicKey:  publicKey,
		Address:    address,
	}

	return wallet, nil
}

// GetAddress はウォレットのアドレスを返します
func (w *Wallet) GetAddress() string {
	return w.Address
}

// Sign はデータに署名します
func (w *Wallet) Sign(data []byte) ([]byte, error) {
	signature, err := common.Sign(w.PrivateKey, data)
	if err != nil {
		return nil, fmt.Errorf("failed to sign data: %w", err)
	}
	return signature, nil
}

// VerifySignature は署名を検証します
func VerifySignature(publicKey *ecdsa.PublicKey, data, signature []byte) bool {
	return common.Verify(publicKey, data, signature)
}

// walletData はウォレット保存用の構造体
type walletData struct {
	PrivateKeyD []byte
	PrivateKeyX []byte
	PrivateKeyY []byte
	Address     string
}

// SaveToFile はウォレットをファイルに保存します
func (w *Wallet) SaveToFile(filename string) error {
	// 秘密鍵のデータを抽出
	data := walletData{
		PrivateKeyD: w.PrivateKey.D.Bytes(),
		PrivateKeyX: w.PrivateKey.X.Bytes(),
		PrivateKeyY: w.PrivateKey.Y.Bytes(),
		Address:     w.Address,
	}

	// エンコード
	var buffer bytes.Buffer
	encoder := gob.NewEncoder(&buffer)
	err := encoder.Encode(data)
	if err != nil {
		return fmt.Errorf("failed to encode wallet: %w", err)
	}

	// ファイルに書き込み
	err = os.WriteFile(filename, buffer.Bytes(), 0600)
	if err != nil {
		return fmt.Errorf("failed to write wallet file: %w", err)
	}

	return nil
}

// LoadWalletFromFile はウォレットをファイルから読み込みます
func LoadWalletFromFile(filename string) (*Wallet, error) {
	// #nosec G304 -- ファイル読み込みは教育目的のため許容
	fileData, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read wallet file: %w", err)
	}

	// デコード
	var data walletData
	decoder := gob.NewDecoder(bytes.NewReader(fileData))
	err = decoder.Decode(&data)
	if err != nil {
		return nil, fmt.Errorf("failed to decode wallet: %w", err)
	}

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
