# Minicoin Development Guidelines

このドキュメントは、Minicoinプロジェクトの開発ルールとガイドラインをまとめたものです。
AIアシスタント（Gemini等）や開発者は、このルールに従って開発を進めてください。

## 1. ブランチ運用 (Branching Strategy)

本プロジェクトでは **GitHub Flow** を採用しています。
`develop` ブランチは使用せず、`main` ブランチを主軸として開発を進めます。

### ブランチの作成
すべての作業用ブランチは `main` ブランチから作成します。

### ブランチ命名規則
ブランチ名は、作業内容を表すプレフィックスと説明を組み合わせて命名します。

- `feat/`: 新機能の追加 (例: `feat/add-p2p-node`)
- `fix/`: バグ修正 (例: `fix/transaction-validation`)
- `hotfix/`: 緊急のバグ修正 (例: `hotfix/ci-failure`)
- `docs/`: ドキュメントの変更 (例: `docs/update-readme`)
- `chore/`: ビルドプロセスやツールの変更 (例: `chore/update-dependencies`)
- `refactor/`: リファクタリング (例: `refactor/wallet-logic`)

## 2. コミットメッセージ (Commit Messages)

**Conventional Commits** 形式に従ってください。

```
<type>(<scope>): <subject>
```

- **type**:
    - `feat`: 新機能
    - `fix`: バグ修正
    - `docs`: ドキュメントのみの変更
    - `style`: コードの動作に影響しない変更（空白、フォーマットなど）
    - `refactor`: バグ修正や機能追加を含まないコードの変更
    - `perf`: パフォーマンス向上のための変更
    - `test`: テストの追加や修正
    - `chore`: ビルドプロセスや補助ツールの変更

- **scope** (省略可): 変更の影響範囲（例: `p2p`, `wallet`, `ci`）
- **subject**: 変更の簡潔な説明（日本語可）

例:
- `feat(p2p): ノード間通信の実装を追加`
- `fix(ci): golangci-lintのバージョンを更新`

## 3. 開発フロー (Development Workflow)

### 作業前の準備
1. `main` ブランチを最新にする: `git checkout main && git pull`
2. 作業用ブランチを作成する: `git checkout -b feat/your-feature-name`

### 品質チェック (Quality Checks)
コミットやプッシュを行う前に、必ずローカルで品質チェックを実行してください。
`Makefile` に定義された以下のコマンドを使用します。

- **必須**: 全品質チェック（フォーマット、Lint、テスト）
  ```bash
  make quality
  ```

- 個別のチェック:
  - `make fmt`: コードフォーマット (`gofmt`)
  - `make lint`:静的解析 (`golangci-lint`)
  - `make test`: 全テスト実行

### プルリクエスト (Pull Request)
1. 作業ブランチをリモートにプッシュします。
2. GitHub上で `main` ブランチに対するPull Request (PR) を作成します。
3. CI (GitHub Actions) がパスすることを確認します。
4. レビューを受けてマージします。

## 4. CI/CD

GitHub Actionsにより、以下のチェックが自動実行されます。
- `test`: ユニットテスト
- `lint`: golangci-lint による静的解析
- `format`: gofmt によるフォーマットチェック
- `coverage`: コードカバレッジの計測

これらのチェックがすべてパスしない限り、マージは推奨されません。
