package handler

import (
	"crypto/aes"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/FireGuo1145/MaiGoDX/internal/database"
	"github.com/FireGuo1145/MaiGoDX/internal/model"
)

const (
	// AimeDBAddress is AquaDX's default AimeDB TCP endpoint. SDGA's Aime daemon
	// connects to this port on the host returned by ALL.Net PowerOn.
	AimeDBDefaultPort = "22345"
	aimeDBKey         = "Copyright(C)SEGA"
)

// StartAimeDB starts the Aime card daemon and keeps accepting one-request TCP
// connections. A failure to bind is logged explicitly so an occupied firewall
// or port is immediately visible in the server console.
func StartAimeDB() {
	port := strings.TrimSpace(os.Getenv("AIME_PORT"))
	if port == "" {
		port = AimeDBDefaultPort
	}
	address := net.JoinHostPort("0.0.0.0", port)
	listener, err := net.Listen("tcp", address)
	if err != nil {
		log.Printf("[MaiGoDX] AimeDB failed to listen on %s: %v", address, err)
		return
	}
	log.Printf("[MaiGoDX] AimeDB listening on %s (AIME_PORT=%s)", address, port)
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("[MaiGoDX] AimeDB accept error: %v", err)
			continue
		}
		go serveAimeDBConnection(conn)
	}
}

func serveAimeDBConnection(conn net.Conn) {
	defer conn.Close()
	remote := conn.RemoteAddr().String()
	if err := conn.SetDeadline(time.Now().Add(5 * time.Second)); err != nil {
		log.Printf("[MaiGoDX] AimeDB deadline error for %s: %v", remote, err)
		return
	}
	request, encrypted, err := readAimeDBRequest(conn)
	if err != nil {
		log.Printf("[MaiGoDX] AimeDB rejected request from %s: %v", remote, err)
		return
	}
	if len(request) < 0x20 {
		log.Printf("[MaiGoDX] AimeDB rejected short plaintext request from %s", remote)
		return
	}

	requestType := binary.LittleEndian.Uint16(request[0x04:0x06])
	gameID := trimAimeASCII(request[0x0a:0x10])
	keychip := trimAimeASCII(request[0x14:0x20])
	if requestType != 0x13 && !aimeDBKeychipExists(keychip) {
		log.Printf("[MaiGoDX] AimeDB rejected: unknown Keychip=%s type=0x%02x game=%s remote=%s", keychip, requestType, gameID, remote)
		return
	}

	response, summary, err := handleAimeDBRequest(requestType, request)
	if err != nil {
		log.Printf("[MaiGoDX] AimeDB request error: type=0x%02x game=%s keychip=%s remote=%s error=%v", requestType, gameID, keychip, remote, err)
		return
	}
	if response == nil {
		log.Printf("[MaiGoDX] AimeDB request complete without response: type=0x%02x game=%s keychip=%s remote=%s", requestType, gameID, keychip, remote)
		return
	}
	if err := writeAimeDBResponse(conn, response, encrypted); err != nil {
		log.Printf("[MaiGoDX] AimeDB response error: type=0x%02x game=%s keychip=%s remote=%s error=%v", requestType, gameID, keychip, remote, err)
		return
	}
	log.Printf("[MaiGoDX] AimeDB %s: type=0x%02x game=%s keychip=%s remote=%s", summary, requestType, gameID, keychip, remote)
}

func readAimeDBRequest(reader io.Reader) ([]byte, bool, error) {
	first := make([]byte, aes.BlockSize)
	if _, err := io.ReadFull(reader, first); err != nil {
		return nil, false, fmt.Errorf("read initial block: %w", err)
	}

	// AquaDX expects encrypted AiMeDB frames, while KanadeDX's daemon uses a
	// PlainRequest path during bootstrap. Detect the standard little-endian
	// magic before decrypting so both client variants share the same listener.
	encrypted := true
	plainHeader := append([]byte{}, first...)
	if binary.LittleEndian.Uint16(plainHeader[0:2]) != 0xa13e {
		var err error
		plainHeader, err = cryptAimeDB(first, false)
		if err != nil {
			return nil, false, err
		}
		if binary.LittleEndian.Uint16(plainHeader[0:2]) != 0xa13e {
			return nil, false, fmt.Errorf("invalid AimeDB magic raw=%x decrypted=%x", first[0:2], plainHeader[0:2])
		}
	} else {
		encrypted = false
	}

	length := int(binary.LittleEndian.Uint16(plainHeader[6:8]))
	if length < 0x20 || length > 4096 || length%aes.BlockSize != 0 {
		return nil, false, fmt.Errorf("invalid AimeDB packet length %d", length)
	}
	packet := append([]byte{}, first...)
	if length > len(packet) {
		rest := make([]byte, length-len(packet))
		if _, err := io.ReadFull(reader, rest); err != nil {
			return nil, false, fmt.Errorf("read request body: %w", err)
		}
		packet = append(packet, rest...)
	}
	if encrypted {
		plain, err := cryptAimeDB(packet, false)
		return plain, true, err
	}
	return packet, false, nil
}

func writeAimeDBResponse(writer io.Writer, response []byte, encrypted bool) error {
	if len(response) < 0x20 || len(response)%aes.BlockSize != 0 {
		return fmt.Errorf("invalid AimeDB response size %d", len(response))
	}
	binary.LittleEndian.PutUint16(response[0x00:0x02], 0xa13e)
	binary.LittleEndian.PutUint16(response[0x02:0x04], 0x3087)
	binary.LittleEndian.PutUint16(response[0x06:0x08], uint16(len(response)))
	payload := response
	if encrypted {
		var err error
		payload, err = cryptAimeDB(response, true)
		if err != nil {
			return err
		}
	}
	_, err := writer.Write(payload)
	return err
}

func cryptAimeDB(source []byte, encrypt bool) ([]byte, error) {
	if len(source) == 0 || len(source)%aes.BlockSize != 0 {
		return nil, fmt.Errorf("AimeDB payload must be a non-empty multiple of %d bytes", aes.BlockSize)
	}
	block, err := aes.NewCipher([]byte(aimeDBKey))
	if err != nil {
		return nil, err
	}
	result := make([]byte, len(source))
	if encrypt {
		block.Encrypt(result[:aes.BlockSize], source[:aes.BlockSize])
		for offset := aes.BlockSize; offset < len(source); offset += aes.BlockSize {
			block.Encrypt(result[offset:offset+aes.BlockSize], source[offset:offset+aes.BlockSize])
		}
	} else {
		block.Decrypt(result[:aes.BlockSize], source[:aes.BlockSize])
		for offset := aes.BlockSize; offset < len(source); offset += aes.BlockSize {
			block.Decrypt(result[offset:offset+aes.BlockSize], source[offset:offset+aes.BlockSize])
		}
	}
	return result, nil
}

func handleAimeDBRequest(requestType uint16, request []byte) ([]byte, string, error) {
	switch requestType {
	case 0x01:
		return aimeFelicaLookupV1(request)
	case 0x04:
		return aimeLookup(request, false)
	case 0x05:
		return aimeRegister(request)
	case 0x09:
		return aimeStaticResponse(0x20, 0x0a, 1), "log", nil
	case 0x0b:
		return aimeStaticResponse(0x200, 0x0c, 1), "campaign", nil
	case 0x0d:
		response := aimeStaticResponse(0x50, 0x0e, 1)
		binary.LittleEndian.PutUint16(response[0x20:0x22], 0x6f)
		binary.LittleEndian.PutUint16(response[0x24:0x26], 0x01)
		return response, "touch", nil
	case 0x0f:
		return aimeLookup(request, true)
	case 0x11:
		return aimeFelicaLookupV2(request)
	case 0x13:
		return aimeStaticResponse(0x40, 0x14, 1), "unknown-19", nil
	case 0x64:
		return aimeStaticResponse(0x20, 0x65, 1), "hello", nil
	case 0x66:
		return nil, "goodbye", nil
	default:
		return nil, "", fmt.Errorf("unsupported request type 0x%02x", requestType)
	}
}

func aimeLookup(request []byte, v2 bool) ([]byte, string, error) {
	accessCode, err := aimeAccessCodeAt(request, 0x20)
	if err != nil {
		return nil, "", err
	}
	aimeID, found, err := aimeCardID(accessCode, false)
	if err != nil {
		return nil, "", err
	}
	if !found {
		aimeID = -1
	}
	responseCode := uint16(0x06)
	if v2 {
		responseCode = 0x10
	}
	response := aimeStaticResponse(0x130, responseCode, 1)
	binary.LittleEndian.PutUint64(response[0x20:0x28], uint64(aimeID))
	response[0x24] = 0
	return response, fmt.Sprintf("lookup access=%s found=%t aimeId=%d", maskAimeAccessCode(accessCode), found, aimeID), nil
}

func aimeRegister(request []byte) ([]byte, string, error) {
	accessCode, err := aimeAccessCodeAt(request, 0x20)
	if err != nil {
		return nil, "", err
	}
	aimeID, created, err := aimeCardID(accessCode, true)
	if err != nil {
		return nil, "", err
	}
	status := uint16(0)
	if created {
		status = 1
	}
	response := aimeStaticResponse(0x30, 0x06, status)
	binary.LittleEndian.PutUint64(response[0x20:0x28], uint64(aimeID))
	return response, fmt.Sprintf("register access=%s created=%t aimeId=%d", maskAimeAccessCode(accessCode), created, aimeID), nil
}

func aimeFelicaLookupV1(request []byte) ([]byte, string, error) {
	accessCode, err := aimeFelicaAccessCode(request, 0x20)
	if err != nil {
		return nil, "", err
	}
	accessCodeBytes, err := hex.DecodeString(accessCode)
	if err != nil {
		return nil, "", fmt.Errorf("decode Felica v1 access code: %w", err)
	}
	response := aimeStaticResponse(0x30, 0x03, 1)
	copy(response[0x24:0x2e], accessCodeBytes)
	return response, "felica-v1 access=" + maskAimeAccessCode(accessCode), nil
}

func aimeFelicaLookupV2(request []byte) ([]byte, string, error) {
	// AquaDX reads the IDm at 0x30 as a signed big-endian 64-bit value,
	// converts its decimal representation to a 20-digit access code, then
	// resolves that card to the game's external user ID.
	accessCode, err := aimeFelicaAccessCode(request, 0x30)
	if err != nil {
		return nil, "", err
	}
	aimeID, found, err := aimeCardID(accessCode, false)
	if err != nil {
		return nil, "", err
	}
	if !found {
		aimeID = -1
	}
	accessCodeBytes, err := hex.DecodeString(accessCode)
	if err != nil {
		return nil, "", fmt.Errorf("decode Felica v2 access code: %w", err)
	}
	response := aimeStaticResponse(0x140, 0x12, 1)
	binary.LittleEndian.PutUint64(response[0x20:0x28], uint64(aimeID))
	binary.LittleEndian.PutUint32(response[0x24:0x28], ^uint32(0))
	binary.LittleEndian.PutUint32(response[0x28:0x2c], ^uint32(0))
	copy(response[0x2c:0x36], accessCodeBytes)
	response[0x37] = 0x01
	return response, fmt.Sprintf("felica-v2 access=%s found=%t aimeId=%d", maskAimeAccessCode(accessCode), found, aimeID), nil
}

func aimeFelicaAccessCode(request []byte, offset int) (string, error) {
	if len(request) < offset+8 {
		return "", errors.New("short Felica request")
	}
	idm := int64(binary.BigEndian.Uint64(request[offset : offset+8]))
	value := strings.TrimPrefix(strconv.FormatInt(idm, 10), "-")
	return strings.Repeat("0", max(0, 20-len(value))) + value, nil
}

func aimeStaticResponse(size int, responseCode, status uint16) []byte {
	response := make([]byte, size)
	binary.LittleEndian.PutUint16(response[0x04:0x06], responseCode)
	binary.LittleEndian.PutUint16(response[0x08:0x0a], status)
	return response
}

func aimeCardID(accessCode string, create bool) (int64, bool, error) {
	var card model.UserCard
	lookup := database.DB.Where("access_code = ?", accessCode).Limit(1).Find(&card)
	if lookup.Error != nil {
		return 0, false, lookup.Error
	}
	if lookup.RowsAffected == 0 {
		if !create {
			return 0, false, nil
		}
		card = model.UserCard{AccessCode: accessCode, CardName: "AimeDB 自动注册"}
		if err := database.DB.Create(&card).Error; err != nil {
			return 0, false, err
		}
		card.GameUserID = int64(card.ID)
		if err := database.DB.Save(&card).Error; err != nil {
			return 0, false, err
		}
		return card.GameUserID, true, nil
	}
	if card.GameUserID <= 0 {
		card.GameUserID = int64(card.ID)
		if err := database.DB.Save(&card).Error; err != nil {
			return 0, false, err
		}
	}
	return card.GameUserID, true, nil
}

func aimeDBKeychipExists(value string) bool {
	terminal, err := findTerminalByKeychip(value)
	return err == nil && terminal.ID != 0 && terminal.IsEnabled
}

func aimeAccessCodeAt(request []byte, offset int) (string, error) {
	if len(request) < offset+10 {
		return "", errors.New("short Aime access-code request")
	}
	return strings.ToUpper(hex.EncodeToString(request[offset : offset+10])), nil
}

func trimAimeASCII(value []byte) string {
	return strings.TrimRight(strings.TrimSpace(string(value)), "\x00")
}

func maskAimeAccessCode(value string) string {
	if len(value) <= 4 {
		return value
	}
	return value[:4] + strings.Repeat("*", len(value)-4)
}
