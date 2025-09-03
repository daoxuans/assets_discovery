package config

import (
	"sync"
	"time"

	"github.com/spf13/viper"
)

var (
	once sync.Once
	cfg  *Config
)

// Config 系统配置结构
type Config struct {
	Capture  CaptureConfig  `yaml:"capture"`
	Parser   ParserConfig   `yaml:"parser"`
	Storage  StorageConfig  `yaml:"storage"`
	Server   ServerConfig   `yaml:"server"`
	Alerting AlertingConfig `yaml:"alerting"`
}

// CaptureConfig 流量捕获配置
type CaptureConfig struct {
	Interface   string        `yaml:"interface"`
	SnapLen     int           `yaml:"snap_len"`
	Promiscuous bool          `yaml:"promiscuous"`
	Timeout     time.Duration `yaml:"timeout"`
	BufferSize  int           `yaml:"buffer_size"`
	Workers     int           `yaml:"workers"`
}

// ParserConfig 协议解析配置
type ParserConfig struct {
	EnabledProtocols []string `yaml:"enabled_protocols"`
	MaxPackets       int      `yaml:"max_packets"`
	AssetTimeout     int      `yaml:"asset_timeout"` // 资产超时时间(分钟)
}

// StorageConfig 存储配置
type StorageConfig struct {
	Type          string     `yaml:"type"` // elasticsearch, file, memory
	Elasticsearch ESConfig   `yaml:"elasticsearch"`
	File          FileConfig `yaml:"file"`
}

// ESConfig Elasticsearch配置
type ESConfig struct {
	URLs     []string `yaml:"urls"`
	Username string   `yaml:"username"`
	Password string   `yaml:"password"`
	Index    string   `yaml:"index"`
}

// FileConfig 文件存储配置
type FileConfig struct {
	OutputDir string `yaml:"output_dir"`
	Format    string `yaml:"format"` // json, csv
}

// ServerConfig Web服务配置
type ServerConfig struct {
	Port    int  `yaml:"port"`
	Enabled bool `yaml:"enabled"`
}

// AlertingConfig 告警配置
type AlertingConfig struct {
	Enabled    bool     `yaml:"enabled"`
	WebhookURL string   `yaml:"webhook_url"`
	EmailTo    []string `yaml:"email_to"`
	AlertRules []string `yaml:"alert_rules"`
}

// GetConfig 获取全局配置
func GetConfig() *Config {
	once.Do(func() {
		cfg = loadConfig()
	})
	return cfg
}

// loadConfig 加载配置
func loadConfig() *Config {
	// 设置默认值
	setDefaults()

	config := &Config{}
	if err := viper.Unmarshal(config); err != nil {
		// 使用默认配置
		config = getDefaultConfig()
	}

	return config
}

// setDefaults 设置默认配置值
func setDefaults() {
	// 捕获配置默认值
	viper.SetDefault("capture.snap_len", 65536)
	viper.SetDefault("capture.promiscuous", true)
	viper.SetDefault("capture.timeout", "30s")
	viper.SetDefault("capture.buffer_size", 2097152) // 2MB
	viper.SetDefault("capture.workers", 4)

	// 解析配置默认值
	viper.SetDefault("parser.enabled_protocols", []string{"arp", "dhcp", "http", "https", "dns", "smb", "mdns"})
	viper.SetDefault("parser.max_packets", 0)    // 0表示无限制
	viper.SetDefault("parser.asset_timeout", 30) // 30分钟

	// 存储配置默认值
	viper.SetDefault("storage.type", "file")
	viper.SetDefault("storage.file.output_dir", "./output")
	viper.SetDefault("storage.file.format", "json")
	viper.SetDefault("storage.elasticsearch.index", "assets")

	// 服务配置默认值
	viper.SetDefault("server.port", 8080)
	viper.SetDefault("server.enabled", true)

	// 告警配置默认值
	viper.SetDefault("alerting.enabled", false)
}

// getDefaultConfig 获取默认配置
func getDefaultConfig() *Config {
	return &Config{
		Capture: CaptureConfig{
			Interface:   "",
			SnapLen:     65536,
			Promiscuous: true,
			Timeout:     30 * time.Second,
			BufferSize:  2097152,
			Workers:     4,
		},
		Parser: ParserConfig{
			EnabledProtocols: []string{"arp", "dhcp", "http", "https", "dns", "smb", "mdns"},
			MaxPackets:       0,
			AssetTimeout:     30,
		},
		Storage: StorageConfig{
			Type: "file",
			File: FileConfig{
				OutputDir: "./output",
				Format:    "json",
			},
		},
		Server: ServerConfig{
			Port:    8080,
			Enabled: true,
		},
		Alerting: AlertingConfig{
			Enabled: false,
		},
	}
}
