package test

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"gitee.com/Sxiaobai/gs/v2/gstool"
)

// TestFpm 测试fpm无session的情况
func TestFpmNoSession(t *testing.T) {
	// 示例JSON
	jsonStr := `{
		"a": [{"xx": 1, "yy": [{"zz": 2}]}],
		"res": 1,
		"user": {
			"info": [{
				"name": "张三",
				"age": 30,
				"addresses": [{
					"city": "北京",
					"street": "长安街"
				}]
			}],
			"active": true
		},
		"items": [
			[{"nested": "value"}],
			{"ignored": "不会被提取"}
		]
	}`

	// 转换为JSON输出
	jsonResult, _ := ExtractJSONPaths(jsonStr)
	fmt.Printf("%s", gstool.JsonFormat(jsonResult))
}

type FlatItem struct {
	Key  string `json:"key"`
	Desc string `json:"desc"`
}

func ExtractJSONPaths(jsonStr string) ([]FlatItem, error) {
	decoder := json.NewDecoder(strings.NewReader(jsonStr))
	decoder.UseNumber()

	var data interface{}
	if err := decoder.Decode(&data); err != nil {
		return nil, err
	}

	results := []FlatItem{}
	extractAllPaths(data, "", &results, true)
	return results, nil
}

// extractAllPaths 提取所有路径，包括中间路径
func extractAllPaths(data interface{}, currentPath string, results *[]FlatItem, addCurrent bool) {
	// 添加当前路径（如果是有效路径）
	if addCurrent && currentPath != "" {
		cleanPath := strings.TrimPrefix(currentPath, ".")
		// 避免重复添加
		exists := false
		for _, item := range *results {
			if item.Key == cleanPath {
				exists = true
				break
			}
		}
		if !exists {
			*results = append(*results, FlatItem{
				Key:  cleanPath,
				Desc: getTypeDesc(data),
			})
		}
	}

	// 递归处理子元素
	switch v := data.(type) {
	case map[string]interface{}:
		keys := getOrderedKeys(v) // 获取有序的keys
		for _, key := range keys {
			val := v[key]
			newPath := currentPath
			if currentPath == "" {
				newPath = key
			} else {
				newPath = currentPath + "." + key
			}
			// 先添加中间路径，再递归处理
			extractAllPaths(val, newPath, results, true)
		}

	case []interface{}:
		if len(v) > 0 {
			// 处理数组中的第一个元素
			newPath := currentPath
			if currentPath != "" && !strings.HasSuffix(currentPath, "]") {
				// 如果路径不是以数组结束，加上数组索引
				newPath = currentPath + "[0]"
			}
			// 添加数组路径本身
			if newPath != currentPath && newPath != "" {
				cleanPath := strings.TrimPrefix(newPath, ".")
				*results = append(*results, FlatItem{
					Key:  cleanPath,
					Desc: "array[0]",
				})
			}
			extractAllPaths(v[0], newPath, results, false)
		}

	default:
		// 基本类型，已经在上面的addCurrent中添加过了
	}
}

// getTypeDesc 获取数据类型的描述
func getTypeDesc(data interface{}) string {
	switch data.(type) {
	case map[string]interface{}:
		return "object"
	case []interface{}:
		return "array"
	case string:
		return "string"
	case json.Number:
		return "number"
	case bool:
		return "boolean"
	case nil:
		return "null"
	default:
		return "unknown"
	}
}

// getOrderedKeys 从map中获取有序的key（按字母顺序排序以保持一致性）
func getOrderedKeys(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for key := range m {
		keys = append(keys, key)
	}
	// 按字母排序以保持一致的输出顺序
	// 注意：这不是原始JSON的顺序，但至少是稳定的
	sortStrings(keys)
	return keys
}

// sortStrings 简单的字符串排序
func sortStrings(strs []string) {
	for i := 0; i < len(strs); i++ {
		for j := i + 1; j < len(strs); j++ {
			if strs[i] > strs[j] {
				strs[i], strs[j] = strs[j], strs[i]
			}
		}
	}
}
