# =============================================================================
# 猪猪记账 - 构建脚本
# Wails v2 + Go + Vue 3 桌面应用构建工具
# =============================================================================

# ============ 项目配置 ============
APP_NAME    := piggy-accounting
BINARY_NAME := 猪猪记账
BUILD_DIR   := build/bin
VERSION     ?= 1.0.0

# ============ 工具检测 ============
GO     := $(shell which go 2>/dev/null)
WAILS  := $(shell which wails 2>/dev/null)
NPM    := $(shell which npm 2>/dev/null)

ifndef GO
$(error "go command not found, please install Go and make sure it's in your PATH")
endif

ifndef WAILS
$(error "wails command not found, please install Wails: go install github.com/wailsapp/wails/v2/cmd/wails@latest")
endif

ifndef NPM
$(error "npm command not found, please install Node.js and make sure npm is in your PATH")
endif

# ============ 版本信息 ============
GIT_COMMIT  := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_TIME  := $(shell date -u '+%Y-%m-%d_%H:%M:%S')
GO_VERSION  := $(shell go version | sed -e 's/^[^0-9.]*//' -e 's/ .*//')
LDFLAGS     := -s -w -buildid= -X main.version=$(VERSION) -X main.commit=$(GIT_COMMIT) -X main.buildTime=$(BUILD_TIME) -X main.goVersion=$(GO_VERSION)

# ============ 平台检测 ============
UNAME_S := $(shell uname -s)
ifeq ($(UNAME_S),Darwin)
    HOST_OS := darwin
    HOST_ARCH := $(shell uname -m)
else ifneq (,$(findstring NT-10.0.1,$(OS)))
    HOST_OS := windows
    HOST_ARCH := amd64
else
    HOST_OS := linux
    HOST_ARCH := $(shell uname -m)
endif

# ============ 颜色输出 ============
BLUE        := \033[0;34m
GREEN       := \033[0;32m
YELLOW      := \033[0;33m
RED         := \033[0;31m
NC          := \033[0m

# ============ 公共目标 ============
.PHONY: help
.DEFAULT_GOAL := help

help: ## 📖 显示此帮助信息
	@echo ""
	@echo "$(BLUE)===========================================================$(NC)"
	@echo "$(BLUE)                     $(BINARY_NAME) - 构建脚本$(NC)"
	@echo "$(BLUE)===========================================================$(NC)"
	@echo ""
	@echo "$(GREEN)开发工具:$(NC)"
	@grep -E '^[a-zA-Z_0-9%-]+:.*?## 🧪 .*$$' $(word 1,$(MAKEFILE_LIST)) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "$(YELLOW)  %-25s$(NC) %s\n", $$1, $$2}'
	@echo ""
	@echo "$(GREEN)构建工具:$(NC)"
	@grep -E '^[a-zA-Z_0-9%-]+:.*?## 🛠️  .*$$' $(word 1,$(MAKEFILE_LIST)) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "$(YELLOW)  %-25s$(NC) %s\n", $$1, $$2}'
	@echo ""
	@echo "$(GREEN)清理工具:$(NC)"
	@grep -E '^[a-zA-Z_0-9%-]+:.*?## 🧹 .*$$' $(word 1,$(MAKEFILE_LIST)) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "$(YELLOW)  %-25s$(NC) %s\n", $$1, $$2}'
	@echo ""
	@echo "$(GREEN)实用工具:$(NC)"
	@grep -E '^[a-zA-Z_0-9%-]+:.*?## 🔧 .*$$' $(word 1,$(MAKEFILE_LIST)) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "$(YELLOW)  %-25s$(NC) %s\n", $$1, $$2}'
	@echo ""

# ============ 开发工具 ============
.PHONY: dev deps tidy
dev: ## 🧪 启动开发模式
	@echo "$(BLUE)🚀 启动开发模式...$(NC)"
	$(WAILS) dev

deps: ## 🧪 安装前端依赖
	@echo "$(BLUE)📦 安装前端依赖...$(NC)"
	cd frontend && $(NPM) install

tidy: ## 🧪 整理 Go 模块依赖
	@echo "$(BLUE)📦 整理 Go 模块依赖...$(NC)"
	$(GO) mod tidy

# ============ 构建工具 ============
.PHONY: build build-release build-debug build-fast build-min
build: ## 🛠️  构建当前平台版本 (Release)
	@echo "$(BLUE)🔨 构建 $(HOST_OS)/$(HOST_ARCH) Release 版本...$(NC)"
	$(WAILS) build \
		-o "$(BINARY_NAME)" \
		-ldflags "$(LDFLAGS)" \
		-trimpath \
		-clean \
		-nosyncgomod

build-min: ## 🛠️  构建最小体积版本 (启用UPX压缩)
	@echo "$(BLUE)📦 构建 $(HOST_OS)/$(HOST_ARCH) 最小体积版本 (启用UPX)...$(NC)"
	@if [ "$(HOST_OS)" = "darwin" ]; then \
		echo "$(YELLOW)⚠️  macOS 平台不支持 UPX 压缩 (与代码签名冲突)，构建普通版本...$(NC)"; \
		$(WAILS) build \
			-o "$(BINARY_NAME)" \
			-ldflags "$(LDFLAGS)" \
			-trimpath \
			-clean \
			-nosyncgomod; \
	else \
		$(WAILS) build \
			-o "$(BINARY_NAME)" \
			-upx \
			-ldflags "$(LDFLAGS)" \
			-trimpath \
			-clean \
			-nosyncgomod || \
		( \
			echo "$(YELLOW)⚠️  UPX 压缩失败，使用未压缩版本...$(NC)"; \
			echo "$(BLUE)📦 构建未压缩版本...$(NC)"; \
			$(WAILS) build \
				-o "$(BINARY_NAME)" \
				-ldflags "$(LDFLAGS)" \
				-trimpath \
				-clean \
				-nosyncgomod \
		); \
	fi

build-release: build ## 🛠️  构建发布版 (同 build)

build-debug: ## 🛠️  构建当前平台 Debug 版本
	@echo "$(YELLOW)🔍 构建 Debug 版本 (带调试信息)...$(NC)"
	$(WAILS) build \
		-o "$(BINARY_NAME)" \
		-debug \
		-devtools \
		-trimpath \
		-clean \
		-nosyncgomod

build-fast: ## 🛠️  快速构建 (跳过前端构建和绑定生成)
	@echo "$(BLUE)⚡ 快速构建 (跳过前端构建)...$(NC)"
	$(WAILS) build \
		-o "$(BINARY_NAME)" \
		-s \
		-skipbindings \
		-ldflags "$(LDFLAGS)" \
		-trimpath \
		-nosyncgomod

# ============ 平台特定构建 ============
.PHONY: build-mac build-mac-arm build-mac-amd build-mac-universal build-win build-win64 build-win-min build-linux

build-mac: build-mac-arm ## 🛠️  构建 macOS (Apple Silicon, 默认)
	@echo "$(BLUE)🍎 构建 macOS (Apple Silicon)...$(NC)"

build-mac-arm: ## 🛠️  构建 macOS arm64 (Apple Silicon)
	@echo "$(BLUE)🍎 构建 macOS arm64 (Apple Silicon)...$(NC)"
	$(WAILS) build \
		-platform "darwin/arm64" \
		-o "$(BINARY_NAME)" \
		-ldflags "$(LDFLAGS)" \
		-trimpath \
		-clean \
		-nosyncgomod

build-mac-amd: ## 🛠️  构建 macOS amd64 (Intel)
	@echo "$(BLUE)🍎 构建 macOS amd64 (Intel)...$(NC)"
	$(WAILS) build \
		-platform "darwin/amd64" \
		-o "$(BINARY_NAME)" \
		-ldflags "$(LDFLAGS)" \
		-trimpath \
		-clean \
		-nosyncgomod

build-mac-universal: ## 🛠️  构建 macOS Universal (arm64 + amd64)
	@echo "$(BLUE)🍎 构建 macOS Universal 二进制文件...$(NC)"
	@echo "$(YELLOW)  步骤 1: 构建 arm64...$(NC)"
	$(WAILS) build \
		-platform "darwin/arm64" \
		-o "$(BINARY_NAME)_arm64" \
		-ldflags "$(LDFLAGS)" \
		-trimpath \
		-nosyncgomod
	@echo "$(YELLOW)  步骤 2: 构建 amd64...$(NC)"
	$(WAILS) build \
		-platform "darwin/amd64" \
		-o "$(BINARY_NAME)_amd64" \
		-ldflags "$(LDFLAGS)" \
		-trimpath \
		-nosyncgomod
	@echo "$(YELLOW)  步骤 3: 合并为 Universal 二进制文件...$(NC)"
	@lipo -create -output "$(BUILD_DIR)/$(BINARY_NAME)" \
		"$(BUILD_DIR)/$(BINARY_NAME)_arm64" \
		"$(BUILD_DIR)/$(BINARY_NAME)_amd64"
	@rm -f "$(BUILD_DIR)/$(BINARY_NAME)_arm64" "$(BUILD_DIR)/$(BINARY_NAME)_amd64"
	@echo "$(GREEN)✅ Universal 二进制文件构建完成: $(BUILD_DIR)/$(BINARY_NAME)$(NC)"

build-win: build-win64 ## 🛠️  构建 Windows (amd64, 默认)

build-win64: ## 🛠️  构建 Windows amd64
	@echo "$(BLUE)🪟 构建 Windows amd64...$(NC)"
	$(WAILS) build \
		-platform "windows/amd64" \
		-o "$(BINARY_NAME).exe" \
		-ldflags "$(LDFLAGS)" \
		-trimpath \
		-clean \
		-nosyncgomod

build-win-min: ## 🛠️  构建最小体积 Windows amd64 (启用UPX压缩)
	@echo "$(BLUE)🪟 构建最小体积 Windows amd64 (启用UPX)...$(NC)"
	$(WAILS) build \
		-platform "windows/amd64" \
		-o "$(BINARY_NAME).exe" \
		-upx \
		-ldflags "$(LDFLAGS)" \
		-trimpath \
		-clean \
		-nosyncgomod || \
	( \
		echo "$(YELLOW)⚠️  UPX 压缩失败，使用未压缩版本...$(NC)"; \
		echo "$(BLUE)📦 构建未压缩 Windows 版本...$(NC)"; \
		$(WAILS) build \
			-platform "windows/amd64" \
			-o "$(BINARY_NAME).exe" \
			-ldflags "$(LDFLAGS)" \
			-trimpath \
			-clean \
			-nosyncgomod \
	)

build-linux: ## 🛠️  构建 Linux amd64 (需在 Linux 上运行)
	@echo "$(BLUE)🐧 构建 Linux amd64...$(NC)"
	@if [ "$(HOST_OS)" != "linux" ]; then \
		echo "$(RED)❌ Wails v2 不支持从 $(HOST_OS) 交叉编译到 Linux$(NC)"; \
		echo "$(YELLOW)请在 Linux 机器上运行此命令$(NC)"; \
		exit 1; \
	fi
	$(WAILS) build \
		-platform "linux/amd64" \
		-o "$(BINARY_NAME)" \
		-ldflags "$(LDFLAGS)" \
		-trimpath \
		-clean \
		-nosyncgomod

build-linux-min: ## 🛠️  构建最小体积 Linux amd64 (启用UPX压缩)
	@echo "$(BLUE)🐧 构建最小体积 Linux amd64 (启用UPX压缩)...$(NC)"
	@if [ "$(HOST_OS)" != "linux" ]; then \
		echo "$(RED)❌ Wails v2 不支持从 $(HOST_OS) 交叉编译到 Linux$(NC)"; \
		echo "$(YELLOW)请在 Linux 机器上运行此命令$(NC)"; \
		exit 1; \
	fi
	$(WAILS) build \
		-platform "linux/amd64" \
		-o "$(BINARY_NAME)" \
		-upx \
		-ldflags "$(LDFLAGS)" \
		-trimpath \
		-clean \
		-nosyncgomod || \
	( \
		echo "$(YELLOW)⚠️  UPX 压缩失败，使用未压缩版本...$(NC)"; \
		echo "$(BLUE)📦 构建未压缩 Linux 版本...$(NC)"; \
		$(WAILS) build \
			-platform "linux/amd64" \
			-o "$(BINARY_NAME)" \
			-ldflags "$(LDFLAGS)" \
			-trimpath \
			-clean \
			-nosyncgomod \
	)

# ============ 全平台构建 ============
.PHONY: build-all build-all-min
build-all: ## 🛠️  构建所有可交叉编译的平台 (macOS + Windows)
	@echo "$(BLUE)🌍 构建所有平台...$(NC)"
	@echo "$(YELLOW)── macOS arm64 ──$(NC)"
	$(WAILS) build \
		-platform "darwin/arm64" \
		-o "$(BINARY_NAME)_mac_arm64" \
		-ldflags "$(LDFLAGS)" \
		-trimpath \
		-nosyncgomod
	@echo "$(YELLOW)── macOS amd64 ──$(NC)"
	$(WAILS) build \
		-platform "darwin/amd64" \
		-o "$(BINARY_NAME)_mac_amd64" \
		-ldflags "$(LDFLAGS)" \
		-trimpath \
		-nosyncgomod
	@echo "$(YELLOW)── Windows amd64 ──$(NC)"
	$(WAILS) build \
		-platform "windows/amd64" \
		-o "$(BINARY_NAME)_win_amd64.exe" \
		-ldflags "$(LDFLAGS)" \
		-trimpath \
		-nosyncgomod
	@echo "$(GREEN)✅ 全平台构建完成！$(NC)"
	@echo ""
	@echo "$(BLUE)产物目录: $(BUILD_DIR)/$(NC)"
	@echo ""
	@ls -lh $(BUILD_DIR)/ 2>/dev/null | tail -n +2
	@if [ "$(HOST_OS)" != "linux" ]; then \
		echo ""; \
		echo "$(YELLOW)⚠️  Linux 版本需要在 Linux 机器上单独构建 (Wails 不支持交叉编译)$(NC)"; \
	fi

build-all-min: ## 🛠️  构建所有可交叉编译平台的最小化版本 (macOS + Windows)
	@echo "$(BLUE)🌍 构建所有平台的最小化版本...$(NC)"
	@echo "$(YELLOW)── macOS arm64 (无UPX压缩) ──$(NC)"
	$(WAILS) build \
		-platform "darwin/arm64" \
		-o "$(BINARY_NAME)_mac_arm64" \
		-ldflags "$(LDFLAGS)" \
		-trimpath \
		-nosyncgomod
	@echo "$(YELLOW)── macOS amd64 (无UPX压缩) ──$(NC)"
	$(WAILS) build \
		-platform "darwin/amd64" \
		-o "$(BINARY_NAME)_mac_amd64" \
		-ldflags "$(LDFLAGS)" \
		-trimpath \
		-nosyncgomod
	@echo "$(YELLOW)── Windows amd64 (启用UPX压缩) ──$(NC)"
	$(WAILS) build \
		-platform "windows/amd64" \
		-o "$(BINARY_NAME)_win_amd64.exe" \
		-upx \
		-ldflags "$(LDFLAGS)" \
		-trimpath \
		-nosyncgomod
	@echo "$(GREEN)✅ 全平台最小化构建完成！$(NC)"
	@echo ""
	@echo "$(BLUE)产物目录: $(BUILD_DIR)/$(NC)"
	@echo ""
	@ls -lh $(BUILD_DIR)/ 2>/dev/null | tail -n +2
	@if [ "$(HOST_OS)" != "linux" ]; then \
		echo ""; \
		echo "$(YELLOW)⚠️  Linux 版本需要在 Linux 机器上单独构建 (Wails 不支持交叉编译)$(NC)"; \
	fi

# ============ 清理工具 ============
.PHONY: clean clean-frontend clean-backend
clean: ## 🧹 清理所有构建产物
	@echo "$(YELLOW)🧹 清理构建目录...$(NC)"
	rm -rf $(BUILD_DIR)
	rm -rf frontend/dist
	rm -rf frontend/node_modules
	@echo "$(GREEN)✅ 清理完成$(NC)"

clean-frontend: ## 🧹 仅清理前端构建产物
	@echo "$(YELLOW)🧹 清理前端构建产物...$(NC)"
	rm -rf frontend/dist
	@echo "$(GREEN)✅ 前端清理完成$(NC)"

clean-backend: ## 🧹 仅清理后端构建产物
	@echo "$(YELLOW)🧹 清理后端构建产物...$(NC)"
	rm -rf $(BUILD_DIR)
	@echo "$(GREEN)✅ 后端清理完成$(NC)"

# ============ 实用工具 ============
.PHONY: size
size: ## 🔧 显示构建产物大小
	@echo "$(BLUE)📏 构建产物大小:$(NC)"
	@find $(BUILD_DIR) -maxdepth 3 -type f \( -name "*.exe" -o -name "*.app" -o -name "$(BINARY_NAME)" -o -perm +111 \) 2>/dev/null | \
		while read f; do \
			if [ -d "$$f" ]; then \
				SIZE=$$(du -sh "$$f" | cut -f1); \
				echo "  $$SIZE  $$f"; \
			else \
				SIZE=$$(ls -lh "$$f" | awk '{print $$5}'); \
				echo "  $$SIZE  $$f"; \
			fi; \
		done
	@echo ""
	@echo "$(BLUE)目录总大小:$(NC)"
	@du -sh $(BUILD_DIR) 2>/dev/null || echo "  构建目录不存在"

# ============ 验证工具 ============
.PHONY: check check-tools check-buildable
check: check-tools ## 🔧 检查环境是否可构建

check-tools: ## 🔧 检查构建所需工具
	@echo "$(BLUE)🔍 检查构建工具...$(NC)"
	@echo "  Go:     $(GO)"
	@echo "  Wails:  $(WAILS)"
	@echo "  NPM:    $(NPM)"
	@echo "  OS:     $(HOST_OS)"
	@echo "  Arch:   $(HOST_ARCH)"
	@echo "$(GREEN)✅ 工具检查完成$(NC)"

check-buildable: ## 🔧 检查是否可以在当前平台上构建
	@echo "$(BLUE)🔍 检查构建能力...$(NC)"
	@echo "  当前平台: $(HOST_OS)/$(HOST_ARCH)"
	@if [ "$(HOST_OS)" = "darwin" ]; then \
		echo "$(GREEN)✅ 支持 macOS (arm64/amd64/universal) 和 Windows (amd64) 构建$(NC)"; \
	elif [ "$(HOST_OS)" = "linux" ]; then \
		echo "$(GREEN)✅ 支持 Linux (当前架构) 和 Windows (amd64) 构建$(NC)"; \
	elif [ "$(HOST_OS)" = "windows" ]; then \
		echo "$(GREEN)✅ 支持 Windows 构建$(NC)"; \
	else \
		echo "$(YELLOW)⚠️  未知平台: $(HOST_OS)$(NC)"; \
	fi