package assets

import (
	"log"
	"sync"
	"time"

	"assets_discovery/internal/config"
	"assets_discovery/internal/storage"
)

// AssetManager 资产管理器
type AssetManager struct {
	config  *config.Config
	storage storage.Storage
	assets  map[string]*Asset // key为资产ID
	mutex   sync.RWMutex
	stopCh  chan struct{}

	// 统计信息
	stats AssetStats
}

// AssetStats 资产统计信息
type AssetStats struct {
	TotalAssets    int            `json:"total_assets"`
	ActiveAssets   int            `json:"active_assets"`
	NewAssets      int            `json:"new_assets"` // 最近发现的资产
	LastUpdate     time.Time      `json:"last_update"`
	DeviceTypes    map[string]int `json:"device_types"`
	OSDistribution map[string]int `json:"os_distribution"`
}

// NewAssetManager 创建新的资产管理器
func NewAssetManager(cfg *config.Config, storage storage.Storage) *AssetManager {
	return &AssetManager{
		config:  cfg,
		storage: storage,
		assets:  make(map[string]*Asset),
		stopCh:  make(chan struct{}),
		stats: AssetStats{
			DeviceTypes:    make(map[string]int),
			OSDistribution: make(map[string]int),
		},
	}
}

// Start 启动资产管理器
func (am *AssetManager) Start() {
	log.Println("资产管理器启动")

	// 从存储中加载现有资产
	am.loadExistingAssets()

	// 启动定期清理任务
	go am.cleanupRoutine()

	// 启动统计更新任务
	go am.statsUpdateRoutine()
}

// Stop 停止资产管理器
func (am *AssetManager) Stop() {
	log.Println("资产管理器停止")
	close(am.stopCh)

	// 保存当前资产状态
	am.saveAllAssets()
}

// UpdateAsset 更新资产信息
func (am *AssetManager) UpdateAsset(assetInfo *AssetInfo) {
	if assetInfo == nil {
		return
	}

	am.mutex.Lock()
	defer am.mutex.Unlock()

	assetID := generateAssetID(assetInfo)

	if existingAsset, exists := am.assets[assetID]; exists {
		// 更新现有资产
		existingAsset.Update(assetInfo)
		log.Printf("更新资产: %s (%s)", assetID, assetInfo.IPAddress)
	} else {
		// 创建新资产
		newAsset := NewAsset(assetInfo)
		am.assets[assetID] = newAsset
		am.stats.NewAssets++
		log.Printf("发现新资产: %s (%s)", assetID, assetInfo.IPAddress)

		// 发送新资产告警
		am.notifyNewAsset(newAsset)
	}

	// 异步保存到存储
	go am.saveAsset(assetID)
}

// GetAsset 获取资产信息
func (am *AssetManager) GetAsset(assetID string) (*Asset, bool) {
	am.mutex.RLock()
	defer am.mutex.RUnlock()

	asset, exists := am.assets[assetID]
	return asset, exists
}

// GetAllAssets 获取所有资产
func (am *AssetManager) GetAllAssets() map[string]*Asset {
	am.mutex.RLock()
	defer am.mutex.RUnlock()

	// 返回副本以避免并发问题
	result := make(map[string]*Asset)
	for id, asset := range am.assets {
		result[id] = asset
	}

	return result
}

// GetActiveAssets 获取活跃资产
func (am *AssetManager) GetActiveAssets() []*Asset {
	am.mutex.RLock()
	defer am.mutex.RUnlock()

	var activeAssets []*Asset
	for _, asset := range am.assets {
		if asset.IsActive {
			activeAssets = append(activeAssets, asset)
		}
	}

	return activeAssets
}

// GetAssetsByType 根据设备类型获取资产
func (am *AssetManager) GetAssetsByType(deviceType string) []*Asset {
	am.mutex.RLock()
	defer am.mutex.RUnlock()

	var assets []*Asset
	for _, asset := range am.assets {
		if asset.DeviceType == deviceType {
			assets = append(assets, asset)
		}
	}

	return assets
}

// GetAssetsByOS 根据操作系统获取资产
func (am *AssetManager) GetAssetsByOS(osFamily string) []*Asset {
	am.mutex.RLock()
	defer am.mutex.RUnlock()

	var assets []*Asset
	for _, asset := range am.assets {
		if asset.OSInfo.Family == osFamily {
			assets = append(assets, asset)
		}
	}

	return assets
}

// GetStats 获取统计信息
func (am *AssetManager) GetStats() AssetStats {
	am.mutex.RLock()
	defer am.mutex.RUnlock()

	return am.stats
}

// SearchAssets 搜索资产
func (am *AssetManager) SearchAssets(query string) []*Asset {
	am.mutex.RLock()
	defer am.mutex.RUnlock()

	var results []*Asset

	for _, asset := range am.assets {
		if am.matchesQuery(asset, query) {
			results = append(results, asset)
		}
	}

	return results
}

// loadExistingAssets 从存储加载现有资产
func (am *AssetManager) loadExistingAssets() {
	assets, err := am.storage.GetAllAssets()
	if err != nil {
		log.Printf("加载现有资产失败: %v", err)
		return
	}

	am.mutex.Lock()
	defer am.mutex.Unlock()

	for _, assetInterface := range assets {
		if asset, ok := assetInterface.(*Asset); ok {
			am.assets[asset.ID] = asset
		}
	}

	log.Printf("加载了 %d 个现有资产", len(assets))
}

// saveAsset 保存单个资产
func (am *AssetManager) saveAsset(assetID string) {
	am.mutex.RLock()
	asset, exists := am.assets[assetID]
	am.mutex.RUnlock()

	if !exists {
		return
	}

	if err := am.storage.SaveAsset(asset); err != nil {
		log.Printf("保存资产失败 %s: %v", assetID, err)
	}
}

// saveAllAssets 保存所有资产
func (am *AssetManager) saveAllAssets() {
	am.mutex.RLock()
	assets := make([]*Asset, 0, len(am.assets))
	for _, asset := range am.assets {
		assets = append(assets, asset)
	}
	am.mutex.RUnlock()

	for _, asset := range assets {
		if err := am.storage.SaveAsset(asset); err != nil {
			log.Printf("保存资产失败 %s: %v", asset.ID, err)
		}
	}

	log.Printf("保存了 %d 个资产", len(assets))
}

// cleanupRoutine 定期清理例程
func (am *AssetManager) cleanupRoutine() {
	ticker := time.NewTicker(5 * time.Minute) // 每5分钟执行一次清理
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			am.cleanupInactiveAssets()
		case <-am.stopCh:
			return
		}
	}
}

// cleanupInactiveAssets 清理非活跃资产
func (am *AssetManager) cleanupInactiveAssets() {
	am.mutex.Lock()
	defer am.mutex.Unlock()

	timeout := time.Duration(am.config.Parser.AssetTimeout) * time.Minute
	cutoff := time.Now().Add(-timeout)

	inactiveCount := 0
	for _, asset := range am.assets {
		if asset.IsActive && asset.LastSeen.Before(cutoff) {
			asset.SetInactive()
			inactiveCount++

			// 保存状态变更
			go am.saveAsset(asset.ID)
		}
	}

	if inactiveCount > 0 {
		log.Printf("标记了 %d 个资产为非活跃状态", inactiveCount)
	}
}

// statsUpdateRoutine 统计信息更新例程
func (am *AssetManager) statsUpdateRoutine() {
	ticker := time.NewTicker(1 * time.Minute) // 每分钟更新统计
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			am.updateStats()
		case <-am.stopCh:
			return
		}
	}
}

// updateStats 更新统计信息
func (am *AssetManager) updateStats() {
	am.mutex.Lock()
	defer am.mutex.Unlock()

	am.stats.TotalAssets = len(am.assets)
	am.stats.ActiveAssets = 0
	am.stats.DeviceTypes = make(map[string]int)
	am.stats.OSDistribution = make(map[string]int)
	am.stats.LastUpdate = time.Now()

	for _, asset := range am.assets {
		if asset.IsActive {
			am.stats.ActiveAssets++
		}

		if asset.DeviceType != "" {
			am.stats.DeviceTypes[asset.DeviceType]++
		}

		if asset.OSInfo.Family != "" {
			am.stats.OSDistribution[asset.OSInfo.Family]++
		}
	}
}

// notifyNewAsset 新资产通知
func (am *AssetManager) notifyNewAsset(asset *Asset) {
	if !am.config.Alerting.Enabled {
		return
	}

	// 这里可以实现各种通知方式：邮件、Webhook、日志等
	log.Printf("新资产告警: %s - %s (%s)", asset.ID, asset.IPAddress, asset.DeviceType)

	// TODO: 实现具体的告警逻辑
	// - 发送邮件
	// - 调用Webhook
	// - 写入告警日志
}

// matchesQuery 检查资产是否匹配查询
func (am *AssetManager) matchesQuery(asset *Asset, query string) bool {
	// 简单的字符串匹配，可以扩展为更复杂的查询语法
	return asset.IPAddress == query ||
		asset.MACAddress == query ||
		asset.Hostname == query ||
		asset.DeviceType == query ||
		asset.OSInfo.Family == query
}

// ExportAssets 导出资产数据
func (am *AssetManager) ExportAssets(format string) ([]byte, error) {
	assets := am.GetAllAssets()

	switch format {
	case "json":
		return am.storage.ExportJSON(assets)
	default:
		return am.storage.ExportJSON(assets)
	}
}
