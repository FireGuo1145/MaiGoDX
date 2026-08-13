package handler

import (
	"encoding/binary"
	"testing"
)

func TestAimeFelicaAccessCodeMatchesAquaDXConversion(t *testing.T) {
	tests := []struct {
		name string
		idm  uint64
		want string
	}{
		{name: "positive IDm", idm: 0x1234, want: "00000000000000004660"},
		// AquaDX reads this as a signed Kotlin Long, removes the minus sign,
		// and pads the remaining decimal representation to 20 characters.
		{name: "negative signed IDm", idm: 0xfffffffffffffffe, want: "00000000000000000002"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := make([]byte, 0x38)
			binary.BigEndian.PutUint64(request[0x30:0x38], test.idm)
			got, err := aimeFelicaAccessCode(request, 0x30)
			if err != nil {
				t.Fatalf("aimeFelicaAccessCode returned error: %v", err)
			}
			if got != test.want {
				t.Fatalf("aimeFelicaAccessCode=%q, want %q", got, test.want)
			}
		})
	}
}
