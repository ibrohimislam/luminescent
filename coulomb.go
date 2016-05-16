package luminescent

import (
	"net"
)

type Coulomb struct{}

func (_ Coulomb) Handshake(_ net.Conn, _ string) {}

func (_ Coulomb) Decapsulate(buf []byte) string {
	return string(buf)
}

func (_ Coulomb) Encapsulate(message string) []byte {
	return []byte(message)
}
