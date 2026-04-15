#!/bin/bash
# =============================================================================
# 猪猪记账 - GitHub Release 发布脚本
# 职责：git 提交 + 打标签 + 推送 + 创建 GitHub Release + 上传产物
# 前置要求：已运行 release.sh 完成版本号更新和 make build-release 构建
# =============================================================================

set -e

# ============ 颜色 ============
BLUE='\033[0;34m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
CYAN='\033[0;36m'
NC='\033[0m'

# ============ 配置 ============
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
cd "$SCRIPT_DIR"

GH_BIN="/usr/local/bin/gh"
BUILD_DIR="build/bin"
BINARY_NAME="猪猪记账"
APP_NAME="piggy-accounting"
WAILS_JSON="wails.json"
REPO="wsczx/piggy-accounting"

# ============ 工具函数 ============
info()    { echo -e "${BLUE}ℹ  $*${NC}"; }
success() { echo -e "${GREEN}✅ $*${NC}"; }
warn()    { echo -e "${YELLOW}⚠️  $*${NC}"; }
error()   { echo -e "${RED}❌ $*${NC}"; exit 1; }
step()    { echo -e "${CYAN}[${1}/${2}]${NC} $3"; }

# 获取当前版本号
get_version() {
    python3 -c "import json; print(json.load(open('$WAILS_JSON'))['info']['productVersion'])" 2>/dev/null \
        || error "无法读取 wails.json 中的版本号"
}

# ============ 检查前置条件 ============
check_requirements() {
    # 检查 gh
    if [ ! -f "$GH_BIN" ]; then
        error "未找到 gh 命令 ($GH_BIN)，请先安装: brew install gh"
    fi

    # 检查 gh 认证
    if ! "$GH_BIN" auth status &>/dev/null; then
        echo ""
        error "gh 未登录，请先运行:\n  gh auth login --git-protocol ssh --web"
    fi

    # 检查 git 状态
    if ! git rev-parse --git-dir &>/dev/null; then
        error "当前目录不是 git 仓库"
    fi

    # 检查 build/bin 是否有产物
    if [ ! -d "$BUILD_DIR" ] || [ -z "$(ls -A "$BUILD_DIR" 2>/dev/null)" ]; then
        error "build/bin 目录为空，请先运行 make build-release 构建产物"
    fi
}

# ============ 收集发布产物 ============
collect_assets() {
    ASSETS=()

    # 按优先级收集：zip > exe > tar.gz（跳过源码包）
    for f in \
        "$BUILD_DIR/${APP_NAME}_macOS_ARM64.zip" \
        "$BUILD_DIR/${APP_NAME}_macOS_AMD64.zip" \
        "$BUILD_DIR/${APP_NAME}.exe"; do
        if [ -f "$f" ]; then
            ASSETS+=("$f")
        fi
    done

    # 如果上面都没有，收集所有 zip/exe（排除源码包）
    if [ ${#ASSETS[@]} -eq 0 ]; then
        while IFS= read -r -d '' f; do
            # 排除源码包（含版本号的 zip/tar.gz）
            basename_f=$(basename "$f")
            if [[ "$basename_f" != "${BINARY_NAME}-"* ]]; then
                ASSETS+=("$f")
            fi
        done < <(find "$BUILD_DIR" -maxdepth 1 \( -name "*.zip" -o -name "*.exe" \) -print0 2>/dev/null)
    fi

    if [ ${#ASSETS[@]} -eq 0 ]; then
        warn "build/bin 中没有找到可上传的产物 (.zip / .exe)"
        warn "Release 将被创建，但不含附件"
    fi
}

# ============ 生成 Release Notes ============
generate_notes() {
    local version=$1
    local prev_tag
    prev_tag=$(git describe --tags --abbrev=0 HEAD~1 2>/dev/null || echo "")

    if [ -n "$prev_tag" ]; then
        # 有上一个 tag，生成 commit log
        echo "## 更新内容"
        echo ""
        git log "${prev_tag}..HEAD" --pretty=format:"- %s" --no-merges 2>/dev/null || true
        echo ""
    else
        echo "## 猪猪记账 ${version}"
        echo ""
        echo "首个正式版本发布。"
        echo ""
    fi

    echo ""
    echo "## 下载说明"
    echo ""
    echo "| 文件 | 平台 |"
    echo "|------|------|"
    echo "| \`${APP_NAME}_macOS_ARM64.zip\` | macOS Apple Silicon (M1/M2/M3) |"
    echo "| \`${APP_NAME}_macOS_AMD64.zip\` | macOS Intel |"
    echo "| \`${APP_NAME}.exe\` | Windows 64位 |"
}

# ============ 主流程 ============
main() {
    # 解析参数
    DRAFT=false
    PRERELEASE=false
    NOTES_FILE=""
    SKIP_GIT=false

    while [[ $# -gt 0 ]]; do
        case "$1" in
            --draft)       DRAFT=true; shift ;;
            --prerelease)  PRERELEASE=true; shift ;;
            --notes)       NOTES_FILE="$2"; shift 2 ;;
            --skip-git)    SKIP_GIT=true; shift ;;
            -h|--help)
                echo "用法: $0 [选项]"
                echo ""
                echo "选项:"
                echo "  --draft        创建草稿 Release（不公开）"
                echo "  --prerelease   标记为预发布版本"
                echo "  --notes FILE   从文件读取 Release Notes"
                echo "  --skip-git     跳过 git commit/tag/push（仅创建 Release）"
                echo "  -h, --help     显示此帮助"
                echo ""
                echo "前置步骤:"
                echo "  1. ./release.sh -v 1.0.0    # 更新版本号"
                echo "  2. make build-release        # 构建所有平台产物"
                echo "  3. ./publish.sh              # 发布到 GitHub Release"
                exit 0
                ;;
            *)
                error "未知参数: $1，使用 --help 查看帮助"
                ;;
        esac
    done

    echo ""
    echo -e "${BLUE}============================================================${NC}"
    echo -e "${BLUE}           猪猪记账 - GitHub Release 发布${NC}"
    echo -e "${BLUE}============================================================${NC}"
    echo ""

    # 前置检查
    check_requirements
    VERSION=$(get_version)
    TAG="v${VERSION}"

    info "当前版本: ${TAG}"
    info "仓库地址: https://github.com/${REPO}"
    echo ""

    # 检查 tag 是否已存在
    if git rev-parse "$TAG" &>/dev/null; then
        warn "tag ${TAG} 已存在"
        echo -n "  是否覆盖已有 tag 继续发布？(y/N): "
        read -r confirm
        if [[ "$confirm" != "y" && "$confirm" != "Y" ]]; then
            echo "已取消"
            exit 0
        fi
        FORCE_TAG=true
    else
        FORCE_TAG=false
    fi

    # 收集发布产物
    collect_assets

    # 显示发布计划
    echo ""
    echo -e "${CYAN}发布计划:${NC}"
    echo "  版本 tag : ${TAG}"
    if [ "$DRAFT" = true ]; then
        echo "  类型     : 草稿（不公开）"
    elif [ "$PRERELEASE" = true ]; then
        echo "  类型     : 预发布"
    else
        echo "  类型     : 正式发布"
    fi
    echo "  附件数量 : ${#ASSETS[@]} 个"
    for asset in "${ASSETS[@]}"; do
        SIZE=$(ls -lh "$asset" 2>/dev/null | awk '{print $5}')
        echo "    - $(basename "$asset")  (${SIZE})"
    done
    echo ""

    echo -n "确认发布? (y/N): "
    read -r confirm
    [[ "$confirm" == "y" || "$confirm" == "Y" ]] || { echo "已取消"; exit 0; }
    echo ""

    # ---- Step 1: git 操作 ----
    if [ "$SKIP_GIT" = false ]; then
        step 1 4 "提交未暂存的变更..."
        if [ -n "$(git status --porcelain)" ]; then
            git add -A
            git commit -m "${TAG}"
            success "已提交: ${TAG}"
        else
            info "工作区干净，跳过 commit"
        fi

        step 2 4 "打标签 ${TAG}..."
        if [ "$FORCE_TAG" = true ]; then
            git tag -d "$TAG" 2>/dev/null || true
            git push origin ":refs/tags/$TAG" 2>/dev/null || true
        fi
        git tag -a "$TAG" -m "${TAG}"
        success "已打标签: ${TAG}"

        step 3 4 "推送到 GitHub..."
        git push origin main --tags
        success "推送完成"
    else
        warn "已跳过 git 操作 (--skip-git)"
    fi

    # ---- Step 2: 创建 GitHub Release ----
    TOTAL_STEPS=$( [ "$SKIP_GIT" = true ] && echo 2 || echo 4 )
    step "$( [ "$SKIP_GIT" = true ] && echo 1 || echo 4 )" "$TOTAL_STEPS" "创建 GitHub Release..."

    # 生成 Notes
    if [ -n "$NOTES_FILE" ] && [ -f "$NOTES_FILE" ]; then
        NOTES_CONTENT=$(cat "$NOTES_FILE")
    else
        NOTES_CONTENT=$(generate_notes "$VERSION")
    fi

    # 构建 gh 命令
    GH_ARGS=("release" "create" "$TAG"
        --repo "$REPO"
        --title "${BINARY_NAME} ${TAG}"
        --notes "$NOTES_CONTENT"
    )
    [ "$DRAFT" = true ]      && GH_ARGS+=(--draft)
    [ "$PRERELEASE" = true ] && GH_ARGS+=(--prerelease)

    # 附加产物
    for asset in "${ASSETS[@]}"; do
        GH_ARGS+=("$asset")
    done

    "$GH_BIN" "${GH_ARGS[@]}"

    echo ""
    success "Release 发布完成！"
    echo ""
    echo -e "${CYAN}查看地址:${NC}"
    echo "  https://github.com/${REPO}/releases/tag/${TAG}"
    echo ""
}

main "$@"
