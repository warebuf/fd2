package main

import (
	"fmt"
	"github.com/google/uuid"
	"strconv"
	"sync"
	"time"
)

func p_write(p *permission_socket) {
	for msg := range p.incoming_message { // special trait of a channel, will block until something is in the channel or it is closed
		if err := p.socket.WriteJSON(msg); err != nil {
			fmt.Println("error writing to socket")
			break
		} else {
			fmt.Println("successful write")
		}
	}
}

func p_read(p *permission_socket) {
	defer func() {
		fmt.Printf("read socket closing")

		pid_to_permissions.mutex.Lock()
		delete(pid_to_permissions.global, p.pid)
		pid_to_permissions.mutex.Unlock()

		if p.open {
			fmt.Println("client left the room", p)
		}
	}()

	for {
		var msg *message
		if err := p.socket.ReadJSON(&msg); err == nil {
			fmt.Printf("JSON: %+v\n", msg)

			msg.When = time.Now()
			msg.Name = p.u.email
			fmt.Println(msg.When, "message~", msg.Name, ", Event: ", msg.Event, ", Message: ", msg.Message)

			// user requests to create match
			if msg.Event == "createMatch" {

				fmt.Println("received createMatch")

				mtch := createMatch(msg)
				go mtch.run()
				globalBroadcast(&message{Event: "newMatch", Message: mtch.game_mode + strconv.Itoa(int(mtch.capacity)), When: time.Now(), MatchID: mtch.mid}) // let everyone know there is a new room
				mtch.gamer_permission_signup <- p

			} else if msg.Event == "participantJoin" { // user request to join match

				fmt.Println("received participantJoin")

				mid_to_match.mutex.Lock()
				mtch, exists := mid_to_match.match[msg.MatchID]
				mid_to_match.mutex.Unlock()

				if exists {
					_, check_uid := mtch.gamer_permission_list[p.u.uid]
					if check_uid == true {
						fmt.Println("participant ws is already is the match!")
					} else if mtch.started == true {
						fmt.Println("match has already started, of course you can't join")
					} else if len(mtch.gamer_permission_list) > int(mtch.capacity) {
						fmt.Println("the match is full")
						msg := &message{Name: p.u.email, Message: "the match is full", Event: "failedMatchJoin", When: time.Now(), MatchID: mtch.mid}
						p.incoming_message <- msg
					} else {
						mtch.gamer_permission_signup <- p
					}
				}
			} else if msg.Event == "participantLeave" {
				mid_to_match.mutex.Lock()
				mtch, exists := mid_to_match.match[msg.MatchID]
				mid_to_match.mutex.Unlock()

				if exists {
					_, check_uid := mtch.gamer_permission_list[p.u.uid]
					if check_uid == true {
						if mtch.started == false {
							mtch.gamer_permission_signout <- p
						} else {
							fmt.Println("cannot leave permission list if match is in progress!")
						}
					} else {
						fmt.Println("cannot leave, user isn't even in the room")
					}
				}
			}

		} else {
			fmt.Println("error reading from socket")
			break
		}
	}
}

func createMatch(msg *message) *match {
	fmt.Println("called 'createMatch'")

	var ans *match
	// create a global room for users to chat in
	not_assigned := true
	for not_assigned {
		random_number := uuid.New()

		mid_to_match.mutex.RLock()
		_, found := mid_to_match.match[random_number]
		mid_to_match.mutex.RUnlock()
		if !found {

			fmt.Println(msg.Message)
			gm := "ffa"

			num := uint(1)
			sd := uint(2)
			if len(msg.Message) < 4 {
				gm = "ffa"
				num = 1
			} else {
				if msg.Message[0:3] == "ffa" {
					gm = "ffa"
				} else if msg.Message[0:3] == "tea" {
					gm = "tea"
				} else if msg.Message[0:3] == "1vx" {
					gm = "1vx"
				} else {
					gm = "ffa"
				}
				num_ops, _ := strconv.Atoi(msg.Message[3:])

				if (num_ops < 100) && (num_ops > 1) {
					num = uint(num_ops)
				} else {
					num = 2
				}

				if gm == "ffa" {
					sd = num
				} else if gm == "tea" {
					sd = 2
				} else if gm == "1vx" {
					sd = 2
				}
			}

			ans = &match{

				mutex: sync.RWMutex{},

				mid:       random_number,
				game_mode: gm,
				capacity:  num,
				sides:     sd,
				started:   false,

				broadcast:      make(chan *message),
				prio_broadcast: make(chan *message),

				gamer_join:                        make(chan *match_socket),
				gamer_uid_to_msid_to_match_socket: make(map[uuid.UUID]map[uuid.UUID]*match_socket),
				gamer_uid_to_user:                 make(map[uuid.UUID]*user),
				gamer_leave:                       make(chan *match_socket),

				gamer_permission_signup:  make(chan *permission_socket),
				gamer_permission_list:    make(map[uuid.UUID]*user),
				gamer_permission_signout: make(chan *permission_socket),

				bot_join:  make(chan *user),
				bot_leave: make(chan *user),

				spectator_join:                        make(chan *match_socket),
				spectator_uid_to_msid_to_match_socket: make(map[uuid.UUID]map[uuid.UUID]*match_socket),
				spectator_uid_to_user:                 make(map[uuid.UUID]*user),

				ticker:         time.NewTicker(2400000 * time.Hour), //will not tick until 100,000 days, or 273 years
				type_of_ticker: 0,
				start_ticker:   make(chan bool),

				message_logs: make([]*message, 0, 16),

				open: make(chan bool),
			}
			fmt.Println("created MID:", random_number)

			mid_to_match.mutex.Lock()
			mid_to_match.match[random_number] = ans
			mid_to_match.mutex.Unlock()

			not_assigned = false
		}
	}

	return ans
}

func (m *match) run() {
	for {

		// dont forget to give timer priority

		select {

		// only add the user, do not add the mmsocket
		case ws := <-m.gamer_permission_signup: // joining the waitroom for matchmaking
			_, check_uid := m.gamer_permission_list[ws.u.uid] // check if the user is in the match object
			fmt.Println("gamer_permission_signup", ws.u.uid)

			// if user is not in the match object, add the user and create a socket object
			m.mutex.Lock()
			if check_uid == false {
				m.gamer_permission_list[ws.u.uid] = ws.u
			}
			m.mutex.Unlock()

			// if this is the first socket the user has joined this room, send a message to everyone that he's joined
			if check_uid == false {
				msg := &message{Name: ws.u.email, Message: "participantJoinSuccess", Event: "participantJoinSuccess", When: time.Now(), MatchID: m.mid}
				go func() {
					pid_to_permissions.mutex.RLock()
					for _, j := range pid_to_permissions.global {
						j.incoming_message <- msg
					}
					pid_to_permissions.mutex.RUnlock()
				}()
			}

			fmt.Println("a psocket has finished permission_signup")
			printAllMatchUserWS()
			continue

		case ws := <-m.gamer_permission_signout: // leaving the waitroom for matchmaking

			_, check_uid := m.gamer_permission_list[ws.u.uid] // check if the user is in the match object
			fmt.Println("gamer_permission_signout", ws.u.uid)

			// if user is in the match object, delete all sockets as well
			m.mutex.Lock()
			if check_uid == true {
				delete(m.gamer_permission_list, ws.u.uid)
			}
			m.mutex.Unlock()

			// if this is the first socket the user has joined this room, send a message to everyone that he's joined
			if check_uid == true {
				msg := &message{Name: ws.u.email, Message: "participantLeaveSuccess", Event: "participantLeaveSuccess", When: time.Now(), MatchID: m.mid}
				go func() {
					pid_to_permissions.mutex.RLock()
					for _, j := range pid_to_permissions.global {
						j.incoming_message <- msg
					}
					pid_to_permissions.mutex.RUnlock()
				}()
			}

			fmt.Println("a psocket has finished permission_signout")
			printAllMatchUserWS()
			continue

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
			for i, j := range m.message_logs {
				ws.incoming_message <- j
				fmt.Println(i, j.Message)
			}

			// let all participants know that a new user has joined
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
			fmt.Println("a socket has left the room")

			check := (len(ws.u.rid_to_sid_to_socket[ws.m.mid]) == 0)

			// remove mid from user object
			uid_to_user.mutex.Lock()
			ws.u.mutex.Lock()
			delete(ws.u.mid_to_msid_to_match_socket[ws.m.mid], ws.msid)
			if len(ws.u.rid_to_sid_to_socket[ws.m.mid]) == 0 {
				delete(ws.u.mid_to_msid_to_match_socket, ws.m.mid)
				delete(ws.u.mid_to_match, ws.m.mid)
			}
			ws.u.mutex.Unlock()
			uid_to_user.mutex.Unlock()

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
			if check == false {
				msg := &message{Name: ws.u.email, Message: "x left the chat", Event: "left", When: time.Now()}
				m.broadcast <- msg
				fmt.Println("left the game")
			}

			// remove/take care of socket object
			ws.u = nil
			ws.m = nil
			close(ws.incoming_message)
			ws.socket.Close()
			ws.open = false

			fmt.Println("a socket has left the match")
			printAllMatchUserWS()
			continue

		case u := <-m.bot_join:

			m.mutex.Lock()
			m.gamer_permission_list[u.uid] = u
			m.gamer_uid_to_user[u.uid] = u
			//m.gamer_uid_to_msid_to_match_socket[u.uid] = make(map[uuid.UUID]*match_socket)
			m.mutex.Unlock()

			// send to all permission sockets that this bot has joined the permission list
			msg := &message{Name: u.email, Message: "participantJoinSuccess", Event: "participantJoinSuccess", When: time.Now(), MatchID: m.mid}
			pid_to_permissions.mutex.RLock()
			for _, j := range pid_to_permissions.global {
				j.incoming_message <- msg
			}
			pid_to_permissions.mutex.RUnlock()

			msg = &message{Name: u.email, Message: "x entered the chat", Event: "entered", When: time.Now()}
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
			fmt.Println("bot has joined the game")

			fmt.Println("added a bot")
			continue

		case u := <-m.bot_leave:

			m.mutex.Lock()
			delete(m.gamer_permission_list, u.uid)
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
			fmt.Println("ticker went off")

		case <-m.start_ticker:
			init_time := time.Now().Add(30 * time.Second)
			msg := &message{Event: "startMatchCountdown", When: time.Now(), MatchID: m.mid}
			m.ticker = time.NewTicker(30 * time.Second) //will tick in 30 s

			// send ticker to everyone
			m.mutex.Lock()
			m.message_logs = append(m.message_logs, msg)
			m.mutex.Unlock()

			fmt.Println("sending:", msg)

			for _, i := range m.gamer_uid_to_msid_to_match_socket {
				for _, j := range i {
					fmt.Println(j.system_time.Sub(j.user_time))
					temp_init_time := init_time
					msg.Message = temp_init_time.Add(j.system_time.Sub(j.user_time)).String()
					fmt.Println(init_time)
					fmt.Println(temp_init_time)
					select {
					case j.incoming_message <- msg:
					}
				}
			}

			for _, i := range m.spectator_uid_to_msid_to_match_socket {
				for _, j := range i {
					fmt.Println(j.system_time.Sub(j.user_time))
					msg.Message = init_time.Add(j.system_time.Sub(j.user_time)).String()
					select {
					case j.incoming_message <- msg:
					}
				}
			}
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

		}

	}
}

func globalBroadcast(msg *message) {
	fmt.Println("global broadcast")
	pid_to_permissions.mutex.RLock()
	for _, i := range pid_to_permissions.global {
		select {
		case i.incoming_message <- msg:
		}
	}

	pid_to_permissions.mutex.RUnlock()
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
