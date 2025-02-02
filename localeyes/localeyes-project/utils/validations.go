package utils

import (
	"github.com/go-playground/validator"
	"localeyes/config"
	"strings"
	"time"
)

func ValidatePassword(fl validator.FieldLevel) bool {
	password := fl.Field().String()
	if len(password) > 5 {
		if strings.Contains(password, "@") || strings.Contains(password, "#") || strings.Contains(password, "$") || strings.Contains(password, "%") || strings.Contains(password, "^") || strings.Contains(password, "*") {
			if strings.Contains(password, "1") || strings.Contains(password, "2") || strings.Contains(password, "3") || strings.Contains(password, "4") || strings.Contains(password, "5") || strings.Contains(password, "6") || strings.Contains(password, "7") || strings.Contains(password, "8") || strings.Contains(password, "9") || strings.Contains(password, "0") {
				return true
			}
		}
	}
	return false
}

func ValidateTime(fl validator.FieldLevel) bool {
	timeValue, ok := fl.Field().Interface().(time.Time)
	if !ok {
		return false
	}
	return !timeValue.IsZero()
}

func ValidateFilter(fl validator.FieldLevel) bool {
	switch config.Filter(fl.Field().String()) {
	case config.Food, config.Travel, config.Shopping:
		return true
	default:
		return false
	}
}

func IsValidFilter(value string) bool {
	switch config.Filter(value) {
	case config.Food, config.Travel, config.Shopping:
		return true
	default:
		return false
	}
}

func SetTag(value float64) string {
	if value > 1.0 {
		return "resident"
	}
	return "newbie"
}
