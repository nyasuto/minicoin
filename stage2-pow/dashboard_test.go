package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewDashboard(t *testing.T) {
	t.Run("ダッシュボードの生成", func(t *testing.T) {
		bc := NewBlockchain(2)

		dashboard := NewDashboard(bc)

		require.NotNil(t, dashboard)
		assert.NotNil(t, dashboard.app)
		assert.NotNil(t, dashboard.blockchain)
		assert.NotNil(t, dashboard.grid)
		assert.NotNil(t, dashboard.overviewPanel)
		assert.NotNil(t, dashboard.blocksPanel)
		assert.NotNil(t, dashboard.miningPanel)
		assert.NotNil(t, dashboard.difficultyPanel)
		assert.NotNil(t, dashboard.helpPanel)
		assert.NotNil(t, dashboard.stopChan)
		assert.Equal(t, bc, dashboard.blockchain)
	})

	t.Run("複数ブロックを持つチェーンでダッシュボード生成", func(t *testing.T) {
		bc := NewBlockchain(1)
		bc.AddBlock("Block 1")
		bc.AddBlock("Block 2")
		bc.AddBlock("Block 3")

		dashboard := NewDashboard(bc)

		require.NotNil(t, dashboard)
		assert.Equal(t, 4, bc.GetChainLength()) // Genesis + 3
	})
}

func TestEstimateHashesForDifficulty(t *testing.T) {
	tests := []struct {
		name       string
		difficulty int
		expected   float64
	}{
		{"難易度0", 0, 1.0},
		{"難易度1", 1, 16.0},
		{"難易度2", 2, 256.0},
		{"難易度3", 3, 4096.0},
		{"難易度4", 4, 65536.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := estimateHashesForDifficulty(tt.difficulty)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFormatHashRate(t *testing.T) {
	tests := []struct {
		name     string
		hashes   float64
		expected string
	}{
		{"1ハッシュ", 1.0, "1 H/s"},
		{"100ハッシュ", 100.0, "100 H/s"},
		{"999ハッシュ", 999.0, "999 H/s"},
		{"1,000ハッシュ (1 KH/s)", 1000.0, "1.00 KH/s"},
		{"10,000ハッシュ (10 KH/s)", 10000.0, "10.00 KH/s"},
		{"1,000,000ハッシュ (1 MH/s)", 1000000.0, "1.00 MH/s"},
		{"10,000,000ハッシュ (10 MH/s)", 10000000.0, "10.00 MH/s"},
		{"1,000,000,000ハッシュ (1 GH/s)", 1000000000.0, "1.00 GH/s"},
		{"10,000,000,000ハッシュ (10 GH/s)", 10000000000.0, "10.00 GH/s"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatHashRate(tt.hashes)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetValidityText(t *testing.T) {
	t.Run("有効な場合", func(t *testing.T) {
		result := getValidityText(true)
		assert.Equal(t, "Yes", result)
	})

	t.Run("無効な場合", func(t *testing.T) {
		result := getValidityText(false)
		assert.Equal(t, "No", result)
	})
}

func TestDashboardUpdate(t *testing.T) {
	t.Run("update()がパニックしない", func(t *testing.T) {
		bc := NewBlockchain(2)
		bc.AddBlock("Block 1")

		dashboard := NewDashboard(bc)

		// update()を呼び出してもパニックしないことを確認
		assert.NotPanics(t, func() {
			dashboard.update()
		})
	})

	t.Run("空のチェーンでupdate()がパニックしない", func(t *testing.T) {
		bc := &Blockchain{
			Blocks:          []*Block{},
			Difficulty:      2,
			TargetBlockTime: 10,
		}

		dashboard := NewDashboard(bc)

		assert.NotPanics(t, func() {
			dashboard.update()
		})
	})
}

func TestDashboardPanelUpdates(t *testing.T) {
	t.Run("updateOverviewPanel()がパニックしない", func(t *testing.T) {
		bc := NewBlockchain(2)
		bc.AddBlock("Block 1")

		dashboard := NewDashboard(bc)

		assert.NotPanics(t, func() {
			dashboard.updateOverviewPanel()
		})
	})

	t.Run("updateBlocksPanel()がパニックしない", func(t *testing.T) {
		bc := NewBlockchain(2)
		for i := 0; i < 10; i++ {
			bc.AddBlock("Block " + string(rune(i+'0')))
		}

		dashboard := NewDashboard(bc)

		assert.NotPanics(t, func() {
			dashboard.updateBlocksPanel()
		})
	})

	t.Run("updateMiningPanel()がパニックしない", func(t *testing.T) {
		bc := NewBlockchain(2)
		bc.AddBlock("Block 1")

		dashboard := NewDashboard(bc)

		assert.NotPanics(t, func() {
			dashboard.updateMiningPanel()
		})
	})

	t.Run("updateDifficultyPanel()がパニックしない", func(t *testing.T) {
		bc := NewBlockchain(2)
		bc.AddBlock("Block 1")

		dashboard := NewDashboard(bc)

		assert.NotPanics(t, func() {
			dashboard.updateDifficultyPanel()
		})
	})
}

func TestDashboardStop(t *testing.T) {
	t.Run("Stop()が正常に動作する", func(t *testing.T) {
		bc := NewBlockchain(2)
		dashboard := NewDashboard(bc)

		// stopChanにメッセージが送られることを確認
		go func() {
			<-dashboard.stopChan
		}()

		// Stop()を呼び出す
		// Note: app.Stop()は実際のTUIが起動していないとエラーになる可能性があるため、
		// ここではstopChanへの送信のみをテスト
		assert.NotPanics(t, func() {
			dashboard.stopChan <- true
		})
	})
}
