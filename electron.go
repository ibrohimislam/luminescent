package luminescent

import (
	"io"
	"log"
	"net"
	"strconv"
	"time"
)

/**
 * Electron
 *
 * Kelas yang bertanggung jawab untuk meneruskan semua command yang diberikan
 * oleh client ke channel-channel tertentu.
 */

type Electron struct {
	Interactor    SocketInteractor
	Conn          net.Conn
	Id            int
	ServerChannel chan<- *Photon
	State         int
}

func (this *Electron) handleClient() {

	defer this.Conn.Close()

	var buf [512]byte
	for this.State >= 0 {

		this.Conn.SetReadDeadline(time.Now().Add(1e9))
		n, err := this.Conn.Read(buf[0:])

		if nil != err {

			if opErr, ok := err.(*net.OpError); ok && opErr.Timeout() {
				continue
			} else if io.EOF == err {
				this.end()
				return
			}

			log.Println(err.Error())
		}

		message := this.Interactor.Decapsulate(buf[0:n])
		photon := *CreatePhoton(message, this)

		this.ServerChannel <- &photon
	}

	log.Println("User #" + strconv.Itoa(this.Id) + " leave.")
}

func (this *Electron) ChangeState(state int) {
	this.State = state
}

func (this *Electron) Emit(message string) {

	buf := this.Interactor.Encapsulate(message)
	this.Conn.Write(buf)

}

func (this *Electron) end() {

	this.Conn.Close()
	log.Println("User #" + strconv.Itoa(this.Id) + " disconnected.")

}
