package helpers

import (
	"fmt"
	"strconv"
	"strings"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

func CapitalizeTitle(title string) string {
	titleCaser := cases.Title(language.English)
	words := strings.Fields(title)
	for i, word := range words {
		words[i] = titleCaser.String(strings.ToLower(word))
	}
	return strings.Join(words, " ")
}

func FormatRoute(route string) (string, error) {
	if route == "" {
		return "", nil
	}

	if !strings.HasPrefix(route, "/dashboard") {
		route = "/dashboard" + route
	}

	words := strings.Fields(route)
	for i, word := range words {
		words[i] = strings.ToLower(word)
	}
	formattedRoute := strings.Join(words, "-")

	return formattedRoute, nil
}

func ConvertToInt(value interface{}) (int, error) {
	switch v := value.(type) {
	case int:
		return v, nil
	case int64:
		return int(v), nil
	case float64:
		return int(v), nil
	case string:
		parsed, err := strconv.Atoi(v)
		if err != nil {
			return 0, fmt.Errorf("failed to convert string to int: %w", err)
		}
		return parsed, nil
	default:
		return 0, fmt.Errorf("unsupported type for conversion: %T", v)
	}
}