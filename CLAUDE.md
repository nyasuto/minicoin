# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## プロジェクト概要

Minicoin は、ブロックチェーン技術を段階的に実装することで暗号通貨の仕組みを学ぶ教育プロジェクトです。4 つのステージで構成され、基本的なハッシュチェーンから P2P ネットワークまで段階的に構築します。

## 開発ポリシー

### 🚫 PR Merge Policy
**ABSOLUTE RULE: Claude MUST NEVER merge PRs automatically**
- ✅ PRの作成のみ可能 (`gh pr create`)
- ❌ PRのマージは禁止 (`gh pr merge`)
- ✅ 必ず人間がレビュー・マージする

### Git Workflow

**必須ルール:**
- **mainブランチに直接コミット禁止**
- 必ずfeatureブランチを作成
- 全ての変更はPR経由

**ブランチ命名:**
- Feature: `feat/issue-X-feature-name`
- Bug fix: `fix/issue-X-description`
- Hotfix: `hotfix/X-description`

**開発フロー:**
1. mainからfeatureブランチ作成
2. 実装
3. **🔴 必須: `make quality` を実行**
4. Conventional Commits形式でコミット
5. リモートにpush
6. PR作成
7. 人間によるレビュー待ち

### Quality Checks

**コミット前に必ず実行:**
```bash
make quality
```

このコマンドは以下を実行：
- `go test ./...` - 全テスト実行
- `go fmt ./...` - フォーマットチェック
- `golangci-lint run` - 静的解析

**CI/CD:**
- 全てのquality checksがCIでパス必須
- マージ前にGitHub Actionsの成功確認

### GitHub Issue/PR

**言語:**
- 全てのIssue・PRは日本語で記述

**必須ラベル:**
- Priority: `priority: critical/high/medium/low`
- Type: `type: feature/bug/enhancement/docs/test/refactor/ci/cd/security`

## 開発コマンド

### テスト実行

```bash
# 全テストを実行
go test ./...

# 特定ステージのテストを実行
go test ./stage1-hash-chain/...

# カバレッジ付きで実行
go test -cover ./...

# マイニングパフォーマンスのベンチマーク
go test -bench=. ./stage2-pow/...
```

### 各ステージの実行

```bash
# ステージ1: ハッシュチェーン
cd stage1-hash-chain
go run main.go

# ステージ2: Proof of Work
cd stage2-pow
go run main.go

# ステージ3: トランザクション
cd stage3-transactions
go run main.go

# ステージ4: P2Pネットワーク
cd stage4-p2p
go run main.go
```

### 可視化ダッシュボード

```bash
# CLIダッシュボード
cd visualization/cli-dashboard
go run dashboard.go

# Webダッシュボードは別途実装予定
```

## アーキテクチャと実装の段階

### ステージ 1: ハッシュチェーンの基礎

- **目的**: ブロック構造、SHA-256 ハッシング、チェーン検証の実装
- **主要ファイル**:
  - `stage1-hash-chain/block.go` - ブロック構造とメソッド
  - `stage1-hash-chain/chain.go` - ブロックチェーンロジック
  - `stage1-hash-chain/main.go` - CLI インターフェース
- **実装内容**: データ、タイムスタンプ、前ブロックハッシュを持つ基本構造

### ステージ 2: Proof of Work

- **目的**: マイニングアルゴリズム、難易度調整、ナンス探索の実装
- **主要ファイル**:
  - `stage2-pow/mining.go` - マイニングアルゴリズム
  - `stage2-pow/difficulty.go` - 難易度調整メカニズム
- **実装内容**: 調整可能な難易度でのマイニング、パフォーマンス指標

### ステージ 3: トランザクションと UTXO

- **目的**: ウォレット、デジタル署名、UTXO モデルの実装
- **主要ファイル**:
  - `stage3-transactions/wallet.go` - 公開鍵・秘密鍵ペア管理
  - `stage3-transactions/transaction.go` - トランザクションロジック
  - `stage3-transactions/utxo.go` - UTXO(未使用トランザクション出力)管理
- **実装内容**: トランザクションの署名・検証、残高計算

### ステージ 4: P2P ネットワーク

- **目的**: ノード発見、ブロック伝播、コンセンサスの実装
- **主要ファイル**:
  - `stage4-p2p/node.go` - ノード実装
  - `stage4-p2p/network.go` - ネットワークプロトコル
  - `stage4-p2p/consensus.go` - 最長チェーンルール
- **実装内容**: マルチノードシミュレーション、ブロードキャスト、フォーク解決

### 共有コンポーネント

- `common/crypto.go` - 暗号化関数(SHA-256, ECDSA 等)
- `common/utils.go` - 共通ヘルパー関数
- `visualization/` - CLI/Web 可視化ツール

## 技術スタック

- **言語**: Go 1.21 以上
- **暗号化**: `crypto/sha256`, `crypto/ecdsa`
- **CLI 可視化**: `termui`, `tcell`
- **Web 可視化**: Echo フレームワーク, D3.js
- **テスト**: Go `testing` パッケージ, `testify`

## 設定ファイル

`config.yaml` で以下を設定:

- `blockchain.difficulty`: マイニング難易度(デフォルト: 4)
- `blockchain.block_time`: ブロック生成目標時間(デフォルト: 10 秒)
- `blockchain.reward`: マイニング報酬(デフォルト: 50)
- `network.port`: ネットワークポート(デフォルト: 8080)
- `visualization.update_interval`: 更新間隔(デフォルト: 1000ms)

## 開発原則

- **シンプルさ優先**: 最適化よりコードの明確さを重視
- **段階的な複雑性**: 各ステージは前のステージの上に構築される
- **視覚的学習**: すべての概念に視覚的表現を用意
- **実践的フォーカス**: 理論的完璧さより動作するコードを優先
- **求められたことのみ実装**: 過剰実装禁止
- **既存ファイルの編集を優先**: 新規ファイルは必要な場合のみ

## 実装時の注意点

1. 各ステージは独立して動作するように実装
2. 共通機能は `common/` ディレクトリに配置
3. 可視化機能は常に実装と並行して開発
4. テストはベンチマークも含めて各ステージごとに作成
5. コードの明確さと教育的価値を最優先
6. **コミット前に必ず `make quality` を実行**
