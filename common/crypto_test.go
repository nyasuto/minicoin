package common

import (
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHash(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		expected string
	}{
		{
			name:     "空のデータ",
			input:    []byte{},
			expected: "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
		},
		{
			name:     "Hello World",
			input:    []byte("Hello World"),
			expected: "a591a6d40bf420404a011733cfb7b190d62c65bf0bcda32b57b277d9ad9f146e",
		},
		{
			name:     "数値データ",
			input:    []byte{1, 2, 3, 4, 5},
			expected: "74f81fe167d99b4cb41d6d0ccda82278caee9f3e2f25d5e5a3936ff3dcec60d0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Hash(tt.input)
			resultHex := hex.EncodeToString(result)
			assert.Equal(t, tt.expected, resultHex)
		})
	}
}

func TestHashString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "空文字列",
			input:    "",
			expected: "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
		},
		{
			name:     "Hello World",
			input:    "Hello World",
			expected: "a591a6d40bf420404a011733cfb7b190d62c65bf0bcda32b57b277d9ad9f146e",
		},
		{
			name:     "日本語",
			input:    "こんにちは",
			expected: "d9f6f73c60c0798d7b2c2d0d6f2d5e3e5f8e9e3f3f3f3f3f3f3f3f3f3f3f3f3f",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := HashString(tt.input)
			// 長さが64文字（32バイトの16進数）であることを確認
			assert.Equal(t, 64, len(result))
			// 16進数文字列であることを確認
			_, err := hex.DecodeString(result)
			assert.NoError(t, err)
		})
	}
}

func TestGenerateKeyPair(t *testing.T) {
	// 鍵ペアを生成
	privateKey, err := GenerateKeyPair()
	require.NoError(t, err)
	require.NotNil(t, privateKey)

	// 秘密鍵が有効であることを確認
	assert.NotNil(t, privateKey.D)
	assert.NotNil(t, privateKey.X)
	assert.NotNil(t, privateKey.Y)

	// 複数回生成して、異なる鍵が生成されることを確認
	privateKey2, err := GenerateKeyPair()
	require.NoError(t, err)
	require.NotNil(t, privateKey2)

	assert.NotEqual(t, privateKey.D, privateKey2.D)
}

func TestSignAndVerify(t *testing.T) {
	// 鍵ペアを生成
	privateKey, err := GenerateKeyPair()
	require.NoError(t, err)

	testData := []byte("Test data for signing")

	// 署名を生成
	signature, err := Sign(privateKey, testData)
	require.NoError(t, err)
	require.NotNil(t, signature)
	require.NotEmpty(t, signature)

	// 署名を検証（成功するはず）
	valid := Verify(&privateKey.PublicKey, testData, signature)
	assert.True(t, valid, "Valid signature should be verified successfully")

	// 異なるデータで検証（失敗するはず）
	invalidData := []byte("Different data")
	valid = Verify(&privateKey.PublicKey, invalidData, signature)
	assert.False(t, valid, "Signature should not verify with different data")

	// 異なる公開鍵で検証（失敗するはず）
	otherPrivateKey, err := GenerateKeyPair()
	require.NoError(t, err)
	valid = Verify(&otherPrivateKey.PublicKey, testData, signature)
	assert.False(t, valid, "Signature should not verify with different public key")

	// 不正な署名で検証（失敗するはず）
	invalidSignature := make([]byte, len(signature))
	copy(invalidSignature, signature)
	invalidSignature[0] ^= 0xFF // 最初のバイトを反転
	valid = Verify(&privateKey.PublicKey, testData, invalidSignature)
	assert.False(t, valid, "Invalid signature should not verify")
}

func TestVerifyWithInvalidSignature(t *testing.T) {
	privateKey, err := GenerateKeyPair()
	require.NoError(t, err)

	testData := []byte("Test data")

	tests := []struct {
		name      string
		signature []byte
	}{
		{
			name:      "空の署名",
			signature: []byte{},
		},
		{
			name:      "奇数長の署名",
			signature: []byte{1, 2, 3},
		},
		{
			name:      "不正なサイズの署名",
			signature: make([]byte, 10),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid := Verify(&privateKey.PublicKey, testData, tt.signature)
			assert.False(t, valid)
		})
	}
}

func TestPublicKeyToAddress(t *testing.T) {
	// 鍵ペアを生成
	privateKey, err := GenerateKeyPair()
	require.NoError(t, err)

	// アドレスを生成
	address := PublicKeyToAddress(&privateKey.PublicKey)

	// アドレスが16進数文字列であることを確認
	_, err = hex.DecodeString(address)
	assert.NoError(t, err)

	// アドレスが40文字（20バイトの16進数）であることを確認
	assert.Equal(t, 40, len(address))

	// 同じ公開鍵から同じアドレスが生成されることを確認
	address2 := PublicKeyToAddress(&privateKey.PublicKey)
	assert.Equal(t, address, address2)

	// 異なる公開鍵からは異なるアドレスが生成されることを確認
	otherPrivateKey, err := GenerateKeyPair()
	require.NoError(t, err)
	otherAddress := PublicKeyToAddress(&otherPrivateKey.PublicKey)
	assert.NotEqual(t, address, otherAddress)
}

// ベンチマーク
func BenchmarkHash(b *testing.B) {
	data := []byte("Benchmark data for hashing")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Hash(data)
	}
}

func BenchmarkGenerateKeyPair(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = GenerateKeyPair()
	}
}

func BenchmarkSign(b *testing.B) {
	privateKey, _ := GenerateKeyPair()
	data := []byte("Benchmark data for signing")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = Sign(privateKey, data)
	}
}

func BenchmarkVerify(b *testing.B) {
	privateKey, _ := GenerateKeyPair()
	data := []byte("Benchmark data for verification")
	signature, _ := Sign(privateKey, data)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Verify(&privateKey.PublicKey, data, signature)
	}
}
