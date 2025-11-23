// Package common provides cryptographic utilities for the blockchain implementation.
// This includes hashing, key generation, and digital signature operations.
package common

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/big"
)

// Hash はSHA-256ハッシュを計算します
func Hash(data []byte) []byte {
	hash := sha256.Sum256(data)
	return hash[:]
}

// HashString は文字列のSHA-256ハッシュを16進数文字列として返します
func HashString(data string) string {
	hash := Hash([]byte(data))
	return hex.EncodeToString(hash)
}

// GenerateKeyPair はECDSA鍵ペアを生成します
func GenerateKeyPair() (*ecdsa.PrivateKey, error) {
	curve := elliptic.P256()
	private, err := ecdsa.GenerateKey(curve, rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("failed to generate key pair: %w", err)
	}
	return private, nil
}

// Sign は秘密鍵を使ってデータに署名します
func Sign(privateKey *ecdsa.PrivateKey, data []byte) ([]byte, error) {
	hash := Hash(data)

	r, s, err := ecdsa.Sign(rand.Reader, privateKey, hash)
	if err != nil {
		return nil, fmt.Errorf("failed to sign data: %w", err)
	}

	// rとsを結合して署名を作成
	signature := append(r.Bytes(), s.Bytes()...)
	return signature, nil
}

// Verify は公開鍵を使って署名を検証します
func Verify(publicKey *ecdsa.PublicKey, data, signature []byte) bool {
	hash := Hash(data)

	// 署名からrとsを抽出
	sigLen := len(signature)
	if sigLen%2 != 0 || sigLen == 0 {
		return false
	}

	r := new(big.Int).SetBytes(signature[:sigLen/2])
	s := new(big.Int).SetBytes(signature[sigLen/2:])

	return ecdsa.Verify(publicKey, hash, r, s)
}

// PublicKeyToAddress は公開鍵からアドレス（16進数文字列）を生成します
func PublicKeyToAddress(publicKey *ecdsa.PublicKey) string {
	// 公開鍵をバイト列に変換
	pubKeyBytes := append(publicKey.X.Bytes(), publicKey.Y.Bytes()...)

	// SHA-256ハッシュを2回適用（Bitcoin式）
	hash1 := Hash(pubKeyBytes)
	hash2 := Hash(hash1)

	// 最初の20バイトを使用してアドレスを生成
	address := hex.EncodeToString(hash2[:20])
	return address
}
