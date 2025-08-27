package storage

// Storage 存储接口
type Storage interface {
	// 保存资产
	SaveAsset(asset interface{}) error

	// 获取资产
	GetAsset(id string) (interface{}, error)

	// 获取所有资产
	GetAllAssets() ([]interface{}, error)

	// 搜索资产
	SearchAssets(query string) ([]interface{}, error)

	// 删除资产
	DeleteAsset(id string) error

	// 导出数据
	ExportJSON(assets interface{}) ([]byte, error)

	// 关闭存储
	Close() error
}
