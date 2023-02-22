package msgtypes

import (
	"fmt"
	"time"
)

type Numeric interface{}

func sumNumeric(a, b Numeric) Numeric {
	switch a.(type) {
	case nil:
		return b
	case int, int8, int16, int32, int64:
		return addToNumericInt(a, b)
	case uint, uint8, uint16, uint32, uint64:
		return addToNumericUint(a, b)
	case float32, float64, Decimal, *Decimal:
		return numericToFloat64(a) + numericToFloat64(b)
	default:
		panic(fmt.Sprintf("invalid value type: %T", a))
	}
}

func addToNumericInt(a, b Numeric) Numeric {
	switch b.(type) {
	case nil:
		return a
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		return numericToInt64(a) + numericToInt64(b)
	case float32, float64, Decimal, *Decimal:
		return numericToFloat64(a) + numericToFloat64(b)
	default:
		panic(fmt.Sprintf("invalid value type: %T", b))
	}
}

func addToNumericUint(a, b Numeric) Numeric {
	switch b.(type) {
	case nil:
		return a
	case int, int8, int16, int32, int64:
		return numericToInt64(a) + numericToInt64(b)
	case uint, uint8, uint16, uint32, uint64:
		return numericToUint64(a) + numericToUint64(b)
	case float32, float64, Decimal, *Decimal:
		return numericToFloat64(a) + numericToFloat64(b)
	default:
		panic(fmt.Sprintf("invalid value type: %T", b))
	}
}

func timeToNumeric(t time.Time) Numeric {
	if t.IsZero() {
		return nil
	}
	if t.Nanosecond() == 0 {
		return t.Unix()
	}
	return float64(t.UnixNano()) / 1e9
}

func numericToTime(v Numeric) time.Time {
	switch v.(type) {
	case nil:
		return time.Time{}
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		return intToTime(numericToInt64(v))
	case float32, float64, Decimal, *Decimal:
		return floatToTime(numericToFloat64(v))
	default:
		panic("unsupported type")
	}
}

func numericToDuration(v Numeric) time.Duration {
	switch v.(type) {
	case nil:
		return 0
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		return time.Duration(numericToInt64(v)) * time.Second
	case float32, float64, Decimal, *Decimal:
		return floatToDuration(numericToFloat64(v))
	default:
		panic("unsupported type")
	}
}

func numericToInt64(v Numeric) int64 {
	switch i := v.(type) {
	case nil:
		return 0
	case int:
		return int64(i)
	case int8:
		return int64(i)
	case int16:
		return int64(i)
	case int32:
		return int64(i)
	case int64:
		return i
	case uint:
		return int64(i)
	case uint8:
		return int64(i)
	case uint16:
		return int64(i)
	case uint32:
		return int64(i)
	case uint64:
		return int64(i)
	case float32:
		return int64(i)
	case float64:
		return int64(i)
	case Decimal:
		return int64(i.Int())
	case *Decimal:
		return int64(i.Int())
	default:
		panic(fmt.Sprintf("invalid value type: %T", i))
	}
}

func numericToUint64(v Numeric) uint64 {
	switch i := v.(type) {
	case nil:
		return 0
	case int:
		return uint64(i)
	case int8:
		return uint64(i)
	case int16:
		return uint64(i)
	case int32:
		return uint64(i)
	case int64:
		return uint64(i)
	case uint:
		return uint64(i)
	case uint8:
		return uint64(i)
	case uint16:
		return uint64(i)
	case uint32:
		return uint64(i)
	case uint64:
		return i
	case float32:
		return uint64(i)
	case float64:
		return uint64(i)
	case Decimal:
		return uint64(i.Int())
	case *Decimal:
		return uint64(i.Int())
	default:
		panic(fmt.Sprintf("invalid value type: %T", i))
	}
}

func numericToFloat64(v Numeric) float64 {
	switch f := v.(type) {
	case nil:
		return 0
	case int:
		return float64(f)
	case int8:
		return float64(f)
	case int16:
		return float64(f)
	case int32:
		return float64(f)
	case int64:
		return float64(f)
	case uint:
		return float64(f)
	case uint8:
		return float64(f)
	case uint16:
		return float64(f)
	case uint32:
		return float64(f)
	case uint64:
		return float64(f)
	case float32:
		return float64(f)
	case float64:
		return f
	case Decimal:
		return f.Float()
	case *Decimal:
		return f.Float()
	default:
		panic(fmt.Sprintf("invalid value type: %T", f))
	}
}
