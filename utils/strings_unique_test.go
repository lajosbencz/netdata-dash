package utils_test

import (
	"testing"

	"github.com/lajosbencz/netdata-dash/utils"
)

func TestStringsUnique(t *testing.T) {
	s := utils.StringsUnique{"foo"}
	if !s.Has("foo") {
		t.Fail()
	}
	if s.Has("bar") {
		t.Fail()
	}
	if s.Add("foo", "bar") != 1 {
		t.Fail()
	}
	if !s.Has("bar") {
		t.Fail()
	}
	if !s.Remove("bar") {
		t.Fail()
	}
	if s.Remove("bar") {
		t.Fail()
	}
	if len(s) != 1 {
		t.Fail()
	}
}
