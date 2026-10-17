package base

import (
	"fmt"
	"testing"
)

type legacyText struct{}

func (legacyText) ToString() string { return "legacy" }

type standardText struct{ legacyText }

func (standardText) String() string { return "standard" }

type errorText struct{ standardText }

func (errorText) Error() string { return "error" }

type formattedText struct{ errorText }

func (formattedText) Format(state fmt.State, verb rune) {
	_, _ = state.Write([]byte("formatted"))
}

type nilText struct{}

func (*nilText) ToString() string { panic("nil method must not run") }

type outerText struct{ Base }

func (*outerText) ToString() string { return "outer legacy" }

type panicText struct{}

func (panicText) ToString() string { panic("legacy panic") }

func TestToStringProtocols(t *testing.T) {
	cases := []struct {
		value any
		want  string
	}{
		{legacyText{}, "legacy"},
		{standardText{}, "standard"},
		{errorText{}, "error"},
		{formattedText{}, "formatted"},
		{nil, "<nil>"},
		{(*nilText)(nil), "<nil>"},
		{(*Base)(nil), "<nil>"},
		{[]int(nil), "<nil>"},
		{map[string]int(nil), "<nil>"},
		{[]int{}, "[]"},
		{"abc", "abc"},
		{true, "true"},
		{int64(-42), "-42"},
		{uint64(42), "42"},
		{float32(1.123456), "1.123456"},
		{struct{ Name string }{"value"}, "{value}"},
	}
	for _, tc := range cases {
		if got := ToString(tc.value); got != tc.want {
			t.Errorf("ToString(%T) = %q, want %q", tc.value, got, tc.want)
		}
	}
	object := &outerText{}
	if ToString(object) != object.String() {
		t.Fatal("promoted String must retain precedence over ToString")
	}
	objectBase := New()
	if ToString(objectBase) != objectBase.ToString() {
		t.Fatal("Base formatting changed")
	}
}

func TestToStringLegacyPanic(t *testing.T) {
	defer func() {
		if recover() != "legacy panic" {
			t.Fatal("legacy panic must propagate")
		}
	}()
	ToString(panicText{})
}
