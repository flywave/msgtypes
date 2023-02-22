package msgtypes

import (
	"math"
	"time"
)

func lcp(l []string) string {
	switch len(l) {
	case 0:
		return ""
	case 1:
		return l[0]
	}
	min, max := l[0], l[0]
	for _, s := range l[1:] {
		switch {
		case s < min:
			min = s
		case s > max:
			max = s
		}
	}
	for i := 0; i < len(min) && i < len(max); i++ {
		if min[i] != max[i] {
			return min[:i]
		}
	}
	return min
}

func maxUnit(units map[Unit]int) (unit Unit) {
	maxV := 1
	for u, c := range units {
		if c > maxV {
			unit = u
			maxV = c
		}
	}

	return
}

func intToTime(t int64) time.Time {
	return time.Unix(t, 0)
}

func floatToTime(t float64) time.Time {
	s, n := math.Modf(t)
	return time.Unix(int64(s), int64(n*1e9))
}

func floatToDuration(d float64) time.Duration {
	return time.Duration(d * float64(time.Second))
}

func parseTime(base, val Numeric, now time.Time) (t time.Time) {
	baseFloat := numericToFloat64(base)
	if base == nil || baseFloat == 0 {
		t = now
	} else if baseFloat >= (1 << 28) {
		t = numericToTime(base)
	} else {
		t = now.Add(numericToDuration(base))
	}

	if val == nil {
		return
	}

	if t.IsZero() {
		return numericToTime(val)
	}

	return t.Add(numericToDuration(val))
}

func pow10(n int) (v int) {
	if n < 0 {
		panic("n must be positive")
	}
	v = 1
	for i := 0; i < n; i++ {
		v *= 10
	}
	return
}
