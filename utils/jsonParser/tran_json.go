package jsonParser

import (
	"encoding/json"
	"fmt"
)

// JSONToMap 将 JSON 字符串解析成 map[string]interface{}
func JSONToMap(jsonStr string) (map[string]interface{}, error) {
	var result map[string]interface{}
	err := json.Unmarshal([]byte(jsonStr), &result)
	if err != nil {
		return nil, fmt.Errorf("json解析失败: %w", err)
	}
	return result, nil
}
