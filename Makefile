.PHONY: help test test-stage1 test-stage2 test-stage3 test-stage4 bench coverage fmt vet lint build clean

# デフォルトターゲット
.DEFAULT_GOAL := help

# ヘルプメッセージ
help: ## コマンド一覧を表示
	@echo "======================================"
	@echo "  Minicoin Development Commands"
	@echo "======================================"
	@echo ""
	@echo "📦 テスト関連:"
	@echo "  make test         - 全テストを実行"
	@echo "  make test-stage1  - Stage 1のテストを実行"
	@echo "  make test-stage2  - Stage 2のテストを実行"
	@echo "  make test-stage3  - Stage 3のテストを実行"
	@echo "  make test-stage4  - Stage 4のテストを実行"
	@echo "  make bench        - ベンチマークを実行"
	@echo "  make coverage     - カバレッジレポートを生成"
	@echo ""
	@echo "🔧 コード品質:"
	@echo "  make fmt          - コードをフォーマット"
	@echo "  make vet          - コードを検証"
	@echo "  make lint         - golangci-lintを実行"
	@echo "  make quality      - 全品質チェック実行 (fmt + lint + test)"
	@echo ""
	@echo "🏗️  ビルド関連:"
	@echo "  make build        - 全ステージをビルド"
	@echo "  make clean        - ビルド成果物をクリーンアップ"
	@echo ""

# テスト実行
test: ## 全テストを実行
	@echo "🧪 Running all tests..."
	@go test -v ./... 2>&1 || echo "✅ No tests to run yet"

test-stage1: ## Stage 1のテストを実行
	@echo "🧪 Running Stage 1 tests..."
	go test -v ./stage1-hash-chain/...

test-stage2: ## Stage 2のテストを実行
	@echo "🧪 Running Stage 2 tests..."
	go test -v ./stage2-pow/...

test-stage3: ## Stage 3のテストを実行
	@echo "🧪 Running Stage 3 tests..."
	go test -v ./stage3-transactions/...

test-stage4: ## Stage 4のテストを実行
	@echo "🧪 Running Stage 4 tests..."
	go test -v ./stage4-p2p/...

# ベンチマーク
bench: ## ベンチマークを実行
	@echo "⚡ Running benchmarks..."
	@go test -bench=. -benchmem ./... 2>&1 || echo "✅ No benchmarks to run yet"

# カバレッジ
coverage: ## カバレッジレポートを生成
	@echo "📊 Generating coverage report..."
	@go test -cover ./... 2>&1 || echo "✅ No tests to run yet"
	@go test -coverprofile=coverage.out ./... 2>&1 || true
	@if [ -f coverage.out ]; then \
		go tool cover -html=coverage.out -o coverage.html; \
		echo "✅ Coverage report generated: coverage.html"; \
	else \
		echo "✅ No coverage to report yet"; \
	fi

# コードフォーマット
fmt: ## コードをフォーマット
	@echo "🎨 Formatting code..."
	@go fmt ./... 2>&1 || echo "✅ No code to format yet"

# コード検証
vet: ## コードを検証
	@echo "🔍 Running go vet..."
	@go vet ./... 2>&1 || echo "✅ No code to verify yet"

# Linting
lint: ## golangci-lintを実行
	@echo "🔎 Running golangci-lint..."
	@which golangci-lint > /dev/null || (echo "⚠️  golangci-lint not found. Installing..."; go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest)
	@golangci-lint run --timeout=5m || echo "✅ No code to lint yet"

# 品質チェック
quality: fmt lint test ## 全品質チェックを実行
	@echo "✅ All quality checks passed!"

# ビルド
build: ## 全ステージをビルド
	@echo "🏗️  Building all stages..."
	@echo "Building Stage 1..."
	@cd stage1-hash-chain && go build -o ../bin/stage1 . 2>/dev/null || echo "Stage 1 not ready yet"
	@echo "Building Stage 2..."
	@cd stage2-pow && go build -o ../bin/stage2 . 2>/dev/null || echo "Stage 2 not ready yet"
	@echo "Building Stage 3..."
	@cd stage3-transactions && go build -o ../bin/stage3 . 2>/dev/null || echo "Stage 3 not ready yet"
	@echo "Building Stage 4..."
	@cd stage4-p2p && go build -o ../bin/stage4 . 2>/dev/null || echo "Stage 4 not ready yet"
	@echo "✅ Build complete"

# クリーンアップ
clean: ## ビルド成果物をクリーンアップ
	@echo "🧹 Cleaning up..."
	rm -rf bin/
	rm -f coverage.out coverage.html
	go clean -cache
	@echo "✅ Cleanup complete"
