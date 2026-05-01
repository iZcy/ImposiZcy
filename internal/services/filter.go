package services

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/iZcy/imposizcy/internal/models"
)

func EvaluateFilters(conditions []models.FilterCondition, data map[string]interface{}) (bool, string) {
	if len(conditions) == 0 {
		return true, ""
	}

	for _, cond := range conditions {
		val, exists := getFieldValue(data, cond.Field)
		passed, reason := evaluateCondition(cond, val, exists)
		if !passed {
			return false, reason
		}
	}

	return true, ""
}

func getFieldValue(data map[string]interface{}, field string) (interface{}, bool) {
	parts := strings.Split(field, ".")
	var current interface{} = data

	for _, part := range parts {
		m, ok := current.(map[string]interface{})
		if !ok {
			return nil, false
		}
		current, ok = m[part]
		if !ok {
			return nil, false
		}
	}

	return current, true
}

func evaluateCondition(cond models.FilterCondition, val interface{}, exists bool) (bool, string) {
	switch cond.Operator {
	case "exists":
		wantExists := cond.Value != "false"
		if wantExists && !exists {
			return false, fmt.Sprintf("field '%s' does not exist", cond.Field)
		}
		if !wantExists && exists {
			return false, fmt.Sprintf("field '%s' exists but should not", cond.Field)
		}
		return true, ""

	case "eq":
		if !exists {
			return false, fmt.Sprintf("field '%s' does not exist", cond.Field)
		}
		if fmt.Sprintf("%v", val) != cond.Value {
			return false, fmt.Sprintf("field '%s' = '%v', expected '%s'", cond.Field, val, cond.Value)
		}
		return true, ""

	case "neq":
		if !exists {
			return false, fmt.Sprintf("field '%s' does not exist", cond.Field)
		}
		if fmt.Sprintf("%v", val) == cond.Value {
			return false, fmt.Sprintf("field '%s' = '%v', expected != '%s'", cond.Field, val, cond.Value)
		}
		return true, ""

	case "contains":
		if !exists {
			return false, fmt.Sprintf("field '%s' does not exist", cond.Field)
		}
		s := fmt.Sprintf("%v", val)
		if !strings.Contains(s, cond.Value) {
			return false, fmt.Sprintf("field '%s' = '%v', expected to contain '%s'", cond.Field, val, cond.Value)
		}
		return true, ""

	case "gt", "gte", "lt", "lte":
		if !exists {
			return false, fmt.Sprintf("field '%s' does not exist", cond.Field)
		}
		return compareNumeric(cond, val)

	default:
		return true, ""
	}
}

func compareNumeric(cond models.FilterCondition, val interface{}) (bool, string) {
	valStr := fmt.Sprintf("%v", val)

	valNum, err1 := strconv.ParseFloat(valStr, 64)
	condNum, err2 := strconv.ParseFloat(cond.Value, 64)

	if err1 == nil && err2 == nil {
		passed := false
		switch cond.Operator {
		case "gt":
			passed = valNum > condNum
		case "gte":
			passed = valNum >= condNum
		case "lt":
			passed = valNum < condNum
		case "lte":
			passed = valNum <= condNum
		}
		if !passed {
			return false, fmt.Sprintf("field '%s' = %v, expected %s %s", cond.Field, valNum, cond.Operator, condNum)
		}
		return true, ""
	}

	passed := false
	switch cond.Operator {
	case "gt":
		passed = valStr > cond.Value
	case "gte":
		passed = valStr >= cond.Value
	case "lt":
		passed = valStr < cond.Value
	case "lte":
		passed = valStr <= cond.Value
	}
	if !passed {
		return false, fmt.Sprintf("field '%s' = '%v', expected %s '%s'", cond.Field, val, cond.Operator, cond.Value)
	}
	return true, ""
}
