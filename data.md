# 设备资产信息数据结构

## JSON格式

```json
{
    "ip": "192.168.19.146",
    "mac": "00:0c:29:a9:ae:c6",
    "uid": "7b2adc86-8185-448b-8438-e00284be78be",
    "last_uid": "ccccccaasdaxxx",
    "imei": "046123d9adf",
    "ipv6": "",
    "eid": "112312ssaaa",
    "expand": [
        {
            "ip": "1.1.1.1",
            "ipv6": "",
            "mac": "aa:bb:cc:dd"
        }
    ],
    "expand_type": 0,
    "mid": "adadfadfadfa",
    "device_id": 1,
    "device": "general purpose",
    "priority": 1,
    "os": {
        "os_id": 1,
        "os_product": "mac_os",
        "os_vendor": "microsoft",
        "os_version": "1",
        "os_update": "sp3",
        "os_priority": 1
    },
    "arch": 0,
    "vendor_id": 1,
    "vendor": "VMware",
    "vendor_priority": 1,
    "hostname": "localhost-txu",
    "hardware": "IPC832-IR3-HP40",
    "port": [80],
    "service": [
        {
            "name": "nginx",
            "port": 80,
            "type": "http"
        }
    ],
    "random_mac": 1,
    "sn": "PF3KLNX0",
    "location": {
        "swcode": "90:5d:7c:eb:48:ab",
        "swip": "192.168.11.128",
        "ifindex": 1,
        "ifname": "GigabitEthernet1/0/1",
        "plmn": "46001",
        "tac": "7678466",
        "mcc": "460",
        "mnc": "30"
    },
    "ietc": [
        {
            "proto_name": "ISTP_IEC104",
            "proto_id": 1107,
            "instrument_type": 17,
            "vendor_id": 23,
            "product_type": "SM39CC",
            "product_year": "19",
            "product_month": "04",
            "product_day": "15",
            "product_range": "10",
            "product_seri": "2136",
            "product_reserve": "",
            "device_id": 17,
            "online_status": 1,
            "first_time": 1625824106,
            "latest_status_change_time": 1625824106
        }
    ],
    "useragent": ["chrome 7.0.1"],
    "ssdp": ["android"],
    "dhcp": {
        "vendor": "MSFT 3.0",
        "param_req_list": "1,3,2,34"
    },
    "status": 1,
    "rate": "100%",
    "linkdetect": {
        "avgrtt": "1.234567ms",
        "packetloss": "100%",
        "level": 1
    },
    "correct_detail": {
        "device_type": 1,
        "os": 1,
        "vendor": 1
    },
    "agent": 1
}
```

## 字段说明

### 基础网络信息
| 字段名 | 类型 | 说明 |
|--------|------|------|
| ip | string | 设备IP地址 |
| mac | string | 设备MAC地址 |
| ipv6 | string | 设备IPv6地址 |
| hostname | string | 主机名 |

### 设备标识
| 字段名 | 类型 | 说明 |
|--------|------|------|
| uid | string | 设备唯一ID（注：不同控制器之间未同步）<br/>4101版本从uuid改为uid，避免控制台升级时索引创建失败 |
| last_uid | string | 上次的临时UUID，需要删除 |
| imei | string | 手机IMEI号 |
| eid | string | 28181、35114协议中的摄像头设备ID |
| mid | string | 设备MID标识 |
| sn | string | 设备序列号（8.6.0.4300新增） |

### 扩展网络信息
| 字段名 | 类型 | 说明 |
|--------|------|------|
| expand | array | 多IP/MAC信息数组 |
| expand.ip | string | 扩展IP地址 |
| expand.ipv6 | string | 扩展IPv6地址 |
| expand.mac | string | 扩展MAC地址（必有，不为空） |
| expand_type | int | 扩展类型：0-多网卡，1-多IP（8.6.0.4110新增） |

### 设备分类
| 字段名 | 类型 | 说明 |
|--------|------|------|
| device_id | int | 对应device_type策略中设备类型配置ID |
| device | string | 设备类型，对应device_type设备类型策略name_en关键字 |
| priority | int | 识别优先级：1-摄像头/打点/工控设备，2-匹配MAC规则 |
| hardware | string | 设备型号（摄像头、工控设备） |

### 操作系统信息
| 字段名 | 类型 | 说明 |
|--------|------|------|
| os.os_id | int | 对应os_type策略中操作系统配置ID |
| os.os_product | string | 操作系统产品名 |
| os.os_vendor | string | 操作系统厂商 |
| os.os_version | string | 操作系统版本 |
| os.os_update | string | 操作系统更新包 |
| os.os_priority | int | 操作系统识别优先级：<br/>1-控制台指定，2-控制台DHCP规则，3-控制台UA规则，<br/>4-默认DHCP或UA规则，5-nmap扫描（8.6.0.3000新增） |

### 硬件架构
| 字段名 | 类型 | 说明 |
|--------|------|------|
| arch | int | 芯片架构：0-unknown，1-x86，2-arm，3-loongarch，4-mips<br/>（8.6.0.4320新增loongarch和mips） |

### 厂商信息
| 字段名 | 类型 | 说明 |
|--------|------|------|
| vendor_id | int | MAC厂商识别后匹配vendor_rule中的ID |
| vendor | string | MAC厂商名称，如果vendor_id=99为其他厂商，需提交完整厂商英文名称 |
| vendor_priority | int | 厂商识别优先级：<br/>1-控制台指定，2-SADP，3-摄像头ONVIF，4-摄像头HTTP，<br/>5-摄像头私有协议，6-NSQ消息上报默认，7-OUI默认库（8.6.0.3000新增） |

### 网络服务
| 字段名 | 类型 | 说明 |
|--------|------|------|
| port | array | 开放端口列表 |
| service | array | 运行服务列表 |
| service.name | string | 服务名称 |
| service.port | int | 服务端口 |
| service.type | string | 服务类型 |

### MAC地址特征
| 字段名 | 类型 | 说明 |
|--------|------|------|
| random_mac | int | 是否为随机MAC：0-未知，1-是，2-不是 |

### 位置信息
| 字段名 | 类型 | 说明 |
|--------|------|------|
| location.swcode | string | 交换机代码 |
| location.swip | string | 交换机IP |
| location.ifindex | int | 接口索引 |
| location.ifname | string | 接口名称 |
| location.plmn | string | 基站运营商（46001-中国联通） |
| location.tac | string | 基站位置编号 |
| location.mcc | string | 移动国家代码 |
| location.mnc | string | 移动网络代码 |

### 工控设备信息
| 字段名 | 类型 | 说明 |
|--------|------|------|
| ietc | array | 工控设备信息数组 |
| ietc.proto_name | string | 应用协议名 |
| ietc.proto_id | int | 应用协议ID |
| ietc.instrument_type | int | 仪器仪表类型 |
| ietc.vendor_id | int | 厂家代码 |
| ietc.product_type | string | 产品型号 |
| ietc.product_year | string | 出厂年份 |
| ietc.product_month | string | 出厂月份 |
| ietc.product_day | string | 出厂日期 |
| ietc.product_range | string | 调校量程 |
| ietc.product_seri | string | 产品序号 |
| ietc.product_reserve | string | 预留字段 |
| ietc.device_id | int | 设备代码（仅供参考） |
| ietc.online_status | int | 在线状态：1-在线，0-离线 |
| ietc.first_time | int | 首次上报时间戳 |
| ietc.latest_status_change_time | int | 最近状态变更时间戳 |

### 网络协议信息（已废弃）
| 字段名 | 类型 | 说明 |
|--------|------|------|
| useragent | array | 用户代理信息（8.6.0.3008、8.6.0.3105已废弃） |
| ssdp | array | SSDP协议信息（8.6.0.3008、8.6.0.3105已废弃） |
| dhcp.vendor | string | DHCP厂商信息（8.6.0.4010、8.6.0.3080不再上报） |
| dhcp.param_req_list | string | DHCP指纹信息（8.6.0.4010、8.6.0.3080不再上报） |

### 识别状态
| 字段名 | 类型 | 说明 |
|--------|------|------|
| status | int | 识别状态：1-确认，2-疑似 |
| rate | string | 置信率百分比 |

### 链路检测
| 字段名 | 类型 | 说明 |
|--------|------|------|
| linkdetect.avgrtt | string | 平均往返时间（毫秒） |
| linkdetect.packetloss | string | 丢包率百分比 |
| linkdetect.level | int | 链路质量等级：<br/>0-未知，1-优秀，2-良好，3-普通，4-较差，5-极差，6-断开 |

### 校准状态
| 字段名 | 类型 | 说明 |
|--------|------|------|
| correct_detail.device_type | int | 设备类型是否被校准：0-否，1-是 |
| correct_detail.os | int | 操作系统是否被校准：0-否，1-是 |
| correct_detail.vendor | int | 厂商是否被校准：0-否，1-是 |

### 安全代理
| 字段名 | 类型 | 说明 |
|--------|------|------|
| agent | int | 安装代理类型：<br/>0-无代理，1-天擎，2-天机，3-信创终端，4-工控，5-准入 |