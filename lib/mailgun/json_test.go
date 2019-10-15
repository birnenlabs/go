package mailgun

import (
	"testing"
)

type findEmailTest struct {
	in   string
	want string
}

func TestFindEmail(t *testing.T) {
	for _, s := range []findEmailTest{
		{
			in:   "",
			want: "",
		},
		{
			in:   "asdfsadf",
			want: "",
		},
		{
			in:   "<>",
			want: "",
		},
		{
			in:   "<a>",
			want: "a",
		},
		{
			in:   "a<b>c",
			want: "b",
		},
		{
			in:   "<abc",
			want: "",
		},
		{
			in:   "abc>",
			want: "",
		},
		{
			in:   "Someone <me@me.com>",
			want: "me@me.com",
		},
		{
			in:   "<a> <b> <c>",
			want: "a",
		},
	} {
		got := findEmail(s.in)
		if got != s.want {
			t.Errorf("got: %q, want: %q", got, s.want)
		}
	}
}
