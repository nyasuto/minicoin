package common

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"regexp"
	"time"
)

// IntToHex は整数を16進数のバイト列に変換します（ビッグエンディアン）
func IntToHex(num int64) []byte {
	buff := new(bytes.Buffer)
	err := binary.Write(buff, binary.BigEndian, num)
	if err != nil {
		// このエラーは通常発生しないはずですが、安全のため
		return []byte{}
	}
	return buff.Bytes()
}

// HexToInt は16進数のバイト列を整数に変換します（ビッグエンディアン）
func HexToInt(hexBytes []byte) int64 {
	var num int64
	buff := bytes.NewReader(hexBytes)
	err := binary.Read(buff, binary.BigEndian, &num)
	if err != nil {
		return 0
	}
	return num
}

// FormatTimestamp はUnixタイムスタンプを人間が読みやすい形式にフォーマットします（UTC）
func FormatTimestamp(timestamp int64) string {
	t := time.Unix(timestamp, 0).UTC()
	return t.Format("2006-01-02 15:04:05")
}

// ValidateHash はハッシュ値が有効な16進数文字列であることを検証します
// SHA-256ハッシュは64文字（32バイト）である必要があります
func ValidateHash(hash string) bool {
	// 長さチェック（SHA-256は64文字）
	if len(hash) != 64 {
		return false
	}

	// 16進数文字列であることを確認
	matched, err := regexp.MatchString("^[a-fA-F0-9]{64}$", hash)
	if err != nil {
		return false
	}

	return matched
}

// MerkleRoot はハッシュのリストからマークルルートを計算します
// 将来的にマークルツリー実装で使用するための基本的な実装
func MerkleRoot(hashes [][]byte) []byte {
	if len(hashes) == 0 {
		return Hash([]byte{})
	}

	if len(hashes) == 1 {
		return hashes[0]
	}

	// ハッシュのペアを結合してハッシュ化
	var newHashes [][]byte

	for i := 0; i < len(hashes); i += 2 {
		if i+1 < len(hashes) {
			// ペアがある場合
			combined := append(hashes[i], hashes[i+1]...)
			newHashes = append(newHashes, Hash(combined))
		} else {
			// 奇数個の場合、最後のハッシュを自分自身と結合
			combined := append(hashes[i], hashes[i]...)
			newHashes = append(newHashes, Hash(combined))
		}
	}

	// 再帰的にマークルルートを計算
	return MerkleRoot(newHashes)
}

// BytesToHex はバイト列を16進数文字列に変換します
func BytesToHex(data []byte) string {
	return hex.EncodeToString(data)
}

// HexToBytes は16進数文字列をバイト列に変換します
func HexToBytes(hexStr string) ([]byte, error) {
	data, err := hex.DecodeString(hexStr)
	if err != nil {
		return nil, fmt.Errorf("invalid hex string: %w", err)
	}
	return data, nil
}

// ReverseBytes はバイト列を逆順にします
func ReverseBytes(data []byte) []byte {
	reversed := make([]byte, len(data))
	for i := 0; i < len(data); i++ {
		reversed[i] = data[len(data)-1-i]
	}
	return reversed
}
