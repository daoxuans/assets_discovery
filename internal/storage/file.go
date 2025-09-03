package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"assets_discovery/internal/config"
)

// FileStorage 文件存储实现
type FileStorage struct {
	config   *config.FileConfig
	data     map[string]interface{}
	mutex    sync.RWMutex
	filePath string
}

// NewFileStorage 创建文件存储
func NewFileStorage(cfg *config.FileConfig) (*FileStorage, error) {
	// 确保输出目录存在
	if err := os.MkdirAll(cfg.OutputDir, 0755); err != nil {
		return nil, fmt.Errorf("创建输出目录失败: %v", err)
	}

	filePath := filepath.Join(cfg.OutputDir, "assets.json")

	fs := &FileStorage{
		config:   cfg,
		data:     make(map[string]interface{}),
		filePath: filePath,
	}

	// 加载现有数据
	fs.loadFromFile()

	return fs, nil
}

// SaveAsset 保存资产
func (fs *FileStorage) SaveAsset(asset interface{}) error {
	fs.mutex.Lock()
	defer fs.mutex.Unlock()

	// 提取资产ID
	if assetMap, ok := asset.(map[string]interface{}); ok {
		if id, exists := assetMap["id"]; exists {
			if idStr, ok := id.(string); ok {
				fs.data[idStr] = asset
				// 立即写入文件
				return fs.saveToFile()
			}
		}
	}

	return fmt.Errorf("无法提取资产ID")
}

// GetAsset 获取资产
func (fs *FileStorage) GetAsset(id string) (interface{}, error) {
	fs.mutex.RLock()
	defer fs.mutex.RUnlock()

	if asset, exists := fs.data[id]; exists {
		return asset, nil
	}

	return nil, fmt.Errorf("资产不存在: %s", id)
}

// GetAllAssets 获取所有资产
func (fs *FileStorage) GetAllAssets() ([]interface{}, error) {
	fs.mutex.RLock()
	defer fs.mutex.RUnlock()

	assets := make([]interface{}, 0, len(fs.data))
	for _, asset := range fs.data {
		assets = append(assets, asset)
	}

	return assets, nil
}

// SearchAssets 搜索资产
func (fs *FileStorage) SearchAssets(query string) ([]interface{}, error) {
	fs.mutex.RLock()
	defer fs.mutex.RUnlock()

	var results []interface{}

	for _, asset := range fs.data {
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
func (fs *FileStorage) DeleteAsset(id string) error {
	fs.mutex.Lock()
	defer fs.mutex.Unlock()

	if _, exists := fs.data[id]; exists {
		delete(fs.data, id)
		return fs.saveToFile()
	}

	return fmt.Errorf("资产不存在: %s", id)
}

// ExportJSON 导出JSON
func (fs *FileStorage) ExportJSON(assets interface{}) ([]byte, error) {
	return json.MarshalIndent(assets, "", "  ")
}

// Close 关闭存储
func (fs *FileStorage) Close() error {
	fs.mutex.Lock()
	defer fs.mutex.Unlock()

	// 最终保存
	return fs.saveToFile()
}

// loadFromFile 从文件加载数据
func (fs *FileStorage) loadFromFile() error {
	if _, err := os.Stat(fs.filePath); os.IsNotExist(err) {
		// 文件不存在，使用空数据
		return nil
	}

	data, err := os.ReadFile(fs.filePath)
	if err != nil {
		return fmt.Errorf("读取文件失败: %v", err)
	}

	if len(data) == 0 {
		return nil
	}

	return json.Unmarshal(data, &fs.data)
}

// saveToFile 保存数据到文件
func (fs *FileStorage) saveToFile() error {
	data, err := json.MarshalIndent(fs.data, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化数据失败: %v", err)
	}

	return os.WriteFile(fs.filePath, data, 0644)
}
