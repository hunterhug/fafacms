package util

import (
	"errors"
	"strconv"
	"strings"
)

// string to int
func SI(s string) (i int, e error) {
	i, e = strconv.Atoi(s)
	return
}

// int to string
func IS(i int) string {
	return strconv.Itoa(i)
}

func ToLower(s string) string {
	return strings.ToLower(s)
}

// change by python
func DivideStringList(files []string, num int) (map[int][]string, error) {
	length := len(files)
	split := map[int][]string{}
	if num <= 0 {
		return split, errors.New("num must not negtive")
	}
	if num > length {
		num = length
		//return split, errors.New("num must not bigger than the list length")
	}
	process := length / num
	for i := 0; i < num; i++ {
		// slice inside has a refer, so must do this append
		//split[i]=files[i*process : (i+1)*process] wrong!
		split[i] = append(split[i], files[i*process:(i+1)*process]...)
	}
	remain := files[num*process:]
	for i := 0; i < len(remain); i++ {
		split[i%num] = append(split[i%num], remain[i])
	}
	return split, nil
}

func InArray(sa []string, a string) bool {
	for _, v := range sa {
		if a == v {
			return true
		}
	}
	return false
}

func Substr(str string, start, length int) string {
	rs := []rune(str)
	rl := len(rs)
	end := 0

	if start < 0 {
		start = rl - 1 + start
	}
	end = start + length

	if start > end {
		start, end = end, start
	}

	if start < 0 {
		start = 0
	}
	if start > rl {
		start = rl
	}
	if end < 0 {
		end = 0
	}
	if end > rl {
		end = rl
	}
	return string(rs[start:end])
}

func MapToArray(m map[int64]struct{}) []int64 {
	b := make([]int64, 0, len(m))
	for k := range m {
		b = append(b, k)
	}

	return b
}
