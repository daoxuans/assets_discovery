# 被动式网络资产识别与分析系统 - 使用指南

## 快速上手指南

### 第一步：环境准备

1. **系统要求**
   - Linux系统 (Ubuntu 18.04+, CentOS 7+)
   - Go 1.21+ 
   - libpcap开发库
   - 网络监听权限

2. **检查当前环境**
   ```bash
   # 查看当前目录结构
   ls -la /opt/assets_discovery/
   ```

### 第二步：安装系统

选择以下任一方式安装：

#### 方式1：自动安装(推荐)
```bash
cd /opt/assets_discovery/src
./install.sh
```

#### 方式2：手动编译
```bash
cd /opt/assets_discovery/src

# 安装依赖
sudo apt-get install libpcap-dev  # Ubuntu/Debian
# 或
sudo yum install libpcap-devel    # CentOS/RHEL

# 编译项目
make build
```

### 第三步：配置系统

1. **创建配置文件**
   ```bash
   cp config.yaml assets_discovery.yaml
   ```

2. **修改配置文件**
   ```yaml
   # 编辑配置文件
   vim assets_discovery.yaml
   
   # 关键配置项
   capture:
     interface: "eth0"        # 指定监听接口
     workers: 4               # 工作协程数
   
   storage:
     type: "file"             # 存储类型
     file:
       output_dir: "./output"
   ```

### 第四步：运行系统

#### 实时监听模式
```bash
# 列出可用网络接口
sudo ./build/assets_discovery live

# 监听指定接口
sudo ./build/assets_discovery live -i eth0

# 使用配置文件
sudo ./build/assets_discovery live --config assets_discovery.yaml
```

#### 离线分析模式  
```bash
# 分析pcap文件
./build/assets_discovery offline -f capture.pcap

# 批量分析多个文件
for file in *.pcap; do
    ./build/assets_discovery offline -f "$file"
done
```

## 实际使用场景

### 场景1：企业网络资产盘点

1. **部署位置**: 核心交换机镜像端口
2. **配置要点**:
   ```yaml
   capture:
     interface: "eth0"
     promiscuous: true
     workers: 8
   
   parser:
     enabled_protocols:
       - "arp"
       - "dhcp" 
       - "http"
       - "dns"
   
   storage:
     type: "elasticsearch"
     elasticsearch:
       urls: ["http://es-cluster:9200"]
       index: "network_assets"
   ```

3. **运行命令**:
   ```bash
   sudo ./assets_discovery live --config production.yaml
   ```

### 场景2：安全事件调查

1. **准备pcap文件**: 从SIEM或网络设备导出
2. **快速分析**:
   ```bash
   # 分析可疑时间段的流量
   ./assets_discovery offline -f incident_20250120.pcap
   
   # 查看发现的新资产
   jq '.[] | select(.is_active == true)' output/assets.json
   ```

### 场景3：网络监控集成

1. **定期扫描**:
   ```bash
   # 创建定时任务
   crontab -e
   
   # 每小时执行一次资产发现
   0 * * * * /opt/assets_discovery/build/assets_discovery live -i eth0 --config /etc/assets_discovery.yaml
   ```

2. **告警集成**:
   ```yaml
   alerting:
     enabled: true
     webhook_url: "http://siem.company.com/webhook"
     alert_rules:
       - "new_asset"
       - "unknown_device"
   ```

## 输出数据说明

### 资产记录格式
```json
{
  "id": "mac_00:50:56:12:34:56",
  "ip_address": "192.168.1.100",
  "mac_address": "00:50:56:12:34:56", 
  "hostname": "web-server-01",
  "vendor": "VMware",
  "device_type": "虚拟机",
  "os_info": {
    "family": "Linux",
    "version": "Ubuntu 20.04",
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
      "name": "apache",
      "version": "2.4.41",
      "port": 80
    }
  ],
  "first_seen": "2025-01-20T10:00:00Z",
  "last_seen": "2025-01-20T15:30:00Z", 
  "is_active": true,
  "confidence": 0.92
}
```

### 数据字段说明

| 字段 | 类型 | 说明 |
|------|------|------|
| `id` | string | 资产唯一标识符 |
| `ip_address` | string | IP地址 |
| `mac_address` | string | MAC地址 |
| `hostname` | string | 主机名 |
| `vendor` | string | 设备厂商 |
| `device_type` | string | 设备类型 |
| `os_info` | object | 操作系统信息 |
| `open_ports` | array | 开放端口列表 |
| `services` | array | 运行服务列表 |
| `first_seen` | datetime | 首次发现时间 |
| `last_seen` | datetime | 最后活跃时间 |
| `is_active` | boolean | 是否活跃 |
| `confidence` | float | 识别置信度 |

## 常见问题解决

### 权限问题
```bash
# 问题：Permission denied
# 解决：设置网络监听权限
sudo setcap cap_net_raw,cap_net_admin=eip ./assets_discovery

# 或者使用sudo运行
sudo ./assets_discovery live -i eth0
```

### 网络接口问题
```bash
# 问题：找不到网络接口
# 解决：列出可用接口
ip link show
./assets_discovery live  # 会显示可用接口列表
```

### 性能优化
```bash
# 问题：处理性能不够
# 解决：调整配置参数
capture:
  workers: 8              # 增加工作协程
  buffer_size: 4194304    # 增大缓冲区
  
parser:
  enabled_protocols:      # 只启用必要协议
    - "arp"
    - "dhcp"
```

### 存储空间
```bash
# 问题：存储空间不足
# 解决：定期清理或使用Elasticsearch
storage:
  type: "elasticsearch"
  # 或设置日志轮转
  file:
    max_size: "100MB"
    max_age: "7d"
```

## 监控和运维

### 日志查看
```bash
# 查看系统日志
tail -f /var/log/assets_discovery.log

# 查看资产发现日志
grep "发现新资产" /var/log/assets_discovery.log
```

### 性能监控
```bash
# 查看处理统计
curl http://localhost:8080/api/stats

# 查看内存使用
ps aux | grep assets_discovery

# 查看网络统计
iftop -i eth0
```

### 数据备份
```bash
# 备份资产数据
cp output/assets.json backup/assets_$(date +%Y%m%d).json

# 导出Elasticsearch数据
curl -X GET "localhost:9200/assets/_search?pretty" > backup/es_assets.json
```

## 扩展开发

### 添加新协议解析
1. 在 `internal/parser/` 中添加新的解析函数
2. 更新配置文件中的 `enabled_protocols`
3. 重新编译和测试

### 自定义存储后端
1. 实现 `storage.Storage` 接口
2. 在 `capture.go` 中添加新的存储类型
3. 更新配置文件格式

### 集成外部系统
1. 使用HTTP API导出数据
2. 配置Webhook告警
3. 集成SIEM和CMDB系统

## 技术支持

如遇到问题，请：
1. 查看日志文件获取详细错误信息
2. 检查配置文件格式和权限设置  
3. 参考README.md和本指南
4. 在GitHub项目页面提交Issue

---

*系统已经过充分测试，可以在生产环境中安全使用。*
