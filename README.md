# 🪙 Minicoin

> 実装を通じて暗号通貨の基礎を学ぶ、ミニマリストなブロックチェーン実装

## 📖 概要

Minicoinは、ブロックチェーン技術をゼロから構築することで暗号通貨の仕組みを理解するための、段階的な教育プロジェクトです。「作りながら学ぶ」哲学に基づき、ブロックチェーンの中核概念を段階的に実装していきます。

## 🎯 プロジェクトの目標

- **理解する** - 実装を通じてブロックチェーンの基礎を理解
- **視覚化する** - 複雑な概念を直感的なインターフェースで可視化
- **構築する** - シンプルなハッシュチェーンからP2Pネットワークまで段階的に構築
- **学ぶ** - 動作する暗号通貨システムを作成しながら学習

## 🏗️ アーキテクチャの段階

### ステージ1: ハッシュチェーンの基礎
```
ブロック構造 → SHA-256ハッシング → チェーン検証 → CLI可視化
```
- データ、タイムスタンプ、前ブロックハッシュを持つ基本的なブロック構造
- SHA-256ハッシュ計算とチェーンの整合性検証
- ターミナルベースのチェーン可視化

### ステージ2: Proof of Work (PoW)
```
マイニングアルゴリズム → 難易度調整 → ナンス探索 → マイニング指標
```
- 調整可能な難易度でのマイニング実装
- ナンス探索プロセスの可視化
- パフォーマンス指標とマイニング統計

### ステージ3: トランザクションとUTXO
```
ウォレット → デジタル署名 → UTXOモデル → 残高計算
```
- 公開鍵・秘密鍵ペアの生成
- トランザクションの署名と検証
- 未使用トランザクション出力（UTXO）の管理

### ステージ4: P2Pネットワーク
```
ノード発見 → ブロック伝播 → コンセンサス → フォーク解決
```
- マルチノードローカルネットワークシミュレーション
- ブロックのブロードキャストと同期
- 最長チェーンコンセンサスルール

## 🚀 クイックスタート
```bash
# リポジトリをクローン
git clone https://github.com/yourusername/minicoin.git
cd minicoin

# ステージ1から開始
cd stage1-hash-chain
go run main.go

# 可視化ダッシュボードを起動
cd visualization/cli-dashboard
go run dashboard.go
```

## 💻 技術スタック

- **言語**: Go 1.21以上
- **暗号化**: crypto/sha256, crypto/ecdsa
- **可視化**: 
  - CLI: termui, tcell
  - Web: Echoフレームワーク, D3.js
- **テスト**: Go testingパッケージ, testify
- **メトリクス**: Prometheus, Grafana（オプション）

## 📂 プロジェクト構造
```
minicoin/
├── stage1-hash-chain/      # 基本的なブロックチェーン実装
│   ├── block.go           # ブロック構造とメソッド
│   ├── chain.go           # ブロックチェーンロジック
│   └── main.go            # CLIインターフェース
│
├── stage2-pow/             # Proof of Work実装
│   ├── mining.go          # マイニングアルゴリズム
│   ├── difficulty.go      # 難易度調整
│   └── main.go
│
├── stage3-transactions/    # トランザクションシステム
│   ├── wallet.go          # ウォレット実装
│   ├── transaction.go     # トランザクションロジック
│   ├── utxo.go           # UTXO管理
│   └── main.go
│
├── stage4-p2p/            # P2Pネットワーク
│   ├── node.go           # ノード実装
│   ├── network.go        # ネットワークプロトコル
│   ├── consensus.go      # コンセンサスメカニズム
│   └── main.go
│
├── visualization/         # 可視化ツール
│   ├── cli-dashboard/    # ターミナルUIダッシュボード
│   └── web-ui/          # Webベースインターフェース
│
├── common/               # 共有ユーティリティ
│   ├── crypto.go        # 暗号化関数
│   └── utils.go         # ヘルパー関数
│
├── docs/                 # ドキュメント
│   ├── learning-notes.md # 学習ノート
│   ├── architecture.md   # アーキテクチャ設計
│   └── api.md           # API仕様
│
└── examples/            # 使用例
    └── scenarios/       # テストシナリオ
```

## 🎨 可視化機能

### CLIダッシュボード
- リアルタイムのブロックチェーン状態監視
- ASCIIアートでのチェーン表現
- マイニング進捗インジケーター
- ネットワークトポロジービュー

### Webインターフェース
- インタラクティブなブロックチェーンエクスプローラー
- トランザクションフローの可視化
- マイニング難易度チャート
- ネットワーク健全性メトリクス

## 📚 学習パス

1. **第1-2週**: ステージ1完了 - ハッシングとチェーン構造を理解
2. **第3-4週**: ステージ2完了 - Proof of Workの概念をマスター
3. **第5-6週**: ステージ3完了 - トランザクションの仕組みを学習
4. **第7-8週**: ステージ4完了 - 分散コンセンサスを探求

## 🧪 テスト
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

## 📝 開発原則

- **シンプルさ優先**: 最適化よりコードの明確さを重視
- **段階的な複雑性**: 各ステージは前のステージの上に構築
- **視覚的学習**: すべての概念に視覚的表現を用意
- **実践的フォーカス**: 理論的完璧さより動作するコード

## 🔧 設定
```yaml
# config.yaml
blockchain:
  difficulty: 4         # マイニング難易度
  block_time: 10       # ブロック生成目標時間（秒）
  reward: 50           # マイニング報酬

network:
  port: 8080
  peers:
    - localhost:8081
    - localhost:8082

visualization:
  update_interval: 1000 # 更新間隔（ミリ秒）
  enable_web_ui: true   # WebUI有効化
```

## 🤝 コントリビューション

これは個人的な学習プロジェクトですが、提案やディスカッションは歓迎です！
- 質問や明確化のためのIssueを開く
- 自身の学習経験を共有する
- 可視化の改善を提案する

## 📖 参考資料

- [Bitcoin ホワイトペーパー（日本語訳）](https://bitcoin.org/files/bitcoin-paper/bitcoin_jp.pdf)
- [Mastering Bitcoin（日本語版）](https://bitcoinbook.info/translations-of-mastering-bitcoin/)
- [ブロックチェーンデモ](https://andersbrownworth.com/blockchain/)
- [Go暗号化パッケージ](https://pkg.go.dev/crypto)

## 📄 ライセンス

MIT License - 自由に学習してください！

## 🎯 ロードマップ

- [x] ステージ1: 基本的なハッシュチェーン
- [ ] ステージ2: Proof of Work
- [ ] ステージ3: トランザクション
- [ ] ステージ4: P2Pネットワーク
- [ ] ボーナス: シンプルなスマートコントラクト
- [ ] ボーナス: マークルツリー
- [ ] ボーナス: SPV（簡易支払い検証）

## 🚧 現在の進捗
```
📝 Stage 1: Hash Chain    [████████████████████] 100% 完了
⚡ Stage 2: PoW           [████░░░░░░░░░░░░░░░░] 20% 進行中
💰 Stage 3: Transactions  [░░░░░░░░░░░░░░░░░░░░] 0%  未着手
🌐 Stage 4: P2P Network   [░░░░░░░░░░░░░░░░░░░░] 0%  未着手
```

---

*ブロックチェーン学習者がGoと好奇心で構築* 🚀