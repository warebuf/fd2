package util

import (
	"fmt"
	"strconv"
	"strings"
)

// Converts a slice of ints to a comma-separated string
func IntsToString(nums []int) string {
	strs := make([]string, len(nums))
	for i, n := range nums {
		strs[i] = strconv.Itoa(n)
	}
	return strings.Join(strs, ",")
}

// Converts a slice of floats to a comma-separated string with given precision
func FloatsToString(nums []float64, precision int) string {
	strs := make([]string, len(nums))
	format := fmt.Sprintf("%%.%df", precision)
	for i, n := range nums {
		strs[i] = fmt.Sprintf(format, n)
	}
	return strings.Join(strs, ",")
}
