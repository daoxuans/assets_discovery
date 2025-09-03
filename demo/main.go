package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"time"
)

// 简化版本的资产信息结构
type SimpleAsset struct {
	IPAddress   string                 `json:"ip_address"`
	MACAddress  string                 `json:"mac_address"`
	Hostname    string                 `json:"hostname,omitempty"`
	Vendor      string                 `json:"vendor,omitempty"`
	DeviceType  string                 `json:"device_type,omitempty"`
	OSGuess     string                 `json:"os_guess,omitempty"`
	OpenPorts   []int                  `json:"open_ports,omitempty"`
	Services    map[string]string      `json:"services,omitempty"`
	FirstSeen   time.Time              `json:"first_seen"`
	LastSeen    time.Time              `json:"last_seen"`
	IsActive    bool                   `json:"is_active"`
	Protocols   map[string]interface{} `json:"protocols,omitempty"`
}

// 简化版本的资产管理器
type SimpleAssetManager struct {
	assets map[string]*SimpleAsset
}

func NewSimpleAssetManager() *SimpleAssetManager {
	return &SimpleAssetManager{
		assets: make(map[string]*SimpleAsset),
	}
}

func (sam *SimpleAssetManager) AddAsset(asset *SimpleAsset) {
	key := asset.IPAddress
	if key == "" {
		key = asset.MACAddress
	}
	
	if existing, exists := sam.assets[key]; exists {
		// 更新现有资产
		existing.LastSeen = time.Now()
		if asset.Hostname != "" {
			existing.Hostname = asset.Hostname
		}
		if asset.Vendor != "" {
			existing.Vendor = asset.Vendor
		}
		if len(asset.OpenPorts) > 0 {
			existing.OpenPorts = mergeIntSlices(existing.OpenPorts, asset.OpenPorts)
		}
		if len(asset.Services) > 0 {
			if existing.Services == nil {
				existing.Services = make(map[string]string)
			}
			for k, v := range asset.Services {
				existing.Services[k] = v
			}
		}
	} else {
		// 新资产
		asset.FirstSeen = time.Now()
		asset.LastSeen = time.Now()
		asset.IsActive = true
		sam.assets[key] = asset
		log.Printf("发现新资产: IP=%s, MAC=%s", asset.IPAddress, asset.MACAddress)
	}
}

func (sam *SimpleAssetManager) ExportJSON() ([]byte, error) {
	return json.MarshalIndent(sam.assets, "", "  ")
}

func (sam *SimpleAssetManager) GetStats() map[string]interface{} {
	stats := map[string]interface{}{
		"total_assets": len(sam.assets),
		"active_assets": 0,
		"device_types": make(map[string]int),
		"os_distribution": make(map[string]int),
		"vendor_distribution": make(map[string]int),
	}
	
	deviceTypes := stats["device_types"].(map[string]int)
	osDistribution := stats["os_distribution"].(map[string]int)
	vendorDistribution := stats["vendor_distribution"].(map[string]int)
	
	for _, asset := range sam.assets {
		if asset.IsActive {
			stats["active_assets"] = stats["active_assets"].(int) + 1
		}
		
		if asset.DeviceType != "" {
			deviceTypes[asset.DeviceType]++
		}
		
		if asset.OSGuess != "" {
			osDistribution[asset.OSGuess]++
		}
		
		if asset.Vendor != "" {
			vendorDistribution[asset.Vendor]++
		}
	}
	
	return stats
}

// 厂商识别函数
func getVendorFromMAC(macStr string) string {
	vendors := map[string]string{
		"00:50:56": "VMware",
		"00:0c:29": "VMware", 
		"08:00:27": "VirtualBox",
		"00:15:5d": "Microsoft Hyper-V",
		"52:54:00": "QEMU/KVM",
		"00:16:3e": "Xen",
		"ec:f4:bb": "NetApp",
		"d4:be:d9": "Dell",
		"98:90:96": "Foxconn",
		"a4:bb:6d": "Intel",
		"00:1b:21": "Intel",
	}
	
	if len(macStr) >= 8 {
		oui := macStr[:8]
		if vendor, ok := vendors[oui]; ok {
			return vendor
		}
	}
	
	return ""
}

// 设备类型分类
func classifyDeviceType(vendor, osGuess string, ports []int) string {
	// 基于厂商判断
	switch vendor {
	case "VMware", "VirtualBox", "Microsoft Hyper-V", "QEMU/KVM", "Xen":
		return "虚拟机"
	}
	
	// 基于端口判断
	hasWebPorts := false
	hasServerPorts := false
	
	for _, port := range ports {
		switch port {
		case 80, 443, 8080, 8443:
			hasWebPorts = true
		case 22, 23, 3389, 21, 25, 53:
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
	switch osGuess {
	case "Linux":
		return "Linux服务器"
	case "Windows":
		return "Windows工作站"
	}
	
	return "未知设备"
}

// 模拟数据生成函数（用于测试）
func generateSampleData() []*SimpleAsset {
	now := time.Now()
	
	return []*SimpleAsset{
		{
			IPAddress:  "192.168.1.10",
			MACAddress: "00:50:56:12:34:56",
			Hostname:   "web-server-01",
			Vendor:     "VMware",
			DeviceType: "虚拟机",
			OSGuess:    "Linux",
			OpenPorts:  []int{22, 80, 443},
			Services:   map[string]string{"http": "Apache/2.4.41", "ssh": "OpenSSH 8.0"},
			FirstSeen:  now.Add(-24 * time.Hour),
			LastSeen:   now,
			IsActive:   true,
		},
		{
			IPAddress:  "192.168.1.20",
			MACAddress: "d4:be:d9:aa:bb:cc",
			Hostname:   "workstation-01",
			Vendor:     "Dell",
			DeviceType: "Windows工作站",
			OSGuess:    "Windows",
			OpenPorts:  []int{135, 445, 3389},
			Services:   map[string]string{"rdp": "Terminal Services", "smb": "Windows SMB"},
			FirstSeen:  now.Add(-12 * time.Hour),
			LastSeen:   now.Add(-5 * time.Minute),
			IsActive:   true,
		},
		{
			IPAddress:  "192.168.1.1",
			MACAddress: "a4:bb:6d:11:22:33",
			Hostname:   "router",
			Vendor:     "Intel",
			DeviceType: "网络设备",
			OSGuess:    "Linux",
			OpenPorts:  []int{22, 23, 80, 443},
			Services:   map[string]string{"http": "lighttpd", "ssh": "Dropbear SSH"},
			FirstSeen:  now.Add(-72 * time.Hour),
			LastSeen:   now.Add(-1 * time.Minute),
			IsActive:   true,
		},
	}
}

// 辅助函数
func mergeIntSlices(a, b []int) []int {
	seen := make(map[int]bool)
	result := make([]int, 0)
	
	for _, v := range a {
		if !seen[v] {
			seen[v] = true
			result = append(result, v)
		}
	}
	
	for _, v := range b {
		if !seen[v] {
			seen[v] = true
			result = append(result, v)
		}
	}
	
	return result
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("被动式网络资产识别与分析系统 - 演示版本")
		fmt.Println("")
		fmt.Println("用法:")
		fmt.Println("  demo generate  - 生成示例资产数据")
		fmt.Println("  demo analyze   - 分析本地网络接口")
		fmt.Println("  demo stats     - 显示统计信息")
		os.Exit(1)
	}
	
	command := os.Args[1]
	
	switch command {
	case "generate":
		generateSampleAssets()
	case "analyze":
		analyzeNetworkInterfaces()
	case "stats":
		showStats()
	default:
		fmt.Printf("未知命令: %s\n", command)
		os.Exit(1)
	}
}

func generateSampleAssets() {
	fmt.Println("生成示例资产数据...")
	
	manager := NewSimpleAssetManager()
	
	// 添加示例资产
	for _, asset := range generateSampleData() {
		manager.AddAsset(asset)
	}
	
	// 导出JSON
	data, err := manager.ExportJSON()
	if err != nil {
		log.Fatalf("导出JSON失败: %v", err)
	}
	
	// 保存到文件
	outputFile := "sample_assets.json"
	if err := os.WriteFile(outputFile, data, 0644); err != nil {
		log.Fatalf("保存文件失败: %v", err)
	}
	
	fmt.Printf("示例资产数据已保存到: %s\n", outputFile)
	
	// 显示统计信息
	stats := manager.GetStats()
	fmt.Println("\n统计信息:")
	fmt.Printf("总资产数: %d\n", stats["total_assets"])
	fmt.Printf("活跃资产: %d\n", stats["active_assets"])
	
	if deviceTypes, ok := stats["device_types"].(map[string]int); ok && len(deviceTypes) > 0 {
		fmt.Println("\n设备类型分布:")
		for dtype, count := range deviceTypes {
			fmt.Printf("  %s: %d\n", dtype, count)
		}
	}
}

func analyzeNetworkInterfaces() {
	fmt.Println("分析本地网络接口...")
	
	interfaces, err := net.Interfaces()
	if err != nil {
		log.Fatalf("获取网络接口失败: %v", err)
	}
	
	manager := NewSimpleAssetManager()
	
	for _, iface := range interfaces {
		if iface.Flags&net.FlagLoopback != 0 {
			continue // 跳过回环接口
		}
		
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}
		
		for _, addr := range addrs {
			if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
				if ipnet.IP.To4() != nil { // IPv4地址
					asset := &SimpleAsset{
						IPAddress:  ipnet.IP.String(),
						MACAddress: iface.HardwareAddr.String(),
						Hostname:   getHostname(),
						Vendor:     getVendorFromMAC(iface.HardwareAddr.String()),
						DeviceType: "本地主机",
						OSGuess:    getLocalOS(),
						IsActive:   true,
					}
					
					if asset.DeviceType == "" {
						asset.DeviceType = classifyDeviceType(asset.Vendor, asset.OSGuess, asset.OpenPorts)
					}
					
					manager.AddAsset(asset)
				}
			}
		}
	}
	
	// 导出结果
	data, err := manager.ExportJSON()
	if err != nil {
		log.Fatalf("导出JSON失败: %v", err)
	}
	
	outputFile := "local_assets.json"
	if err := os.WriteFile(outputFile, data, 0644); err != nil {
		log.Fatalf("保存文件失败: %v", err)
	}
	
	fmt.Printf("本地资产信息已保存到: %s\n", outputFile)
}

func showStats() {
	// 尝试读取现有的资产文件
	files := []string{"sample_assets.json", "local_assets.json", "assets.json"}
	
	for _, file := range files {
		if _, err := os.Stat(file); err == nil {
			showStatsFromFile(file)
			return
		}
	}
	
	fmt.Println("未找到资产数据文件，请先运行 'demo generate' 或 'demo analyze'")
}

func showStatsFromFile(filename string) {
	fmt.Printf("从文件读取统计信息: %s\n", filename)
	
	data, err := os.ReadFile(filename)
	if err != nil {
		log.Fatalf("读取文件失败: %v", err)
	}
	
	var assets map[string]*SimpleAsset
	if err := json.Unmarshal(data, &assets); err != nil {
		log.Fatalf("解析JSON失败: %v", err)
	}
	
	// 统计信息
	totalAssets := len(assets)
	activeAssets := 0
	deviceTypes := make(map[string]int)
	osDistribution := make(map[string]int)
	vendorDistribution := make(map[string]int)
	
	for _, asset := range assets {
		if asset.IsActive {
			activeAssets++
		}
		
		if asset.DeviceType != "" {
			deviceTypes[asset.DeviceType]++
		}
		
		if asset.OSGuess != "" {
			osDistribution[asset.OSGuess]++
		}
		
		if asset.Vendor != "" {
			vendorDistribution[asset.Vendor]++
		}
	}
	
	fmt.Printf("\n=== 资产统计报告 ===\n")
	fmt.Printf("总资产数: %d\n", totalAssets)
	fmt.Printf("活跃资产: %d\n", activeAssets)
	
	if len(deviceTypes) > 0 {
		fmt.Println("\n设备类型分布:")
		for dtype, count := range deviceTypes {
			fmt.Printf("  %-15s: %d\n", dtype, count)
		}
	}
	
	if len(osDistribution) > 0 {
		fmt.Println("\n操作系统分布:")
		for os, count := range osDistribution {
			fmt.Printf("  %-15s: %d\n", os, count)
		}
	}
	
	if len(vendorDistribution) > 0 {
		fmt.Println("\n厂商分布:")
		for vendor, count := range vendorDistribution {
			fmt.Printf("  %-15s: %d\n", vendor, count)
		}
	}
	
	fmt.Println("\n最近发现的资产:")
	count := 0
	for _, asset := range assets {
		if count >= 5 {
			break
		}
		fmt.Printf("  %s (%s) - %s\n", asset.IPAddress, asset.MACAddress, asset.DeviceType)
		count++
	}
}

func getHostname() string {
	hostname, err := os.Hostname()
	if err != nil {
		return ""
	}
	return hostname
}

func getLocalOS() string {
	// 简单的操作系统检测
	if _, err := os.Stat("/proc/version"); err == nil {
		return "Linux"
	} else if _, err := os.Stat("/System/Library/CoreServices/SystemVersion.plist"); err == nil {
		return "macOS"
	} else if _, err := os.Stat("C:\\Windows"); err == nil {
		return "Windows"
	}
	return "Unknown"
}
