package util

import (
	"fmt"
	"log"
	"math"
	"runtime"
)

const earthRadius = float64(6378137)

// In meters
func Haversine(fromLat, fromLong, toLat, toLong float64) (distance float64) {
	var deltaLat = (toLat - fromLat) * (math.Pi / 180)
	var deltaLon = (toLong - fromLong) * (math.Pi / 180)

	var a = math.Sin(deltaLat/2)*math.Sin(deltaLat/2) +
		math.Cos(fromLat*(math.Pi/180))*math.Cos(toLat*(math.Pi/180))*
			math.Sin(deltaLon/2)*math.Sin(deltaLon/2)
	var c = 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	distance = earthRadius * c
	return
}

func RoundUp(val float64, precision int) float64 {
	return math.Ceil(val*(math.Pow10(precision))) / math.Pow10(precision)
}

func Contains[T comparable](elems []T, v T) bool {
	for _, s := range elems {
		if v == s {
			return true
		}
	}
	return false
}

func LogErr(err error, args ...any) {
	if err == nil {
		return
	}
	_, file, line, _ := runtime.Caller(1)
	critical := false
	for _, arg := range args {
		if arg == "critical" {
			critical = true
		}
	}
	criticalStr := "Critical Error"
	if !critical {
		criticalStr = "Error"
	}
	errStr := fmt.Sprintf("%v\n%s occurred at %s:%d\n", err, criticalStr, file, line)
	log.Print(errStr)
}
