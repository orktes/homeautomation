package util

import (
	"regexp"
	"strings"
)

var arrAccessRegex = regexp.MustCompile("\\[([0-9]+)\\]")
var numberPropAccessRegex = regexp.MustCompile("\\.([0-9]+)")

func ConvertDotIDToJavascript(id string) string {
	return numberPropAccessRegex.ReplaceAllStringFunc(id, func(part string) string {
		return "[" + part[1:] + "]"
	})
}

func ConvertJSIDToDotID(id string) string {
	return arrAccessRegex.ReplaceAllStringFunc(id, func(part string) string {
		return "." + part[1:len(part)-1]
	})
}

func SplitID(id string) []string {
	return strings.Split(ConvertJSIDToDotID(id), ".")
}
