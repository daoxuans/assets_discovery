package assets

import (
	"sync"
	"time"
)

// AssetInfo 资产信息结构
type AssetInfo struct {
	// 基本信息
	IPAddress  string    `json:"ip_address"`
	MACAddress string    `json:"mac_address"`
	Hostname   string    `json:"hostname"`
	Vendor     string    `json:"vendor"`
	DeviceType string    `json:"device_type"`
	OSGuess    string    `json:"os_guess"`
	Timestamp  time.Time `json:"timestamp"`

	// 网络信息
	OpenPorts []int                  `json:"open_ports"`
	Services  map[string]interface{} `json:"services"`
	Protocols map[string]interface{} `json:"protocols"`

	// 状态信息
	FirstSeen  time.Time `json:"first_seen"`
	LastSeen   time.Time `json:"last_seen"`
	IsActive   bool      `json:"is_active"`
	Confidence float64   `json:"confidence"` // 识别置信度
}

// Asset 完整的资产信息
type Asset struct {
	ID         string `json:"id"`
	IPAddress  string `json:"ip_address"`
	MACAddress string `json:"mac_address"`
	Hostname   string `json:"hostname"`
	Vendor     string `json:"vendor"`
	DeviceType string `json:"device_type"`
	OSInfo     OSInfo `json:"os_info"`

	// 网络服务信息
	OpenPorts []PortInfo    `json:"open_ports"`
	Services  []ServiceInfo `json:"services"`

	// 协议信息
	Protocols map[string]interface{} `json:"protocols"`

	// 统计信息
	FirstSeen  time.Time `json:"first_seen"`
	LastSeen   time.Time `json:"last_seen"`
	LastUpdate time.Time `json:"last_update"`
	IsActive   bool      `json:"is_active"`
	Confidence float64   `json:"confidence"`

	// 变更历史
	Changes []ChangeRecord `json:"changes"`

	mu sync.RWMutex `json:"-"`
}

// OSInfo 操作系统信息
type OSInfo struct {
	Family     string   `json:"family"`     // Windows, Linux, macOS等
	Version    string   `json:"version"`    // 版本号
	Kernel     string   `json:"kernel"`     // 内核版本
	Detection  []string `json:"detection"`  // 检测方法
	Confidence float64  `json:"confidence"` // 识别置信度
}

// PortInfo 端口信息
type PortInfo struct {
	Port      int       `json:"port"`
	Protocol  string    `json:"protocol"` // tcp, udp
	State     string    `json:"state"`    // open, closed, filtered
	Service   string    `json:"service"`  // 服务名称
	Version   string    `json:"version"`  // 服务版本
	Banner    string    `json:"banner"`   // 服务横幅
	FirstSeen time.Time `json:"first_seen"`
	LastSeen  time.Time `json:"last_seen"`
}

// ServiceInfo 服务信息
type ServiceInfo struct {
	Name      string                 `json:"name"`
	Version   string                 `json:"version"`
	Port      int                    `json:"port"`
	Protocol  string                 `json:"protocol"`
	Banner    string                 `json:"banner"`
	Headers   map[string]interface{} `json:"headers"`
	FirstSeen time.Time              `json:"first_seen"`
	LastSeen  time.Time              `json:"last_seen"`
}

// ChangeRecord 变更记录
type ChangeRecord struct {
	Timestamp   time.Time   `json:"timestamp"`
	ChangeType  string      `json:"change_type"` // ip_change, service_change, status_change等
	OldValue    interface{} `json:"old_value"`
	NewValue    interface{} `json:"new_value"`
	Description string      `json:"description"`
}

// NewAsset 创建新资产
func NewAsset(assetInfo *AssetInfo) *Asset {
	now := time.Now()

	asset := &Asset{
		ID:         generateAssetID(assetInfo),
		IPAddress:  assetInfo.IPAddress,
		MACAddress: assetInfo.MACAddress,
		Hostname:   assetInfo.Hostname,
		Vendor:     assetInfo.Vendor,
		DeviceType: classifyDeviceType(assetInfo),
		OSInfo:     extractOSInfo(assetInfo),
		OpenPorts:  convertPorts(assetInfo.OpenPorts),
		Services:   convertServices(assetInfo.Services),
		Protocols:  assetInfo.Protocols,
		FirstSeen:  now,
		LastSeen:   now,
		LastUpdate: now,
		IsActive:   true,
		Confidence: calculateConfidence(assetInfo),
		Changes:    []ChangeRecord{},
	}

	return asset
}

// Update 更新资产信息
func (a *Asset) Update(assetInfo *AssetInfo) {
	a.mu.Lock()
	defer a.mu.Unlock()

	now := time.Now()
	changes := []ChangeRecord{}

	// 检查IP地址变更
	if assetInfo.IPAddress != "" && assetInfo.IPAddress != a.IPAddress {
		changes = append(changes, ChangeRecord{
			Timestamp:   now,
			ChangeType:  "ip_change",
			OldValue:    a.IPAddress,
			NewValue:    assetInfo.IPAddress,
			Description: "IP地址发生变更",
		})
		a.IPAddress = assetInfo.IPAddress
	}

	// 检查主机名变更
	if assetInfo.Hostname != "" && assetInfo.Hostname != a.Hostname {
		changes = append(changes, ChangeRecord{
			Timestamp:   now,
			ChangeType:  "hostname_change",
			OldValue:    a.Hostname,
			NewValue:    assetInfo.Hostname,
			Description: "主机名发生变更",
		})
		a.Hostname = assetInfo.Hostname
	}

	// 更新端口信息
	if len(assetInfo.OpenPorts) > 0 {
		newPorts := convertPorts(assetInfo.OpenPorts)
		if !equalPorts(a.OpenPorts, newPorts) {
			changes = append(changes, ChangeRecord{
				Timestamp:   now,
				ChangeType:  "ports_change",
				OldValue:    a.OpenPorts,
				NewValue:    newPorts,
				Description: "开放端口发生变更",
			})
			a.OpenPorts = mergePorts(a.OpenPorts, newPorts)
		}
	}

	// 更新服务信息
	if len(assetInfo.Services) > 0 {
		newServices := convertServices(assetInfo.Services)
		a.Services = mergeServices(a.Services, newServices)
	}

	// 更新协议信息
	if len(assetInfo.Protocols) > 0 {
		a.Protocols = mergeProtocols(a.Protocols, assetInfo.Protocols)
	}

	// 更新操作系统信息
	if assetInfo.OSGuess != "" {
		newOSInfo := extractOSInfo(assetInfo)
		if newOSInfo.Family != "" && newOSInfo.Family != a.OSInfo.Family {
			changes = append(changes, ChangeRecord{
				Timestamp:   now,
				ChangeType:  "os_change",
				OldValue:    a.OSInfo,
				NewValue:    newOSInfo,
				Description: "操作系统信息发生变更",
			})
			a.OSInfo = mergeOSInfo(a.OSInfo, newOSInfo)
		}
	}

	// 更新设备类型
	if newDeviceType := classifyDeviceType(assetInfo); newDeviceType != "" && newDeviceType != a.DeviceType {
		changes = append(changes, ChangeRecord{
			Timestamp:   now,
			ChangeType:  "device_type_change",
			OldValue:    a.DeviceType,
			NewValue:    newDeviceType,
			Description: "设备类型发生变更",
		})
		a.DeviceType = newDeviceType
	}

	// 添加变更记录
	a.Changes = append(a.Changes, changes...)

	// 更新时间戳
	a.LastSeen = now
	a.LastUpdate = now
	a.IsActive = true

	// 重新计算置信度
	a.Confidence = calculateConfidence(assetInfo)
}

// SetInactive 设置资产为非活跃状态
func (a *Asset) SetInactive() {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.IsActive {
		a.IsActive = false
		a.LastUpdate = time.Now()

		a.Changes = append(a.Changes, ChangeRecord{
			Timestamp:   time.Now(),
			ChangeType:  "status_change",
			OldValue:    true,
			NewValue:    false,
			Description: "资产变为非活跃状态",
		})
	}
}

// GetSummary 获取资产摘要信息
func (a *Asset) GetSummary() map[string]interface{} {
	a.mu.RLock()
	defer a.mu.RUnlock()

	return map[string]interface{}{
		"id":             a.ID,
		"ip_address":     a.IPAddress,
		"mac_address":    a.MACAddress,
		"hostname":       a.Hostname,
		"vendor":         a.Vendor,
		"device_type":    a.DeviceType,
		"os_family":      a.OSInfo.Family,
		"ports_count":    len(a.OpenPorts),
		"services_count": len(a.Services),
		"first_seen":     a.FirstSeen,
		"last_seen":      a.LastSeen,
		"is_active":      a.IsActive,
		"confidence":     a.Confidence,
	}
}

// 辅助函数
func generateAssetID(assetInfo *AssetInfo) string {
	// 使用MAC地址作为主要标识符，如果没有则使用IP地址
	if assetInfo.MACAddress != "" {
		return "mac_" + assetInfo.MACAddress
	}
	if assetInfo.IPAddress != "" {
		return "ip_" + assetInfo.IPAddress
	}
	return "unknown_" + time.Now().Format("20060102150405")
}

func classifyDeviceType(assetInfo *AssetInfo) string {
	// 基于厂商信息判断设备类型
	switch assetInfo.Vendor {
	case "VMware":
		return "虚拟机"
	case "VirtualBox", "QEMU/KVM", "Microsoft Hyper-V", "Xen":
		return "虚拟机"
	}

	// 基于开放端口判断设备类型
	hasWebPorts := false
	hasServerPorts := false

	for _, port := range assetInfo.OpenPorts {
		switch port {
		case 80, 443, 8080, 8443:
			hasWebPorts = true
		case 22, 23, 3389:
			hasServerPorts = true
		case 21, 25, 53, 110, 143:
			hasServerPorts = true
		}
	}

	if hasWebPorts && hasServerPorts {
		return "服务器"
	} else if hasWebPorts {
		return "Web设备"
	} else if hasServerPorts {
		return "服务器"
	}

	// 基于操作系统判断
	switch assetInfo.OSGuess {
	case "Linux/Unix":
		return "服务器"
	case "Windows":
		return "工作站"
	case "Cisco/Network Device":
		return "网络设备"
	}

	return "未知设备"
}

func extractOSInfo(assetInfo *AssetInfo) OSInfo {
	osInfo := OSInfo{
		Family:     assetInfo.OSGuess,
		Detection:  []string{},
		Confidence: 0.5,
	}

	// 从不同来源提取操作系统信息
	if assetInfo.OSGuess != "" {
		osInfo.Detection = append(osInfo.Detection, "ttl_analysis")
	}

	// 从HTTP User-Agent提取
	if protocols, ok := assetInfo.Protocols["http"]; ok {
		if headers, ok := protocols.(map[string]interface{}); ok {
			if _, ok := headers["user-agent"]; ok {
				osInfo.Detection = append(osInfo.Detection, "user_agent")
				// 可以进一步解析User-Agent获取更详细的版本信息
			}
		}
	}

	// 从DHCP信息提取
	if protocols, ok := assetInfo.Protocols["dhcp"]; ok {
		if dhcp, ok := protocols.(map[string]interface{}); ok {
			if vendorClass, ok := dhcp["vendor_class"]; ok {
				osInfo.Detection = append(osInfo.Detection, "dhcp_vendor_class")
				osInfo.Version = vendorClass.(string)
			}
		}
	}

	return osInfo
}

func convertPorts(ports []int) []PortInfo {
	result := make([]PortInfo, 0, len(ports))
	now := time.Now()

	for _, port := range ports {
		result = append(result, PortInfo{
			Port:      port,
			Protocol:  "tcp",
			State:     "open",
			FirstSeen: now,
			LastSeen:  now,
		})
	}

	return result
}

func convertServices(services map[string]interface{}) []ServiceInfo {
	result := make([]ServiceInfo, 0, len(services))
	now := time.Now()

	for name, info := range services {
		serviceInfo := ServiceInfo{
			Name:      name,
			FirstSeen: now,
			LastSeen:  now,
		}

		if infoStr, ok := info.(string); ok {
			serviceInfo.Version = infoStr
		}

		result = append(result, serviceInfo)
	}

	return result
}

func calculateConfidence(assetInfo *AssetInfo) float64 {
	confidence := 0.0

	// 基于可用信息计算置信度
	if assetInfo.MACAddress != "" {
		confidence += 0.3
	}
	if assetInfo.IPAddress != "" {
		confidence += 0.2
	}
	if assetInfo.Hostname != "" {
		confidence += 0.2
	}
	if len(assetInfo.OpenPorts) > 0 {
		confidence += 0.1
	}
	if len(assetInfo.Services) > 0 {
		confidence += 0.1
	}
	if assetInfo.OSGuess != "" {
		confidence += 0.1
	}

	if confidence > 1.0 {
		confidence = 1.0
	}

	return confidence
}

// 合并函数
func equalPorts(a, b []PortInfo) bool {
	if len(a) != len(b) {
		return false
	}

	portMapA := make(map[int]bool)
	for _, port := range a {
		portMapA[port.Port] = true
	}

	for _, port := range b {
		if !portMapA[port.Port] {
			return false
		}
	}

	return true
}

func mergePorts(existing, new []PortInfo) []PortInfo {
	portMap := make(map[int]PortInfo)

	// 添加现有端口
	for _, port := range existing {
		portMap[port.Port] = port
	}

	// 合并新端口
	for _, port := range new {
		if existingPort, exists := portMap[port.Port]; exists {
			existingPort.LastSeen = port.LastSeen
			portMap[port.Port] = existingPort
		} else {
			portMap[port.Port] = port
		}
	}

	// 转换回切片
	result := make([]PortInfo, 0, len(portMap))
	for _, port := range portMap {
		result = append(result, port)
	}

	return result
}

func mergeServices(existing, new []ServiceInfo) []ServiceInfo {
	serviceMap := make(map[string]ServiceInfo)

	// 添加现有服务
	for _, service := range existing {
		serviceMap[service.Name] = service
	}

	// 合并新服务
	for _, service := range new {
		if existingService, exists := serviceMap[service.Name]; exists {
			existingService.LastSeen = service.LastSeen
			if service.Version != "" {
				existingService.Version = service.Version
			}
			serviceMap[service.Name] = existingService
		} else {
			serviceMap[service.Name] = service
		}
	}

	// 转换回切片
	result := make([]ServiceInfo, 0, len(serviceMap))
	for _, service := range serviceMap {
		result = append(result, service)
	}

	return result
}

func mergeProtocols(existing, new map[string]interface{}) map[string]interface{} {
	if existing == nil {
		existing = make(map[string]interface{})
	}

	for key, value := range new {
		existing[key] = value
	}

	return existing
}

func mergeOSInfo(existing, new OSInfo) OSInfo {
	if new.Family != "" {
		existing.Family = new.Family
	}
	if new.Version != "" {
		existing.Version = new.Version
	}
	if new.Kernel != "" {
		existing.Kernel = new.Kernel
	}

	// 合并检测方法
	detectionMap := make(map[string]bool)
	for _, method := range existing.Detection {
		detectionMap[method] = true
	}
	for _, method := range new.Detection {
		detectionMap[method] = true
	}

	detection := make([]string, 0, len(detectionMap))
	for method := range detectionMap {
		detection = append(detection, method)
	}
	existing.Detection = detection

	// 使用较高的置信度
	if new.Confidence > existing.Confidence {
		existing.Confidence = new.Confidence
	}

	return existing
}
