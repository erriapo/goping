// Package thirdparty are code that are licensed
// according to their respective authors.
package thirdparty

import (
	"fmt"
	"github.com/erriapo/stats"
	"math"
)

// round will round up the nearest integer.
// Source: https://github.com/golang/go/issues/4594#issuecomment-135336012
// Author: James Hartig https://github.com/fastest963
func round(f float64) int {
	if math.Abs(f) < 0.5 {
		return 0
	}
	return int(f + math.Copysign(0.5, f))
}

// ToFixed truncates a float into a specified precision
// Source: https://stackoverflow.com/a/29786394
// Author: David Calhoun (Apr 22 15)
// License: http://creativecommons.org/licenses/by-sa/3.0/
func ToFixed(num float64, precision int) float64 {
	output := math.Pow(10, float64(precision))
	return float64(round(num*output)) / output
}

func Format(sink *stats.WelfordSink) string {
	return fmt.Sprintf("rtt min/avg/max/mdev = %v/%v/%v/%v ms", ToFixed(sink.Min(), 3),
		ToFixed(sink.Mean(), 3), ToFixed(sink.Max(), 3), ToFixed(sink.StandardDeviation(), 3))
}
