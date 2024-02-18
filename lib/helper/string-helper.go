package helper

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

type ForbiddenWord string

const (
	SELECT   = "select"
	FROM     = "from"
	STAR     = "*"
	DROP     = "drop"
	DELETE   = "delete"
	UNION    = "union"
	TABLE    = "table"
	SCHEMA   = "schema"
	DATABASE = "database"
	INDEX    = "index"
)

func SanitizeAll(texts []string) (err error) {
	for _, str := range texts {
		if err = Sanitize(str); err != nil {
			return
		}
	}
	return
}

func Sanitize(str string) error {
	lowStr := strings.ToLower(str)
	err := fmt.Errorf("suspicious sql injection \"%s\"", str)
	if strings.Contains(lowStr, FROM) && ContainsAnyStr(lowStr, []string{SELECT, STAR, DELETE, UNION}) {
		return err
	}
	if strings.Contains(lowStr, DROP) && ContainsAnyStr(lowStr, []string{TABLE, SCHEMA, DATABASE, INDEX}) {
		return err
	}
	return nil
}

func ValidateLengthStr(str string, minLength int, maxlength int) (err error) {
	if utf8.RuneCountInString(str) > maxlength {
		err = fmt.Errorf("too large text: \"%s\"", str)
		return
	}
	if utf8.RuneCountInString(str) < minLength {
		err = fmt.Errorf("too short text: \"%s\"", str)
		return
	}
	return
}

func ContainsAnyStr(str string, words []string) bool {
	res := false
	for _, w := range words {
		res = res || strings.Contains(str, w)
	}
	return res
}

func ContainsAllStr(str string, words []string) bool {
	res := true
	for _, w := range words {
		res = res && strings.Contains(str, w)
	}
	return res
}
