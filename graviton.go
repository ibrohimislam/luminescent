package amaryllis

import (
	"math/rand"
	"strconv"
	"strings"

	"encoding/json"
)

type Graviton struct {
	roomId      int
	roomMaster  int
	roomChannel <-chan *Photon
	users       []*Electron
	status      int
	playerCount int
	board       [20][20]int
	turn        int
	play        bool
}

func CreateGraviton(_roomId int, _roomMaster int, _roomChannel <-chan *Photon) *Graviton {
	return &Graviton{_roomId, _roomMaster, _roomChannel, []*Electron{}, 0, 0, [20][20]int{}, 0, false}
}

func (this *Graviton) Join(user *Electron) {
	this.users = append(this.users, user)
	this.playerCount = this.playerCount + 1

	this.broadcast(this.list())
}

func (this *Graviton) RoomRoutine() {

	for {

		for !this.play {
			current := <-this.roomChannel

			token := strings.Split(current.Message, " ")

			if len(token) > 0 {
				switch {
				case token[0] == "play":
					if current.Electron.Id == this.roomMaster {
						this.broadcast("play")
						this.play = true
						current.Electron.Emit("OK")
					} else {
						current.Electron.Emit("FAIL")
					}
				case token[0] == "leave":
					this.implLEAVE(current, token)
				}

			}
		}

		this.initPlay()

		for this.play {
			current := <-this.roomChannel
			token := strings.Split(current.Message, " ")

			if current.Electron.Id == this.users[this.turn%this.playerCount].Id {
				switch {
				case token[0] == "move":
					this.implMOVE(current, token)
				}
			}
		}

	}
}

func (this *Graviton) implLEAVE(current *Photon, token []string) {
	//kick all
	if current.Electron.Id == this.roomMaster {

		// randomize user turn
		for _, v := range this.users {
			v.ChangeState(0)
			v.RoomChannel = nil
			v.Emit("leave")
		}

		ProtonInstance.implDESTROY(this.roomId)

	} else {

		current.Electron.ChangeState(0)
		current.Electron.RoomChannel = nil
		current.Electron.Emit("leave")

		index := -1

		for i := 0; i < len(this.users); i++ {
			if this.users[i].Id == current.Electron.Id {
				index = i
			}
		}

		this.users = append(this.users[:index], this.users[index+1:]...)

		this.broadcast(this.list())

	}
}

func (this *Graviton) list() string {
	var players []User

	for _, user := range this.users {
		players = append(players, User{user.Id, ProtonInstance.userIdNick[user.Id]})
	}

	message, _ := json.Marshal(players)
	return string(message)
}

func (this *Graviton) implMOVE(current *Photon, token []string) {

	x, _ := strconv.Atoi(token[1])
	y, _ := strconv.Atoi(token[2])

	if this.board[x][y] == 0 {

		current.Electron.Emit("OK")
		this.board[x][y] = current.Electron.Id

		state, _ := json.Marshal(this.board)
		this.broadcast(string(state))

		winner := this.checkWinner(x, y, current.Electron.Id)

		if winner {

			this.play = false
			this.broadcast("win " + strconv.Itoa(current.Electron.Id))

		} else {

			this.turn = this.turn + 1
			this.broadcast("turn " + strconv.Itoa(this.users[this.turn%this.playerCount].Id))

		}

	}
}

func (this *Graviton) checkWinner(x int, y int, player int) bool {

	dx := [4]int{1, 1, 0, 1}
	dy := [4]int{0, 1, 1, -1}

	for i := range dx {
		count := 0

		nx, ny := x, y
		for (nx >= 0 && nx < 20) && (ny >= 0 && ny < 20) && this.board[nx][ny] == player {
			count++
			nx, ny = nx+dx[i], ny+dy[i]
		}

		nx, ny = x-dx[i], y-dy[i]
		for (nx >= 0 && nx < 20) && (ny >= 0 && ny < 20) && this.board[nx][ny] == player {
			count++
			nx, ny = nx-dx[i], ny-dy[i]
		}

		if count == 5 {
			return true
		}
	}

	return false
}

func (this *Graviton) initPlay() {
	//playerCount:=
	// randomize user turn
	for i := range this.users {
		j := rand.Intn(i + 1)
		this.users[i], this.users[j] = this.users[j], this.users[i]
	}

	this.board = [20][20]int{}
	this.turn = 0
	this.broadcast("turn " + strconv.Itoa(this.users[this.turn%this.playerCount].Id))

}

func (this *Graviton) broadcast(message string) {
	for _, v := range this.users {
		v.Emit(message)
	}
}
