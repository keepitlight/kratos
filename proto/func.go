package proto

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
)

// UnmarshalJSONForGeneratedMessage 全局函数：通过反射智能解析 JSON 到 Protobuf 生成的结构体
func UnmarshalJSONForGeneratedMessage(data []byte, msg interface{}) error {
	// 1. 先用标准库解析到 map，以保留所有字段的原始 JSON 类型
	var rawMap map[string]json.RawMessage
	if err := json.Unmarshal(data, &rawMap); err != nil {
		return err
	}

	// 2. 获取目标结构体的反射值
	v := reflect.ValueOf(msg).Elem()
	t := v.Type()

	// 3. 遍历结构体的所有字段
	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		fieldType := t.Field(i)

		// 获取 JSON 标签名（protobuf生成代码通常包含 `json:”field_name”`）
		jsonTag := fieldType.Tag.Get("json")
		if jsonTag == "" || jsonTag == "-" {
			continue
		}
		jsonFieldName := strings.Split(jsonTag, ",")[0] // 处理 `json:"name,omitempty"`

		// 检查 JSON 数据中是否有这个字段
		if rawValue, ok := rawMap[jsonFieldName]; ok {
			// 调用处理单个字段的函数
			if err := setFieldFromJSON(field, rawValue); err != nil {
				return fmt.Errorf("failed to parse field %s: %v", jsonFieldName, err)
			}
		}
	}
	return nil
}

// 核心：根据字段类型（特别是枚举）设置值
func setFieldFromJSON(field reflect.Value, rawValue json.RawMessage) error {
	// 处理指针类型的字段（protobuf消息字段通常是指针）
	if field.Kind() == reflect.Ptr {
		if string(rawValue) == "null" {
			return nil // JSON 的 null 对应 Go 的 nil 指针
		}
		if field.IsNil() {
			field.Set(reflect.New(field.Type().Elem())) // 创建新实例
		}
		// 递归处理指针指向的值
		return setFieldFromJSON(field.Elem(), rawValue)
	}

	// 处理枚举类型（protobuf生成的枚举底层是 int32）
	if field.Type().Name() == "Status" { // 这里需要为你的每个枚举类型添加判断
		// 尝试解析为数字
		var intVal int32
		if err := json.Unmarshal(rawValue, &intVal); err == nil {
			field.SetInt(int64(intVal))
			return nil
		}
		// 尝试解析为字符串，并映射到枚举值
		var strVal string
		if err := json.Unmarshal(rawValue, &strVal); err == nil {
			// 关键：使用 Protobuf 生成的映射表
			// 假设你有一个全局函数或映射来查询，例如：GetEnumValue(“Status”, strVal)
			// 或者为每个枚举类型单独判断：
			if strings.ToUpper(strVal) == "ENABLED" {
				field.SetInt(1) // 对应 Status_Enabled
			} else if strings.ToUpper(strVal) == "DISABLED" {
				field.SetInt(0) // 对应 Status_Disabled
			} else {
				return fmt.Errorf("unknown enum value: %s", strVal)
			}
			return nil
		}
		return fmt.Errorf("enum field must be int or string")
	}

	// 对于非枚举的普通字段，直接使用标准库反序列化
	return json.Unmarshal(rawValue, field.Addr().Interface())
}
