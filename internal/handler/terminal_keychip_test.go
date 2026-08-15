package handler

import "testing"

func TestKeychipRegistrationFormat(t *testing.T) {
	valid := []string{"A123-45678901234", "ABCD-12345678901"}
	for _, value := range valid {
		if !isKeychipRegistrationFormat(value) {
			t.Fatalf("expected valid keychip %q", value)
		}
	}
	invalid := []string{"A12345678901234", "A123-4567890123", "B123-45678901234", "A123-456789012!X", "A123-4567-8901234"}
	for _, value := range invalid {
		if isKeychipRegistrationFormat(value) {
			t.Fatalf("expected invalid keychip %q", value)
		}
	}
}

func TestKeychipMatchPrefixIgnoresHyphenAndLastFour(t *testing.T) {
	if got, want := keychipMatchPrefix("A123-45678901234"), "A1234567890"; got != want {
		t.Fatalf("formatted prefix=%q, want %q", got, want)
	}
	if got, want := keychipMatchPrefix("A12345678909999"), "A1234567890"; got != want {
		t.Fatalf("reported prefix=%q, want %q", got, want)
	}
}
