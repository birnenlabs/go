package icy

import (
	"testing"
)

type streamTitleTest struct {
	in   []byte
	want string
}

func TestFindTitle(t *testing.T) {
	for _, s := range []streamTitleTest{
		{
			in:   []byte("StreamTitle='';"),
			want: "",
		},
		{
			in:   []byte("StreamTitle='a';"),
			want: "a",
		},
		{
			in:   []byte("StreamTitle=StreamTitle='abcdef';"),
			want: "abcdef",
		},
	} {
		got := findStreamTitle(s.in)
		if *got != s.want {
			t.Errorf("got: %q, want: %q", *got, s.want)
		}
	}
}
