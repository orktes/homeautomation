package util

import "strings"

func ConvertValueToTopic(str string, typ string) string {
	parts := strings.Split(str, "/")

	fullTopic := make([]string, len(parts)+1)
	fullTopic[0] = parts[0]
	fullTopic[1] = typ
	copy(fullTopic[2:], parts[1:])

	return strings.Join(fullTopic, "/")
}

func ConvertFloatValueToRange(inputRange, outputRange []float64, val float64) (float64, error) {
	if len(inputRange) < 2 || len(outputRange) < 2 {
		return val, nil
	}

	inputRangeStart := inputRange[0]
	inputRangeEnd := inputRange[1]

	outputRangeStart := outputRange[0]
	outputRangeEnd := outputRange[1]

	if inputRangeStart == outputRangeStart && inputRangeEnd == outputRangeEnd {
		return val, nil
	}

	inputPos := (val - inputRangeStart) / (inputRangeEnd - inputRangeStart)

	outputRangeDelta := outputRangeEnd - outputRangeStart

	if inputPos < 0 {
		inputPos = 0
	} else if inputPos > 1 {
		inputPos = 1
	}

	return outputRangeStart + (inputPos * outputRangeDelta), nil
}
