package x

import (
	"strconv"
	"strings"
)

func ToString(obj any) (string, error) {
	if v, ok := obj.(string); ok {
		return v, nil
	} else {
		return "", Errorf("not a string")
	}
}

func ToBool(obj any) (bool, error) {
	if v, ok := obj.(bool); ok {
		return v, nil
	} else {
		return false, Errorf("not a bool")
	}
}

func ToInt(obj any) (int, error) {
	if v, ok := obj.(int); ok {
		return v, nil
	} else {
		return 0, Errorf("not an int")
	}
}

func ToInt64(obj any) (int64, error) {
	if v, ok := obj.(int64); ok {
		return v, nil
	} else {
		return 0, Errorf("not a int64")
	}
}

func ToInt32(obj any) (int32, error) {
	if v, ok := obj.(int32); ok {
		return v, nil
	} else {
		return 0, Errorf("not a int32")
	}
}

func ToInt16(obj any) (int16, error) {
	if v, ok := obj.(int16); ok {
		return v, nil
	} else {
		return 0, Errorf("not a int16")
	}
}

func ToInt8(obj any) (int8, error) {
	if v, ok := obj.(int8); ok {
		return v, nil
	} else {
		return 0, Errorf("not a int8")
	}
}

func ToUint(obj any) (uint, error) {
	if v, ok := obj.(uint); ok {
		return v, nil
	} else {
		return 0, Errorf("not a uint")
	}
}

func ToUint64(obj any) (uint64, error) {
	if v, ok := obj.(uint64); ok {
		return v, nil
	} else {
		return 0, Errorf("not a uint64")
	}
}

func ToUint32(obj any) (uint32, error) {
	if v, ok := obj.(uint32); ok {
		return v, nil
	} else {
		return 0, Errorf("not a uint32")
	}
}

func ToUint16(obj any) (uint16, error) {
	if v, ok := obj.(uint16); ok {
		return v, nil
	} else {
		return 0, Errorf("not a uint16")
	}
}

func ToUint8(obj any) (uint8, error) {
	if v, ok := obj.(uint8); ok {
		return v, nil
	} else {
		return 0, Errorf("not a uint8")
	}
}

func ToFloat64(obj any) (float64, error) {
	if v, ok := obj.(float64); ok {
		return v, nil
	} else {
		return 0, Errorf("not a float64")
	}
}

func ToFloat32(obj any) (float32, error) {
	if v, ok := obj.(float32); ok {
		return v, nil
	} else {
		return 0, Errorf("not a float32")
	}
}

func ToJsonMap(obj any) (map[string]any, error) {
	if v, ok := obj.(map[string]any); ok {
		return v, nil
	} else {
		return nil, Errorf("not a map")
	}
}

func ToJsonArray(obj any) ([]any, error) {
	if v, ok := obj.([]any); ok {
		return v, nil
	} else {
		return nil, Errorf("not an array")
	}
}

func JsonPath(obj any, jPath string) (any, error) {
	attrArray := strings.Split(jPath, ".")

	iter := obj
	for _, attr := range attrArray {
		attrName := ""
		attrIndex := int(-1)
		if strings.Contains(attr, "[") {
			if !strings.HasSuffix(attr, "]") {
				return nil, Errorf("invalid index: %s", attr)
			}

			// parse array like attrName[index]
			parts := strings.Split(attr, "[")
			attrName = parts[0]
			indexStr := strings.TrimRight(parts[1], "]")
			index, err := strconv.Atoi(indexStr)
			if err != nil {
				return nil, Errorf("invalid index: %s", indexStr)
			}
			attrIndex = index
		} else {
			attrName = attr
		}

		if o, err := ToJsonMap(iter); err != nil {
			return nil, err
		} else if subObj, ok := o[attrName]; !ok {
			return nil, Errorf("not found: %s", attrName)
		} else {
			iter = subObj
			if attrIndex >= 0 {
				if a, err := ToJsonArray(iter); err != nil {
					return nil, err
				} else if attrIndex >= len(a) {
					return nil, Errorf("index out of range: %d", attrIndex)
				} else {
					iter = a[attrIndex]
				}
			}
		}
	}

	return iter, nil
}

func JsonPath_ToString(obj any, jPath string) (string, error) {
	if v, err := JsonPath(obj, jPath); err != nil {
		return "", err
	} else {
		return ToString(v)
	}
}

func JsonPath_ToStringArray(obj any, jPath string) ([]string, error) {
	if arrObj, err := JsonPath(obj, jPath); err != nil {
		return nil, err
	} else if arrRaw, err := ToJsonArray(arrObj); err != nil {
		return nil, err
	} else {
		ret := make([]string, 0)

		for _, v := range arrRaw {
			if s, err := ToString(v); err != nil {
				return nil, err
			} else {
				ret = append(ret, s)
			}
		}

		return ret, nil
	}
}

func JsonPath_ToStringMap(obj any, jPath string) (map[string]string, error) {
	if mapObj, err := JsonPath(obj, jPath); err != nil {
		return nil, err
	} else if mapRaw, err := ToJsonMap(mapObj); err != nil {
		return nil, err
	} else {
		ret := make(map[string]string)

		for k, v := range mapRaw {
			if s, err := ToString(v); err != nil {
				return nil, err
			} else {
				ret[k] = s
			}
		}

		return ret, nil
	}
}

func JsonPath_ToBool(obj any, jPath string) (bool, error) {
	if v, err := JsonPath(obj, jPath); err != nil {
		return false, err
	} else {
		return ToBool(v)
	}
}

func JsonPath_ToInt(obj any, jPath string) (int, error) {
	if v, err := JsonPath(obj, jPath); err != nil {
		return 0, err
	} else {
		return ToInt(v)
	}
}

func JsonPath_ToInt64(obj any, jPath string) (int64, error) {
	if v, err := JsonPath(obj, jPath); err != nil {
		return 0, err
	} else {
		return ToInt64(v)
	}
}

func JsonPath_ToInt32(obj any, jPath string) (int32, error) {
	if v, err := JsonPath(obj, jPath); err != nil {
		return 0, err
	} else {
		return ToInt32(v)
	}
}

func JsonPath_ToInt16(obj any, jPath string) (int16, error) {
	if v, err := JsonPath(obj, jPath); err != nil {
		return 0, err
	} else {
		return ToInt16(v)
	}
}

func JsonPath_ToInt8(obj any, jPath string) (int8, error) {
	if v, err := JsonPath(obj, jPath); err != nil {
		return 0, err
	} else {
		return ToInt8(v)
	}
}

func JsonPath_ToUint(obj any, jPath string) (uint, error) {
	if v, err := JsonPath(obj, jPath); err != nil {
		return 0, err
	} else {
		return ToUint(v)
	}
}

func JsonPath_ToUint64(obj any, jPath string) (uint64, error) {
	if v, err := JsonPath(obj, jPath); err != nil {
		return 0, err
	} else {
		return ToUint64(v)
	}
}

func JsonPath_ToUint32(obj any, jPath string) (uint32, error) {
	if v, err := JsonPath(obj, jPath); err != nil {
		return 0, err
	} else {
		return ToUint32(v)
	}
}

func JsonPath_ToUint16(obj any, jPath string) (uint16, error) {
	if v, err := JsonPath(obj, jPath); err != nil {
		return 0, err
	} else {
		return ToUint16(v)
	}
}

func JsonPath_ToUint8(obj any, jPath string) (uint8, error) {
	if v, err := JsonPath(obj, jPath); err != nil {
		return 0, err
	} else {
		return ToUint8(v)
	}
}

func JsonPath_ToFloat64(obj any, jPath string) (float64, error) {
	if v, err := JsonPath(obj, jPath); err != nil {
		return 0, err
	} else {
		return ToFloat64(v)
	}
}

func JsonPath_ToFloat32(obj any, jPath string) (float32, error) {
	if v, err := JsonPath(obj, jPath); err != nil {
		return 0, err
	} else {
		return ToFloat32(v)
	}
}
