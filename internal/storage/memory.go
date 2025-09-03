package storage

import (
	"encoding/json"
	"fmt"
	"sync"
)

// MemoryStorage 内存存储实现
type MemoryStorage struct {
	data  map[string]interface{}
	mutex sync.RWMutex
}

// NewMemoryStorage 创建内存存储
func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		data: make(map[string]interface{}),
	}
}

// SaveAsset 保存资产
func (ms *MemoryStorage) SaveAsset(asset interface{}) error {
	ms.mutex.Lock()
	defer ms.mutex.Unlock()

	// 这里需要从asset中提取ID，为了避免循环导入，使用反射或类型断言
	if assetMap, ok := asset.(map[string]interface{}); ok {
		if id, exists := assetMap["id"]; exists {
			if idStr, ok := id.(string); ok {
				ms.data[idStr] = asset
				return nil
			}
		}
	}

	return fmt.Errorf("无法提取资产ID")
}

// GetAsset 获取资产
func (ms *MemoryStorage) GetAsset(id string) (interface{}, error) {
	ms.mutex.RLock()
	defer ms.mutex.RUnlock()

	if asset, exists := ms.data[id]; exists {
		return asset, nil
	}

	return nil, fmt.Errorf("资产不存在: %s", id)
}

// GetAllAssets 获取所有资产
func (ms *MemoryStorage) GetAllAssets() ([]interface{}, error) {
	ms.mutex.RLock()
	defer ms.mutex.RUnlock()

	assets := make([]interface{}, 0, len(ms.data))
	for _, asset := range ms.data {
		assets = append(assets, asset)
	}

	return assets, nil
}

// SearchAssets 搜索资产
func (ms *MemoryStorage) SearchAssets(query string) ([]interface{}, error) {
	ms.mutex.RLock()
	defer ms.mutex.RUnlock()

	var results []interface{}

	for _, asset := range ms.data {
		// 简单的字符串匹配搜索
		if assetBytes, err := json.Marshal(asset); err == nil {
			if contains(string(assetBytes), query) {
				results = append(results, asset)
			}
		}
	}

	return results, nil
}

// DeleteAsset 删除资产
func (ms *MemoryStorage) DeleteAsset(id string) error {
	ms.mutex.Lock()
	defer ms.mutex.Unlock()

	if _, exists := ms.data[id]; exists {
		delete(ms.data, id)
		return nil
	}

	return fmt.Errorf("资产不存在: %s", id)
}

// ExportJSON 导出JSON
func (ms *MemoryStorage) ExportJSON(assets interface{}) ([]byte, error) {
	return json.MarshalIndent(assets, "", "  ")
}

// Close 关闭存储
func (ms *MemoryStorage) Close() error {
	ms.mutex.Lock()
	defer ms.mutex.Unlock()

	ms.data = nil
	return nil
}

// 辅助函数
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
			containsSubstring(s, substr))))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
