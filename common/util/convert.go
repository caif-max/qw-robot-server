package util

import (
	"encoding/json"
)

func MustString(str interface{}) string {
	result, ok := str.(string)
	if ok {
		return result
	} else {
		return ""
	}
}

// 转换从redis获取的数据
func ConvertStringToMap(base map[string]string) map[string]interface{} {
	resultMap := make(map[string]interface{})
	for k, v := range base {
		var dat map[string]interface{}
		if err := json.Unmarshal([]byte(v), &dat); err == nil {
			resultMap[k] = dat
		} else {
			resultMap[k] = v
		}
	}
	return resultMap
}
