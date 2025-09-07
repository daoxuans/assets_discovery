package config

import (
	"fmt"
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
	Capture  CaptureConfig  `yaml:"capture" mapstructure:"capture"`
	Parser   ParserConfig   `yaml:"parser" mapstructure:"parser"`
	Storage  StorageConfig  `yaml:"storage" mapstructure:"storage"`
	Server   ServerConfig   `yaml:"server" mapstructure:"server"`
	Alerting AlertingConfig `yaml:"alerting" mapstructure:"alerting"`
}

// CaptureConfig 流量捕获配置
type CaptureConfig struct {
	Interface   string        `yaml:"interface" mapstructure:"interface"`
	SnapLen     int           `yaml:"snap_len" mapstructure:"snap_len"`
	Promiscuous bool          `yaml:"promiscuous" mapstructure:"promiscuous"`
	Timeout     time.Duration `yaml:"timeout" mapstructure:"timeout"`
	BufferSize  int           `yaml:"buffer_size" mapstructure:"buffer_size"`
	Workers     int           `yaml:"workers" mapstructure:"workers"`
}

// ParserConfig 协议解析配置
type ParserConfig struct {
	EnabledProtocols []string `yaml:"enabled_protocols" mapstructure:"enabled_protocols"`
	MaxPackets       int      `yaml:"max_packets" mapstructure:"max_packets"`
	AssetTimeout     int      `yaml:"asset_timeout" mapstructure:"asset_timeout"` // 资产超时时间(分钟)
}

// StorageConfig 存储配置
type StorageConfig struct {
	Type          string     `yaml:"type" mapstructure:"type"` // elasticsearch, file, memory
	Elasticsearch ESConfig   `yaml:"elasticsearch" mapstructure:"elasticsearch"`
	File          FileConfig `yaml:"file" mapstructure:"file"`
}

// ESConfig Elasticsearch配置
type ESConfig struct {
	URLs     []string `yaml:"urls" mapstructure:"urls"`
	Username string   `yaml:"username" mapstructure:"username"`
	Password string   `yaml:"password" mapstructure:"password"`
	Index    string   `yaml:"index" mapstructure:"index"`
}

// FileConfig 文件存储配置
type FileConfig struct {
	OutputDir string `yaml:"output_dir" mapstructure:"output_dir"`
	Format    string `yaml:"format" mapstructure:"format"` // json, csv
}

// ServerConfig Web服务配置
type ServerConfig struct {
	Port    int  `yaml:"port" mapstructure:"port"`
	Enabled bool `yaml:"enabled" mapstructure:"enabled"`
}

// AlertingConfig 告警配置
type AlertingConfig struct {
	Enabled    bool     `yaml:"enabled" mapstructure:"enabled"`
	WebhookURL string   `yaml:"webhook_url" mapstructure:"webhook_url"`
	EmailTo    []string `yaml:"email_to" mapstructure:"email_to"`
	AlertRules []string `yaml:"alert_rules" mapstructure:"alert_rules"`
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
	config := &Config{}
	if err := viper.Unmarshal(config); err != nil {
		fmt.Printf("警告: 配置解析失败 (%v)，使用硬编码默认配置\n", err)
		config = getDefaultConfig()
	}

	return config
}

// SetDefaults 设置默认配置值
func SetDefaults() {
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
