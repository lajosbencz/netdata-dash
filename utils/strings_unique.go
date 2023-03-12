package utils

import "strings"

type StringsUnique []string

func (r *StringsUnique) String() string {
	return strings.Join(*r, ",")
}

func (r *StringsUnique) Set(value string) error {
	*r = strings.Split(value, ",")
	return nil
}

func (r *StringsUnique) Has(str string) bool {
	for _, v := range *r {
		if v == str {
			return true
		}
	}
	return false
}

func (r *StringsUnique) Add(list ...string) int {
	n := 0
	for _, str := range list {
		if r.Has(str) {
			continue
		}
		*r = append(*r, str)
		n++
	}
	return n
}

func (r *StringsUnique) Remove(list ...string) int {
	n := 0
	for _, str := range list {
		if !r.Has(str) {
			continue
		}
		for k, v := range *r {
			if v == str {
				(*r) = append((*r)[:k], (*r)[k+1:]...)
				break
			}
		}
		n++
	}
	return n
}
