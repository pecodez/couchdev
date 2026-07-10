package tmux

import "testing"

func TestLoginWrap(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		{"claude", "bash --login -c 'claude'"},
		{`claude --rc "proj/s1"`, `bash --login -c 'claude --rc "proj/s1"'`},
		{"cmd with 'single quotes'", `bash --login -c 'cmd with '\''single quotes'\'''`},
	}
	for _, tc := range cases {
		got := loginWrap(tc.in)
		if got != tc.want {
			t.Errorf("loginWrap(%q)\n got  %q\n want %q", tc.in, got, tc.want)
		}
	}
}
