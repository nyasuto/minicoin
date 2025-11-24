// Package main implements a TUI dashboard for the blockchain.
// This includes real-time monitoring and visualization.
package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/nyasuto/minicoin/common"
	"github.com/rivo/tview"
)

// Dashboard はターミナルUIダッシュボード
type Dashboard struct {
	app        *tview.Application
	blockchain *Blockchain
	grid       *tview.Grid

	// パネル
	overviewPanel   *tview.TextView
	blocksPanel     *tview.TextView
	miningPanel     *tview.TextView
	difficultyPanel *tview.TextView
	helpPanel       *tview.TextView

	// 更新制御
	updateInterval time.Duration
	stopChan       chan bool

	// マイニング制御
	isMining       bool
	miningStopChan chan bool
	miningCounter  int
}

// NewDashboard は新しいダッシュボードを作成します
func NewDashboard(bc *Blockchain) *Dashboard {
	app := tview.NewApplication()

	d := &Dashboard{
		app:            app,
		blockchain:     bc,
		updateInterval: 1 * time.Second,
		stopChan:       make(chan bool),
		miningStopChan: make(chan bool),
		isMining:       false,
	}

	// パネルの作成
	d.overviewPanel = d.createPanel("Blockchain Overview")
	d.blocksPanel = d.createPanel("Latest Blocks")
	d.miningPanel = d.createPanel("Mining Stats")
	d.difficultyPanel = d.createPanel("Difficulty Adjustment")
	d.helpPanel = d.createHelpPanel()

	// グリッドレイアウトの作成
	d.grid = tview.NewGrid().
		SetRows(8, 10, 8, 8, 3).
		SetColumns(0).
		SetBorders(false)

	// パネルの配置
	d.grid.AddItem(d.overviewPanel, 0, 0, 1, 1, 0, 0, false)
	d.grid.AddItem(d.blocksPanel, 1, 0, 1, 1, 0, 0, false)
	d.grid.AddItem(d.miningPanel, 2, 0, 1, 1, 0, 0, false)
	d.grid.AddItem(d.difficultyPanel, 3, 0, 1, 1, 0, 0, false)
	d.grid.AddItem(d.helpPanel, 4, 0, 1, 1, 0, 0, false)

	// キーボード入力処理
	d.grid.SetInputCapture(d.handleKeyPress)

	d.app.SetRoot(d.grid, true)

	return d
}

// createPanel は基本的なパネルを作成します
func (d *Dashboard) createPanel(title string) *tview.TextView {
	panel := tview.NewTextView().
		SetDynamicColors(true).
		SetScrollable(false)

	panel.SetBorder(true).
		SetTitle(fmt.Sprintf(" %s ", title)).
		SetTitleAlign(tview.AlignLeft).
		SetBorderPadding(0, 0, 1, 1)

	return panel
}

// createHelpPanel はヘルプパネルを作成します
func (d *Dashboard) createHelpPanel() *tview.TextView {
	panel := tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignCenter).
		SetText("[yellow]Keys:[white] [green]q[white] Quit | [green]r[white] Refresh | [green]m[white] Mining Start/Stop | [green]Ctrl+C[white] Exit")

	panel.SetBorder(false)

	return panel
}

// handleKeyPress はキーボード入力を処理します
func (d *Dashboard) handleKeyPress(event *tcell.EventKey) *tcell.EventKey {
	switch event.Rune() {
	case 'q', 'Q':
		d.Stop()
		return nil
	case 'r', 'R':
		d.update()
		return nil
	case 'm', 'M':
		d.toggleMining()
		return nil
	}

	// Ctrl+Cの処理
	if event.Key() == tcell.KeyCtrlC {
		d.Stop()
		return nil
	}

	return event
}

// Run はダッシュボードを起動します
func (d *Dashboard) Run() error {
	// 初期更新
	d.update()

	// 自動更新ゴルーチンを開始
	go d.autoUpdate()

	// アプリケーションを実行
	return d.app.Run()
}

// Stop はダッシュボードを停止します
func (d *Dashboard) Stop() {
	if d.isMining {
		d.stopMining()
	}
	d.stopChan <- true
	d.app.Stop()
}

// autoUpdate は定期的にダッシュボードを更新します
func (d *Dashboard) autoUpdate() {
	ticker := time.NewTicker(d.updateInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			d.app.QueueUpdateDraw(func() {
				d.update()
			})
		case <-d.stopChan:
			return
		}
	}
}

// update は全パネルを更新します
func (d *Dashboard) update() {
	d.updateOverviewPanel()
	d.updateBlocksPanel()
	d.updateMiningPanel()
	d.updateDifficultyPanel()
}

// updateOverviewPanel はチェーン概要パネルを更新します
func (d *Dashboard) updateOverviewPanel() {
	bc := d.blockchain
	bc.mutex.RLock()
	defer bc.mutex.RUnlock()

	totalBlocks := len(bc.Blocks)
	currentDifficulty := bc.Difficulty
	isValid := bc.IsValid()
	validIcon := "✗"
	validColor := "red"
	if isValid {
		validIcon = "✓"
		validColor = "green"
	}

	var lastBlockTime string
	if totalBlocks > 0 {
		lastBlock := bc.Blocks[totalBlocks-1]
		lastBlockTime = common.FormatTimestamp(lastBlock.Timestamp)
	} else {
		lastBlockTime = "N/A"
	}

	content := fmt.Sprintf(
		"[white]Total Blocks:      [cyan]%d[white]\n"+
			"Current Difficulty: [yellow]%d[white]\n"+
			"Chain Valid:        [%s]%s %s[white]\n"+
			"Last Block Time:    [cyan]%s[white]",
		totalBlocks,
		currentDifficulty,
		validColor, validIcon, getValidityText(isValid),
		lastBlockTime,
	)

	d.overviewPanel.SetText(content)
}

// getValidityText は有効性のテキストを返します
func getValidityText(isValid bool) string {
	if isValid {
		return "Yes"
	}
	return "No"
}

// updateBlocksPanel は最新ブロックパネルを更新します
func (d *Dashboard) updateBlocksPanel() {
	bc := d.blockchain
	bc.mutex.RLock()
	defer bc.mutex.RUnlock()

	var lines []string

	// 最新5ブロックを表示
	numToShow := 5
	totalBlocks := len(bc.Blocks)
	startIdx := totalBlocks - numToShow
	if startIdx < 0 {
		startIdx = 0
	}

	for i := totalBlocks - 1; i >= startIdx; i-- {
		block := bc.Blocks[i]
		shortHash := block.Hash
		if len(shortHash) > 12 {
			shortHash = shortHash[:12] + "..."
		}

		timeStr := common.FormatTimestamp(block.Timestamp)
		// 時刻部分のみ抽出
		if len(timeStr) > 11 {
			timeStr = timeStr[11:]
		}

		var blockType string
		if block.Index == 0 {
			blockType = "[yellow](Genesis)[white]"
		} else {
			blockType = ""
		}

		line := fmt.Sprintf(
			"[cyan]#%-4d[white] %s Hash: [green]%s[white]  Time: [yellow]%s[white]  Nonce: [magenta]%d[white]",
			block.Index,
			blockType,
			shortHash,
			timeStr,
			block.Nonce,
		)
		lines = append(lines, line)
	}

	d.blocksPanel.SetText(strings.Join(lines, "\n"))
}

// updateMiningPanel はマイニング統計パネルを更新します
func (d *Dashboard) updateMiningPanel() {
	bc := d.blockchain
	bc.mutex.RLock()
	defer bc.mutex.RUnlock()

	totalBlocks := len(bc.Blocks)

	// 平均ブロック時間を計算
	avgBlockTime := 0.0
	if totalBlocks > 1 {
		avgBlockTime = GetAverageBlockTime(bc, 10)
	}

	// 最新ブロックのハッシュレートを推定（仮想値）
	hashRate := "N/A"
	if totalBlocks > 0 {
		// 難易度に基づいた推定ハッシュレート
		estimatedHashes := estimateHashesForDifficulty(bc.Difficulty)
		hashRate = formatHashRate(estimatedHashes)
	}

	// マイニング状態
	miningStatus := "[red]● Stopped[white]"
	miningInfo := ""
	if d.isMining {
		miningStatus = "[green]● Mining...[white]"
		miningInfo = fmt.Sprintf("\nAuto-mined:         [cyan]%d blocks[white]", d.miningCounter)
	}

	content := fmt.Sprintf(
		"[white]Mining Status:      %s"+
			"%s\n"+
			"Hash Rate (est):    [cyan]%s[white]\n"+
			"Avg Block Time:     [yellow]%.2f s[white]\n"+
			"Target Block Time:  [green]%d s[white]\n"+
			"Total Blocks:       [cyan]%d[white]",
		miningStatus,
		miningInfo,
		hashRate,
		avgBlockTime,
		bc.TargetBlockTime,
		totalBlocks,
	)

	d.miningPanel.SetText(content)
}

// updateDifficultyPanel は難易度調整パネルを更新します
func (d *Dashboard) updateDifficultyPanel() {
	bc := d.blockchain
	bc.mutex.RLock()
	defer bc.mutex.RUnlock()

	stats := GetDifficultyStatsFromChain(bc)

	// パフォーマンス評価
	var status string
	var statusColor string
	if stats.AverageBlockTime > 0 {
		ratio := stats.AverageBlockTime / float64(stats.TargetBlockTime)
		if ratio > 1.2 {
			status = "⚠  Slow"
			statusColor = "yellow"
		} else if ratio < 0.8 {
			status = "⚡ Fast"
			statusColor = "cyan"
		} else {
			status = "✓ Optimal"
			statusColor = "green"
		}
	} else {
		status = "N/A"
		statusColor = "white"
	}

	content := fmt.Sprintf(
		"[white]Current Difficulty: [yellow]%d[white]\n"+
			"Next Adjustment:    [cyan]%d blocks[white]\n"+
			"Status:             [%s]%s[white]\n"+
			"Adjustment Interval: [cyan]%d blocks[white]",
		stats.CurrentDifficulty,
		stats.NextAdjustment,
		statusColor, status,
		AdjustmentInterval,
	)

	d.difficultyPanel.SetText(content)
}

// estimateHashesForDifficulty は難易度から推定ハッシュ数を計算します
func estimateHashesForDifficulty(difficulty int) float64 {
	// 難易度0: 平均1回
	// 難易度1: 平均16回
	// 難易度2: 平均256回
	// ...
	if difficulty == 0 {
		return 1.0
	}
	// 2^(4*difficulty) の概算
	result := 1.0
	for i := 0; i < difficulty; i++ {
		result *= 16.0
	}
	return result
}

// formatHashRate はハッシュレートを読みやすい形式にフォーマットします
func formatHashRate(hashes float64) string {
	if hashes < 1000 {
		return fmt.Sprintf("%.0f H/s", hashes)
	} else if hashes < 1000000 {
		return fmt.Sprintf("%.2f KH/s", hashes/1000)
	} else if hashes < 1000000000 {
		return fmt.Sprintf("%.2f MH/s", hashes/1000000)
	}
	return fmt.Sprintf("%.2f GH/s", hashes/1000000000)
}

// toggleMining はマイニングを開始/停止します
func (d *Dashboard) toggleMining() {
	if d.isMining {
		d.stopMining()
	} else {
		d.startMining()
	}
}

// startMining はバックグラウンドマイニングを開始します
func (d *Dashboard) startMining() {
	if d.isMining {
		return
	}

	d.isMining = true
	d.miningStopChan = make(chan bool)
	go d.miningLoop()
}

// stopMining はバックグラウンドマイニングを停止します
func (d *Dashboard) stopMining() {
	if !d.isMining {
		return
	}

	d.isMining = false
	d.miningStopChan <- true
}

// miningLoop はバックグラウンドでブロックをマイニングし続けます
func (d *Dashboard) miningLoop() {
	for d.isMining {
		select {
		case <-d.miningStopChan:
			return
		default:
			// ブロックをマイニング
			d.miningCounter++
			data := fmt.Sprintf("Auto-mined block #%d", d.miningCounter)

			_, err := d.blockchain.AddBlock(data)
			if err != nil {
				// エラーが発生した場合は少し待機
				time.Sleep(100 * time.Millisecond)
				continue
			}

			// UIを更新
			d.app.QueueUpdateDraw(func() {
				d.update()
			})

			// 少し待機してから次のブロックをマイニング
			time.Sleep(100 * time.Millisecond)
		}
	}
}
