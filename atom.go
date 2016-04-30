// http://tools.ietf.org/html/rfc6455

package luminescent

import (
	"log"
	"strconv"

	"net"
	"os"
)

type SocketInteractor interface {
	Handshake(conn net.Conn, data string)
	Decapsulate(buf []byte) string
	Encapsulate(message string) []byte
}

type Atom struct {
	Interactor SocketInteractor
	Proton     Proton
}

func (this *Atom) Run() {
	var serverChannel chan *Photon = make(chan *Photon)
	this.Proton = *CreateProton(serverChannel)
	go this.Proton.serverRoutine()

	logger := log.New(os.Stderr, "", 0)

	service := ":8080"
	tcpAddr, err := net.ResolveTCPAddr("tcp4", service)
	checkError(err)
	listener, err := net.ListenTCP("tcp", tcpAddr)
	checkError(err)

	logger.Print("Listen on " + tcpAddr.String())

	clientCount := 0

	for {
		conn, err := listener.Accept()

		if err != nil {
			continue
		}

		clientCount += 1
		logger.Print("User #" + strconv.Itoa(clientCount) + " connected.")
		electron := &Electron{this.Interactor, conn, clientCount, serverChannel, nil, 0}
		go electron.handleClient()
	}
}

func checkError(err error) {
	if err != nil {
		log.New(os.Stderr, "", 0).Panic(os.Stderr, "fatal error: %s", err.Error())
		os.Exit(1)
	}
}
