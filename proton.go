package amaryllis

import (
	"log"
	"os"

	"strconv"
	"strings"

	"encoding/json"
)

var ProtonInstance *Proton

type User struct {
	UserId int    `json:"userId"`
	Nick   string `json:"nick"`
}

type Room struct {
	RoomId     int    `json:"roomId"`
	RoomMaster int    `json:"roomMaster"`
	Players    []User `json:"players"`
}

type Proton struct {
	inputChannel <-chan *Photon

	userIdNick  map[int]string
	rooms       map[int]*Graviton
	roomChannel map[int]chan<- *Photon
	roomCount   int
}

func CreateProton(inputChannel <-chan *Photon) *Proton {

	ProtonInstance = &Proton{inputChannel, map[int]string{}, map[int]*Graviton{}, map[int]chan<- *Photon{}, 0}
	return ProtonInstance

}

func (this *Proton) serverRoutine() {
	l := log.New(os.Stderr, "", 0)

	for {
		current := <-this.inputChannel

		token := strings.Split(current.Message, " ")

		if len(token) > 0 {

			l.Println("[User #" + strconv.Itoa(current.Electron.Id) + "]" + token[0])

			switch {
			case token[0] == "nick":
				this.implNICK(current, token)
			case token[0] == "join":
				this.implJOIN(current, token)
			case token[0] == "create":
				this.implCREATE(current, token)
			case token[0] == "list":
				this.implLIST(current, token)
			}

		}
	}
}

func (this *Proton) implNICK(current *Photon, token []string) {
	this.userIdNick[current.Electron.Id] = token[1]
	current.Electron.Emit(current.Electron.Id)
	current.Electron.ChangeState(1)
}

func (this *Proton) implJOIN(current *Photon, token []string) {
	roomId, _ := strconv.Atoi(token[1])

	if room, ok := this.rooms[roomId]; ok {
		room.Join(current.Electron)

		current.Electron.RoomChannel, _ = this.roomChannel[roomId]

		current.Electron.Emit("OK")
		current.Electron.ChangeState(1)
	} else {
		current.Electron.Emit("FAIL")
	}
}

func (this *Proton) implCREATE(current *Photon, token []string) {
	this.roomCount += 1
	roomId := this.roomCount
	userId := current.Electron.Id

	newRoomChannel := make(chan *Photon)

	newRoom := CreateGraviton(roomId, userId, newRoomChannel)
	newRoom.Join(current.Electron)
	go newRoom.RoomRoutine()

	current.Electron.RoomChannel = newRoomChannel

	this.rooms[roomId] = newRoom
	this.roomChannel[roomId] = newRoomChannel

	current.Electron.Emit("OK")
	current.Electron.ChangeState(2)
}

func (this *Proton) implLIST(current *Photon, token []string) {
	rooms := []Room{}

	for key, value := range this.rooms {
		var players []User

		for _, user := range value.users {
			players = append(players, User{user.Id, this.userIdNick[user.Id]})
		}

		room := Room{key, value.roomMaster, players}
		rooms = append(rooms, room)
	}

	result, _ := json.Marshal(rooms)
	current.Electron.Emit(string(result))
}

func (this *Proton) implDESTROY(roomId int) {
	delete(this.rooms, roomId)
	delete(this.roomChannel, roomId)
}
