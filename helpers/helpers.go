package helpers

import (
	"fmt"
	"strings"
	"time"
)

func Contains[T comparable](s []T, e T) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func Insert[T any](a []T, index int, value T) []T {
	if len(a) == index { // nil or empty slice or after last element
		return append(a, value)
	}
	a = append(a[:index+1], a[index:]...) // index < len(a)
	a[index] = value
	return a
}

func Remove[T any](s []T, i int) []T {
	s[i] = s[len(s)-1]
	return s[:len(s)-1]
}

func FmtSanitize[T any](toSanitize T) string {
	escaped := strings.Replace(fmt.Sprint(toSanitize), "\n", "", -1)
	escaped = strings.Replace(escaped, "\r", "", -1)
	return escaped
}

func GetCurrentSchoolYear() string {
	d := time.Now()
	year := d.Year()
	if d.Month() < time.August || (d.Month() == time.August && d.Day() < 23) {
		year--
	}
	year2 := year + 1
	return fmt.Sprintf("%d/%d", year, year2)
}
