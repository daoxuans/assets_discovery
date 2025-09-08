package capture

import (
	"fmt"
	"log"
	"sync"

	"github.com/google/gopacket"
	"github.com/google/gopacket/pcap"

	"assets_discovery/internal/assets"
	"assets_discovery/internal/config"
	"assets_discovery/internal/parser"
	"assets_discovery/internal/storage"
)

// CaptureEngine 流量捕获引擎
type CaptureEngine struct {
	config       *config.Config
	parser       *parser.PacketParser
	assetManager *assets.AssetManager
	storage      storage.Storage
	wg           sync.WaitGroup
	stopCh       chan struct{}
}

// NewCaptureEngine 创建新的捕获引擎
func NewCaptureEngine(cfg *config.Config) *CaptureEngine {
	// 初始化存储
	var stor storage.Storage
	var err error

	switch cfg.Storage.Type {
	case "elasticsearch":
		stor, err = storage.NewElasticsearchStorage(&cfg.Storage.Elasticsearch)
	case "file":
		stor, err = storage.NewFileStorage(&cfg.Storage.File)
	default:
		stor = storage.NewMemoryStorage()
	}

	if err != nil {
		log.Printf("初始化存储失败，使用内存存储: %v", err)
		stor = storage.NewMemoryStorage()
	}

	assetMgr := assets.NewAssetManager(cfg, stor)

	return &CaptureEngine{
		config:       cfg,
		parser:       parser.NewPacketParser(cfg),
		assetManager: assetMgr,
		storage:      stor,
		stopCh:       make(chan struct{}),
	}
}

// StartLiveCapture 开始实时流量捕获
func (ce *CaptureEngine) StartLiveCapture() error {
	if ce.config.Capture.Interface == "" {
		// 如果没有指定接口，列出可用接口
		return ce.listInterfaces()
	}

	log.Printf("开始监听网络接口: %s", ce.config.Capture.Interface)

	// 打开网络接口
	handle, err := pcap.OpenLive(
		ce.config.Capture.Interface,
		int32(ce.config.Capture.SnapLen),
		ce.config.Capture.Promiscuous,
		ce.config.Capture.Timeout,
	)
	if err != nil {
		return fmt.Errorf("打开网络接口失败: %v", err)
	}
	defer handle.Close()

	// 设置BPF过滤器（可选）
	if err := ce.setBPFFilter(handle); err != nil {
		log.Printf("设置BPF过滤器失败: %v", err)
	}

	// 启动资产管理器
	ce.assetManager.Start()
	defer ce.assetManager.Stop()

	// 启动数据包处理
	return ce.processPackets(handle)
}

// StartOfflineCapture 开始离线pcap文件分析
func (ce *CaptureEngine) StartOfflineCapture(pcapFile string) error {
	log.Printf("开始分析pcap文件: %s", pcapFile)

	// 打开pcap文件
	handle, err := pcap.OpenOffline(pcapFile)
	if err != nil {
		return fmt.Errorf("打开pcap文件失败: %v", err)
	}
	defer handle.Close()

	// 启动资产管理器
	ce.assetManager.Start()
	defer ce.assetManager.Stop()

	// 处理数据包
	return ce.processPackets(handle)
}

// processPackets 处理数据包
func (ce *CaptureEngine) processPackets(handle *pcap.Handle) error {
	packetSource := gopacket.NewPacketSource(handle, handle.LinkType())
	packetChan := packetSource.Packets()

	// 启动多个工作协程处理数据包
	for i := 0; i < ce.config.Capture.Workers; i++ {
		ce.wg.Add(1)
		go ce.packetWorker(packetChan)
	}

	log.Printf("流量捕获已启动，使用 %d 个工作协程", ce.config.Capture.Workers)

	// 等待停止信号或工作协程结束
	<-ce.stopCh
	log.Println("收到停止信号")

	ce.wg.Wait()
	log.Println("流量捕获已停止")
	return nil
}

// packetWorker 数据包处理工作协程
func (ce *CaptureEngine) packetWorker(packetChan chan gopacket.Packet) {
	defer ce.wg.Done()

	packetsProcessed := 0
	for {
		select {
		case packet, ok := <-packetChan:
			if !ok {
				log.Printf("数据包通道已关闭，工作协程退出. 已处理 %d 个数据包", packetsProcessed)
				return
			}

			// 解析数据包
			if assetInfo := ce.parser.ParsePacket(packet); assetInfo != nil {
				// 更新资产信息
				ce.assetManager.UpdateAsset(assetInfo)
			}

			packetsProcessed++

			// 检查是否达到最大处理包数
			if ce.config.Parser.MaxPackets > 0 && packetsProcessed >= ce.config.Parser.MaxPackets {
				log.Printf("已达到最大处理包数 %d，工作协程退出", ce.config.Parser.MaxPackets)
				return
			}

		case <-ce.stopCh:
			log.Printf("工作协程收到停止信号，已处理 %d 个数据包", packetsProcessed)
			return
		}
	}
}

// Stop 停止捕获
func (ce *CaptureEngine) Stop() {
	close(ce.stopCh)
}

// listInterfaces 列出可用的网络接口
func (ce *CaptureEngine) listInterfaces() error {
	devices, err := pcap.FindAllDevs()
	if err != nil {
		return fmt.Errorf("获取网络接口列表失败: %v", err)
	}

	if len(devices) == 0 {
		return fmt.Errorf("未找到可用的网络接口")
	}

	fmt.Println("可用的网络接口:")
	for _, device := range devices {
		fmt.Printf("  %s", device.Name)
		if device.Description != "" {
			fmt.Printf(" (%s)", device.Description)
		}
		fmt.Println()

		for _, addr := range device.Addresses {
			fmt.Printf("    IP: %s", addr.IP)
			if addr.Netmask != nil {
				fmt.Printf("/%s", addr.Netmask)
			}
			fmt.Println()
		}
	}

	fmt.Println("\n请使用 -i 参数指定网络接口，例如:")
	fmt.Printf("  %s live -i %s\n", "assets_discovery", devices[0].Name)

	return nil
}

// setBPFFilter 设置BPF过滤器
func (ce *CaptureEngine) setBPFFilter(handle *pcap.Handle) error {
	// 构建BPF过滤器，只捕获我们关心的协议
	filters := []string{}

	for _, protocol := range ce.config.Parser.EnabledProtocols {
		switch protocol {
		case "arp":
			filters = append(filters, "arp")
		case "dhcp":
			filters = append(filters, "port 67 or port 68")
		case "dns":
			filters = append(filters, "port 53")
		case "http":
			filters = append(filters, "port 80")
		case "https":
			filters = append(filters, "port 443")
		case "smb":
			filters = append(filters, "port 445 or port 139")
		case "mdns":
			filters = append(filters, "port 5353")
		}
	}

	if len(filters) > 0 {
		// 如果有过滤器，就设置；否则捕获所有流量
		filter := fmt.Sprintf("(%s)", joinFilters(filters))
		log.Printf("设置BPF过滤器: %s", filter)
		return handle.SetBPFFilter(filter)
	}

	return nil
}

// joinFilters 连接过滤器
func joinFilters(filters []string) string {
	if len(filters) == 0 {
		return ""
	}
	if len(filters) == 1 {
		return filters[0]
	}

	result := filters[0]
	for i := 1; i < len(filters); i++ {
		result += " or " + filters[i]
	}
	return result
}
