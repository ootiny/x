package x

import (
	"encoding/json"
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

func ToFloat64(obj any) (float64, error) {
	if v, ok := obj.(float64); ok {
		return v, nil
	} else {
		return 0, Errorf("not a float64")
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

func JsonPath(v any, jPath string) (any, error) {
	var obj any

	if vBytes, err := json.Marshal(v); err != nil {
		return nil, err
	} else if err := json.Unmarshal(vBytes, &obj); err != nil {
		return nil, err
	} else {

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
	} else if floatv, err := ToFloat64(v); err != nil {
		return 0, err
	} else {
		return int(floatv), nil
	}
}

func JsonPath_ToFloat64(obj any, jPath string) (float64, error) {
	if v, err := JsonPath(obj, jPath); err != nil {
		return 0, err
	} else {
		return ToFloat64(v)
	}
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
	if mapObj, err := JsonPath_ToMap(obj, jPath); err != nil {
		return nil, err
	} else {
		ret := make(map[string]string)

		for k, v := range mapObj {
			if s, err := ToString(v); err != nil {
				return nil, err
			} else {
				ret[k] = s
			}
		}

		return ret, nil
	}
}

func JsonPath_ToMap(obj any, jPath string) (map[string]any, error) {
	if mapObj, err := JsonPath(obj, jPath); err != nil {
		return nil, err
	} else if mapRaw, err := ToJsonMap(mapObj); err != nil {
		return nil, err
	} else {
		return mapRaw, nil
	}
}
