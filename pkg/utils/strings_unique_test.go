package utils_test

import (
	"testing"

	"github.com/lajosbencz/netdata-dash/pkg/utils"
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
	if s.Remove("bar") != 1 {
		t.Fail()
	}
	if s.Remove("bar") != 0 {
		t.Fail()
	}
	if len(s) != 1 {
		t.Fail()
	}
}
