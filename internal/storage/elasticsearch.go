package storage

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"

	"assets_discovery/internal/config"
)

// ElasticsearchStorage Elasticsearch存储实现
type ElasticsearchStorage struct {
	client *elasticsearch.Client
	index  string
}

// NewElasticsearchStorage 创建Elasticsearch存储
func NewElasticsearchStorage(cfg *config.ESConfig) (*ElasticsearchStorage, error) {
	esCfg := elasticsearch.Config{
		Addresses: cfg.URLs,
	}

	if cfg.Username != "" && cfg.Password != "" {
		esCfg.Username = cfg.Username
		esCfg.Password = cfg.Password
	}

	client, err := elasticsearch.NewClient(esCfg)
	if err != nil {
		return nil, fmt.Errorf("创建Elasticsearch客户端失败: %v", err)
	}

	es := &ElasticsearchStorage{
		client: client,
		index:  cfg.Index,
	}

	// 创建索引和映射
	if err := es.createIndex(); err != nil {
		return nil, fmt.Errorf("创建索引失败: %v", err)
	}

	return es, nil
}

// SaveAsset 保存资产
func (es *ElasticsearchStorage) SaveAsset(asset interface{}) error {
	// 提取资产ID
	var assetID string
	if assetMap, ok := asset.(map[string]interface{}); ok {
		if id, exists := assetMap["id"]; exists {
			if idStr, ok := id.(string); ok {
				assetID = idStr
			}
		}
	}

	if assetID == "" {
		return fmt.Errorf("无法提取资产ID")
	}

	// 序列化资产数据
	assetBytes, err := json.Marshal(asset)
	if err != nil {
		return fmt.Errorf("序列化资产失败: %v", err)
	}

	// 索引文档
	req := esapi.IndexRequest{
		Index:      es.index,
		DocumentID: assetID,
		Body:       bytes.NewReader(assetBytes),
		Refresh:    "true",
	}

	res, err := req.Do(context.Background(), es.client)
	if err != nil {
		return fmt.Errorf("索引文档失败: %v", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("Elasticsearch错误: %s", res.Status())
	}

	return nil
}

// GetAsset 获取资产
func (es *ElasticsearchStorage) GetAsset(id string) (interface{}, error) {
	req := esapi.GetRequest{
		Index:      es.index,
		DocumentID: id,
	}

	res, err := req.Do(context.Background(), es.client)
	if err != nil {
		return nil, fmt.Errorf("获取文档失败: %v", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		if res.StatusCode == 404 {
			return nil, fmt.Errorf("资产不存在: %s", id)
		}
		return nil, fmt.Errorf("Elasticsearch错误: %s", res.Status())
	}

	var result map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("解析响应失败: %v", err)
	}

	if source, ok := result["_source"]; ok {
		return source, nil
	}

	return nil, fmt.Errorf("响应中没有_source字段")
}

// GetAllAssets 获取所有资产
func (es *ElasticsearchStorage) GetAllAssets() ([]interface{}, error) {
	query := map[string]interface{}{
		"query": map[string]interface{}{
			"match_all": map[string]interface{}{},
		},
		"size": 10000, // 限制返回数量
	}

	queryBytes, err := json.Marshal(query)
	if err != nil {
		return nil, fmt.Errorf("构建查询失败: %v", err)
	}

	req := esapi.SearchRequest{
		Index: []string{es.index},
		Body:  bytes.NewReader(queryBytes),
	}

	res, err := req.Do(context.Background(), es.client)
	if err != nil {
		return nil, fmt.Errorf("搜索失败: %v", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, fmt.Errorf("Elasticsearch错误: %s", res.Status())
	}

	var result map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("解析响应失败: %v", err)
	}

	hits, ok := result["hits"].(map[string]interface{})["hits"].([]interface{})
	if !ok {
		return []interface{}{}, nil
	}

	assets := make([]interface{}, 0, len(hits))
	for _, hit := range hits {
		if hitMap, ok := hit.(map[string]interface{}); ok {
			if source, ok := hitMap["_source"]; ok {
				assets = append(assets, source)
			}
		}
	}

	return assets, nil
}

// SearchAssets 搜索资产
func (es *ElasticsearchStorage) SearchAssets(query string) ([]interface{}, error) {
	searchQuery := map[string]interface{}{
		"query": map[string]interface{}{
			"multi_match": map[string]interface{}{
				"query":  query,
				"fields": []string{"ip_address", "mac_address", "hostname", "device_type", "os_info.family"},
			},
		},
		"size": 1000,
	}

	queryBytes, err := json.Marshal(searchQuery)
	if err != nil {
		return nil, fmt.Errorf("构建查询失败: %v", err)
	}

	req := esapi.SearchRequest{
		Index: []string{es.index},
		Body:  bytes.NewReader(queryBytes),
	}

	res, err := req.Do(context.Background(), es.client)
	if err != nil {
		return nil, fmt.Errorf("搜索失败: %v", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, fmt.Errorf("Elasticsearch错误: %s", res.Status())
	}

	var result map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("解析响应失败: %v", err)
	}

	hits, ok := result["hits"].(map[string]interface{})["hits"].([]interface{})
	if !ok {
		return []interface{}{}, nil
	}

	assets := make([]interface{}, 0, len(hits))
	for _, hit := range hits {
		if hitMap, ok := hit.(map[string]interface{}); ok {
			if source, ok := hitMap["_source"]; ok {
				assets = append(assets, source)
			}
		}
	}

	return assets, nil
}

// DeleteAsset 删除资产
func (es *ElasticsearchStorage) DeleteAsset(id string) error {
	req := esapi.DeleteRequest{
		Index:      es.index,
		DocumentID: id,
		Refresh:    "true",
	}

	res, err := req.Do(context.Background(), es.client)
	if err != nil {
		return fmt.Errorf("删除文档失败: %v", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		if res.StatusCode == 404 {
			return fmt.Errorf("资产不存在: %s", id)
		}
		return fmt.Errorf("Elasticsearch错误: %s", res.Status())
	}

	return nil
}

// ExportJSON 导出JSON
func (es *ElasticsearchStorage) ExportJSON(assets interface{}) ([]byte, error) {
	return json.MarshalIndent(assets, "", "  ")
}

// Close 关闭存储
func (es *ElasticsearchStorage) Close() error {
	// Elasticsearch客户端不需要显式关闭
	return nil
}

// createIndex 创建索引和映射
func (es *ElasticsearchStorage) createIndex() error {
	// 检查索引是否存在
	req := esapi.IndicesExistsRequest{
		Index: []string{es.index},
	}

	res, err := req.Do(context.Background(), es.client)
	if err != nil {
		return fmt.Errorf("检查索引失败: %v", err)
	}
	res.Body.Close()

	if res.StatusCode == 200 {
		// 索引已存在
		return nil
	}

	// 创建索引映射
	mapping := map[string]interface{}{
		"mappings": map[string]interface{}{
			"properties": map[string]interface{}{
				"ip_address": map[string]interface{}{
					"type": "ip",
				},
				"mac_address": map[string]interface{}{
					"type": "keyword",
				},
				"hostname": map[string]interface{}{
					"type": "text",
					"fields": map[string]interface{}{
						"keyword": map[string]interface{}{
							"type": "keyword",
						},
					},
				},
				"device_type": map[string]interface{}{
					"type": "keyword",
				},
				"os_info": map[string]interface{}{
					"properties": map[string]interface{}{
						"family": map[string]interface{}{
							"type": "keyword",
						},
						"version": map[string]interface{}{
							"type": "text",
						},
					},
				},
				"first_seen": map[string]interface{}{
					"type": "date",
				},
				"last_seen": map[string]interface{}{
					"type": "date",
				},
				"is_active": map[string]interface{}{
					"type": "boolean",
				},
			},
		},
	}

	mappingBytes, err := json.Marshal(mapping)
	if err != nil {
		return fmt.Errorf("构建映射失败: %v", err)
	}

	// 创建索引
	createReq := esapi.IndicesCreateRequest{
		Index: es.index,
		Body:  strings.NewReader(string(mappingBytes)),
	}

	createRes, err := createReq.Do(context.Background(), es.client)
	if err != nil {
		return fmt.Errorf("创建索引失败: %v", err)
	}
	defer createRes.Body.Close()

	if createRes.IsError() {
		return fmt.Errorf("创建索引错误: %s", createRes.Status())
	}

	return nil
}
