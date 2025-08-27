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

## 安装说明

### 依赖要求

- Go 1.21+
- libpcap开发库
- 可选：Elasticsearch（用于高级存储和搜索）

### 编译安装

```bash
# 1. 克隆代码
cd /opt/assets_discovery/src

# 2. 安装依赖
go mod tidy

# 3. 编译
go build -o assets_discovery main.go

# 4. 安装libpcap（Ubuntu/Debian）
sudo apt-get install libpcap-dev

# 或者（CentOS/RHEL）
sudo yum install libpcap-devel
```

## 使用方法

### 1. 基本使用

```bash
# 列出可用的网络接口
sudo ./assets_discovery live

# 监听指定网络接口
sudo ./assets_discovery live -i eth0

# 分析离线pcap文件
./assets_discovery offline -f capture.pcap
```

### 2. 配置文件

创建配置文件 `config.yaml`：

```yaml
capture:
  interface: "eth0"
  workers: 4

storage:
  type: "file"
  file:
    output_dir: "./output"
    format: "json"

parser:
  enabled_protocols:
    - "arp"
    - "dhcp"
    - "http"
    - "dns"
```

使用配置文件：

```bash
./assets_discovery live --config config.yaml
```

### 3. 高级功能

#### Elasticsearch存储

```yaml
storage:
  type: "elasticsearch"
  elasticsearch:
    urls:
      - "http://localhost:9200"
    index: "network_assets"
```

#### 告警配置

```yaml
alerting:
  enabled: true
  webhook_url: "http://your-webhook-url"
  alert_rules:
    - "new_asset"
    - "unknown_device"
```

## 输出格式

系统会生成包含以下信息的资产记录：

```json
{
  "id": "mac_aa:bb:cc:dd:ee:ff",
  "ip_address": "192.168.1.100",
  "mac_address": "aa:bb:cc:dd:ee:ff",
  "hostname": "workstation-01",
  "vendor": "Dell Inc.",
  "device_type": "工作站",
  "os_info": {
    "family": "Windows",
    "version": "Windows 10",
    "confidence": 0.85
  },
  "open_ports": [
    {
      "port": 80,
      "protocol": "tcp",
      "service": "HTTP",
      "state": "open"
    }
  ],
  "services": [
    {
      "name": "http",
      "version": "Apache/2.4.41",
      "port": 80
    }
  ],
  "first_seen": "2025-01-20T10:00:00Z",
  "last_seen": "2025-01-20T15:30:00Z",
  "is_active": true,
  "confidence": 0.92
}
```

## 支持的协议

- **ARP**: 获取IP-MAC地址映射
- **DHCP**: 提取主机名、操作系统指纹
- **HTTP/HTTPS**: 分析User-Agent、Server头、SSL证书
- **DNS**: 域名解析记录
- **SMB**: Windows网络共享信息
- **mDNS**: 局域网服务发现

## 安全考虑

1. **权限要求**: 监听网络接口需要root权限
2. **数据隐私**: 系统仅提取协议头信息，不存储应用层敏感数据
3. **合规使用**: 确保获得网络监控授权

## 性能优化

- 使用多个工作协程并行处理数据包
- BPF过滤器减少不必要的数据包处理
- 支持PF_RING等高性能抓包技术
- 异步存储避免阻塞处理流程

## 故障排查

### 常见问题

1. **权限不足**
   ```bash
   sudo setcap cap_net_raw,cap_net_admin=eip ./assets_discovery
   ```

2. **找不到网络接口**
   ```bash
   ip link show
   ```

3. **pcap文件格式错误**
   确保pcap文件是标准格式，可用tcpdump或Wireshark验证

### 调试模式

```bash
# 启用详细日志
./assets_discovery live -i eth0 --verbose

# 限制处理包数量进行测试
./assets_discovery offline -f test.pcap --max-packets 1000
```

## 扩展开发

系统采用模块化设计，支持以下扩展：

- 添加新的协议解析器
- 实现自定义存储后端
- 扩展资产识别规则
- 集成外部威胁情报

## 许可证

本项目基于MIT许可证开源。

## 贡献

欢迎提交Issue和Pull Request来改进项目。

## 联系方式

如有问题请联系项目维护者。
