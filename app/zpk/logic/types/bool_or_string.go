package types

import (
	"encoding/json"
	"strconv"
	"strings"
)

// BoolOrString 是一个可以接受布尔值或字符串形式布尔值的类型
type BoolOrString struct {
	value bool
}

// UnmarshalJSON 实现自定义的JSON解析逻辑
func (b *BoolOrString) UnmarshalJSON(data []byte) error {
	// 尝试解析为布尔值
	var boolValue bool
	if err := json.Unmarshal(data, &boolValue); err == nil {
		b.value = boolValue
		return nil
	}

	// 尝试解析为字符串
	var strValue string
	if err := json.Unmarshal(data, &strValue); err != nil {
		return err
	}

	// 将字符串转换为布尔值
	strValue = strings.ToLower(strValue)
	if strValue == "true" || strValue == "1" || strValue == "yes" {
		b.value = true
	} else if strValue == "false" || strValue == "0" || strValue == "no" || strValue == "" {
		b.value = false
	} else {
		// 尝试将字符串解析为整数
		if intValue, err := strconv.Atoi(strValue); err == nil {
			b.value = intValue != 0
		} else {
			// 无法解析，默认为false
			b.value = false
		}
	}

	return nil
}

// MarshalJSON 实现自定义的JSON序列化逻辑
func (b BoolOrString) MarshalJSON() ([]byte, error) {
	return json.Marshal(b.value)
}

// Bool 返回布尔值
func (b BoolOrString) Bool() bool {
	return b.value
}

// String 返回字符串表示
func (b BoolOrString) String() string {
	return strconv.FormatBool(b.value)
}
