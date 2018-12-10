package main

import (
	"testing"
)

func TestCalcOk1(t *testing.T) {
	tables := []struct {
		str    string
		result float64
	}{
		{"1 =", 1},
		{"1 1 + =", 2},
		{"1 2 + 3 4 + * =", 21},
		{"2 3 4 5 6 * + - / =", -0.06451612903225806},
		{"2 3 * 4 5 * + =", 26},
	}

	for _, table := range tables {
		ans, err := Calc(table.str)
		if ans != table.result {
			t.Errorf("expected: %v, result: %v", table.result, ans)
		}
		if err != nil {
			t.Error(err)
		}
	}
}

func TestCalcError(t *testing.T) {
	tables := []struct {
		str string
		err string
	}{
		{"1 1 +", "Expression must end with ="},
		{"1", "Too short expression"},
		{"2 3 4 5 6 * + - / = = =", "Only numbers and ariphmetical operators are valid"},
		{"2 3 * /4 5 * + =", "Only numbers and ariphmetical operators are valid"},
		{"1 1 + * =", "Invalid expression"},
		{"1 + 1 1 + =", "Invalid expression"},
		{"+ = ", "Invalid expression"},
	}

	for _, table := range tables {
		_, err := Calc(table.str)
		if err.Error() != table.err {
			t.Errorf("expected: %v, result: %v", table.err, err.Error())
		}
	}
}
