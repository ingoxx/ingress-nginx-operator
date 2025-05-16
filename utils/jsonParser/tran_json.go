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

// JSONToStruct 将 JSON 字符串转换为指定类型的结构体（需要传指针）
func JSONToStruct[T any](jsonStr string, out *T) error {
	if err := json.Unmarshal([]byte(jsonStr), out); err != nil {
		return fmt.Errorf("json parse failed: %w", err)
	}

	return nil
}
