package main

import (
    "encoding/json"
    "fmt"
    "math/rand"
    "net"
    "net/http"
    "sync"
    "time"

    "golang.org/x/net/websocket"
)

const (
    PlayerCount = 4
    CardCount = 13
    TotalCardIndex = 52

    CheckTick = time.Millisecond * 100
    PrepareCheckCount = 50
    PlayTime = time.Second * 30
    ShowTime = time.Second * 15
)

type ClientMsg struct {
    Type    string          // request type
    Name    string          // join name
    Cards   [CardCount]int  // set cards
}

type LocationMsg struct {
    Type        string  // response type
    Location    int     // location code
    Name        string  // location name
}

type RoomInfo struct {
    Name    string              // room name
    Players [PlayerCount]string // players name
}
type RoomMsg struct {
    Type    string      // response type
    Rooms   []RoomInfo  // rooms information
}

type CardMsg struct {
    Type    string                      // response type
    Cards   [PlayerCount][CardCount]int // cards information
}

type Player struct {
    name        string          // player nickname
    location    int             // player room index, -1 is in lobby
    ws          *websocket.Conn // player websocket connection
}

type Room struct {
    name    string                      // room name
    state   int                         // room state, 0: waiting, 1: ready, 2: playing
    players [PlayerCount]string         // player nickname
    cards   [PlayerCount][CardCount]int // player cards
}

type GameServer struct {
    playerList  []Player
    roomList    []Room
    mtx         *sync.Mutex
}

func (gs *GameServer) sendLocationMsg(ws *websocket.Conn) {
    var msg LocationMsg
    msg.Type = "LocationMsg"
    msg.Location = 0
    msg.Name = "Logout"
    gs.mtx.Lock()
    for i := 0; i < len(gs.playerList); i++ {
        if ws.Request().RemoteAddr == gs.playerList[i].ws.Request().RemoteAddr {
            if gs.playerList[i].location < 0 || gs.playerList[i].location >= len(gs.roomList) {
                msg.Location = 1
                msg.Name = "Lobby"
            } else {
                msg.Location = 2
                msg.Name = gs.roomList[gs.playerList[i].location].name
            }
        }
    }
    gs.mtx.Unlock()
    jsonBytes, _ := json.Marshal(msg)
    websocket.Message.Send(ws, string(jsonBytes))
}

func (gs *GameServer) sendRoomMsg(ws *websocket.Conn) {
    var msg RoomMsg
    msg.Type = "RoomMsg"
    msg.Rooms = make([]RoomInfo, 0)
    gs.mtx.Lock()
    for i := 0; i < len(gs.playerList); i++ {
        if ws.Request().RemoteAddr == gs.playerList[i].ws.Request().RemoteAddr {
            if gs.playerList[i].location < 0 || gs.playerList[i].location >= len(gs.roomList) {
                for j := 0; j < len(gs.roomList); j++ {
                    active := false
                    var room RoomInfo
                    room.Name = gs.roomList[j].name
                    for k := 0; k < PlayerCount; k++ {
                        room.Players[k] = gs.roomList[j].players[k]
                        if room.Players[k] != "" {
                            active = true
                        }
                    }
                    if active {
                        msg.Rooms = append(msg.Rooms, room)
                    }
                }
            } else {
                var room RoomInfo
                room.Name = gs.roomList[gs.playerList[i].location].name
                for j := 0; j < PlayerCount; j++ {
                    room.Players[j] = gs.roomList[gs.playerList[i].location].players[j]
                }
                msg.Rooms = append(msg.Rooms, room)
            }
            jsonBytes, _ := json.Marshal(msg)
            websocket.Message.Send(ws, string(jsonBytes))
            break
        }
    }
    gs.mtx.Unlock()
}

func (gs *GameServer) sendCardMsg(ws *websocket.Conn) {
    var msg CardMsg
    msg.Type = "CardMsg"
    for i := 0; i < PlayerCount; i++ {
        for j := 0; j < CardCount; j++ {
            msg.Cards[i][j] = -1
        }
    }
    gs.mtx.Lock()
    for i := 0; i < len(gs.playerList); i++ {
        if ws.Request().RemoteAddr == gs.playerList[i].ws.Request().RemoteAddr {
            if gs.playerList[i].location >= 0 && gs.playerList[i].location < len(gs.roomList) {
                if gs.roomList[gs.playerList[i].location].state == 2 {
                    for j := 0; j < PlayerCount; j++ {
                        if gs.roomList[gs.playerList[i].location].players[j] == gs.playerList[i].name {
                            for k := 0; k < CardCount; k++ {
                                msg.Cards[j][k] = gs.roomList[gs.playerList[i].location].cards[j][k]
                            }
                        }
                    }
                } else {
                    for j := 0; j < PlayerCount; j++ {
                        for k := 0; k < CardCount; k++ {
                            msg.Cards[j][k] = gs.roomList[gs.playerList[i].location].cards[j][k]
                        }
                    }
                }
            }
        }
    }
    gs.mtx.Unlock()
    jsonBytes, _ := json.Marshal(msg)
    websocket.Message.Send(ws, string(jsonBytes))
}

func (gs *GameServer) startGame(index int) {
    for i := 0; i < PrepareCheckCount; i++ {
        gs.mtx.Lock()
        if gs.roomList[index].state > 0 && i == 0 {
            gs.mtx.Unlock()
            return
        }
        for j := 0; j < PlayerCount; j++ {
            if gs.roomList[index].players[j] == "" {
                gs.roomList[index].state = 0
                gs.mtx.Unlock()
                return
            }
        }
        gs.roomList[index].state = 1
        gs.mtx.Unlock()
        time.Sleep(CheckTick)
    }
    for {
        gs.mtx.Lock()
        gs.roomList[index].state = 2
        newCards := make([]int, TotalCardIndex)
        for i := 0; i < TotalCardIndex; i++ {
            newCards[i] = i
        }
        for i := 0; i < TotalCardIndex; i++ {
            j := rand.Intn(TotalCardIndex)
            newCards[i], newCards[j] = newCards[j], newCards[i]
        }
        for i := 0; i < PlayerCount; i++ {
            for j := 0; j < CardCount; j++ {
                gs.roomList[index].cards[i][j] = newCards[i * CardCount + j]
            }
        }
        for i := 0; i < len(gs.playerList); i++ {
            if gs.playerList[i].location == index {
                gs.mtx.Unlock()
                gs.sendCardMsg(gs.playerList[i].ws)
                gs.mtx.Lock()
            }
        }
        gs.mtx.Unlock()
        time.Sleep(PlayTime)
        gs.mtx.Lock()
        gs.roomList[index].state = 1
        for i := 0; i < len(gs.playerList); i++ {
            if gs.playerList[i].location == index {
                gs.mtx.Unlock()
                gs.sendCardMsg(gs.playerList[i].ws)
                gs.mtx.Lock()
            }
        }
        gs.mtx.Unlock()
        time.Sleep(ShowTime)
        gs.mtx.Lock()
        for i := 0; i < PlayerCount; i++ {
            if gs.roomList[index].players[i] == "" {
                gs.roomList[index].state = 0
                gs.mtx.Unlock()
                return
            }
        }
        gs.mtx.Unlock()
    }
}

func (gs *GameServer) updateRoom(index int) {
    if index < 0 || index >= len(gs.roomList) {
        return
    }
    gs.mtx.Lock()
    for i := 0; i < len(gs.playerList); i++ {
        if gs.playerList[i].location < 0 || gs.playerList[i].location == index {
            defer gs.sendRoomMsg(gs.playerList[i].ws)
        }
    }
    gs.mtx.Unlock()
    go gs.startGame(index)
}

func (gs *GameServer) joinLobby(name string, ws *websocket.Conn) {
    defer gameServer.sendRoomMsg(ws)
    defer gameServer.sendLocationMsg(ws)
    if name == "" {
        return
    }
    if len(name) > 16 {
        name = name[:16]
    }
    gs.mtx.Lock()
    for i := 0; i < len(gs.playerList); i++ {
        if name == gs.playerList[i].name || ws.Request().RemoteAddr == gs.playerList[i].ws.Request().RemoteAddr {
            gs.mtx.Unlock()
            return
        }
    }
    var newPlayer Player
    newPlayer.name = name
    newPlayer.location = -1
    newPlayer.ws = ws
    gs.playerList = append(gs.playerList, newPlayer)
    gs.mtx.Unlock()
}

func (gs *GameServer) exitLobby(ws *websocket.Conn) {
    defer gameServer.sendLocationMsg(ws)
    gs.mtx.Lock()
    for i := 0; i < len(gs.playerList); i++ {
        if ws.Request().RemoteAddr == gs.playerList[i].ws.Request().RemoteAddr {
            if gs.playerList[i].location >= 0 && gs.playerList[i].location < len(gs.roomList) {
                for j := 0; j < PlayerCount; j++ {
                    if gs.roomList[gs.playerList[i].location].players[j] == gs.playerList[i].name {
                        gs.roomList[gs.playerList[i].location].players[j] = ""
                    }
                }
                defer gs.updateRoom(gs.playerList[i].location)
            }
            gs.playerList = append(gs.playerList[:i], gs.playerList[i+1:]...)
            break
        }
    }
    gs.mtx.Unlock()
}

func (gs *GameServer) joinRoom(name string, ws *websocket.Conn) {
    if name == "" {
        gameServer.sendLocationMsg(ws)
        return
    }
    if len(name) > 16 {
        name = name[:16]
    }
    gs.mtx.Lock()
    for i := 0; i < len(gs.playerList); i++ {
        if ws.Request().RemoteAddr == gs.playerList[i].ws.Request().RemoteAddr && gs.playerList[i].location == -1 {
            for j := 0; j < len(gs.roomList); j++ {
                if name == gs.roomList[j].name {
                    if gs.roomList[j].state == 0 {
                        for k := 0; k < PlayerCount; k++ {
                            if gs.roomList[j].players[k] == "" {
                                gs.roomList[j].players[k] = gs.playerList[i].name
                                gs.playerList[i].location = j
                                defer gs.updateRoom(gs.playerList[i].location)
                                break
                            }
                        }
                    }
                    gs.mtx.Unlock()
                    gameServer.sendLocationMsg(ws)
                    return
                }
            }
            var newRoom Room
            newRoom.name = name
            newRoom.players[0] = gs.playerList[i].name
            gs.playerList[i].location = len(gs.roomList)
            gs.roomList = append(gs.roomList, newRoom)
            defer gs.updateRoom(gs.playerList[i].location)
            break
        }
    }
    gs.mtx.Unlock()
    gameServer.sendLocationMsg(ws)
}

func (gs *GameServer) exitRoom(ws *websocket.Conn) {
    gs.mtx.Lock()
    for i := 0; i < len(gs.playerList); i++ {
        if ws.Request().RemoteAddr == gs.playerList[i].ws.Request().RemoteAddr {
            if gs.playerList[i].location >= 0 && gs.playerList[i].location < len(gs.roomList) {
                for j := 0; j < PlayerCount; j++ {
                    if gs.roomList[gs.playerList[i].location].players[j] == gs.playerList[i].name {
                        gs.roomList[gs.playerList[i].location].players[j] = ""
                    }
                }
                defer gs.updateRoom(gs.playerList[i].location)
            }
            gs.playerList[i].location = -1
            break
        }
    }
    gs.mtx.Unlock()
    gameServer.sendLocationMsg(ws)
}

func (gs *GameServer) setCard(cards [CardCount]int, ws *websocket.Conn) {
    gs.mtx.Lock()
    for i := 0; i < len(gs.playerList); i++ {
        if ws.Request().RemoteAddr == gs.playerList[i].ws.Request().RemoteAddr && gs.playerList[i].location >= 0 {
            if gs.playerList[i].location < len(gs.roomList) && gs.roomList[gs.playerList[i].location].state == 2 {
                for j := 0; j < PlayerCount; j++ {
                    if gs.roomList[gs.playerList[i].location].players[j] == gs.playerList[i].name {
                        sum := 0
                        product := 1
                        for k := 0; k < CardCount; k++ {
                            sum += gs.roomList[gs.playerList[i].location].cards[j][k]
                            product *= gs.roomList[gs.playerList[i].location].cards[j][k]
                        }
                        positive := true
                        newSum := 0
                        newProduct := 1
                        for k := 0; k < CardCount; k++ {
                            if cards[k] < 0 {
                                positive = false
                            }
                            newSum += cards[k]
                            newProduct *= cards[k]
                        }
                        if positive && sum == newSum && product == newProduct {
                            for k := 0; k < CardCount; k++ {
                                gs.roomList[gs.playerList[i].location].cards[j][k] = cards[k]
                            }
                        }
                    }
                }
            }
        }
    }
    gs.mtx.Unlock()
    gs.sendCardMsg(ws)
}

var gameServer GameServer

func Client(ws *websocket.Conn) {
    fmt.Println(ws.Request().RemoteAddr, "connect.")
    gameServer.sendLocationMsg(ws)
    for {
        var request string
        if err := websocket.Message.Receive(ws, &request); err != nil {
            break
        }
        fmt.Println(request, "Receive from", ws.Request().RemoteAddr)
        var msg ClientMsg
        json.Unmarshal([]byte(request), &msg)
        if msg.Type == "join.lobby" {
            gameServer.joinLobby(msg.Name, ws)
        } else if msg.Type == "exit.lobby" {
            gameServer.exitLobby(ws)
        } else if msg.Type == "join.room" {
            gameServer.joinRoom(msg.Name, ws)
        } else if msg.Type == "exit.room" {
            gameServer.exitRoom(ws)
            
        } else if msg.Type == "set.card" {
            gameServer.setCard(msg.Cards, ws)
        } else {
            if err := websocket.Message.Send(ws, "unknown request type"); err != nil {
                break
            }
        }
    }
    gameServer.exitLobby(ws)
    fmt.Println(ws.Request().RemoteAddr, "disconnect.")
}

func initServer() {
    gameServer.playerList = make([]Player, 0)
    gameServer.roomList = make([]Room, 0)
    gameServer.mtx = new(sync.Mutex)
}

func IP() string {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		fmt.Println(err)
	}
	defer conn.Close()
	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return localAddr.IP.To4().String()
}

func main() {
    rand.Seed(time.Now().UnixNano())
    ip := IP()
    port := ":2345"
    initServer()
    http.Handle("/", websocket.Handler(Client))
    fmt.Println("Game Server uri:", ip + port)
    if err := http.ListenAndServeTLS(ip + port, "./cert.pem", "./key.pem", nil); err != nil {
        fmt.Println("ListenAndServeTLS:", err)
    }
}