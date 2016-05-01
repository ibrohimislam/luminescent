package luminescent

import (
	"log"
	"math"
	"strconv"
	"strings"
	"time"

	"encoding/json"
	"math/rand"
)

var ProtonInstance *Proton

type Proton struct {
	inputChannel <-chan *Photon

	// game players
	playerCount int
	usernameId  map[string]int    // map username -> id
	idUsername  map[int]string    // map id -> username
	idAddress   map[int]string    // map id -> address
	idPort      map[int]int64     // map id -> port
	idConn      map[int]*Electron // map id -> connection

	// game state
	gameState     int         // game state {0: waiting, 1: on game}
	idState       map[int]int // map id -> state {0: not yet ready, 1: ready, 2: dead}
	day           int
	kpuId         int64
	werewolfKills int
	civilianKills int
}

func CreateProton(inputChannel <-chan *Photon) *Proton {

	ProtonInstance = &Proton{inputChannel, 0, map[string]int{}, map[int]string{}, map[int]string{}, map[int]int64{}, map[int]*Electron{}, 0, map[int]int{}, 1, 0, 0, 0}
	return ProtonInstance

}

func (this *Proton) serverRoutine() {
	for {
		current := <-this.inputChannel

		var req map[string]interface{}
		var res map[string]interface{}

		d := json.NewDecoder(strings.NewReader(current.Message))
		d.UseNumber()

		err := d.Decode(&req)

		if err != nil {
			log.Println("Parsing error: " + err.Error())
			continue
		}

		switch {
		case req["method"] == "join":
			res = this.implJoin(current, req)
		case req["method"] == "leave":
			res = this.implLeave(current, req)
		case req["method"] == "ready":
			res = this.implReady(current, req)
		case req["method"] == "client_address":
			res = this.implClientAddress(current, req)
		case req["method"] == "accepted_proposal":
			res = this.implAcceptProposal(current, req)
		case req["method"] == "vote_result_werewolf":
			res = this.implVoteResultWerewolf(current, req)
		case req["method"] == "vote_result_civilian":
			res = this.implVoteResultCivilian(current, req)
		default:
			res = map[string]interface{}{"status": "error", "description": "method not found"}
		}

		res_data, _ := json.Marshal(res)
		current.Electron.Emit(string(res_data))
	}
}

func (this *Proton) implJoin(current *Photon, req map[string]interface{}) map[string]interface{} {

	res := map[string]interface{}{}

	username, usernameFound := req["username"]
	udpAddressInterface, udpAddressFound := req["udp_address"]
	udpPortInterface, udpPortFound := req["udp_port"]

	if !usernameFound || !udpAddressFound || !udpPortFound {
		res["status"] = "error"
		res["description"] = "wrong request"
		return res
	}

	_, isUsernameUsed := this.usernameId[username.(string)]

	if isUsernameUsed {
		res["status"] = "fail"
		res["description"] = "user exists"
		return res
	}

	if this.gameState == 1 {
		res["status"] = "fail"
		res["description"] = "please wait, game is currently running"
		return res
	}

	this.playerCount = this.playerCount + 1

	this.idConn[current.Electron.Id] = current.Electron

	this.usernameId[username.(string)] = current.Electron.Id
	this.idUsername[current.Electron.Id] = username.(string)

	udpAddress := udpAddressInterface.(string)
	this.idAddress[current.Electron.Id] = udpAddress

	port, isNumber := udpPortInterface.(json.Number)
	if !isNumber {
		res["status"] = "error"
		res["description"] = "port must be an integer"
		return res
	}
	udpPort, _ := port.Int64()
	this.idPort[current.Electron.Id] = udpPort

	this.idState[current.Electron.Id] = 0

	log.Print(username.(string) + "(" + strconv.Itoa(current.Electron.Id) + ") joins [" + udpAddress + ":" + strconv.Itoa(int(udpPort)) + "]")

	res["status"] = "ok"
	res["player_id"] = current.Electron.Id

	return res
}

func (this *Proton) implLeave(current *Photon, _ map[string]interface{}) map[string]interface{} {

	username := this.idUsername[current.Electron.Id]

	log.Print(username + "(" + strconv.Itoa(current.Electron.Id) + ") leaves")

	this.playerCount = this.playerCount - 1

	delete(this.usernameId, username)
	delete(this.idUsername, current.Electron.Id)
	delete(this.idAddress, current.Electron.Id)
	delete(this.idPort, current.Electron.Id)
	delete(this.idState, current.Electron.Id)
	delete(this.idConn, current.Electron.Id)

	if this.playerCount >= 6 {
		go this.checkPlayerStates()
	}

	res := map[string]interface{}{}

	res["status"] = "ok"

	current.Electron.ChangeState(-1)

	return res
}

func (this *Proton) implReady(current *Photon, _ map[string]interface{}) map[string]interface{} {

	if this.playerCount >= 6 {
		go this.checkPlayerStates()
	}

	this.idState[current.Electron.Id] = 1
	log.Println(strconv.Itoa(current.Electron.Id) + " is ready")

	res := map[string]interface{}{}

	res["status"] = "ok"
	res["description"] = "waiting for other player to start"

	return res
}

func (this *Proton) checkPlayerStates() {

	isEveryoneReady := true

	// checking
	for _, value := range this.idState {
		isEveryoneReady = isEveryoneReady && (value == 1)
	}

	if isEveryoneReady {
		time.Sleep(666)
		log.Println("Everyone is ready, game start!")
		this.gameStart()
	}
}

func (this *Proton) gameStart() {

	this.gameState = 1

	werewolfCount := int(math.Sqrt(float64(this.playerCount)))
	perm := rand.Perm(this.playerCount)

	werewolfUsernames := []string{}

	i := 0

	for _, username := range this.idUsername {
		if perm[i] < werewolfCount {
			werewolfUsernames = append(werewolfUsernames, username)
		}
		i++
	}

	i = 0

	for _, client := range this.idConn {

		msg := map[string]interface{}{}

		msg["method"] = "start"
		msg["time"] = "day"

		if perm[i] < werewolfCount || this.idUsername[client.Id] == "wolf" {
			msg["role"] = "werewolf"
			msg["friend"] = werewolfUsernames
		} else {
			msg["role"] = "villager"
			msg["friend"] = []string{}
		}

		msg["description"] = "game is started"

		res_data, _ := json.Marshal(msg)
		client.Emit(string(res_data))

		i++

	}

}

func (this *Proton) implClientAddress(current *Photon, _ map[string]interface{}) map[string]interface{} {
	res := map[string]interface{}{}

	res["status"] = "ok"
	res["clients"] = []map[string]interface{}{}

	for _, id := range this.usernameId {
		res["clients"] = append(res["clients"].([]map[string]interface{}), this.getClient(id))
	}

	return res
}

func (this *Proton) getClient(clientId int) map[string]interface{} {
	client := map[string]interface{}{}
	client["player_id"] = clientId

	if this.idState[clientId] == 1 {
		client["is_alive"] = 1
	} else {
		client["is_alive"] = 0
	}

	client["address"] = this.idAddress[clientId]
	client["port"] = this.idPort[clientId]
	client["username"] = this.idUsername[clientId]

	return client
}

func (this *Proton) implAcceptProposal(current *Photon, req map[string]interface{}) map[string]interface{} {

	res := map[string]interface{}{}

	kpuId, isNumber := req["kpu_id"].(json.Number)

	if !isNumber {
		res["status"] = "error"
		res["description"] = "kpu_id must be an integer"
		return res
	}

	this.kpuId, _ = kpuId.Int64()

	log.Println(strconv.Itoa(int(this.kpuId)) + " is selected as KPU")

	res["status"] = "ok"
	res["description"] = ""

	return res
}

func (this *Proton) implVoteResultWerewolf(current *Photon, req map[string]interface{}) map[string]interface{} {

	res := map[string]interface{}{}

	playerKilled, isNumber := req["player_killed"].(json.Number)

	if !isNumber {
		res["status"] = "error"
		res["description"] = "player_killed must be an integer"
		return res
	}

	werewolfKills, _ := playerKilled.Int64()
	this.werewolfKills = int(werewolfKills)

	log.Println("werewolves selected " + strconv.Itoa(this.werewolfKills) + " to be killed.")

	res["status"] = "ok"
	res["description"] = ""

	return res
}

func (this *Proton) implVoteResultCivilian(current *Photon, req map[string]interface{}) map[string]interface{} {

	res := map[string]interface{}{}

	playerKilled, isNumber := req["player_killed"].(json.Number)

	if !isNumber {
		res["status"] = "error"
		res["description"] = "player_killed must be an integer"
		return res
	}

	civilianKills, _ := playerKilled.Int64()
	this.civilianKills = int(civilianKills)

	log.Println("werewolves selected " + strconv.Itoa(this.civilianKills) + " to be killed.")

	res["status"] = "ok"
	res["description"] = ""

	return res
}
