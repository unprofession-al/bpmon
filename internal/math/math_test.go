package math

import "testing"

type mathTestSet struct {
	op  string
	val []bool
	res bool
}

var MathTestSets = []mathTestSet{
	{"AND", []bool{true, true, true}, true},
	{"AND", []bool{true}, true},
	{"AND", []bool{}, true},
	{"AND", []bool{true, false, true}, false},
	{"AND", []bool{false}, false},
	{"OR", []bool{true, true, true}, true},
	{"OR", []bool{true}, true},
	{"OR", []bool{}, true},
	{"OR", []bool{true, false, true}, true},
	{"OR", []bool{true, false, false}, true},
	{"OR", []bool{false, false}, false},
	{"OR", []bool{false}, false},
	{"MIN 3", []bool{true, true, true}, true},
	{"MIN 2", []bool{true}, false},
	{"MIN 0", []bool{}, true},
	{"MIN 3", []bool{true, false, true}, false},
	{"MINPERCENT 100", []bool{true, true, true}, true},
	{"MINPERCENT 33.34", []bool{true, false, false}, false},
	{"MINPERCENT 33.33", []bool{true, false, false}, true},
	{"MINPERCENT 0", []bool{false, false}, true},
	{"MINPERCENT 0", []bool{}, true},
	{"MINPERCENT 100", []bool{}, true},
}

func TestOperations(t *testing.T) {
	for _, test := range MathTestSets {
		res, _ := Calculate(test.op, test.val)
		if res != test.res {
			t.Errorf("Expected operation '%s' with data %v to be %v, is %v", test.op, test.val, test.res, res)
		}
	}
}

func TestUnknownOperation(t *testing.T) {
	op := "UNKNOWN"
	_, err := Calculate(op, []bool{true, true})
	if err == nil {
		t.Errorf("Unknown operation '%s' did not fail", op)
	}
}

func TestEmptyOperation(t *testing.T) {
	_, err := Calculate("", []bool{true, true})
	if err == nil {
		t.Errorf("Empty operation did not return error")
	}
}

func TestMalformedOperation(t *testing.T) {
	_, err := Calculate("MIN 4,3", []bool{true, true})
	if err == nil {
		t.Errorf("Malformed operation did not return error")
	}
}
