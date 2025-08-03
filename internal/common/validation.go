package common

import (
	"fmt"
	"reflect"
	"strings"
)

// ValidationRule represents a validation rule
type ValidationRule func(value interface{}) error

// Validator provides common validation functionality
type Validator struct {
	rules map[string][]ValidationRule
}

// NewValidator creates a new validator
func NewValidator() *Validator {
	return &Validator{
		rules: make(map[string][]ValidationRule),
	}
}

// Required validates that a field is not empty
func Required(fieldName string) ValidationRule {
	return func(value interface{}) error {
		if value == nil {
			return fmt.Errorf("%s cannot be empty", fieldName)
		}

		v := reflect.ValueOf(value)
		switch v.Kind() {
		case reflect.String:
			if v.String() == "" {
				return fmt.Errorf("%s cannot be empty", fieldName)
			}
		case reflect.Slice, reflect.Map:
			if v.Len() == 0 {
				return fmt.Errorf("%s cannot be empty", fieldName)
			}
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			if v.Int() == 0 {
				return fmt.Errorf("%s cannot be zero", fieldName)
			}
		case reflect.Float32, reflect.Float64:
			if v.Float() == 0 {
				return fmt.Errorf("%s cannot be zero", fieldName)
			}
		}

		return nil
	}
}

// Positive validates that a numeric field is positive
func Positive(fieldName string) ValidationRule {
	return func(value interface{}) error {
		v := reflect.ValueOf(value)
		switch v.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			if v.Int() <= 0 {
				return fmt.Errorf("%s must be positive", fieldName)
			}
		case reflect.Float32, reflect.Float64:
			if v.Float() <= 0 {
				return fmt.Errorf("%s must be positive", fieldName)
			}
		default:
			return fmt.Errorf("%s must be a numeric type", fieldName)
		}
		return nil
	}
}

// Range validates that a numeric field is within a specific range
func Range(fieldName string, min, max float64) ValidationRule {
	return func(value interface{}) error {
		v := reflect.ValueOf(value)
		var floatVal float64

		switch v.Kind() {
		case reflect.Float32, reflect.Float64:
			floatVal = v.Float()
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			floatVal = float64(v.Int())
		default:
			return fmt.Errorf("%s must be a numeric type", fieldName)
		}

		if floatVal < min || floatVal > max {
			// Use integer format if both min and max are whole numbers
			if min == float64(int(min)) && max == float64(int(max)) {
				return fmt.Errorf("%s must be between %d and %d", fieldName, int(min), int(max))
			}
			return fmt.Errorf("%s must be between %.1f and %.1f", fieldName, min, max)
		}
		return nil
	}
}

// OneOf validates that a string field is one of the allowed values
func OneOf(fieldName string, allowedValues ...string) ValidationRule {
	return func(value interface{}) error {
		str, ok := value.(string)
		if !ok {
			return fmt.Errorf("%s must be a string", fieldName)
		}

		for _, allowed := range allowedValues {
			if str == allowed {
				return nil
			}
		}

		return fmt.Errorf("invalid %s: %s, must be one of: %s",
			fieldName, str, strings.Join(allowedValues, ", "))
	}
}

// Validate runs all validation rules for the given fields
func (v *Validator) Validate(data map[string]interface{}) error {
	for fieldName, value := range data {
		if rules, exists := v.rules[fieldName]; exists {
			for _, rule := range rules {
				if err := rule(value); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

// AddRule adds a validation rule for a field
func (v *Validator) AddRule(fieldName string, rule ValidationRule) {
	v.rules[fieldName] = append(v.rules[fieldName], rule)
}
