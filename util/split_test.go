package util

import (
	"reflect"
	"strings"
	"testing"
)

func TestSplitID(t *testing.T) {
	partsA := SplitID("foo.bar[1].biz")
	partsB := SplitID("foo.bar.1.biz")

	if !reflect.DeepEqual(partsA, strings.Split("foo.bar.1.biz", ".")) {
		t.Error("Should parse correctly", partsA)
	}

	if !reflect.DeepEqual(partsA, partsB) {
		t.Error("Should parse in a similar way", partsB)
	}
}

func TestSplitIDWithMultipleArrayAccess(t *testing.T) {
	partsA := SplitID("foo.bar[1][2].biz")
	partsB := SplitID("foo.bar.1.2.biz")

	if !reflect.DeepEqual(partsA, strings.Split("foo.bar.1.2.biz", ".")) {
		t.Error("Should parse correctly", partsA)
	}

	if !reflect.DeepEqual(partsA, partsB) {
		t.Error("Should parse in a similar way", partsB)
	}
}

func TestConvert(t *testing.T) {
	jsID := ConvertDotIDToJavascript("foo.bar.1.biz")
	if jsID != "foo.bar[1].biz" {
		t.Error("Wrong id returned", jsID)
	}

	dotID := ConvertJSIDToDotID(jsID)
	if dotID != "foo.bar.1.biz" {
		t.Error("Wrong id returned", dotID)
	}

	jsID = ConvertDotIDToJavascript("foo.bar.1")
	if jsID != "foo.bar[1]" {
		t.Error("Wrong id returned", jsID)
	}

	dotID = ConvertJSIDToDotID(jsID)
	if dotID != "foo.bar.1" {
		t.Error("Wrong id returned", dotID)
	}

	jsID = ConvertDotIDToJavascript("foo.bar[1].1")
	if jsID != "foo.bar[1][1]" {
		t.Error("Wrong id returned", jsID)
	}

	dotID = ConvertJSIDToDotID(jsID)
	if dotID != "foo.bar.1.1" {
		t.Error("Wrong id returned", dotID)
	}

	jsID = ConvertDotIDToJavascript("foo.bar.1.1")
	if jsID != "foo.bar[1][1]" {
		t.Error("Wrong id returned", jsID)
	}

	dotID = ConvertJSIDToDotID(jsID)
	if dotID != "foo.bar.1.1" {
		t.Error("Wrong id returned", dotID)
	}

}
