package cache

import (
	"fmt"
	core_utils "github.com/siper92/core-utils"
	"strconv"
)

func ToMapValue(rawVal any) string {
	return _toRedisMapValue(rawVal, 1)
}

func _toRedisMapValue(rawVal any, depth int) string {
	if depth > 3 {
		return ""
	}

	switch val := rawVal.(type) {
	case string, *string,
		[]byte, *[]byte,
		int, int32, int64, uint, uint32, uint64,
		*int, *int32, *int64, *uint, *uint32, *uint64,
		float32, float64, *float32, *float64,
		bool, *bool,
		fmt.Stringer:
		return core_utils.ToString(val)
	case CacheableObject:
		return val.CacheKey()
	}

	core_utils.Debug("unsupported type: %T", rawVal)
	return fmt.Sprintf("%v", rawVal)
}

func StrToInt(val string) int {
	i, err := strconv.Atoi(val)
	if err != nil {
		return 0
	}

	return i
}

func StrToInt64(val string) int64 {
	i, err := strconv.ParseInt(val, 10, 64)
	if err != nil {
		return 0
	}

	return i
}

func StrToUint64(val string) uint64 {
	i, err := strconv.ParseUint(val, 10, 64)
	if err != nil {
		return 0
	}

	return i
}

func StrToFloat64(val string) float64 {
	i, err := strconv.ParseFloat(val, 64)
	if err != nil {
		return 0
	}

	return i
}

func StrToBool(val string) bool {
	b, err := strconv.ParseBool(val)
	if err != nil {
		return false
	}

	return b
}
