package main

import (
	"errors"
	"strconv"
	"strings"
)

var validChars = [...]string{"1", "2", "3", "4", "5", "6", "7", "8", "9", "0", "+", "-", "*", "/"}

func Calc(spacedChars string) (float64, error) {
	chars := strings.Fields(spacedChars)
	if len(chars) == 1 {
		return 0, errors.New("Too short expression")
	}
	if chars[len(chars)-1] != "=" {
		return 0, errors.New("Expression must end with =")
	}
	if !validateExpression(chars) {
		return 0, errors.New("Only numbers and ariphmetical operators are valid")
	}
	i := 0
	for chars[i] != "=" {
		for str := chars[i]; str != "+" && str != "-" && str != "*" && str != "/"; {
			i++
			str = chars[i]
			if str == "=" && len(chars) == 2 {
				a, _ := strconv.ParseFloat(chars[0], 64)
				return a, nil
			}
		}
		// i stands for index of ariphmetical operator. i<2 means that we have less than 2 operands
		if i < 2 {
			return 0, errors.New("Invalid expression")
		}
		a, err1 := strconv.ParseFloat(chars[i-2], 64)
		b, err2 := strconv.ParseFloat(chars[i-1], 64)
		//check if two previos chars are numbers
		if err1 != nil || err2 != nil {
			return 0, errors.New("Invalid expression")
		}
		switch string(chars[i]) {
		case "+":
			chars = removeAndPushChar(chars, i, a+b)
		case "-":
			chars = removeAndPushChar(chars, i, a-b)
		case "*":
			chars = removeAndPushChar(chars, i, a*b)
		case "/":
			chars = removeAndPushChar(chars, i, a/b)
		}
		i--
	}
	if len(chars) > 2 {
		return 0, errors.New("Invalid expression")
	}
	a, _ := strconv.ParseFloat(chars[0], 64)
	return a, nil
}

func validateExpression(chars []string) bool {
	for idx := range chars[:len(chars)-1] {
		if _, err := strconv.ParseFloat(chars[idx],
			64); err != nil && chars[idx] != "+" && chars[idx] != "-" && chars[idx] != "*" && chars[idx] != "/" {
			return false
		}
	}
	return true
}

func removeAndPushChar(chars []string, idx int, newNumber float64) []string {
	chars = append(chars[:idx-2], chars[idx:]...)
	chars[idx-2] = strconv.FormatFloat(newNumber, 'f', -1, 64)
	return chars
}
