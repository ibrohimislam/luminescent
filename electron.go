package luminescent

import (
	"net"

	"regexp"
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
	RoomChannel   chan<- *Photon
	State         int
}

func (this *Electron) handleClient() {

	defer this.Conn.Close()

	var buf [512]byte
	for {

		n, err := this.Conn.Read(buf[0:])
		if err != nil {
			return
		}

		data := string(buf[0:n])
		first, _ := regexp.MatchString("^GET", data)

		if first {

			this.Interactor.Handshake(this.Conn, data)

		} else {

			if buf[0] == 129 {

				message := this.Interactor.Decapsulate(buf[0:n])
				photon := *CreatePhoton(message, this)

				if this.RoomChannel != nil {
					this.RoomChannel <- &photon
				} else {
					this.ServerChannel <- &photon
				}

			} else if buf[0] == 136 {

				return

			}

		}
	}

}

func (this *Electron) ChangeState(state int) {
	this.State = state
}

func (this *Electron) Emit(message string) {

	buf := this.Interactor.Encapsulate(message)
	this.Conn.Write(buf)

}
