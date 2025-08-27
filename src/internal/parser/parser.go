package parser

import (
	"fmt"
	"net"
	"strings"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"

	"assets_discovery/internal/assets"
	"assets_discovery/internal/config"
)

// PacketParser 数据包解析器
type PacketParser struct {
	config           *config.Config
	enabledProtocols map[string]bool
}

// NewPacketParser 创建新的数据包解析器
func NewPacketParser(cfg *config.Config) *PacketParser {
	enabled := make(map[string]bool)
	for _, protocol := range cfg.Parser.EnabledProtocols {
		enabled[protocol] = true
	}

	return &PacketParser{
		config:           cfg,
		enabledProtocols: enabled,
	}
}

// ParsePacket 解析数据包并提取资产信息
func (pp *PacketParser) ParsePacket(packet gopacket.Packet) *assets.AssetInfo {
	if packet == nil {
		return nil
	}

	assetInfo := &assets.AssetInfo{
		Timestamp: packet.Metadata().Timestamp,
		Protocols: make(map[string]interface{}),
	}

	// 解析以太网层
	if ethLayer := packet.Layer(layers.LayerTypeEthernet); ethLayer != nil {
		eth, _ := ethLayer.(*layers.Ethernet)
		pp.parseEthernet(assetInfo, eth)
	}

	// 解析ARP
	if pp.enabledProtocols["arp"] {
		if arpLayer := packet.Layer(layers.LayerTypeARP); arpLayer != nil {
			arp, _ := arpLayer.(*layers.ARP)
			pp.parseARP(assetInfo, arp)
		}
	}

	// 解析IPv4层
	if ipLayer := packet.Layer(layers.LayerTypeIPv4); ipLayer != nil {
		ip, _ := ipLayer.(*layers.IPv4)
		pp.parseIPv4(assetInfo, ip)

		// 解析TCP层
		if tcpLayer := packet.Layer(layers.LayerTypeTCP); tcpLayer != nil {
			tcp, _ := tcpLayer.(*layers.TCP)
			pp.parseTCP(assetInfo, tcp, packet.ApplicationLayer())
		}

		// 解析UDP层
		if udpLayer := packet.Layer(layers.LayerTypeUDP); udpLayer != nil {
			udp, _ := udpLayer.(*layers.UDP)
			pp.parseUDP(assetInfo, udp, packet.ApplicationLayer())
		}
	}

	// 只返回包含有用信息的资产信息
	if pp.hasUsefulInfo(assetInfo) {
		return assetInfo
	}

	return nil
}

// parseEthernet 解析以太网层
func (pp *PacketParser) parseEthernet(assetInfo *assets.AssetInfo, eth *layers.Ethernet) {
	// 提取源MAC地址
	if !pp.isMulticastMAC(eth.SrcMAC) {
		assetInfo.MACAddress = eth.SrcMAC.String()
		assetInfo.Vendor = pp.getVendorFromMAC(eth.SrcMAC)
	}
}

// parseARP 解析ARP协议
func (pp *PacketParser) parseARP(assetInfo *assets.AssetInfo, arp *layers.ARP) {
	if arp.Operation == layers.ARPRequest || arp.Operation == layers.ARPReply {
		// ARP请求或响应
		srcIP := net.IP(arp.SourceProtAddress).String()
		srcMAC := net.HardwareAddr(arp.SourceHwAddress).String()

		assetInfo.IPAddress = srcIP
		assetInfo.MACAddress = srcMAC
		assetInfo.Vendor = pp.getVendorFromMAC(net.HardwareAddr(arp.SourceHwAddress))

		assetInfo.Protocols["arp"] = map[string]interface{}{
			"operation": arp.Operation,
			"src_ip":    srcIP,
			"src_mac":   srcMAC,
			"dst_ip":    net.IP(arp.DstProtAddress).String(),
			"dst_mac":   net.HardwareAddr(arp.DstHwAddress).String(),
		}
	}
}

// parseIPv4 解析IPv4层
func (pp *PacketParser) parseIPv4(assetInfo *assets.AssetInfo, ip *layers.IPv4) {
	assetInfo.IPAddress = ip.SrcIP.String()

	// 基于TTL值推测操作系统
	assetInfo.OSGuess = pp.guessOSFromTTL(ip.TTL)

	assetInfo.Protocols["ipv4"] = map[string]interface{}{
		"src_ip":   ip.SrcIP.String(),
		"dst_ip":   ip.DstIP.String(),
		"ttl":      ip.TTL,
		"protocol": ip.Protocol,
		"length":   ip.Length,
	}
}

// parseTCP 解析TCP层
func (pp *PacketParser) parseTCP(assetInfo *assets.AssetInfo, tcp *layers.TCP, appLayer gopacket.ApplicationLayer) {
	srcPort := int(tcp.SrcPort)
	dstPort := int(tcp.DstPort)

	// 记录开放的端口
	if tcp.SYN && tcp.ACK {
		// SYN-ACK 响应表示端口开放
		assetInfo.OpenPorts = append(assetInfo.OpenPorts, srcPort)
	}

	// 识别服务
	service := pp.identifyService(srcPort, dstPort, appLayer)
	if service != "" {
		if assetInfo.Services == nil {
			assetInfo.Services = make(map[string]interface{})
		}
		assetInfo.Services[fmt.Sprintf("%d/tcp", srcPort)] = service
	}

	assetInfo.Protocols["tcp"] = map[string]interface{}{
		"src_port": srcPort,
		"dst_port": dstPort,
		"flags": map[string]bool{
			"syn": tcp.SYN,
			"ack": tcp.ACK,
			"fin": tcp.FIN,
			"rst": tcp.RST,
		},
	}

	// 解析HTTP协议
	if pp.enabledProtocols["http"] && (srcPort == 80 || dstPort == 80) && appLayer != nil {
		pp.parseHTTP(assetInfo, appLayer.Payload())
	}
}

// parseUDP 解析UDP层
func (pp *PacketParser) parseUDP(assetInfo *assets.AssetInfo, udp *layers.UDP, appLayer gopacket.ApplicationLayer) {
	srcPort := int(udp.SrcPort)
	dstPort := int(udp.DstPort)

	assetInfo.Protocols["udp"] = map[string]interface{}{
		"src_port": srcPort,
		"dst_port": dstPort,
	}

	// 解析DHCP
	if pp.enabledProtocols["dhcp"] && (srcPort == 67 || srcPort == 68 || dstPort == 67 || dstPort == 68) {
		if appLayer != nil {
			pp.parseDHCP(assetInfo, appLayer.Payload())
		}
	}

	// 解析DNS
	if pp.enabledProtocols["dns"] && (srcPort == 53 || dstPort == 53) {
		if appLayer != nil {
			pp.parseDNS(assetInfo, appLayer.Payload())
		}
	}

	// 解析mDNS
	if pp.enabledProtocols["mdns"] && (srcPort == 5353 || dstPort == 5353) {
		if appLayer != nil {
			pp.parseMDNS(assetInfo, appLayer.Payload())
		}
	}
}

// parseHTTP 解析HTTP协议
func (pp *PacketParser) parseHTTP(assetInfo *assets.AssetInfo, payload []byte) {
	if len(payload) == 0 {
		return
	}

	httpData := string(payload)
	headers := pp.parseHTTPHeaders(httpData)

	if len(headers) > 0 {
		assetInfo.Protocols["http"] = headers

		// 提取关键信息
		if userAgent, ok := headers["user-agent"]; ok {
			assetInfo.OSGuess = pp.guessOSFromUserAgent(userAgent.(string))
		}

		if server, ok := headers["server"]; ok {
			if assetInfo.Services == nil {
				assetInfo.Services = make(map[string]interface{})
			}
			assetInfo.Services["http"] = server
		}

		if host, ok := headers["host"]; ok {
			assetInfo.Hostname = host.(string)
		}
	}
}

// parseDHCP 解析DHCP协议
func (pp *PacketParser) parseDHCP(assetInfo *assets.AssetInfo, payload []byte) {
	// 简化的DHCP解析
	if len(payload) < 240 {
		return
	}

	// 提取客户端MAC地址 (offset 28, length 6)
	if payload[0] == 1 { // DHCP Request
		mac := net.HardwareAddr(payload[28:34])
		assetInfo.MACAddress = mac.String()
		assetInfo.Vendor = pp.getVendorFromMAC(mac)

		// 解析DHCP选项中的主机名等信息
		options := pp.parseDHCPOptions(payload[240:])
		if len(options) > 0 {
			assetInfo.Protocols["dhcp"] = options

			if hostname, ok := options["hostname"]; ok {
				assetInfo.Hostname = hostname.(string)
			}
		}
	}
}

// parseDNS 解析DNS协议
func (pp *PacketParser) parseDNS(assetInfo *assets.AssetInfo, payload []byte) {
	// 简化的DNS解析
	if len(payload) < 12 {
		return
	}

	// 这里可以解析DNS查询和响应，提取域名信息
	assetInfo.Protocols["dns"] = map[string]interface{}{
		"packet_length": len(payload),
	}
}

// parseMDNS 解析mDNS协议
func (pp *PacketParser) parseMDNS(assetInfo *assets.AssetInfo, payload []byte) {
	// 简化的mDNS解析
	if len(payload) < 12 {
		return
	}

	// mDNS通常包含服务发现信息
	assetInfo.Protocols["mdns"] = map[string]interface{}{
		"packet_length": len(payload),
	}
}

// parseHTTPHeaders 解析HTTP头部
func (pp *PacketParser) parseHTTPHeaders(httpData string) map[string]interface{} {
	headers := make(map[string]interface{})
	lines := strings.Split(httpData, "\r\n")

	for _, line := range lines {
		if strings.Contains(line, ":") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				key := strings.ToLower(strings.TrimSpace(parts[0]))
				value := strings.TrimSpace(parts[1])
				headers[key] = value
			}
		}
	}

	return headers
}

// parseDHCPOptions 解析DHCP选项
func (pp *PacketParser) parseDHCPOptions(options []byte) map[string]interface{} {
	result := make(map[string]interface{})

	for i := 0; i < len(options); {
		if options[i] == 255 { // End option
			break
		}
		if options[i] == 0 { // Pad option
			i++
			continue
		}

		optionType := options[i]
		if i+1 >= len(options) {
			break
		}

		optionLen := int(options[i+1])
		if i+2+optionLen > len(options) {
			break
		}

		optionData := options[i+2 : i+2+optionLen]

		switch optionType {
		case 12: // Hostname
			result["hostname"] = string(optionData)
		case 15: // Domain name
			result["domain"] = string(optionData)
		case 60: // Vendor class identifier
			result["vendor_class"] = string(optionData)
		}

		i += 2 + optionLen
	}

	return result
}

// 辅助函数
func (pp *PacketParser) isMulticastMAC(mac net.HardwareAddr) bool {
	return len(mac) > 0 && (mac[0]&0x01) != 0
}

func (pp *PacketParser) getVendorFromMAC(mac net.HardwareAddr) string {
	// 简化的厂商识别，基于OUI
	if len(mac) < 3 {
		return ""
	}

	oui := fmt.Sprintf("%02x:%02x:%02x", mac[0], mac[1], mac[2])

	// 常见厂商OUI映射
	vendors := map[string]string{
		"00:50:56": "VMware",
		"00:0c:29": "VMware",
		"08:00:27": "VirtualBox",
		"00:15:5d": "Microsoft Hyper-V",
		"52:54:00": "QEMU/KVM",
		"00:16:3e": "Xen",
		"ec:f4:bb": "NetApp",
		"00:90:27": "Intel",
		"d4:be:d9": "Dell",
		"98:90:96": "Foxconn",
	}

	if vendor, ok := vendors[oui]; ok {
		return vendor
	}

	return ""
}

func (pp *PacketParser) guessOSFromTTL(ttl uint8) string {
	// 基于TTL值推测操作系统
	switch {
	case ttl <= 64:
		return "Linux/Unix"
	case ttl <= 128:
		return "Windows"
	case ttl <= 255:
		return "Cisco/Network Device"
	default:
		return ""
	}
}

func (pp *PacketParser) guessOSFromUserAgent(userAgent string) string {
	userAgent = strings.ToLower(userAgent)

	if strings.Contains(userAgent, "windows") {
		return "Windows"
	} else if strings.Contains(userAgent, "mac os x") || strings.Contains(userAgent, "macos") {
		return "macOS"
	} else if strings.Contains(userAgent, "linux") {
		return "Linux"
	} else if strings.Contains(userAgent, "android") {
		return "Android"
	} else if strings.Contains(userAgent, "iphone") || strings.Contains(userAgent, "ipad") {
		return "iOS"
	}

	return ""
}

func (pp *PacketParser) identifyService(srcPort, dstPort int, appLayer gopacket.ApplicationLayer) string {
	// 识别常见服务
	services := map[int]string{
		80:    "HTTP",
		443:   "HTTPS",
		22:    "SSH",
		23:    "Telnet",
		21:    "FTP",
		25:    "SMTP",
		110:   "POP3",
		143:   "IMAP",
		993:   "IMAPS",
		995:   "POP3S",
		3389:  "RDP",
		5432:  "PostgreSQL",
		3306:  "MySQL",
		1433:  "MSSQL",
		6379:  "Redis",
		27017: "MongoDB",
	}

	if service, ok := services[srcPort]; ok {
		return service
	}
	if service, ok := services[dstPort]; ok {
		return service
	}

	return ""
}

func (pp *PacketParser) hasUsefulInfo(assetInfo *assets.AssetInfo) bool {
	return assetInfo.IPAddress != "" || assetInfo.MACAddress != "" ||
		assetInfo.Hostname != "" || len(assetInfo.OpenPorts) > 0 ||
		len(assetInfo.Services) > 0 || len(assetInfo.Protocols) > 0
}
