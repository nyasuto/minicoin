package main

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"math/big"
	"time"

	"github.com/nyasuto/minicoin/common"
)

// Transaction はトランザクションを表します
type Transaction struct {
	ID        []byte     // トランザクションID（ハッシュ）
	Inputs    []TxInput  // 入力
	Outputs   []TxOutput // 出力
	Timestamp int64      // タイムスタンプ
}

// TxInput はトランザクション入力を表します
type TxInput struct {
	TxID      []byte // 参照するトランザクションID
	OutIndex  int    // 参照する出力のインデックス
	Signature []byte // 署名
	PubKey    []byte // 公開鍵
}

// TxOutput はトランザクション出力を表します
type TxOutput struct {
	Value      int    // 送金額
	PubKeyHash []byte // 受取人の公開鍵ハッシュ
}

// NewCoinbaseTx はコインベーストランザクション（マイニング報酬）を作成します
func NewCoinbaseTx(to string, data string) *Transaction {
	if data == "" {
		data = fmt.Sprintf("Reward to '%s'", to)
	}

	// コインベーストランザクションは入力なし
	txIn := TxInput{
		TxID:      []byte{},
		OutIndex:  -1,
		Signature: nil,
		PubKey:    []byte(data),
	}

	// アドレスを公開鍵ハッシュに変換
	pubKeyHash, err := hex.DecodeString(to)
	if err != nil {
		pubKeyHash = []byte(to)
	}

	txOut := TxOutput{
		Value:      50, // マイニング報酬
		PubKeyHash: pubKeyHash,
	}

	tx := &Transaction{
		Inputs:    []TxInput{txIn},
		Outputs:   []TxOutput{txOut},
		Timestamp: time.Now().Unix(),
	}

	tx.ID = tx.Hash()

	return tx
}

// NewTransaction は新しいトランザクションを作成します
// 注意: この実装は簡略版です。Issue #11でUTXO検索機能を追加します
func NewTransaction(from, to string, amount int, blockchain interface{}) (*Transaction, error) {
	if amount <= 0 {
		return nil, fmt.Errorf("amount must be positive")
	}

	// 現時点では簡単な実装（Issue #11でUTXOロジックを追加）
	// ここでは基本的な構造のみ作成
	inputs := []TxInput{}
	outputs := []TxOutput{}

	// from の公開鍵ハッシュ
	fromPubKeyHash, err := hex.DecodeString(from)
	if err != nil {
		return nil, fmt.Errorf("invalid from address: %w", err)
	}

	// to の公開鍵ハッシュ
	toPubKeyHash, err := hex.DecodeString(to)
	if err != nil {
		return nil, fmt.Errorf("invalid to address: %w", err)
	}

	// 出力を作成
	outputs = append(outputs, TxOutput{
		Value:      amount,
		PubKeyHash: toPubKeyHash,
	})

	// おつりの出力（今は簡略化のため省略、Issue #11で実装）

	tx := &Transaction{
		Inputs:    inputs,
		Outputs:   outputs,
		Timestamp: time.Now().Unix(),
	}

	tx.ID = tx.Hash()

	// 署名は Issue #11 で実装
	_ = fromPubKeyHash

	return tx, nil
}

// Hash はトランザクションのハッシュを計算します
func (tx *Transaction) Hash() []byte {
	txCopy := *tx
	txCopy.ID = []byte{}

	return common.Hash(tx.serialize())
}

// serialize はトランザクションをバイト列にシリアライズします
func (tx *Transaction) serialize() []byte {
	var buffer bytes.Buffer

	encoder := gob.NewEncoder(&buffer)
	err := encoder.Encode(tx)
	if err != nil {
		return []byte{}
	}

	return buffer.Bytes()
}

// IsCoinbase はコインベーストランザクションかどうかを判定します
func (tx *Transaction) IsCoinbase() bool {
	return len(tx.Inputs) == 1 && len(tx.Inputs[0].TxID) == 0 && tx.Inputs[0].OutIndex == -1
}

// Sign はトランザクションに署名します
// prevTxs: 参照する前トランザクションのマップ（TxID(hex) -> Transaction）
func (tx *Transaction) Sign(wallet *Wallet, prevTxs map[string]*Transaction) error {
	if tx.IsCoinbase() {
		return nil // コインベーストランザクションは署名不要
	}

	// 各入力について前トランザクションが存在するか確認
	for _, input := range tx.Inputs {
		if prevTxs[hex.EncodeToString(input.TxID)] == nil {
			return fmt.Errorf("previous transaction not found")
		}
	}

	// トランザクションのコピーを作成
	txCopy := tx.trimmedCopy()

	// 各入力に署名
	for i, input := range txCopy.Inputs {
		prevTx := prevTxs[hex.EncodeToString(input.TxID)]
		txCopy.Inputs[i].Signature = nil
		txCopy.Inputs[i].PubKey = prevTx.Outputs[input.OutIndex].PubKeyHash
		txCopy.ID = txCopy.Hash()
		txCopy.Inputs[i].PubKey = nil

		// 署名を生成
		signature, err := wallet.Sign(txCopy.ID)
		if err != nil {
			return fmt.Errorf("failed to sign transaction: %w", err)
		}

		tx.Inputs[i].Signature = signature
		tx.Inputs[i].PubKey = publicKeyToBytes(wallet.PublicKey)
	}

	return nil
}

// Verify はトランザクションの署名を検証します
func (tx *Transaction) Verify(prevTxs map[string]*Transaction) bool {
	if tx.IsCoinbase() {
		return true // コインベーストランザクションは常に有効
	}

	// 各入力について前トランザクションが存在するか確認
	for _, input := range tx.Inputs {
		if prevTxs[hex.EncodeToString(input.TxID)] == nil {
			return false
		}
	}

	// トランザクションのコピーを作成
	txCopy := tx.trimmedCopy()

	// 各入力の署名を検証
	for i, input := range tx.Inputs {
		prevTx := prevTxs[hex.EncodeToString(input.TxID)]
		txCopy.Inputs[i].Signature = nil
		txCopy.Inputs[i].PubKey = prevTx.Outputs[input.OutIndex].PubKeyHash
		txCopy.ID = txCopy.Hash()
		txCopy.Inputs[i].PubKey = nil

		// 公開鍵を復元
		pubKey, err := bytesToPublicKey(input.PubKey)
		if err != nil {
			return false
		}

		// 署名を検証
		if !VerifySignature(pubKey, txCopy.ID, input.Signature) {
			return false
		}
	}

	return true
}

// trimmedCopy は署名用にトリムされたトランザクションのコピーを返します
func (tx *Transaction) trimmedCopy() Transaction {
	var inputs []TxInput
	var outputs []TxOutput

	for _, input := range tx.Inputs {
		inputs = append(inputs, TxInput{
			TxID:      input.TxID,
			OutIndex:  input.OutIndex,
			Signature: nil,
			PubKey:    nil,
		})
	}

	for _, output := range tx.Outputs {
		outputs = append(outputs, TxOutput{
			Value:      output.Value,
			PubKeyHash: output.PubKeyHash,
		})
	}

	return Transaction{
		ID:        tx.ID,
		Inputs:    inputs,
		Outputs:   outputs,
		Timestamp: tx.Timestamp,
	}
}

// String はトランザクションの文字列表現を返します
func (tx *Transaction) String() string {
	var lines []string

	lines = append(lines, fmt.Sprintf("Transaction %s:", hex.EncodeToString(tx.ID)))
	lines = append(lines, fmt.Sprintf("  Timestamp: %s", time.Unix(tx.Timestamp, 0).Format("2006-01-02 15:04:05")))

	if tx.IsCoinbase() {
		lines = append(lines, "  Type: Coinbase (Mining Reward)")
	}

	lines = append(lines, fmt.Sprintf("  Inputs: %d", len(tx.Inputs)))
	for i, input := range tx.Inputs {
		if tx.IsCoinbase() {
			lines = append(lines, fmt.Sprintf("    [%d] Coinbase data: %s", i, string(input.PubKey)))
		} else {
			lines = append(lines, fmt.Sprintf("    [%d] TxID: %s, OutIndex: %d", i, hex.EncodeToString(input.TxID), input.OutIndex))
		}
	}

	lines = append(lines, fmt.Sprintf("  Outputs: %d", len(tx.Outputs)))
	for i, output := range tx.Outputs {
		lines = append(lines, fmt.Sprintf("    [%d] Value: %d, To: %s", i, output.Value, hex.EncodeToString(output.PubKeyHash)))
	}

	result := ""
	for _, line := range lines {
		result += line + "\n"
	}

	return result
}

// publicKeyToBytes は公開鍵をバイト列に変換します
func publicKeyToBytes(pubKey *ecdsa.PublicKey) []byte {
	return append(pubKey.X.Bytes(), pubKey.Y.Bytes()...)
}

// bytesToPublicKey はバイト列から公開鍵を復元します
func bytesToPublicKey(pubKeyBytes []byte) (*ecdsa.PublicKey, error) {
	if len(pubKeyBytes) == 0 {
		return nil, fmt.Errorf("empty public key")
	}

	// バイト列が不正な長さの場合
	if len(pubKeyBytes)%2 != 0 {
		return nil, fmt.Errorf("invalid public key length")
	}

	keyLen := len(pubKeyBytes) / 2
	pubKey := &ecdsa.PublicKey{
		Curve: elliptic.P256(),
		X:     new(big.Int).SetBytes(pubKeyBytes[:keyLen]),
		Y:     new(big.Int).SetBytes(pubKeyBytes[keyLen:]),
	}

	return pubKey, nil
}
