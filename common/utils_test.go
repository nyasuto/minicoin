package common

import (
	"encoding/hex"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIntToHex(t *testing.T) {
	tests := []struct {
		name     string
		input    int64
		expected string
	}{
		{
			name:     "ゼロ",
			input:    0,
			expected: "0000000000000000",
		},
		{
			name:     "正の数",
			input:    12345,
			expected: "0000000000003039",
		},
		{
			name:     "負の数",
			input:    -1,
			expected: "ffffffffffffffff",
		},
		{
			name:     "大きな数",
			input:    1000000,
			expected: "00000000000f4240",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IntToHex(tt.input)
			resultHex := hex.EncodeToString(result)
			assert.Equal(t, tt.expected, resultHex)
		})
	}
}

func TestHexToInt(t *testing.T) {
	tests := []struct {
		name     string
		inputHex string
		expected int64
	}{
		{
			name:     "ゼロ",
			inputHex: "0000000000000000",
			expected: 0,
		},
		{
			name:     "正の数",
			inputHex: "0000000000003039",
			expected: 12345,
		},
		{
			name:     "負の数",
			inputHex: "ffffffffffffffff",
			expected: -1,
		},
		{
			name:     "大きな数",
			inputHex: "00000000000f4240",
			expected: 1000000,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input, err := hex.DecodeString(tt.inputHex)
			require.NoError(t, err)
			result := HexToInt(input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIntToHexAndHexToInt(t *testing.T) {
	// 変換の往復テスト
	testValues := []int64{0, 1, -1, 100, -100, 12345, -12345, 1000000}

	for _, val := range testValues {
		hexBytes := IntToHex(val)
		result := HexToInt(hexBytes)
		assert.Equal(t, val, result, "往復変換で元の値に戻るべき")
	}
}

func TestFormatTimestamp(t *testing.T) {
	tests := []struct {
		name      string
		timestamp int64
		expected  string
	}{
		{
			name:      "Unix epoch",
			timestamp: 0,
			expected:  "1970-01-01 00:00:00",
		},
		{
			name:      "特定の日時",
			timestamp: 1609459200, // 2021-01-01 00:00:00 UTC
			expected:  "2021-01-01 00:00:00",
		},
		{
			name:      "現在に近い時刻",
			timestamp: time.Now().Unix(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatTimestamp(tt.timestamp)
			if tt.expected != "" {
				assert.Equal(t, tt.expected, result)
			} else {
				// フォーマットが正しいことだけ確認
				assert.Len(t, result, 19) // "YYYY-MM-DD HH:MM:SS" = 19文字
			}
		})
	}
}

func TestValidateHash(t *testing.T) {
	tests := []struct {
		name     string
		hash     string
		expected bool
	}{
		{
			name:     "有効なハッシュ",
			hash:     "a591a6d40bf420404a011733cfb7b190d62c65bf0bcda32b57b277d9ad9f146e",
			expected: true,
		},
		{
			name:     "有効なハッシュ（大文字）",
			hash:     "A591A6D40BF420404A011733CFB7B190D62C65BF0BCDA32B57B277D9AD9F146E",
			expected: true,
		},
		{
			name:     "有効なハッシュ（混在）",
			hash:     "A591a6D40BF420404a011733cfb7b190D62C65BF0BCDA32B57b277d9ad9f146E",
			expected: true,
		},
		{
			name:     "短すぎるハッシュ",
			hash:     "a591a6d40bf420404a011733cfb7b190",
			expected: false,
		},
		{
			name:     "長すぎるハッシュ",
			hash:     "a591a6d40bf420404a011733cfb7b190d62c65bf0bcda32b57b277d9ad9f146e00",
			expected: false,
		},
		{
			name:     "無効な文字を含む",
			hash:     "g591a6d40bf420404a011733cfb7b190d62c65bf0bcda32b57b277d9ad9f146e",
			expected: false,
		},
		{
			name:     "空文字列",
			hash:     "",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ValidateHash(tt.hash)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestMerkleRoot(t *testing.T) {
	t.Run("空のリスト", func(t *testing.T) {
		result := MerkleRoot([][]byte{})
		assert.NotNil(t, result)
		assert.NotEmpty(t, result)
	})

	t.Run("単一のハッシュ", func(t *testing.T) {
		hash1 := Hash([]byte("data1"))
		result := MerkleRoot([][]byte{hash1})
		assert.Equal(t, hash1, result)
	})

	t.Run("2つのハッシュ", func(t *testing.T) {
		hash1 := Hash([]byte("data1"))
		hash2 := Hash([]byte("data2"))
		result := MerkleRoot([][]byte{hash1, hash2})

		// 期待される結果を手動で計算
		combined := append(hash1, hash2...)
		expected := Hash(combined)

		assert.Equal(t, expected, result)
	})

	t.Run("3つのハッシュ（奇数）", func(t *testing.T) {
		hash1 := Hash([]byte("data1"))
		hash2 := Hash([]byte("data2"))
		hash3 := Hash([]byte("data3"))

		result := MerkleRoot([][]byte{hash1, hash2, hash3})
		assert.NotNil(t, result)
		assert.Len(t, result, 32) // SHA-256は32バイト
	})

	t.Run("4つのハッシュ", func(t *testing.T) {
		hash1 := Hash([]byte("data1"))
		hash2 := Hash([]byte("data2"))
		hash3 := Hash([]byte("data3"))
		hash4 := Hash([]byte("data4"))

		result := MerkleRoot([][]byte{hash1, hash2, hash3, hash4})
		assert.NotNil(t, result)
		assert.Len(t, result, 32)
	})

	t.Run("同じ入力で同じ結果", func(t *testing.T) {
		hash1 := Hash([]byte("data1"))
		hash2 := Hash([]byte("data2"))

		result1 := MerkleRoot([][]byte{hash1, hash2})
		result2 := MerkleRoot([][]byte{hash1, hash2})

		assert.Equal(t, result1, result2)
	})

	t.Run("異なる順序で異なる結果", func(t *testing.T) {
		hash1 := Hash([]byte("data1"))
		hash2 := Hash([]byte("data2"))

		result1 := MerkleRoot([][]byte{hash1, hash2})
		result2 := MerkleRoot([][]byte{hash2, hash1})

		assert.NotEqual(t, result1, result2)
	})
}

func TestBytesToHex(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		expected string
	}{
		{
			name:     "空のバイト列",
			input:    []byte{},
			expected: "",
		},
		{
			name:     "単一バイト",
			input:    []byte{0xFF},
			expected: "ff",
		},
		{
			name:     "複数バイト",
			input:    []byte{0x01, 0x02, 0x03},
			expected: "010203",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := BytesToHex(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestHexToBytes(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expected    []byte
		expectError bool
	}{
		{
			name:        "空文字列",
			input:       "",
			expected:    []byte{},
			expectError: false,
		},
		{
			name:        "有効な16進数",
			input:       "010203",
			expected:    []byte{0x01, 0x02, 0x03},
			expectError: false,
		},
		{
			name:        "無効な16進数",
			input:       "xyz",
			expected:    nil,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := HexToBytes(tt.input)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestReverseBytes(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		expected []byte
	}{
		{
			name:     "空のバイト列",
			input:    []byte{},
			expected: []byte{},
		},
		{
			name:     "単一バイト",
			input:    []byte{0x01},
			expected: []byte{0x01},
		},
		{
			name:     "複数バイト",
			input:    []byte{0x01, 0x02, 0x03},
			expected: []byte{0x03, 0x02, 0x01},
		},
		{
			name:     "偶数個",
			input:    []byte{0x01, 0x02, 0x03, 0x04},
			expected: []byte{0x04, 0x03, 0x02, 0x01},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ReverseBytes(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// ベンチマーク
func BenchmarkIntToHex(b *testing.B) {
	for i := 0; i < b.N; i++ {
		IntToHex(12345)
	}
}

func BenchmarkHexToInt(b *testing.B) {
	hexBytes := IntToHex(12345)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		HexToInt(hexBytes)
	}
}

func BenchmarkMerkleRoot(b *testing.B) {
	hashes := make([][]byte, 100)
	for i := 0; i < 100; i++ {
		hashes[i] = Hash([]byte{byte(i)})
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		MerkleRoot(hashes)
	}
}
