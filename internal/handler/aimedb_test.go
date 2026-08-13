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

func TestAimeRegisterDuplicateMatchesAquaDX(t *testing.T) {
	setupMaimaiTestDB(t)
	request := make([]byte, 0x30)
	copy(request[0x20:0x2a], []byte{0x50, 0x10, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01})

	first, _, err := aimeRegister(request)
	if err != nil {
		t.Fatalf("first registration: %v", err)
	}
	if status := binary.LittleEndian.Uint16(first[0x08:0x0a]); status != 1 {
		t.Fatalf("first registration status=%d, want 1", status)
	}
	if got := binary.LittleEndian.Uint64(first[0x20:0x28]); got == 0 {
		t.Fatal("first registration returned zero Aime ID")
	}

	duplicate, _, err := aimeRegister(request)
	if err != nil {
		t.Fatalf("duplicate registration: %v", err)
	}
	if status := binary.LittleEndian.Uint16(duplicate[0x08:0x0a]); status != 0 {
		t.Fatalf("duplicate registration status=%d, want 0", status)
	}
	if got := binary.LittleEndian.Uint64(duplicate[0x20:0x28]); got != 0 {
		t.Fatalf("duplicate registration Aime ID=%d, want 0", got)
	}
}
