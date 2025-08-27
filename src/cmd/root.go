package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"assets_discovery/internal/capture"
	"assets_discovery/internal/config"
)

var cfgFile string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "assets_discovery",
	Short: "被动式网络资产识别与分析系统",
	Long: `被动式网络资产识别与分析系统 - 通过监听网络流量来识别和分析网络资产

支持在线监听网络接口和离线分析pcap文件，能够识别设备类型、操作系统、服务等信息。`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initConfig)

	// 全局配置文件标志
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.assets_discovery.yaml)")

	// 子命令
	rootCmd.AddCommand(liveCmd)
	rootCmd.AddCommand(offlineCmd)
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		viper.AddConfigPath(home)
		viper.AddConfigPath(".")
		viper.SetConfigType("yaml")
		viper.SetConfigName(".assets_discovery")
	}

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}
}

// liveCmd represents the live command
var liveCmd = &cobra.Command{
	Use:   "live",
	Short: "实时监听网络接口",
	Long:  `实时监听指定的网络接口，分析流量并识别资产`,
	Run: func(cmd *cobra.Command, args []string) {
		cfg := config.GetConfig()

		iface, _ := cmd.Flags().GetString("interface")
		if iface != "" {
			cfg.Capture.Interface = iface
		}

		captureEngine := capture.NewCaptureEngine(cfg)
		if err := captureEngine.StartLiveCapture(); err != nil {
			fmt.Printf("启动实时捕获失败: %v\n", err)
			os.Exit(1)
		}
	},
}

// offlineCmd represents the offline command
var offlineCmd = &cobra.Command{
	Use:   "offline",
	Short: "离线分析pcap文件",
	Long:  `分析离线的pcap文件，识别其中的网络资产`,
	Run: func(cmd *cobra.Command, args []string) {
		cfg := config.GetConfig()

		pcapFile, _ := cmd.Flags().GetString("file")
		if pcapFile == "" {
			fmt.Println("请指定pcap文件路径")
			os.Exit(1)
		}

		captureEngine := capture.NewCaptureEngine(cfg)
		if err := captureEngine.StartOfflineCapture(pcapFile); err != nil {
			fmt.Printf("离线分析失败: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	// live命令标志
	liveCmd.Flags().StringP("interface", "i", "", "网络接口名称 (例如: eth0)")

	// offline命令标志
	offlineCmd.Flags().StringP("file", "f", "", "pcap文件路径")
	offlineCmd.MarkFlagRequired("file")
}
