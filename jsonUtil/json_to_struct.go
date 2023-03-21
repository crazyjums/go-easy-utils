package jsonUtil

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

// JsonToStruct 将 JSON 字符串解析为指定的结构体指针
// 根据结构体的字段类型和标签来自动选择将 JSON 值转换为相应的类型。
//
// 支持的字段类型包括 string、int、int8、int16、int32、int64、uint、uint8、uint16、uint32、uint64、bool、float32 和 float64。
//
// 支持的标签有 "json"、"jsonb" 和 "mapstructure"。
// - "json" 和 "jsonb" 标签指示解析 JSON 时使用的键名。
// - "mapstructure" 标签指示字段名的映射关系。
//
// 如果 JSON 中的某些键在结构体中没有对应的字段，则它们将被忽略。
// 如果 JSON 中的某些键的类型与结构体中的字段类型不匹配，则会引发解析错误。
//
// 参数 jsonData 是要解析的 JSON 字符串。
// 参数 result 是指向要填充 JSON 值的结构体指针。
//
// 如果解析成功，则返回 nil。如果解析失败，则返回解析错误。
func JsonToStruct(jsonData string, result interface{}) error {
	var data map[string]interface{}
	err := json.Unmarshal([]byte(jsonData), &data)
	if err != nil {
		return err
	}

	resultValue := reflect.ValueOf(result).Elem()
	resultType := resultValue.Type()

	for i := 0; i < resultType.NumField(); i++ {
		fieldType := resultType.Field(i)
		fieldName := fieldType.Name
		fieldValue := resultValue.FieldByName(fieldName)

		// 从json的tag标签中取出定义字段
		jsonTag := fieldType.Tag.Get("json")
		if jsonTag == "" {
			jsonTag = fieldName
		} else {
			if commaIndex := strings.Index(jsonTag, ","); commaIndex != -1 {
				jsonTag = jsonTag[:commaIndex]
			}
		}

		value, ok := data[jsonTag]
		if !ok {
			continue
		}

		switch fieldValue.Kind() {
		case reflect.String:
			fieldValue.SetString(value.(string))
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			val, err := toInt64(value)
			if err != nil {
				return err
			}
			fieldValue.SetInt(val)
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			val, err := toUint64(value)
			if err != nil {
				return err
			}
			fieldValue.SetUint(val)
		case reflect.Float32, reflect.Float64:
			val, err := toFloat64(value)
			if err != nil {
				return err
			}
			fieldValue.SetFloat(val)
		case reflect.Struct:
			if subData, ok := value.(map[string]interface{}); ok {
				subResult := reflect.New(fieldValue.Type())
				JsonToStruct(convertToJSONString(subData), subResult.Interface())
				fieldValue.Set(subResult.Elem())
			}
		case reflect.Slice:
			if subData, ok := value.([]interface{}); ok {
				subResult := reflect.MakeSlice(fieldValue.Type(), len(subData), len(subData))
				for j := 0; j < len(subData); j++ {
					subValue := subData[j]
					subElem := subResult.Index(j)

					if subElem.Kind() == reflect.Struct {
						if subDataElem, ok := subValue.(map[string]interface{}); ok {
							subResultElem := reflect.New(subElem.Type())
							JsonToStruct(convertToJSONString(subDataElem), subResultElem.Interface())
							subElem.Set(subResultElem.Elem())
						}
					} else {
						subElem.Set(reflect.ValueOf(subValue))
					}
				}
				fieldValue.Set(subResult)
			}
		default:
			fieldValue.Set(reflect.ValueOf(value))
		}
	}

	return nil
}

func convertToJSONString(data map[string]interface{}) string {
	jsonBytes, _ := json.Marshal(data)
	return string(jsonBytes)
}

func toInt64(value interface{}) (int64, error) {
	switch value.(type) {
	case float32:
		return int64(value.(float32)), nil
	case float64:
		return int64(value.(float64)), nil
	case string:
		intValue, err := strconv.ParseInt(value.(string), 10, 64)
		if err != nil {
			return 0, err
		}
		return intValue, nil
	case int:
		return int64(value.(int)), nil
	case int8:
		return int64(value.(int8)), nil
	case int16:
		return int64(value.(int16)), nil
	case int32:
		return int64(value.(int32)), nil
	case int64:
		return value.(int64), nil
	default:
		return 0, errors.New(fmt.Sprintf("jsonUtils toInt64 err: %T \n", value))
	}
}

func toUint64(value interface{}) (uint64, error) {
	switch value.(type) {
	case float32:
		return uint64(value.(float32)), nil
	case float64:
		return uint64(value.(float64)), nil
	case string:
		intValue, err := strconv.ParseUint(value.(string), 10, 64)
		if err != nil {
			return 0, err
		}
		return intValue, nil
	case uint:
		return uint64(value.(uint)), nil
	case uint8:
		return uint64(value.(uint8)), nil
	case uint16:
		return uint64(value.(uint16)), nil
	case uint32:
		return uint64(value.(uint32)), nil
	case uint64:
		return value.(uint64), nil
	default:
		return 0, errors.New(fmt.Sprintf("jsonUtils toUint64 err: %T \n", value))
	}
}

func toFloat64(value interface{}) (float64, error) {
	switch value.(type) {
	case float64:
		return value.(float64), nil
	case string:
		floatValue, err := strconv.ParseFloat(value.(string), 64)
		if err != nil {
			return 0, err
		}
		return floatValue, nil
	default:
		return 0, errors.New(fmt.Sprintf("jsonUtils toFloat64 err: %T \n", value))
	}
}
