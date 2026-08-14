package handler

import (
	"encoding/binary"
	"testing"

	"github.com/FireGuo1145/MaiGoDX/internal/database"
	"github.com/FireGuo1145/MaiGoDX/internal/model"
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

func TestAimeFelicaLookupV2FindsCardUsingAquaDXByteOrder(t *testing.T) {
	setupMaimaiTestDB(t)
	const accessCode = "00000000000000004660" // 0x1234, as AquaDX formats it
	if err := database.DB.Create(&model.UserCard{AccessCode: accessCode, GameUserID: 7654}).Error; err != nil {
		t.Fatalf("create card: %v", err)
	}

	request := make([]byte, 0x40)
	binary.BigEndian.PutUint64(request[0x30:0x38], 0x1234)
	response, _, err := aimeFelicaLookupV2(request)
	if err != nil {
		t.Fatalf("Felica lookup: %v", err)
	}
	// AquaDX overwrites bytes 0x24..0x27 with 0xffffffff, so its client reads
	// the 32-bit external ID from offset 0x20.
	if got := binary.LittleEndian.Uint32(response[0x20:0x24]); got != 7654 {
		t.Fatalf("Felica lookup returned user ID %d, want 7654", got)
	}
}

func TestAimeStaticResponsesMatchAquaDXCapacities(t *testing.T) {
	for _, test := range []struct {
		requestType uint16
		wantSize    int
		wantCode    uint16
	}{
		{requestType: 0x09, wantSize: 0x20, wantCode: 0x0a},  // Log
		{requestType: 0x0b, wantSize: 0x200, wantCode: 0x0c}, // Campaign
		{requestType: 0x64, wantSize: 0x20, wantCode: 0x65},  // Hello
	} {
		response, _, err := handleAimeDBRequest(test.requestType, make([]byte, 0x20))
		if err != nil {
			t.Fatalf("type 0x%02x: %v", test.requestType, err)
		}
		if len(response) != test.wantSize {
			t.Fatalf("type 0x%02x size=%#x, want %#x", test.requestType, len(response), test.wantSize)
		}
		if got := binary.LittleEndian.Uint16(response[0x04:0x06]); got != test.wantCode {
			t.Fatalf("type 0x%02x response code=%#x, want %#x", test.requestType, got, test.wantCode)
		}
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
