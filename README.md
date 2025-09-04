# 被动式网络资产识别与分析系统

这是一个基于Go语言开发的被动式网络资产识别与分析系统，通过监听网络流量来自动发现和识别网络中的各种资产。

## 功能特性

- **被动流量分析**: 通过镜像端口监听网络流量，不干扰业务系统
- **多协议解析**: 支持ARP、DHCP、HTTP/HTTPS、DNS、SMB、mDNS等协议
- **智能资产识别**: 自动识别设备类型、操作系统、服务版本等信息
- **实时监控**: 7x24小时实时监控，及时发现新资产和状态变化
- **多种存储**: 支持文件、内存、Elasticsearch等多种存储方式
- **灵活部署**: 支持在线监听和离线pcap文件分析

## 系统架构

```
镜像流量 → 流量捕获 → 协议解析 → 特征提取 → 资产关联 → 存储输出
```

## 快速开始

### 安装依赖

```bash
# Ubuntu/Debian
sudo apt-get update
sudo apt-get install -y libpcap-dev build-essential

# CentOS/RHEL
sudo yum install -y libpcap-devel gcc make
```

### 编译程序

```bash
# 进入项目目录
cd /opt/assets_discovery

# 编译
make build

# 或者使用Go直接编译
go build -o build/assets_discovery main.go
```

### 基本使用

#### 1. 实时监听网络接口

```bash
# 查看可用网络接口
sudo ./build/assets_discovery live

# 监听指定接口
sudo ./build/assets_discovery live -i eth0

# 使用配置文件
sudo ./build/assets_discovery live --config config.yaml
```

#### 2. 离线分析pcap文件

```bash
# 分析单个pcap文件
./build/assets_discovery offline -f capture.pcap

# 分析多个文件
./build/assets_discovery offline -f "*.pcap"
```

## 配置说明

主要配置文件 `config.yaml`:

```yaml
# 流量捕获配置
capture:
  interface: "eth0"          # 网络接口名称
  snap_len: 65536           # 捕获包长度
  promiscuous: true         # 混杂模式
  timeout: "30s"            # 超时时间
  buffer_size: 2097152      # 缓冲区大小
  workers: 4                # 工作协程数

# 协议解析配置
parser:
  enabled_protocols:        # 启用的协议
    - "arp"
    - "dhcp"
    - "http"
    - "https"
    - "dns"
    - "smb"
    - "mdns"
  max_packets: 0           # 最大处理包数(0=无限制)
  asset_timeout: 30        # 资产超时时间(分钟)

# 存储配置
storage:
  type: "file"             # 存储类型: file, elasticsearch, memory
  file:
    output_dir: "./output"
    format: "json"
  elasticsearch:
    urls: ["http://localhost:9200"]
    index: "assets"

# 服务配置
server:
  port: 8080
  enabled: true

# 告警配置
alerting:
  enabled: false
  webhook_url: ""
```

## 数据输出格式

系统输出标准JSON格式的资产信息：

```json
{
  "ip_address": "192.168.1.10",
  "mac_address": "00:50:56:12:34:56",
  "hostname": "web-server-01",
  "vendor": "VMware",
  "device_type": "虚拟机",
  "os_guess": "Linux",
  "open_ports": [22, 80, 443],
  "services": {
    "http": "Apache/2.4.41",
    "ssh": "OpenSSH 8.0"
  },
  "first_seen": "2025-01-01T00:00:00Z",
  "last_seen": "2025-01-01T12:00:00Z",
  "is_active": true,
  "confidence": 0.95
}
```

## 支持的协议和识别能力

### 协议解析
- **ARP**: IP-MAC地址映射
- **DHCP**: 主机名、操作系统指纹
- **HTTP/HTTPS**: User-Agent、Server头、SSL证书信息
- **DNS**: 域名解析记录
- **SMB**: Windows网络共享信息
- **mDNS**: 局域网服务发现

### 资产识别
- **厂商识别**: 基于MAC地址OUI数据库
- **操作系统**: Windows、Linux、macOS等
- **设备类型**: 服务器、工作站、虚拟机、网络设备
- **服务识别**: Web服务、数据库、远程管理等

## 部署建议

### 网络部署
1. 将系统部署在核心交换机的镜像端口
2. 确保镜像端口配置正确，能够复制全网流量
3. 使用专用的管理网络进行远程管理

### 性能调优
1. 根据网络流量调整工作协程数量
2. 适当设置缓冲区大小避免丢包
3. 使用SSD存储提高I/O性能
4. 考虑使用Elasticsearch集群提高存储性能

### 安全考虑
1. 系统只解析协议头信息，不存储敏感数据
2. 可配置MAC地址哈希化保护隐私
3. 建议运行在隔离的管理网络中
4. 定期更新指纹库和规则

## 常见问题

### 1. 权限问题
```bash
# 需要root权限或设置capabilities
sudo setcap cap_net_raw,cap_net_admin=eip ./assets_discovery
```

### 2. 找不到网络接口
```bash
# 查看可用接口
ip link show
./assets_discovery live  # 会列出可用接口
```

### 3. 性能问题
- 调整workers数量：建议设置为CPU核心数
- 增大缓冲区：在高流量环境下增大buffer_size
- 使用BPF过滤器：只捕获需要的流量

### 4. 存储问题
- 文件存储：确保有足够的磁盘空间
- Elasticsearch：检查集群状态和索引配置

## 开发和贡献

### 项目结构
```
├── main.go              # 主程序入口
├── cmd/                 # 命令行界面
├── internal/            # 核心业务逻辑
│   ├── capture/        # 流量捕获
│   ├── parser/         # 协议解析
│   ├── assets/         # 资产管理
│   ├── storage/        # 存储层
│   └── config/         # 配置管理
├── config.yaml         # 配置文件示例
├── Makefile           # 构建脚本
└── README.md          # 说明文档
```

### 添加新协议支持
1. 在 `internal/parser/` 中实现协议解析器
2. 在 `internal/assets/` 中添加资产识别规则
3. 更新配置文件中的协议列表
4. 添加相应的测试用例

## 许可证

本项目采用MIT许可证，详见LICENSE文件。

## 联系方式

- 项目地址: https://github.com/your-org/assets_discovery
- 问题反馈: https://github.com/your-org/assets_discovery/issues
- 技术支持: support@example.com
