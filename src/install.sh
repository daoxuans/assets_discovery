#!/bin/bash

# 被动式网络资产识别与分析系统安装脚本
# Assets Discovery System Installation Script

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 打印函数
print_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# 检查是否为root用户
check_root() {
    if [[ $EUID -eq 0 ]]; then
        print_warning "建议不要以root用户运行安装脚本"
        print_info "系统将在需要时提示输入sudo密码"
    fi
}

# 检测操作系统
detect_os() {
    if [[ -f /etc/os-release ]]; then
        . /etc/os-release
        OS=$NAME
        VER=$VERSION_ID
    elif type lsb_release >/dev/null 2>&1; then
        OS=$(lsb_release -si)
        VER=$(lsb_release -sr)
    elif [[ -f /etc/redhat-release ]]; then
        OS="Red Hat"
        VER=$(cat /etc/redhat-release | grep -oE '[0-9]+\.[0-9]+')
    else
        OS=$(uname -s)
        VER=$(uname -r)
    fi
    
    print_info "检测到操作系统: $OS $VER"
}

# 检查Go环境
check_go() {
    if command -v go &> /dev/null; then
        GO_VERSION=$(go version | grep -oE 'go[0-9]+\.[0-9]+\.[0-9]+')
        print_success "Go已安装: $GO_VERSION"
        
        # 检查Go版本是否满足要求
        GO_MAJOR=$(echo $GO_VERSION | cut -d'.' -f1 | cut -d'o' -f2)
        GO_MINOR=$(echo $GO_VERSION | cut -d'.' -f2)
        
        if [[ $GO_MAJOR -gt 1 ]] || [[ $GO_MAJOR -eq 1 && $GO_MINOR -ge 21 ]]; then
            print_success "Go版本满足要求 (>= 1.21)"
        else
            print_error "Go版本过低，需要 >= 1.21"
            install_go
        fi
    else
        print_warning "未检测到Go环境"
        install_go
    fi
}

# 安装Go
install_go() {
    print_info "开始安装Go..."
    
    GO_VERSION="1.21.5"
    GO_TARBALL="go${GO_VERSION}.linux-amd64.tar.gz"
    GO_URL="https://golang.org/dl/${GO_TARBALL}"
    
    # 下载Go
    print_info "下载Go ${GO_VERSION}..."
    wget -O "/tmp/${GO_TARBALL}" "${GO_URL}"
    
    # 安装Go
    print_info "安装Go到 /usr/local/"
    sudo rm -rf /usr/local/go
    sudo tar -C /usr/local -xzf "/tmp/${GO_TARBALL}"
    
    # 设置环境变量
    if ! grep -q "/usr/local/go/bin" ~/.bashrc; then
        echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
        echo 'export GOPATH=$HOME/go' >> ~/.bashrc
        echo 'export PATH=$PATH:$GOPATH/bin' >> ~/.bashrc
    fi
    
    # 重新加载环境变量
    export PATH=$PATH:/usr/local/go/bin
    export GOPATH=$HOME/go
    
    print_success "Go安装完成"
}

# 安装系统依赖
install_system_deps() {
    print_info "安装系统依赖..."
    
    case "$OS" in
        *"Ubuntu"*|*"Debian"*)
            print_info "安装Ubuntu/Debian依赖..."
            sudo apt-get update
            sudo apt-get install -y libpcap-dev build-essential wget curl git
            ;;
        *"CentOS"*|*"Red Hat"*|*"Fedora"*)
            print_info "安装CentOS/RHEL/Fedora依赖..."
            sudo yum install -y libpcap-devel gcc make wget curl git
            ;;
        *"Arch"*)
            print_info "安装Arch Linux依赖..."
            sudo pacman -S libpcap gcc make wget curl git
            ;;
        *)
            print_warning "未识别的操作系统，请手动安装libpcap开发库"
            ;;
    esac
    
    print_success "系统依赖安装完成"
}

# 编译项目
build_project() {
    print_info "开始编译项目..."
    
    # 安装Go依赖
    print_info "安装Go模块依赖..."
    go mod tidy
    
    # 编译项目
    print_info "编译可执行文件..."
    make build
    
    if [[ -f "./build/assets_discovery" ]]; then
        print_success "编译完成: ./build/assets_discovery"
    else
        print_error "编译失败"
        exit 1
    fi
}

# 安装到系统
install_to_system() {
    read -p "是否将程序安装到系统路径 (/usr/local/bin)? [y/N]: " -n 1 -r
    echo
    
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        print_info "安装到系统路径..."
        make install
        print_success "安装完成，现在可以在任何地方使用 'assets_discovery' 命令"
    else
        print_info "程序编译完成，可执行文件位于: ./build/assets_discovery"
    fi
}

# 创建配置文件
create_config() {
    print_info "创建配置文件..."
    
    if [[ ! -f "assets_discovery.yaml" ]]; then
        cp config.yaml assets_discovery.yaml
        print_success "配置文件已创建: assets_discovery.yaml"
        print_info "请根据需要修改配置文件"
    else
        print_info "配置文件已存在: assets_discovery.yaml"
    fi
}

# 设置权限
setup_permissions() {
    read -p "是否设置网络监听权限 (避免每次都需要sudo)? [y/N]: " -n 1 -r
    echo
    
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        print_info "设置网络监听权限..."
        if [[ -f "./build/assets_discovery" ]]; then
            sudo setcap cap_net_raw,cap_net_admin=eip ./build/assets_discovery
            print_success "权限设置完成"
        else
            print_warning "可执行文件不存在，请先编译项目"
        fi
    fi
}

# 显示使用说明
show_usage() {
    print_info "安装完成！"
    echo
    echo "使用方法:"
    echo "  1. 列出网络接口:    ./build/assets_discovery live"
    echo "  2. 监听指定接口:    ./build/assets_discovery live -i eth0"
    echo "  3. 分析pcap文件:    ./build/assets_discovery offline -f capture.pcap"
    echo "  4. 使用配置文件:    ./build/assets_discovery live --config assets_discovery.yaml"
    echo
    echo "注意事项:"
    echo "  - 监听网络接口需要root权限或设置了相应的capabilities"
    echo "  - 修改 assets_discovery.yaml 配置文件以自定义行为"
    echo "  - 查看 README.md 获取详细使用说明"
    echo
}

# 主安装流程
main() {
    print_info "开始安装被动式网络资产识别与分析系统..."
    echo
    
    check_root
    detect_os
    check_go
    install_system_deps
    build_project
    install_to_system
    create_config
    setup_permissions
    show_usage
    
    print_success "安装流程完成！"
}

# 处理命令行参数
case "${1:-}" in
    --help|-h)
        echo "被动式网络资产识别与分析系统安装脚本"
        echo
        echo "用法: $0 [选项]"
        echo
        echo "选项:"
        echo "  --help, -h     显示此帮助信息"
        echo "  --deps-only    仅安装依赖"
        echo "  --build-only   仅编译项目"
        echo
        exit 0
        ;;
    --deps-only)
        detect_os
        check_go
        install_system_deps
        ;;
    --build-only)
        build_project
        ;;
    *)
        main
        ;;
esac
