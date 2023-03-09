package utils

type StringsUnique []string

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

func (r *StringsUnique) Remove(str string) bool {
	if !r.Has(str) {
		return false
	}
	for k, v := range *r {
		if v == str {
			(*r) = append((*r)[:k], (*r)[k+1:]...)
			break
		}
	}
	return true
}
