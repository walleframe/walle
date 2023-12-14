package rdconv

import (
	"fmt"
	"strconv"

	"github.com/walleframe/walle/util"
)

func AnyToInt8(val interface{}) (int8, error) {
	switch val := val.(type) {
	case int64:
		return int8(val), nil
	case string:
		v, err := strconv.ParseInt(val, 10, 8)
		if err != nil {
			return 0, err
		}
		return int8(v), nil
	default:
		return 0, fmt.Errorf("redis: unexpected type=%T for int8", val)
	}
}

func AnyToInt16(val interface{}) (int16, error) {
	switch val := val.(type) {
	case int64:
		return int16(val), nil
	case string:
		v, err := strconv.ParseInt(val, 10, 16)
		if err != nil {
			return 0, err
		}
		return int16(v), nil
	default:
		return 0, fmt.Errorf("redis: unexpected type=%T for int16", val)
	}
}

func AnyToInt32(val interface{}) (int32, error) {
	switch val := val.(type) {
	case int64:
		return int32(val), nil
	case string:
		v, err := strconv.ParseInt(val, 10, 32)
		if err != nil {
			return 0, err
		}
		return int32(v), nil
	default:
		return 0, fmt.Errorf("redis: unexpected type=%T for int32", val)
	}
}

func AnyToInt64(val interface{}) (int64, error) {
	switch val := val.(type) {
	case int64:
		return int64(val), nil
	case string:
		v, err := strconv.ParseInt(val, 10, 64)
		if err != nil {
			return 0, err
		}
		return v, nil
	default:
		return 0, fmt.Errorf("redis: unexpected type=%T for int64", val)
	}
}

func AnyToUint8(val interface{}) (uint8, error) {
	switch val := val.(type) {
	case int64:
		return uint8(val), nil
	case string:
		v, err := strconv.ParseUint(val, 10, 8)
		if err != nil {
			return 0, err
		}
		return uint8(v), nil
	default:
		return 0, fmt.Errorf("redis: unexpected type=%T for uint8", val)
	}
}

func AnyToUint16(val interface{}) (uint16, error) {
	switch val := val.(type) {
	case int64:
		return uint16(val), nil
	case string:
		v, err := strconv.ParseUint(val, 10, 16)
		if err != nil {
			return 0, err
		}
		return uint16(v), nil
	default:
		return 0, fmt.Errorf("redis: unexpected type=%T for uint16", val)
	}
}

func AnyToUint32(val interface{}) (uint32, error) {
	switch val := val.(type) {
	case int64:
		return uint32(val), nil
	case string:
		v, err := strconv.ParseUint(val, 10, 32)
		if err != nil {
			return 0, err
		}
		return uint32(v), nil
	default:
		return 0, fmt.Errorf("redis: unexpected type=%T for uint32", val)
	}
}

func AnyToUint64(val interface{}) (uint64, error) {
	switch val := val.(type) {
	case int64:
		return uint64(val), nil
	case string:
		v, err := strconv.ParseUint(val, 10, 64)
		if err != nil {
			return 0, err
		}
		return v, nil
	default:
		return 0, fmt.Errorf("redis: unexpected type=%T for uint64", val)
	}
}

func AnyToBool(val interface{}) (bool, error) {
	switch val := val.(type) {
	case bool:
		return val, nil
	case int64:
		return val != 0, nil
	case string:
		return strconv.ParseBool(val)
	default:
		return false, fmt.Errorf("redis: unexpected type=%T for bool", val)
	}
}

func AnyToBinary(val interface{}) ([]byte, error) {
	switch val := val.(type) {
	case string:
		// 此处需要拷贝内存,不能使用 StringToBytes
		return []byte(val), nil
	default:
		return nil, fmt.Errorf("redis: unexpected type=%T for bool", val)
	}
}

func AnyToString(val interface{}) (string, error) {
	switch val := val.(type) {
	case bool:
		return strconv.FormatBool(val), nil
	case int64:
		return strconv.FormatInt(val, 10), nil
	case string:
		return val, nil
	default:
		return "", fmt.Errorf("redis: unexpected type=%T for bool", val)
	}
}

func AnyToFloat32(val interface{}) (float32, error) {
	switch val := val.(type) {
	case int64:
		return float32(val), nil
	case string:
		v, err := strconv.ParseFloat(val, 32)
		if err != nil {
			return 0, err
		}
		return float32(v), nil
	default:
		return 0, fmt.Errorf("redis: unexpected type=%T for uint64", val)
	}
}

func AnyToFloat64(val interface{}) (float64, error) {
	switch val := val.(type) {
	case int64:
		return float64(val), nil
	case string:
		v, err := strconv.ParseFloat(val, 64)
		if err != nil {
			return 0, err
		}
		return float64(v), nil
	default:
		return 0, fmt.Errorf("redis: unexpected type=%T for uint64", val)
	}
}

// 用于临时转换数据结构,[]byte 只读.
func AnyToBytes(val interface{}) ([]byte, error) {
	switch val := val.(type) {
	case string:
		return util.StringToBytes(val), nil
	default:
		return nil, fmt.Errorf("redis: unexpected type=%T for bool", val)
	}
}

func StringToInt8(val string) (int8, error) {
	v, err := strconv.ParseInt(val, 10, 8)
	if err != nil {
		return 0, err
	}
	return int8(v), nil
}

func StringToInt16(val string) (int16, error) {
	v, err := strconv.ParseInt(val, 10, 16)
	if err != nil {
		return 0, err
	}
	return int16(v), nil
}

func StringToInt32(val string) (int32, error) {
	v, err := strconv.ParseInt(val, 10, 32)
	if err != nil {
		return 0, err
	}
	return int32(v), nil
}

func StringToInt64(val string) (int64, error) {
	v, err := strconv.ParseInt(val, 10, 64)
	if err != nil {
		return 0, err
	}
	return v, nil
}

func StringToUint8(val string) (uint8, error) {
	v, err := strconv.ParseUint(val, 10, 8)
	if err != nil {
		return 0, err
	}
	return uint8(v), nil
}

func StringToUint16(val string) (uint16, error) {
	v, err := strconv.ParseUint(val, 10, 16)
	if err != nil {
		return 0, err
	}
	return uint16(v), nil
}

func StringToUint32(val string) (uint32, error) {

	v, err := strconv.ParseUint(val, 10, 32)
	if err != nil {
		return 0, err
	}
	return uint32(v), nil
}

func StringToUint64(val string) (uint64, error) {
	v, err := strconv.ParseUint(val, 10, 64)
	if err != nil {
		return 0, err
	}
	return v, nil
}

func StringToBool(val string) (bool, error) {
	return strconv.ParseBool(val)
}

func StringToBinary(val string) ([]byte, error) {
	// 此处需要拷贝内存,不能使用 StringToBytes
	return []byte(val), nil
}

func StringToFloat32(val string) (float32, error) {
	v, err := strconv.ParseFloat(val, 32)
	if err != nil {
		return 0, err
	}
	return float32(v), nil
}

func StringToFloat64(val string) (float64, error) {
	v, err := strconv.ParseFloat(val, 64)
	if err != nil {
		return 0, err
	}
	return float64(v), nil
}

func Int8ToString(val int8) string {
	return strconv.FormatInt(int64(val), 10)
}

func Int16ToString(val int16) string {
	return strconv.FormatInt(int64(val), 10)
}
func Int32ToString(val int32) string {
	return strconv.FormatInt(int64(val), 10)
}
func Int64ToString(val int64) string {
	return strconv.FormatInt(val, 10)
}

func Uint8ToString(val uint8) string {
	return strconv.FormatUint(uint64(val), 10)
}

func Uint16ToString(val uint16) string {
	return strconv.FormatUint(uint64(val), 10)
}
func Uint32ToString(val uint32) string {
	return strconv.FormatUint(uint64(val), 10)
}
func Uint64ToString(val uint64) string {
	return strconv.FormatUint(val, 10)
}

func BinaryToString(val []byte) string {
	return util.BytesToString(val)
}

func Float32ToString(val float32) string {
	return strconv.FormatFloat(float64(val), 'f', -1, 32)
}

func Float64ToString(val float64) string {
	return strconv.FormatFloat(val, 'f', -1, 64)
}

func BoolToString(val bool) string {
	if val {
		return "1"
	}
	return "0"
}

func StringToString(val string) string {
	return val
}
