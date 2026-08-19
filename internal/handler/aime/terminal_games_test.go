package aime

import "testing"

func TestNormalizeTerminalGameID(t *testing.T) {
	for _, test := range []struct {
		input string
		want  string
	}{
		{input: "SDHD", want: "SDHD"},
		{input: "sdez", want: "SDEZ"},
		{input: " SDGA ", want: "SDGA"},
		{input: "SDED", want: "SDED"},
		{input: "SDDT", want: "SDDT"},
		{input: "SBZV", want: "SBZV"},
		{input: "SDFE", want: "SDFE"},
	} {
		got, supported := normalizeTerminalGameID(test.input)
		if !supported || got != test.want {
			t.Fatalf("normalizeTerminalGameID(%q) = (%q, %t), want (%q, true)", test.input, got, supported, test.want)
		}
	}
}

func TestNormalizeTerminalGameIDRejectsUnsupportedGames(t *testing.T) {
	for _, gameID := range []string{"", "SDEY", "MAIMAI"} {
		if _, supported := normalizeTerminalGameID(gameID); supported {
			t.Errorf("normalizeTerminalGameID(%q) unexpectedly accepted", gameID)
		}
	}
}
