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
		fmt.Printf("read socket closing: %+v\n", p)

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
			}

			ans = &match{

				mutex: sync.RWMutex{},

				mid:       random_number,
				game_mode: gm,
				capacity:  num,
				started:   false,

				broadcast: make(chan *message),

				gamer_join:                        make(chan *match_socket),
				gamer_uid_to_msid_to_match_socket: make(map[uuid.UUID]map[uuid.UUID]*match_socket),
				gamer_uid_to_user:                 make(map[uuid.UUID]*user),
				leave:                             make(chan *match_socket),

				gamer_permission_signup:  make(chan *permission_socket),
				gamer_permission_list:    make(map[uuid.UUID]*user),
				gamer_permission_signout: make(chan *permission_socket),

				bot_join:  make(chan *user),
				bot_leave: make(chan *user),

				spectator_join:                        make(chan *match_socket),
				spectator_uid_to_msid_to_match_socket: make(map[uuid.UUID]map[uuid.UUID]*match_socket),
				spectator_uid_to_user:                 make(map[uuid.UUID]*user),

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
		select {

		// only add the user, do not add the mmsocket
		case ws := <-m.gamer_permission_signup: // joining the waitroom for matchmaking
			_, check_uid := m.gamer_permission_list[ws.u.uid] // check if the user is in the match object
			_, check_mid := ws.u.mid_to_match[m.mid]          // check if the match is to the user object
			fmt.Println("MID JOIN:", m.mid)
			fmt.Println("UID JOIN:", ws.u.uid)

			// if the user is not in the match already and we are at capacity
			if (check_uid == false) && (len(m.gamer_permission_list) > int(m.capacity)) {
				fmt.Println("the match is full")
				msg := &message{Name: ws.u.email, Message: "the match is full", Event: "failedMatchJoin", When: time.Now(), MatchID: m.mid}
				ws.incoming_message <- msg
			} else { // add the match socket to the room

				// if user is not in the match object, add the user and create a socket object
				m.mutex.Lock()
				if check_uid == false {
					m.gamer_permission_list[ws.u.uid] = ws.u
				}
				m.mutex.Unlock()

				// if match is not in the user object, add the match and create a match_socket object
				ws.u.mutex.Lock()
				if check_mid == false {
					ws.u.mid_to_match[m.mid] = m
				}
				ws.u.mutex.Unlock()

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

				fmt.Println("a mmsocket has joined the match")
				printAllMatchUserWS()
			}

		case ws := <-m.gamer_permission_signout: // leaving the waitroom for matchmaking

			// cannot leave the permission list, the match has already started
			if m.started == true {

			} else {
				_, check_uid := m.gamer_permission_list[ws.u.uid] // check if the user is in the match object
				_, check_mid := ws.u.mid_to_match[m.mid]          // check if the match is to the user object
				fmt.Println("MID LEAVE:", m.mid)
				fmt.Println("UID LEAVE:", ws.u.uid)

				// if user is in the match object, delete all sockets as well
				m.mutex.Lock()
				if check_uid == true {
					delete(m.gamer_permission_list, ws.u.uid)
					delete(m.gamer_uid_to_msid_to_match_socket, ws.u.uid)
				}
				m.mutex.Unlock()

				// if match is not in the user object, add the match and create a match_socket object
				ws.u.mutex.Lock()
				if check_mid == true {
					delete(ws.u.mid_to_match, m.mid)
					delete(ws.u.mid_to_msid_to_match_socket, m.mid)
				}
				ws.u.mutex.Unlock()

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

				fmt.Println("a mmsocket has left the match")
				printAllMatchUserWS()
			}

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

			// if this is the first socket the user has joined this room, send a message to everyone that he's joined
			if check_uid == false {
				msg := &message{Name: ws.u.email, Message: "participantJoinSuccess", Event: "participantJoinSuccess", When: time.Now(), MatchID: m.mid}
				go func() {
					msid_to_sock.mutex.RLock()
					for _, j := range msid_to_sock.msid_to_sock {
						j.incoming_message <- msg
					}
					msid_to_sock.mutex.RUnlock()
				}()
			}

			fmt.Println("a socket has joined the match")
			printAllMatchUserWS()

		case u := <-m.bot_join:

			m.mutex.Lock()
			m.gamer_permission_list[u.uid] = u
			m.gamer_uid_to_user[u.uid] = u
			//m.gamer_uid_to_msid_to_match_socket[u.uid] = make(map[uuid.UUID]*match_socket)
			m.mutex.Unlock()

			msg := &message{Name: u.email, Message: "participantJoinSuccess", Event: "participantJoinSuccess", When: time.Now(), MatchID: m.mid}
			go func() {
				pid_to_permissions.mutex.RLock()
				for _, j := range pid_to_permissions.global {
					j.incoming_message <- msg
				}
				pid_to_permissions.mutex.RUnlock()
			}()

			fmt.Println("added a bot")

		case u := <-m.bot_leave:

			m.mutex.Lock()
			delete(m.gamer_permission_list, u.uid)
			delete(m.gamer_uid_to_user, u.uid)
			//delete(m.gamer_uid_to_msid_to_match_socket, u.uid)
			m.mutex.Unlock()

			fmt.Println("removed a bot")

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
		fmt.Println("MD:", i)
		for k, _ := range j.gamer_uid_to_user {
			fmt.Println("UID:", k)
		}
	}
}
