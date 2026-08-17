package util

import (
	"fmt"
	"regexp"
	"strings"
	"strconv"
)

var verbose = true

func PrintVerbose(format string, a ...interface{}) {
	if verbose {
		fmt.Printf(format+"\n", a...)
	}
}

func SplitNumbers(s string) ([]int64, error) {
	re := regexp.MustCompile(`(?:Rp\.?\s*)?([\d.]+)(?:,\d+)?`)
	matches := re.FindAllStringSubmatch(s, -1)

	var numbers []int64

	for _, match := range matches {
		if len(match) < 2 {
			continue
		}

		value := strings.ReplaceAll(match[1], ".", "")

		num, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			continue
		}

		numbers = append(numbers, num)
	}

	return numbers, nil
}

