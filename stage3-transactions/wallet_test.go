package main

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewWallet(t *testing.T) {
	t.Run("ウォレットの生成", func(t *testing.T) {
		wallet, err := NewWallet()

		require.NoError(t, err)
		require.NotNil(t, wallet)
		assert.NotNil(t, wallet.PrivateKey)
		assert.NotNil(t, wallet.PublicKey)
		assert.NotEmpty(t, wallet.Address)
	})

	t.Run("複数ウォレットのアドレス一意性", func(t *testing.T) {
		wallet1, err := NewWallet()
		require.NoError(t, err)

		wallet2, err := NewWallet()
		require.NoError(t, err)

		// アドレスが異なることを確認
		assert.NotEqual(t, wallet1.Address, wallet2.Address)
	})
}

func TestWalletGetAddress(t *testing.T) {
	t.Run("アドレスの取得", func(t *testing.T) {
		wallet, err := NewWallet()
		require.NoError(t, err)

		address := wallet.GetAddress()

		assert.NotEmpty(t, address)
		assert.Equal(t, wallet.Address, address)
	})
}

func TestWalletSign(t *testing.T) {
	t.Run("データへの署名", func(t *testing.T) {
		wallet, err := NewWallet()
		require.NoError(t, err)

		data := []byte("test data")
		signature, err := wallet.Sign(data)

		require.NoError(t, err)
		assert.NotNil(t, signature)
		assert.NotEmpty(t, signature)
	})
}

func TestVerifySignature(t *testing.T) {
	t.Run("正しい署名の検証", func(t *testing.T) {
		wallet, err := NewWallet()
		require.NoError(t, err)

		data := []byte("test data")
		signature, err := wallet.Sign(data)
		require.NoError(t, err)

		// 署名を検証
		valid := VerifySignature(wallet.PublicKey, data, signature)

		assert.True(t, valid)
	})

	t.Run("異なるデータでの署名検証失敗", func(t *testing.T) {
		wallet, err := NewWallet()
		require.NoError(t, err)

		data := []byte("test data")
		signature, err := wallet.Sign(data)
		require.NoError(t, err)

		// 異なるデータで検証
		differentData := []byte("different data")
		valid := VerifySignature(wallet.PublicKey, differentData, signature)

		assert.False(t, valid)
	})

	t.Run("異なる公開鍵での署名検証失敗", func(t *testing.T) {
		wallet1, err := NewWallet()
		require.NoError(t, err)

		wallet2, err := NewWallet()
		require.NoError(t, err)

		data := []byte("test data")
		signature, err := wallet1.Sign(data)
		require.NoError(t, err)

		// wallet2の公開鍵で検証
		valid := VerifySignature(wallet2.PublicKey, data, signature)

		assert.False(t, valid)
	})

	t.Run("不正な署名での検証失敗", func(t *testing.T) {
		wallet, err := NewWallet()
		require.NoError(t, err)

		data := []byte("test data")
		invalidSignature := []byte("invalid signature")

		valid := VerifySignature(wallet.PublicKey, data, invalidSignature)

		assert.False(t, valid)
	})
}

func TestWalletSaveAndLoad(t *testing.T) {
	t.Run("ウォレットの保存と読み込み", func(t *testing.T) {
		wallet, err := NewWallet()
		require.NoError(t, err)

		filename := "/tmp/test_wallet.dat"
		defer func() { _ = os.Remove(filename) }()

		// 保存
		err = wallet.SaveToFile(filename)
		require.NoError(t, err)

		// 読み込み
		loadedWallet, err := LoadWalletFromFile(filename)
		require.NoError(t, err)

		// 検証
		assert.Equal(t, wallet.Address, loadedWallet.Address)
		assert.Equal(t, wallet.PrivateKey.D.Bytes(), loadedWallet.PrivateKey.D.Bytes())
		assert.Equal(t, wallet.PublicKey.X.Bytes(), loadedWallet.PublicKey.X.Bytes())
		assert.Equal(t, wallet.PublicKey.Y.Bytes(), loadedWallet.PublicKey.Y.Bytes())
	})

	t.Run("署名の互換性確認", func(t *testing.T) {
		wallet, err := NewWallet()
		require.NoError(t, err)

		filename := "/tmp/test_wallet_sign.dat"
		defer func() { _ = os.Remove(filename) }()

		// データに署名
		data := []byte("test data")
		signature, err := wallet.Sign(data)
		require.NoError(t, err)

		// 保存して読み込み
		err = wallet.SaveToFile(filename)
		require.NoError(t, err)

		loadedWallet, err := LoadWalletFromFile(filename)
		require.NoError(t, err)

		// 読み込んだウォレットで署名を検証
		valid := VerifySignature(loadedWallet.PublicKey, data, signature)
		assert.True(t, valid)
	})

	t.Run("存在しないファイルからの読み込みエラー", func(t *testing.T) {
		_, err := LoadWalletFromFile("/tmp/nonexistent_wallet.dat")

		assert.Error(t, err)
	})
}

func TestNewWallets(t *testing.T) {
	t.Run("ウォレットコレクションの生成", func(t *testing.T) {
		wallets := NewWallets()

		require.NotNil(t, wallets)
		assert.NotNil(t, wallets.Wallets)
		assert.Equal(t, 0, len(wallets.Wallets))
	})
}

func TestWalletsCreateWallet(t *testing.T) {
	t.Run("新規ウォレットの作成", func(t *testing.T) {
		wallets := NewWallets()

		address, err := wallets.CreateWallet()

		require.NoError(t, err)
		assert.NotEmpty(t, address)
		assert.Equal(t, 1, len(wallets.Wallets))
		assert.Contains(t, wallets.Wallets, address)
	})

	t.Run("複数ウォレットの作成", func(t *testing.T) {
		wallets := NewWallets()

		address1, err := wallets.CreateWallet()
		require.NoError(t, err)

		address2, err := wallets.CreateWallet()
		require.NoError(t, err)

		address3, err := wallets.CreateWallet()
		require.NoError(t, err)

		assert.Equal(t, 3, len(wallets.Wallets))
		assert.NotEqual(t, address1, address2)
		assert.NotEqual(t, address2, address3)
		assert.NotEqual(t, address1, address3)
	})
}

func TestWalletsGetWallet(t *testing.T) {
	t.Run("ウォレットの取得", func(t *testing.T) {
		wallets := NewWallets()

		address, err := wallets.CreateWallet()
		require.NoError(t, err)

		wallet, err := wallets.GetWallet(address)

		require.NoError(t, err)
		assert.NotNil(t, wallet)
		assert.Equal(t, address, wallet.Address)
	})

	t.Run("存在しないウォレットの取得エラー", func(t *testing.T) {
		wallets := NewWallets()

		_, err := wallets.GetWallet("nonexistent_address")

		assert.Error(t, err)
	})
}

func TestWalletsGetAddresses(t *testing.T) {
	t.Run("全アドレスの取得", func(t *testing.T) {
		wallets := NewWallets()

		address1, _ := wallets.CreateWallet()
		address2, _ := wallets.CreateWallet()
		address3, _ := wallets.CreateWallet()

		addresses := wallets.GetAddresses()

		assert.Equal(t, 3, len(addresses))
		assert.Contains(t, addresses, address1)
		assert.Contains(t, addresses, address2)
		assert.Contains(t, addresses, address3)
	})

	t.Run("空のコレクションからのアドレス取得", func(t *testing.T) {
		wallets := NewWallets()

		addresses := wallets.GetAddresses()

		assert.Equal(t, 0, len(addresses))
	})
}

func TestWalletsSaveAndLoad(t *testing.T) {
	t.Run("ウォレットコレクションの保存と読み込み", func(t *testing.T) {
		wallets := NewWallets()

		address1, _ := wallets.CreateWallet()
		address2, _ := wallets.CreateWallet()
		address3, _ := wallets.CreateWallet()

		filename := "/tmp/test_wallets.dat"
		defer func() { _ = os.Remove(filename) }()

		// 保存
		err := wallets.SaveToFile(filename)
		require.NoError(t, err)

		// 読み込み
		loadedWallets, err := LoadWalletsFromFile(filename)
		require.NoError(t, err)

		// 検証
		assert.Equal(t, 3, len(loadedWallets.Wallets))
		assert.Contains(t, loadedWallets.Wallets, address1)
		assert.Contains(t, loadedWallets.Wallets, address2)
		assert.Contains(t, loadedWallets.Wallets, address3)
	})

	t.Run("署名の互換性確認", func(t *testing.T) {
		wallets := NewWallets()

		address, _ := wallets.CreateWallet()
		wallet, _ := wallets.GetWallet(address)

		data := []byte("test data")
		signature, err := wallet.Sign(data)
		require.NoError(t, err)

		filename := "/tmp/test_wallets_sign.dat"
		defer func() { _ = os.Remove(filename) }()

		// 保存して読み込み
		err = wallets.SaveToFile(filename)
		require.NoError(t, err)

		loadedWallets, err := LoadWalletsFromFile(filename)
		require.NoError(t, err)

		loadedWallet, err := loadedWallets.GetWallet(address)
		require.NoError(t, err)

		// 署名を検証
		valid := VerifySignature(loadedWallet.PublicKey, data, signature)
		assert.True(t, valid)
	})

	t.Run("存在しないファイルからの読み込み（空のコレクション）", func(t *testing.T) {
		loadedWallets, err := LoadWalletsFromFile("/tmp/nonexistent_wallets.dat")

		require.NoError(t, err)
		assert.NotNil(t, loadedWallets)
		assert.Equal(t, 0, len(loadedWallets.Wallets))
	})
}
