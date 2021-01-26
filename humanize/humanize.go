package humanize

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

func logN(n, b float64) float64 {
	return math.Log(n) / math.Log(b)
}

func HumValue(s uint64, base float64, sizes []string) string {
	if s < 10 {
		return fmt.Sprintf("%d B", s)
	}
	e := math.Floor(logN(float64(s), base))
	suffix := sizes[int(e)]
	val := math.Floor(float64(s)/math.Pow(base, e)*10+0.5) / 10
	f := "%.0f %s"
	if val < 10 {
		f = "%.1f %s"
	}

	return fmt.Sprintf(f, val, suffix)
}

var (
	sufShortSI = []string{"B", "KB", "MB", "GB", "TB", "PB", "EB"}
	sufLongSI  = []string{"byte", "Kbyte", "Mbyte", "Gbyte", "Tbyte", "Pbyte", "Ebyte"}
)

// produces a human readable representation of an SI size. (82854982) -> 83 MB
func BytesSi(s uint64) string {
	return HumValue(s, 1000, sufShortSI)
}

// produces a human readable representation of an IEC size. (82854982) -> 79 MiB
func Bytes(s uint64) string {
	return HumValue(s, 1024, sufLongSI)
}

func Comma(v int64) string {
	sign := ""

	// Min int64 can't be negated to a usable value, so it has to be special cased.
	if v == math.MinInt64 {
		return "-9,223,372,036,854,775,808"
	}

	if v < 0 {
		sign = "-"
		v = 0 - v
	}

	parts := []string{"", "", "", "", "", "", ""}
	j := len(parts) - 1

	for v > 999 {
		parts[j] = strconv.FormatInt(v%1000, 10)
		switch len(parts[j]) {
		case 2:
			parts[j] = "0" + parts[j]
		case 1:
			parts[j] = "00" + parts[j]
		}
		v = v / 1000
		j--
	}
	parts[j] = strconv.Itoa(int(v))
	return sign + strings.Join(parts[j:], ".")
}
