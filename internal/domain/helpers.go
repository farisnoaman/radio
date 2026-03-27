package domain

import (
	"strconv"
	"strings"
)

func splitAndTrim(s, sep string) []string {
	parts := strings.Split(s, sep)
	var result []string
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

func parseInt(s string) int {
	val, err := strconv.Atoi(s)
	if err != nil {
		return 0
	}
	return val
}

func parseThresholds(s string) []int {
	var result []int
	for _, part := range splitAndTrim(s, ",") {
		val := parseInt(part)
		if val > 0 {
			result = append(result, val)
		}
	}
	return result
}
