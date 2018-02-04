package util

import "testing"

type rangeTestData struct {
	inputRange  []float64
	outputRange []float64
	input       float64
	output      float64
}

func TestConvertFloatValueToRange(t *testing.T) {
	testData := []rangeTestData{
		{[]float64{0, 0}, []float64{0, 0}, 0, 0},
		{[]float64{0, 100}, []float64{100, 0}, 0, 100},
		{[]float64{0, 100}, []float64{100, 0}, -10, 100},
		{[]float64{100, 0}, []float64{100, 0}, 0, 0},
		{[]float64{0, 50}, []float64{0, 100}, 50, 100},
		{[]float64{0, 100}, []float64{0, 50}, 100, 50},
		{[]float64{0, 100}, []float64{-90, 0}, 100, 0},
		{[]float64{0, 100}, []float64{-90, 0}, 50, -45},
		{[]float64{0, 100}, []float64{-90, 0}, -50, -90},
	}

	for _, test := range testData {
		res, err := ConvertFloatValueToRange(test.inputRange, test.outputRange, test.input)
		if err != nil {
			t.Error(err)
		}

		if res != test.output {
			t.Errorf("ConvertValueToRange returned wrong result %+v got %f\n", test, res)
		}
	}
}
