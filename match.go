package main

import (
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"

	"math/rand"
	"strconv"
	"sync"
	"time"
)

type match struct {
	mutex sync.RWMutex

	mid       uuid.UUID
	game_mode string
	sides     uint
	capacity  uint
	ended     bool

	// each client is represented as an int
	team_client_hero [][][]*hero
	uuid_to_team_int map[uuid.UUID]pair
	turn             string // tells you turn #

	broadcast      chan *message // a channel is a thread-safe queue, incoming messages
	prio_broadcast chan *message // gets priority over normal broadcast ^

	gamer_join                        chan *match_socket
	gamer_uid_to_user                 map[uuid.UUID]*user
	gamer_uid_to_msid_to_match_socket map[uuid.UUID]map[uuid.UUID]*match_socket
	gamer_leave                       chan *match_socket // a channel for clients wishing to leave

	bot_join  chan *user
	bot_leave chan *user

	spectator_join                        chan *match_socket // a channel for clients wishing to join
	spectator_uid_to_msid_to_match_socket map[uuid.UUID]map[uuid.UUID]*match_socket
	spectator_uid_to_user                 map[uuid.UUID]*user

	phase string

	ticker       *time.Ticker
	start_ticker chan bool

	next_time time.Time

	char_sel_done   map[uuid.UUID]bool
	char_sel_ticker chan bool

	message_logs []*message

	simulate chan bool
}

type match_socket struct {
	mutex sync.RWMutex

	socket *websocket.Conn

	msid uuid.UUID

	u *user
	m *match

	user_time   time.Time
	system_time time.Time

	incoming_message chan *message // send is a channel on which messages are sent.

	open bool
}

func m_write(m *match_socket) {
	for msg := range m.incoming_message { // special trait of a channel, will block until something is in the channel or it is closed
		if err := m.socket.WriteJSON(msg); err != nil {
			fmt.Println("error writing to socket")
			break
		} else {
			fmt.Println("successful write")
		}
	}
}

func m_read(m *match_socket) {
	defer func() {
		fmt.Printf("read socket closing")
		m.m.gamer_leave <- m

		if m.open {
			fmt.Println("client left the match (match_socket)")
		}
	}()

	for {
		var msg *message
		if err := m.socket.ReadJSON(&msg); err == nil {
			fmt.Printf("JSON: %+v\n", msg)

			msg.When = time.Now()
			msg.Name = m.u.email
			fmt.Println(msg.When, "message~", msg.Name, ", Event: ", msg.Event, ", Message: ", msg.Message)

			// user requests to create match
			if msg.Event == "act" {
				fmt.Println("received act")
				// check if the match has started and the action is to a user's own bot
				if m.m.phase == "TURN" && m.m.ended == false && m.m.uuid_to_team_int[m.u.uid].a == msg.Team_index && m.m.uuid_to_team_int[m.u.uid].b == msg.Client_index {
					// check if the bot is alive
					if m.m.team_client_hero[msg.Team_index][msg.Client_index][msg.Hero_index].H.HP > 0 && m.m.team_client_hero[msg.Team_index][msg.Client_index][msg.Hero_index].Position == 0 {
						if msg.Message == "HEAD" || msg.Message == "LARM" || msg.Message == "RARM" || msg.Message == "BOTTOM" {

							if msg.Message == "HEAD" {
								m.m.team_client_hero[msg.Team_index][msg.Client_index][msg.Hero_index].Move = 0
							} else if msg.Message == "LARM" {
								m.m.team_client_hero[msg.Team_index][msg.Client_index][msg.Hero_index].Move = 1
							} else if msg.Message == "RARM" {
								m.m.team_client_hero[msg.Team_index][msg.Client_index][msg.Hero_index].Move = 2
							} else if msg.Message == "BOTTOM" {
								m.m.team_client_hero[msg.Team_index][msg.Client_index][msg.Hero_index].Move = 3
							}
							m.m.team_client_hero[msg.Team_index][msg.Client_index][msg.Hero_index].Direction = 1

							// check if all bots have commands, if so, calculate next position
							sim_check := true
							for i := 0; i < len(m.m.team_client_hero); i++ {
								for j := 0; j < len(m.m.team_client_hero[i]); j++ {
									for k := 0; k < len(m.m.team_client_hero[i][j]); k++ {
										if (m.m.team_client_hero[i][j][k].Bot == false) &&
											(m.m.team_client_hero[i][j][k].Move < 0) &&
											(m.m.team_client_hero[i][j][k].Position == 0) &&
											(m.m.team_client_hero[i][j][k].H.HP > 0) {
											sim_check = false
											fmt.Println("sim_check", i, j, k)
										}
									}
								}
							}
							fmt.Println("sim_check", sim_check)
							if sim_check == true {
								m.m.simulate <- true
							} else {
								m.m.sharepos(nil)
							}
						}
					}

				}

			} else if msg.Event == "clockSyncResponse" {
				test, _ := strconv.ParseInt(msg.Message, 10, 64)
				m.user_time = time.UnixMilli(test)
				fmt.Println("dif", m.system_time.Sub(m.user_time))

				msg := &message{Event: "ticker_start", When: time.Now(), MatchID: m.m.mid}
				msg.Message = m.m.next_time.Add(m.user_time.Sub(m.system_time)).UTC().String()
				m.incoming_message <- msg

			} else if msg.Event == "endCharSel" {

				m.m.char_sel_done[m.u.uid] = true
				all_done := true
				for _, j := range m.m.char_sel_done {
					if j == false {
						all_done = false
						break
					}
				}
				if all_done {
					m.m.start_ticker <- true
				}
			}

		} else {
			fmt.Println("error reading from socket")
			break
		}
	}
}

func createMatch(pl *permission_list) *match {
	fmt.Println("called 'createMatch'")

	var ans *match

	mid_to_match.mutex.RLock()
	_, found := mid_to_match.match[pl.plid] // USE SAME PLID FOR MID
	mid_to_match.mutex.RUnlock()
	if !found {

		ans = &match{

			mutex: sync.RWMutex{},

			mid:       pl.plid,
			game_mode: pl.game_mode,
			capacity:  pl.capacity,
			sides:     pl.sides,
			ended:     false,

			team_client_hero: make([][][]*hero, 0, 0),
			//TCH_JSON:         make([][][]string, 0, 0),
			uuid_to_team_int: make(map[uuid.UUID]pair),

			broadcast:      make(chan *message),
			prio_broadcast: make(chan *message),

			gamer_join:                        make(chan *match_socket),
			gamer_uid_to_msid_to_match_socket: make(map[uuid.UUID]map[uuid.UUID]*match_socket),
			gamer_uid_to_user:                 make(map[uuid.UUID]*user),
			gamer_leave:                       make(chan *match_socket),

			bot_join:  make(chan *user),
			bot_leave: make(chan *user),

			spectator_join:                        make(chan *match_socket),
			spectator_uid_to_msid_to_match_socket: make(map[uuid.UUID]map[uuid.UUID]*match_socket),
			spectator_uid_to_user:                 make(map[uuid.UUID]*user),

			phase: "null",

			ticker:       time.NewTicker(2400000 * time.Hour), //will not tick until 100,000 days, or 273 years
			turn:         "null",
			start_ticker: make(chan bool),

			char_sel_done:   make(map[uuid.UUID]bool),
			char_sel_ticker: make(chan bool),

			message_logs: make([]*message, 0, 16),

			simulate: make(chan bool),
		}

		// create each team
		for i := 0; i < int(ans.sides); i++ {
			ans.team_client_hero = append(ans.team_client_hero, make([][]*hero, 0))
		}

		// assign clients and heroes to each team
		team_int := 0
		client_int := 0
		for i, j := range pl.gamer_permission_list {

			ans.char_sel_done[i] = false

			ans.uuid_to_team_int[i] = pair{team_int, client_int, strconv.Itoa(team_int) + ";" + strconv.Itoa(client_int)}
			ans.team_client_hero[team_int] = append(ans.team_client_hero[team_int], make([]*hero, 0, 5))

			for y := 0; y < 5; y++ {
				h := head{
					SERIAL: 0,

					HP:          100,
					ATK:         0,
					DEF:         0,
					ACC:         0,
					CRT:         0,
					MOB:         0,
					CD:          0,
					CLU:         0,
					Use_current: 0,
					Use_outof:   0,
					Weight:      1,
				}
				larm := arm{
					SERIAL: 0,
					LORR:   false,

					HP:          100,
					ATK:         0,
					DEF:         0,
					ACC:         0,
					CRT:         0,
					MOB:         0,
					CD:          0,
					CLU:         0,
					Use_current: 0,
					Use_outof:   0,
					Weight:      1,
				}
				rarm := arm{
					SERIAL: 0,
					LORR:   true,

					HP:          100,
					ATK:         0,
					DEF:         0,
					ACC:         0,
					CRT:         0,
					MOB:         0,
					CD:          0,
					CLU:         0,
					Use_current: 0,
					Use_outof:   0,
					Weight:      1,
				}
				btm := bottom{
					SERIAL: 0,

					HP:          100,
					ATK:         0,
					DEF:         0,
					ACC:         0,
					CRT:         0,
					MOB:         0,
					CD:          0,
					CLU:         0,
					Use_current: 0,
					Use_outof:   0,
					Weight:      1,

					DOG: 0,
					SPD: rand.Intn(10) * 10,
					ACL: 0,
					ANT: 0,
					END: 0,
				}
				temp := &hero{
					Bot:       j.bot_status,
					Position:  0,
					Direction: 0,
					Move:      -1,
					H:         h,
					L:         larm,
					R:         rarm,
					B:         btm,
				}
				ans.team_client_hero[team_int][client_int] = append(ans.team_client_hero[team_int][client_int], temp)
			}

			if ans.game_mode == "ffa" {
				team_int++
			} else if ans.game_mode == "tea" {
				if team_int == 0 {
					team_int = 1
				} else {
					team_int = 0
				}
			} else if ans.game_mode == "1vx" {
				team_int = 1
			}
			if len(ans.team_client_hero) > team_int {
				client_int = len(ans.team_client_hero[team_int])
			}
		}

		fmt.Println("created MID:", ans.mid)

		mid_to_match.mutex.Lock()
		mid_to_match.match[ans.mid] = ans
		mid_to_match.mutex.Unlock()
	}

	return ans
}

func (m *match) run() {
	for {

		// dont forget to give timer priority

		select {

		case ws := <-m.gamer_join: // joining

			_, check_uid := m.gamer_uid_to_user[ws.u.uid] // check if the user is in the match object
			_, check_mid := ws.u.mid_to_match[m.mid]      // check if the match is to the user object
			fmt.Println("MID JOIN:", m.mid)
			fmt.Println("UID JOIN:", ws.u.uid)

			// if user is not in the match object, add the user and create a socket object
			m.mutex.Lock()
			if check_uid == false {
				m.gamer_uid_to_user[ws.u.uid] = ws.u
				m.gamer_uid_to_msid_to_match_socket[ws.u.uid] = make(map[uuid.UUID]*match_socket)
			}
			m.gamer_uid_to_msid_to_match_socket[ws.u.uid][ws.msid] = ws
			m.mutex.Unlock()

			// if match is not in the user object, add the match and create a match_socket object
			ws.u.mutex.Lock()
			if check_mid == false {
				ws.u.mid_to_match[m.mid] = m
				ws.u.mid_to_msid_to_match_socket[m.mid] = make(map[uuid.UUID]*match_socket)
			}
			ws.u.mid_to_msid_to_match_socket[m.mid][ws.msid] = ws
			ws.u.mutex.Unlock()

			fmt.Println("joined the game")

			// send the joined user a list of all historic messages
			for _, j := range m.message_logs {
				ws.incoming_message <- j
			}

			if m.ended == true {
				ws.incoming_message <- &message{Event: "game_over"}
			}

			// let all participants know that a new user has joined (if only socket for this user that has joined)
			if check_uid == false {
				msg := &message{Name: ws.u.email, Message: "x entered the chat", Event: "entered", When: time.Now()}

				m.mutex.Lock()
				m.message_logs = append(m.message_logs, msg)
				m.mutex.Unlock()

				fmt.Println("sending:", msg)

				for _, i := range m.gamer_uid_to_msid_to_match_socket {
					for _, j := range i {
						select {
						case j.incoming_message <- msg:
						}
					}
				}

				for _, i := range m.spectator_uid_to_msid_to_match_socket {
					for _, j := range i {
						select {
						case j.incoming_message <- msg:
						}
					}
				}
			}

			fmt.Println("a socket has joined the match")
			printAllMatchUserWS()

			// start communication to sync clocks
			ws.system_time = time.Now()
			msg := &message{Event: "clockSyncRequest", Message: ws.system_time.String(), When: time.Now()}
			ws.incoming_message <- msg

			continue

		case ws := <-m.gamer_leave: // leaving
			fmt.Println("<-m.gamer_leave")

			// remove mid from user object
			ws.u.mutex.Lock()
			delete(ws.u.mid_to_msid_to_match_socket[ws.m.mid], ws.msid)
			if len(ws.u.mid_to_msid_to_match_socket[ws.m.mid]) == 0 {
				delete(ws.u.mid_to_msid_to_match_socket, ws.m.mid)
				delete(ws.u.mid_to_match, ws.m.mid)
			}
			ws.u.mutex.Unlock()

			// remove uid from match object
			rid_to_room.mutex.Lock()
			ws.m.mutex.Lock()
			delete(ws.m.gamer_uid_to_msid_to_match_socket[ws.u.uid], ws.msid)
			if len(ws.m.gamer_uid_to_msid_to_match_socket[ws.u.uid]) == 0 {
				delete(ws.m.gamer_uid_to_msid_to_match_socket, ws.u.uid)
				delete(ws.m.gamer_uid_to_user, ws.u.uid)
			}
			ws.m.mutex.Unlock()
			rid_to_room.mutex.Unlock()

			// broadcast to all users of the room that the user has left
			if len(ws.m.gamer_uid_to_msid_to_match_socket[ws.u.uid]) == 0 {
				msg := &message{Name: ws.u.email, Message: "x left the chat", Event: "left", When: time.Now()}
				m.mutex.Lock()
				m.message_logs = append(m.message_logs, msg)
				m.mutex.Unlock()

				fmt.Println("sending:", msg)

				for _, i := range m.gamer_uid_to_msid_to_match_socket {
					for _, j := range i {
						select {
						case j.incoming_message <- msg:
						}
					}
				}

				for _, i := range m.spectator_uid_to_msid_to_match_socket {
					for _, j := range i {
						select {
						case j.incoming_message <- msg:
						}
					}
				}
			}

			// remove/take care of socket object
			ws.u = nil
			ws.m = nil
			close(ws.incoming_message)
			ws.socket.Close()
			ws.open = false

			// if empty now, delete match, clear all users and global variables of objects related to this match
			if m.ended == true && len(m.gamer_uid_to_msid_to_match_socket) == 0 {
				delete(mid_to_match.match, m.mid)
				//permissionlistBroadcast(&pmessage{Event: "removeMatch", Message: m.mid.String()}) // let everyone know there is a new room
				break
			} else {
				printAllMatchUserWS()
			}
			continue

		case u := <-m.bot_join:

			m.mutex.Lock()
			m.gamer_uid_to_user[u.uid] = u
			//m.gamer_uid_to_msid_to_match_socket[u.uid] = make(map[uuid.UUID]*match_socket)
			m.mutex.Unlock()

			// send to all permission sockets that this bot has joined the permission list
			msg1 := &pmessage{Name: u.email, Message: "participantJoinSuccess", Event: "participantJoinSuccess", When: time.Now(), PLID: m.mid}
			pid_to_permissions.mutex.RLock()
			for _, j := range pid_to_permissions.global {
				j.incoming_message <- msg1
			}
			pid_to_permissions.mutex.RUnlock()

			// send to all match sockets that the bot has entered the match
			msg2 := &message{Name: u.email, Message: "x entered the match", Event: "entered", When: time.Now()}
			m.mutex.Lock()
			m.message_logs = append(m.message_logs, msg2)
			m.mutex.Unlock()
			fmt.Println("sending:", msg2)
			for _, i := range m.gamer_uid_to_msid_to_match_socket {
				for _, j := range i {
					select {
					case j.incoming_message <- msg2:
					}
				}
			}
			for _, i := range m.spectator_uid_to_msid_to_match_socket {
				for _, j := range i {
					select {
					case j.incoming_message <- msg2:
					}
				}
			}
			fmt.Println("bot has joined the game")

			// check if everyone is set up to begin the game countdown
			if len(m.gamer_uid_to_user) == int(m.capacity) {
				allset := true
				for i, j := range m.gamer_uid_to_msid_to_match_socket {
					if uid_to_user.users[i].bot_status == true {

					} else {
						for _, l := range j {
							if (l.user_time == time.Time{}) {
								allset = false
								break
							}
						}
					}
					if allset == false {
						break
					}
				}
				if allset == true && m.ended == false {
					fmt.Println("started gamecountdown from bot join")
					m.char_sel_ticker <- true
				}
			}

			fmt.Println("added a bot")
			continue

		case u := <-m.bot_leave:

			m.mutex.Lock()
			delete(m.gamer_uid_to_user, u.uid)
			//delete(m.gamer_uid_to_msid_to_match_socket, u.uid)
			m.mutex.Unlock()

			go func(m *match, email string) {
				msg := &message{Name: email, Message: "x left the chat", Event: "left", When: time.Now()}
				m.broadcast <- msg
				fmt.Println("bot has left the game")
			}(m, u.email)

			fmt.Println("removed a bot")
			continue

		case <-m.ticker.C: // ticker goes off

			fmt.Println("m.Ticker.C")

			if m.phase == "CHARACTER SELECTION" {
				m.start_ticker <- true
			} else if m.turn[0:4] == "TURN" {
				num, _ := strconv.Atoi(m.turn[5:])
				m.turn = "TURN " + strconv.Itoa(num+1)

				msg := &message{Event: "ticker_start", When: time.Now(), MatchID: m.mid}

				m.next_time = time.Now().Add(120 * time.Second)
				m.ticker = time.NewTicker(m.next_time.Sub(time.Now())) //will tick in 30 s

				// send ticker to everyone
				m.mutex.Lock()
				m.message_logs = append(m.message_logs, msg)
				m.mutex.Unlock()

				fmt.Println("sending:", msg)

				for _, i := range m.gamer_uid_to_msid_to_match_socket {
					for _, j := range i {
						msg.Message = m.next_time.Add(j.user_time.Sub(j.system_time)).UTC().String()
						select {
						case j.incoming_message <- msg:
						}
					}
				}

				for _, i := range m.spectator_uid_to_msid_to_match_socket {
					for _, j := range i {
						msg.Message = m.next_time.Add(j.system_time.Sub(j.user_time)).String()
						select {
						case j.incoming_message <- msg:
						}
					}
				}
			}

		case <-m.char_sel_ticker:

			m.phase = "CHARACTER SELECTION"
			msg := &message{Event: "ticker_start", When: time.Now(), MatchID: m.mid}
			msg2 := &message{Event: "update_phase", Phase: m.phase, MatchID: m.mid}
			m.next_time = time.Now().Add(600 * time.Second)
			m.ticker = time.NewTicker(m.next_time.Sub(time.Now()))

			// send ticker to everyone
			fmt.Println("sending:", msg)

			for _, i := range m.gamer_uid_to_msid_to_match_socket {
				for _, j := range i {
					msg.Message = m.next_time.Add(j.user_time.Sub(j.system_time)).UTC().String()
					select {
					case j.incoming_message <- msg:
						j.incoming_message <- msg2
					}
				}
			}

			for _, i := range m.spectator_uid_to_msid_to_match_socket {
				for _, j := range i {
					msg.Message = m.next_time.Add(j.system_time.Sub(j.user_time)).String()
					select {
					case j.incoming_message <- msg:
						j.incoming_message <- msg2
					}
				}
			}
			continue

		case <-m.start_ticker:
			m.phase = "TURN"
			m.turn = "TURN 0"
			msg := &message{Event: "ticker_start", When: time.Now(), MatchID: m.mid}

			m.next_time = time.Now().Add(120 * time.Second)
			m.ticker = time.NewTicker(m.next_time.Sub(time.Now())) //will tick in 30 s

			// send ticker to everyone
			m.mutex.Lock()
			m.message_logs = append(m.message_logs, msg)
			m.mutex.Unlock()

			fmt.Println("sending:", msg)

			for _, i := range m.gamer_uid_to_msid_to_match_socket {
				for _, j := range i {
					msg.Message = m.next_time.Add(j.user_time.Sub(j.system_time)).UTC().String()
					select {
					case j.incoming_message <- msg:
					}
				}
			}

			for _, i := range m.spectator_uid_to_msid_to_match_socket {
				for _, j := range i {
					msg.Message = m.next_time.Add(j.system_time.Sub(j.user_time)).String()
					select {
					case j.incoming_message <- msg:
					}
				}
			}

			m.sharepos(nil)

			continue

		case msg := <-m.broadcast: // forward message to all clients

			m.mutex.Lock()
			m.message_logs = append(m.message_logs, msg)
			m.mutex.Unlock()

			fmt.Println("sending:", msg)

			for _, i := range m.gamer_uid_to_msid_to_match_socket {
				for _, j := range i {
					select {
					case j.incoming_message <- msg:
					}
				}
			}

			for _, i := range m.spectator_uid_to_msid_to_match_socket {
				for _, j := range i {
					select {
					case j.incoming_message <- msg:
					}
				}
			}
			continue

		case <-m.simulate:

			fmt.Println("SIMULATE")

			m.ticker.Stop()

			// give all bots an action
			gave_bots_acts := false
			for i := 0; i < len(m.team_client_hero); i++ {
				for j := 0; j < len(m.team_client_hero[i]); j++ {
					for k := 0; k < len(m.team_client_hero[i][j]); k++ {
						if (m.team_client_hero[i][j][k].Position == 0) && (m.team_client_hero[i][j][k].Direction == 0) && (m.team_client_hero[i][j][k].Move < 0) && (m.team_client_hero[i][j][k].Bot == true) {
							m.team_client_hero[i][j][k].Move = rand.Intn(4)
							m.team_client_hero[i][j][k].Direction = 1
							gave_bots_acts = true
						}
					}
				}
			}
			if gave_bots_acts == true {
				m.sharepos(nil)
			}

			// calculate min unit of time to action
			min_units := 999999.99 //9,223,372,036,854,775,807 (9 quintillion)
			for i := 0; i < len(m.team_client_hero); i++ {
				for j := 0; j < len(m.team_client_hero[i]); j++ {
					for k := 0; k < len(m.team_client_hero[i][j]); k++ {
						if m.team_client_hero[i][j][k].H.HP > 0 {
							if m.team_client_hero[i][j][k].Direction == 0 {
								units_of_time := 999999.99
								if m.team_client_hero[i][j][k].B.SPD != 0 {
									units_of_time = toFixed(m.team_client_hero[i][j][k].Position/float64(m.team_client_hero[i][j][k].B.SPD), 3)
								}
								//fmt.Println(m.team_client_hero[i][j][k].Position, m.team_client_hero[i][j][k].B.SPD, units_of_time)
								if units_of_time < min_units {
									min_units = units_of_time
								}
							} else if m.team_client_hero[i][j][k].Direction == 1 {
								units_of_time := 999999.99
								if m.team_client_hero[i][j][k].B.SPD != 0 {
									units_of_time = toFixed((100-m.team_client_hero[i][j][k].Position)/float64(m.team_client_hero[i][j][k].B.SPD), 3)
								}
								//fmt.Println(m.team_client_hero[i][j][k].Position, m.team_client_hero[i][j][k].B.SPD, units_of_time)
								if units_of_time < min_units {
									min_units = units_of_time
								}
							}
						}
					}
				}
			}
			fmt.Println("min", min_units)

			has_cmds := false
			only_bot_cmds := true

			list_of_attackers := [][]int{}

			// move all unit positions
			for i := 0; i < len(m.team_client_hero); i++ {
				for j := 0; j < len(m.team_client_hero[i]); j++ {
					for k := 0; k < len(m.team_client_hero[i][j]); k++ {
						if m.team_client_hero[i][j][k].H.HP > 0 {
							if m.team_client_hero[i][j][k].Direction == 0 {

								new_pos := toFixed(m.team_client_hero[i][j][k].Position-(min_units*float64(m.team_client_hero[i][j][k].B.SPD)), 3)
								if new_pos <= 0.01 {
									new_pos = 0
									has_cmds = true
									if m.team_client_hero[i][j][k].Bot == false {
										only_bot_cmds = false
									}
								}
								m.team_client_hero[i][j][k].Position = new_pos

							} else if m.team_client_hero[i][j][k].Direction == 1 {
								new_pos := toFixed(m.team_client_hero[i][j][k].Position+(min_units*float64(m.team_client_hero[i][j][k].B.SPD)), 3)
								if new_pos >= 99.99 {
									new_pos = 100
									list_of_attackers = append(list_of_attackers, []int{i, j, k})
								}
								m.team_client_hero[i][j][k].Position = new_pos
							}
						}
					}
				}
			}
			fmt.Println("finished unit movement")
			m.sharepos(nil)

			// simulate all attacks
			atk_list := make([]*attack, 0)
			for i := 0; i < len(list_of_attackers); i++ {
				a := &attack{
					Attacker: make([]int, 3),
					Defender: make([][]int, 0),
					Damage:   make([][]string, 0),
				}
				a.Attacker[0] = list_of_attackers[i][0]
				a.Attacker[1] = list_of_attackers[i][1]
				a.Attacker[2] = list_of_attackers[i][2]
				a.Defender = closest_enemies(m.team_client_hero, a.Attacker[0], a.Attacker[1], a.Attacker[2])
				a.Damage = close_attack(m.team_client_hero, a.Attacker[0], a.Attacker[1], a.Attacker[2], a.Defender)

				if len(a.Defender) > 0 {
					atk_list = append(atk_list, a)
				}

				m.team_client_hero[a.Attacker[0]][a.Attacker[1]][a.Attacker[2]].Direction = 0
				m.team_client_hero[a.Attacker[0]][a.Attacker[1]][a.Attacker[2]].Move = -1
			}

			num, _ := strconv.Atoi(m.turn[5:])
			if len(atk_list) > 0 {
				m.turn = "TURN " + strconv.Itoa(num+len(atk_list))
			} else {
				m.turn = "TURN " + strconv.Itoa(num+1)
			}

			if len(atk_list) > 0 {
				m.sharepos(atk_list)
			}

			if game_over_check(m.team_client_hero) {
				// have to handle a game over

				// just setting this for now to make sure we don't call 'SIMULATE' anymore
				has_cmds = true
				only_bot_cmds = false

				m.ended = true
				m.ticker.Stop()

				msg := &message{Event: "game_over"}

				for _, i := range m.gamer_uid_to_msid_to_match_socket {
					for _, j := range i {
						select {
						case j.incoming_message <- msg:
						}
					}
				}

				for _, i := range m.spectator_uid_to_msid_to_match_socket {
					for _, j := range i {
						select {
						case j.incoming_message <- msg:
						}
					}
				}

			}

			// send new timer to everyone
			timer1_length := time.Second * 120
			timer2_length := time.Second * 121

			init_time := time.Now().Add(timer1_length)
			msg := &message{Event: "ticker_start", When: time.Now(), MatchID: m.mid}
			m.ticker = time.NewTicker(timer2_length)
			msg2 := &message{Event: "unitsOfTime", When: time.Now(), Message: strconv.FormatFloat(min_units, 'E', 3, 64)}

			// send ticker to everyone
			m.mutex.Lock()
			m.message_logs = append(m.message_logs, msg)
			m.message_logs = append(m.message_logs, msg2)
			m.mutex.Unlock()

			fmt.Println("sending:", msg)

			for _, i := range m.gamer_uid_to_msid_to_match_socket {
				for _, j := range i {
					msg.Message = init_time.Add(j.user_time.Sub(j.system_time)).UTC().String()
					select {
					case j.incoming_message <- msg:
						j.incoming_message <- msg2
					}
				}
			}

			for _, i := range m.spectator_uid_to_msid_to_match_socket {
				for _, j := range i {
					msg.Message = init_time.Add(j.system_time.Sub(j.user_time)).String()
					select {
					case j.incoming_message <- msg:
						j.incoming_message <- msg2
					}
				}
			}

			if (has_cmds == false) || (only_bot_cmds == true) {
				fmt.Println("concurrent SIMULATE called")
				go func() { m.simulate <- true }()
			}

		}
	}
}

func printAllMatchUserWS() {
	fmt.Println("printing everything")

	for i, j := range mid_to_match.match {
		fmt.Println("MID:", i)
		for k, l := range j.gamer_uid_to_user {
			fmt.Println("UID:", k, l.email)
		}
	}
}
func (m *match) sharepos(a []*attack) {
	fmt.Println("sharepos", a)

	atk_temp := make([]string, 0)
	if a == nil {
		atk_temp = nil
	} else {
		for i := 0; i < len(a); i++ {
			marshalled, _ := json.Marshal(a[i])
			atk_temp = append(atk_temp, string(marshalled))
		}
	}

	// convert game state to JSON
	temp := make([][][]string, len(m.team_client_hero))
	for i := 0; i < len(m.team_client_hero); i++ {
		temp[i] = make([][]string, len(m.team_client_hero[i]))
		for j := 0; j < len(m.team_client_hero[i]); j++ {
			temp[i][j] = make([]string, len(m.team_client_hero[i][j]))
			for k := 0; k < len(m.team_client_hero[i][j]); k++ {
				marshalled, _ := json.Marshal(m.team_client_hero[i][j][k])
				//fmt.Println(string(marshalled))
				//m.TCH_JSON[i][j][k] = string(marshalled)
				temp[i][j][k] = string(marshalled)
			}
		}
	}

	// send updated positions to everyone
	for k, i := range m.gamer_uid_to_msid_to_match_socket {
		for _, j := range i {
			select {
			case j.incoming_message <- &message{Event: "game_state", TCH: temp, Atk: atk_temp, Message: m.uuid_to_team_int[k].ab, Turn: m.turn, When: time.Now(), MatchID: m.mid}:
			}
		}
	}
	for _, i := range m.spectator_uid_to_msid_to_match_socket {
		for _, j := range i {
			select {
			case j.incoming_message <- &message{Event: "game_state", TCH: temp, Atk: atk_temp, When: time.Now(), MatchID: m.mid}:
			}
		}
	}
}
func closest_enemies(state [][][]*hero, atk_t int, atk_u int, atk_b int) [][]int {

	ans := [][]int{}

	closest := 999999.99

	// calculate closest enemy
	for i := 0; i < len(state); i++ {
		if i != atk_t {
			for j := 0; j < len(state[i]); j++ {
				for k := 0; k < len(state[i][j]); k++ {
					if state[i][j][k].H.HP > 0 {
						if 100-state[i][j][k].Position == closest {
							closest = 100 - state[i][j][k].Position
							ans = append(ans, []int{i, j, k})
						} else if 100-state[i][j][k].Position < closest {
							closest = 100 - state[i][j][k].Position
							ans = append([][]int{}, []int{i, j, k})
						}
					}
				}
			}
		}
	}

	return ans
}
func close_attack(state [][][]*hero, atk_t int, atk_u int, atk_b int, def [][]int) [][]string {

	fmt.Println("called close attack")

	if len(def) == 0 {
		fmt.Println("no more enemies to attack")
		return [][]string{}
	}

	closest_i := def[0][0]
	closest_j := def[0][1]
	closest_k := def[0][2]

	// attack first closest enemy
	dmg := 100
	hweight := int(0)
	lweight := int(0)
	rweight := int(0)
	bweight := int(0)

	dmg_list := [][]string{} // each unit attacked, each part attacked

	if state[closest_i][closest_j][closest_k].H.HP > 0 {
		hweight = int(100 * (float64(state[closest_i][closest_j][closest_k].H.Weight) / float64(state[closest_i][closest_j][closest_k].H.Weight+state[closest_i][closest_j][closest_k].L.Weight+state[closest_i][closest_j][closest_k].R.Weight+state[closest_i][closest_j][closest_k].B.Weight)))
	}
	if state[closest_i][closest_j][closest_k].L.HP > 0 {
		lweight = int(100 * (float64(state[closest_i][closest_j][closest_k].L.Weight) / float64(state[closest_i][closest_j][closest_k].H.Weight+state[closest_i][closest_j][closest_k].L.Weight+state[closest_i][closest_j][closest_k].R.Weight+state[closest_i][closest_j][closest_k].B.Weight)))
	}
	if state[closest_i][closest_j][closest_k].R.HP > 0 {
		rweight = int(100 * (float64(state[closest_i][closest_j][closest_k].R.Weight) / float64(state[closest_i][closest_j][closest_k].H.Weight+state[closest_i][closest_j][closest_k].L.Weight+state[closest_i][closest_j][closest_k].R.Weight+state[closest_i][closest_j][closest_k].B.Weight)))
	}
	if state[closest_i][closest_j][closest_k].B.HP > 0 {
		bweight = int(100 * (float64(state[closest_i][closest_j][closest_k].B.Weight) / float64(state[closest_i][closest_j][closest_k].H.Weight+state[closest_i][closest_j][closest_k].L.Weight+state[closest_i][closest_j][closest_k].R.Weight+state[closest_i][closest_j][closest_k].B.Weight)))
	}

	random_number := rand.Intn(hweight + lweight + rweight + bweight)

	if random_number < hweight {
		state[closest_i][closest_j][closest_k].H.HP = state[closest_i][closest_j][closest_k].H.HP - dmg
		dmg_list = append(dmg_list, []string{"H;100;" + strconv.Itoa(hweight) + ";" + strconv.Itoa(lweight) + ";" + strconv.Itoa(rweight) + ";" + strconv.Itoa(bweight) + ";" + strconv.Itoa(random_number)})
	} else if random_number < hweight+lweight {
		state[closest_i][closest_j][closest_k].L.HP = state[closest_i][closest_j][closest_k].L.HP - dmg
		dmg_list = append(dmg_list, []string{"L;100;" + strconv.Itoa(hweight) + ";" + strconv.Itoa(lweight) + ";" + strconv.Itoa(rweight) + ";" + strconv.Itoa(bweight) + ";" + strconv.Itoa(random_number)})
	} else if random_number < hweight+lweight+rweight {
		state[closest_i][closest_j][closest_k].R.HP = state[closest_i][closest_j][closest_k].R.HP - dmg
		dmg_list = append(dmg_list, []string{"R;100;" + strconv.Itoa(hweight) + ";" + strconv.Itoa(lweight) + ";" + strconv.Itoa(rweight) + ";" + strconv.Itoa(bweight) + ";" + strconv.Itoa(random_number)})
	} else {
		state[closest_i][closest_j][closest_k].B.HP = state[closest_i][closest_j][closest_k].B.HP - dmg
		dmg_list = append(dmg_list, []string{"B;100;" + strconv.Itoa(hweight) + ";" + strconv.Itoa(lweight) + ";" + strconv.Itoa(rweight) + ";" + strconv.Itoa(bweight) + ";" + strconv.Itoa(random_number)})
	}

	return dmg_list
}
func game_over_check(state [][][]*hero) bool {

	// check all teams, if there are two+ teams with alive heroes, then it is not game over
	counter := 0
	for i := 0; i < len(state); i++ {
		at_least_one_alive := false
		for j := 0; j < len(state[i]); j++ {
			for k := 0; k < len(state[i][j]); k++ {
				if state[i][j][k].H.HP > 0 {
					at_least_one_alive = true
					break
				}
			}
			if at_least_one_alive == true {
				break
			}
		}
		if at_least_one_alive == true {
			counter++
			if counter >= 2 {
				return false
			}
		}
	}

	return true
}
