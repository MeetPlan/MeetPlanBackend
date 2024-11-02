package helpers

import (
	"fmt"
	"strconv"
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

func CheckIfLeapYear(year int) bool {
	if year%400 == 0 {
		return true
	}
	if year%100 == 0 {
		return false
	}
	return year%4 == 0
}

func GetMaxDayForMonth(month int, year int) int {
	if month < 1 || month > 12 {
		return -1
	}
	if month == 1 || month == 3 || month == 5 || month == 7 || month == 8 || month == 10 || month == 12 {
		return 31
	}
	if month == 2 {
		if CheckIfLeapYear(year) {
			return 29
		} else {
			return 28
		}
	}
	return 30
}

var EMSO_POMNOZITEV = []int{
	7,
	6,
	5,
	4,
	3,
	2,
	7,
	6,
	5,
	4,
	3,
	2,
}

// gender:
// male = 0
// female = 1
// Grozote
func VerifyEMSO(emso string, gender string) bool {
	if gender != "male" && gender != "female" {
		return false
	}

	emso = strings.TrimSpace(emso)

	if len(emso) != 13 {
		return false
	}

	day, err := strconv.Atoi(strings.TrimLeft(emso[0:2], "0"))
	if err != nil {
		return false
	}
	month, err := strconv.Atoi(strings.TrimLeft(emso[2:4], "0"))
	if err != nil {
		return false
	}
	year, err := strconv.Atoi(strings.TrimLeft(emso[4:7], "0"))
	if err != nil {
		return false
	}
	g, err := strconv.Atoi(strings.TrimLeft(emso[9:12], "0"))
	if err != nil {
		return false
	}

	// future proofano za naslednjih 800 let
	if year >= 900 {
		year += 1000
	} else {
		year += 2000
	}
	if year > time.Now().Year() {
		return false
	}

	if month < 1 || month > 12 {
		return false
	}

	if day < 1 || day > GetMaxDayForMonth(month, year) {
		return false
	}

	// ZZZ - zaporedna številka, odvisna od spola
	if g < 0 {
		return false
	}
	if gender == "male" && g >= 500 {
		return false
	}
	if gender == "female" && g < 500 {
		return false
	}

	// Kontrolna številka
	e := 0
	for i, v := range []rune(emso)[0:12] {
		e += int(v-'0') * EMSO_POMNOZITEV[i]
	}
	kontrolnaSt := 11 - (e % 11)

	if kontrolnaSt != int(emso[12]-'0') {
		return false
	}

	return true
}
