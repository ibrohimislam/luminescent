package amaryllis

import (
	"net"

	"crypto/sha1"
	"encoding/base64"
	"regexp"
	"strings"
)

type Coulomb struct{}

func (_ Coulomb) Handshake(conn net.Conn, data string) {

	websocketKey := regexp.MustCompile("Sec-WebSocket-Key: (.*)").FindStringSubmatch(data)

	websocketAcceptKeyUInt8 := sha1.Sum([]byte(strings.TrimSpace(websocketKey[1]) + "258EAFA5-E914-47DA-95CA-C5AB0DC85B11"))
	websocketAcceptKey := base64.StdEncoding.EncodeToString(websocketAcceptKeyUInt8[0:])

	response := []byte("HTTP/1.1 101 Switching Protocols\n" +
		"Connection: Upgrade\n" +
		"Upgrade: websocket\n" +
		"Sec-WebSocket-Accept: " + websocketAcceptKey + "\n\n")

	go conn.Write(response[0:])
}

func (_ Coulomb) Decapsulate(buf []byte) string {
	var length int
	var offset int

	if buf[1] < 254 {
		length = int(buf[1] & 127)
		offset = 2
	} else {
		length = int((buf[2] << 8) + buf[3])
		offset = 4
	}

	var key [4]byte

	for i := 0; i < 4; i++ {
		key[i] = buf[offset+i]
	}

	var decoded []byte

	payload_data_offset := offset + 4
	for i := 0; i < length; i++ {
		decoded = append(decoded, byte(buf[payload_data_offset+i]^key[i%4]))
	}

	return string(decoded)
}

func (_ Coulomb) Encapsulate(message string) []byte {

	var response []byte

	response = append(response, byte(129))

	if len(message) < 126 {

		response = append(response, byte(len(message)))

	} else {

		response = append(response, byte(126))
		response = append(response, byte(len(message)>>8))
		response = append(response, byte(len(message)&255))

	}

	byte_string := []byte(message)

	response = append(response, byte_string...)

	return response
}
