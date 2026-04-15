#!/bin/bash
# =============================================================================
# 猪猪记账 - 版本发布脚本
# 用于版本号管理、构建和发布
# =============================================================================

set -e  # 遇到错误立即退出

# 颜色定义
BLUE='\033[0;34m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# 项目路径
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
cd "$SCRIPT_DIR"

WAILS_JSON="wails.json"
PACKAGE_JSON="frontend/package.json"
PROFILE_VIEW="frontend/src/views/ProfileView.vue"
ABOUT_MODAL="frontend/src/components/modals/AboutModal.vue"
MAKEFILE="Makefile"

# 获取当前版本 - 优先从wails.json获取
get_current_version() {
    python3 -c "import json; print(json.load(open('$WAILS_JSON'))['info']['productVersion'])" 2>/dev/null || \
    grep -E "^VERSION\s*\?=" Makefile | cut -d'=' -f2 | tr -d ' '
}

# 更新版本号
update_version() {
    local new_version=$1
    local current_version=$(get_current_version)
    
    if [ -z "$current_version" ] || [ -z "$new_version" ]; then
        echo -e "${RED}错误: 无法获取当前版本或未指定新版本${NC}"
        exit 1
    fi
    
    echo -e "${BLUE}更新版本号: $current_version -> $new_version${NC}"
    
    # 更新 wails.json
    echo -e "${YELLOW}[1/5]${NC} 更新 wails.json"
    python3 -c "
import json
with open('$WAILS_JSON', 'r') as f:
    data = json.load(f)
data['info']['productVersion'] = '$new_version'
with open('$WAILS_JSON', 'w') as f:
    json.dump(data, f, indent=2, ensure_ascii=False)
    f.write('\n')
"
    
    # 更新 package.json
    echo -e "${YELLOW}[2/5]${NC} 更新 package.json"
    python3 -c "
import json
with open('$PACKAGE_JSON', 'r') as f:
    data = json.load(f)
data['version'] = '$new_version'
with open('$PACKAGE_JSON', 'w') as f:
    json.dump(data, f, indent=2, ensure_ascii=False)
    f.write('\n')
"
    
    # 更新 Makefile
    if [ -f "$MAKEFILE" ]; then
        echo -e "${YELLOW}[3/5]${NC} 更新 Makefile"
        sed -i.bak "s/^VERSION\s*?=\s*${current_version}$/VERSION ?= ${new_version}/" Makefile && rm Makefile.bak
    else
        echo -e "${YELLOW}[3/5]${NC} 跳过 Makefile (文件不存在)"
    fi
    
    # 更新 Go 代码中的版本变量（如果存在）
    if [ -f "main.go" ]; then
        sed -i.bak "s/version = \"${current_version}\"/version = \"${new_version}\"/" main.go && rm main.go.bak
    elif [ -f "app.go" ]; then
        sed -i.bak "s/version = \"${current_version}\"/version = \"${new_version}\"/" app.go && rm app.go.bak
    fi
    
    # 更新前端组件
    if [ -f "$PROFILE_VIEW" ]; then
        echo -e "${YELLOW}[4/5]${NC} 更新 ProfileView.vue"
        sed -i.bak "s/v${current_version}/v${new_version}/g" "$PROFILE_VIEW" && rm "$PROFILE_VIEW.bak"
    fi
    
    if [ -f "$ABOUT_MODAL" ]; then
        echo -e "${YELLOW}[5/5]${NC} 更新 AboutModal.vue"
        sed -i.bak "s/v${current_version}/v${new_version}/g" "$ABOUT_MODAL" && rm "$ABOUT_MODAL.bak"
    fi
    
    echo -e "${GREEN}版本已更新为: $new_version${NC}"
}

# 显示帮助信息
show_help() {
    echo "猪猪记账 - 版本发布脚本"
    echo ""
    echo "用法: $0 [选项] [版本号]"
    echo ""
    echo "选项:"
    echo "  -v, --version <version>    设置新的版本号"
    echo "  -m, --major               递增大版本号 (1.0.0 -> 2.0.0)"
    echo "  -n, --minor               递增次版本号 (1.0.0 -> 1.1.0)"
    echo "  -p, --patch               递增修订版本号 (1.0.0 -> 1.0.1)"
    echo "  -c, --current             显示当前版本号"
    echo "  -b, --build               构建当前版本"
    echo "  -r, --release             构建发布版本（包括所有平台）"
    echo "  -h, --help                显示此帮助信息"
}

# 解析版本号
parse_version() {
    local version=$1
    local major=$(echo "$version" | cut -d. -f1)
    local minor=$(echo "$version" | cut -d. -f2)
    local patch=$(echo "$version" | cut -d. -f3 | cut -d- -f1)
    echo "$major $minor $patch"
}

# 递增版本号
increment_version() {
    local version=$1
    local part=$2
    
    local major=$(echo "$version" | cut -d. -f1)
    local minor=$(echo "$version" | cut -d. -f2)
    local patch=$(echo "$version" | cut -d. -f3 | cut -d- -f1)
    
    case $part in
        "major")
            major=$((major + 1))
            minor=0
            patch=0
            ;;
        "minor")
            minor=$((minor + 1))
            patch=0
            ;;
        "patch")
            patch=$((patch + 1))
            ;;
    esac
    
    echo "$major.$minor.$patch"
}

# 主函数
main() {
    if [ $# -eq 0 ]; then
        # 如果没有参数，显示当前版本
        current_version=$(get_current_version)
        echo -e "${CYAN}猪猪记账 - 版本发布工具${NC}"
        echo -e "当前版本: ${GREEN}v${current_version}${NC}"
        echo ""
        echo "使用 --help 查看可用选项"
        exit 0
    fi

    while [[ $# -gt 0 ]]; do
        key="$1"
        
        case $key in
            -v|--version)
                NEW_VERSION="$2"
                shift
                shift
                ;;
            -m|--major)
                UPDATE_MAJOR="true"
                shift
                ;;
            -n|--minor)
                UPDATE_MINOR="true"
                shift
                ;;
            -p|--patch)
                UPDATE_PATCH="true"
                shift
                ;;
            -c|--current)
                CURRENT_ONLY="true"
                shift
                ;;
            -b|--build)
                BUILD_CURRENT="true"
                shift
                ;;
            -r|--release)
                BUILD_RELEASE="true"
                shift
                ;;
            -h|--help)
                show_help
                exit 0
                ;;
            *)
                # 如果是纯数字版本参数，当作版本号处理
                if [[ $1 =~ ^[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
                    NEW_VERSION="$1"
                    shift
                else
                    echo -e "${RED}未知参数: $1${NC}"
                    show_help
                    exit 1
                fi
                ;;
        esac
    done

    # 处理版本显示
    if [ "$CURRENT_ONLY" = "true" ]; then
        current_version=$(get_current_version)
        echo -e "${GREEN}当前版本: $current_version${NC}"
        exit 0
    fi

    # 处理版本更新
    if [ -n "$NEW_VERSION" ]; then
        # 校验版本格式
        if ! echo "$NEW_VERSION" | grep -qE '^[0-9]+\.[0-9]+\.[0-9]+$'; then
            echo -e "${RED}错误: 版本号格式不正确，请使用 x.y.z 格式 (如 1.0.0)${NC}"
            exit 1
        fi
        
        # 校验版本是否递增
        current_version=$(get_current_version)
        if [ "$NEW_VERSION" = "$current_version" ]; then
            echo -e "${RED}错误: 新版本号与当前版本相同 (${NEW_VERSION})${NC}"
            exit 1
        fi

        if [ "$(printf '%s\n' "$NEW_VERSION" "$current_version" | sort -V | head -n1)" = "$NEW_VERSION" ]; then
            echo -e "${RED}错误: 新版本号 (${NEW_VERSION}) 不能低于当前版本 (${current_version})${NC}"
            exit 1
        fi
        
        update_version "$NEW_VERSION"
    elif [ "$UPDATE_MAJOR" = "true" ] || [ "$UPDATE_MINOR" = "true" ] || [ "$UPDATE_PATCH" = "true" ]; then
        current_version=$(get_current_version)
        if [ "$UPDATE_MAJOR" = "true" ]; then
            new_version=$(increment_version "$current_version" "major")
        elif [ "$UPDATE_MINOR" = "true" ]; then
            new_version=$(increment_version "$current_version" "minor")
        elif [ "$UPDATE_PATCH" = "true" ]; then
            new_version=$(increment_version "$current_version" "patch")
        fi
        update_version "$new_version"
    fi

    # 处理构建
    if [ "$BUILD_CURRENT" = "true" ]; then
        echo -e "${BLUE}开始构建当前版本...${NC}"
        make build
    fi

    if [ "$BUILD_RELEASE" = "true" ]; then
        echo -e "${BLUE}开始构建发布版本...${NC}"
        make build-release
    fi
}

# 调用主函数，传入所有参数
main "$@"
