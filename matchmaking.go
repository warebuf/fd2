package main

import (
	"fmt"
	"github.com/google/uuid"
	"strconv"
	"sync"
	"time"
)

func mm_write(mm *mmsocket) {
	for msg := range mm.incoming_message { // special trait of a channel, will block until something is in the channel or it is closed
		if err := mm.socket.WriteJSON(msg); err != nil {
			fmt.Println("error writing to socket")
			break
		} else {
			fmt.Println("successful write")
		}
	}
}

func mm_read(mm *mmsocket) {
	defer func() {
		fmt.Printf("read socket closing: %+v\n", mm)
		if mm.open {
			fmt.Println("client left the room", mm)
		}
	}()

	for {
		var msg *message
		if err := mm.socket.ReadJSON(&msg); err == nil {
			fmt.Printf("JSON: %+v\n", msg)

			msg.When = time.Now()
			msg.Name = mm.u.email
			fmt.Println(msg.When, "message~", msg.Name, ", Event: ", msg.Event, ", Message: ", msg.Message)

			// user requests to create match
			if msg.Event == "createMatch" {

				fmt.Println("received createMatch")

				mtch := createMatch(mm, msg)
				go mtch.run()
				globalBroadcast(&message{Event: "newMatch", Message: mtch.game_mode + strconv.Itoa(int(mtch.capacity)), When: time.Now(), MatchID: mtch.mid}) // when a room is created, send it to all sockets (not just sockets in room)
				mtch.participant_signup <- mm

			} else if msg.Event == "participantJoin" { // user request to join match

				fmt.Println("received participantJoin")

				mid_to_match.mutex.Lock()
				mtch, exists := mid_to_match.match[msg.MatchID]
				mid_to_match.mutex.Unlock()

				if exists {
					_, check_uid := mtch.participant_uid_to_user[mm.u.uid]
					if check_uid == true {
						fmt.Println("participant ws is already is the match!")
					} else {
						mtch.participant_signup <- mm
					}
				}
			} else if msg.Event == "participantLeave" {
				mid_to_match.mutex.Lock()
				mtch, exists := mid_to_match.match[msg.MatchID]
				mid_to_match.mutex.Unlock()

				if exists {
					mtch.participant_signout <- mm
				}
			}

		} else {
			fmt.Println("error reading from socket")
			break
		}
	}
}

func createMatch(mm *mmsocket, msg *message) *match {
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

				if (num_ops < 100) && (num_ops > 0) {
					num = uint(num_ops)
				} else {
					num = 1
				}
			}

			ans = &match{

				mutex: sync.RWMutex{},

				mid:       random_number,
				game_mode: gm,
				capacity:  num,

				broadcast:        make(chan *message),
				participant_join: make(chan *match_socket),
				spectator_join:   make(chan *match_socket),
				leave:            make(chan *match_socket),

				participant_signup:  make(chan *mmsocket),
				participant_signout: make(chan *mmsocket),

				participant_uid_to_msid_to_match_socket: make(map[uuid.UUID]map[uuid.UUID]*match_socket),
				participant_uid_to_user:                 make(map[uuid.UUID]*user),

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
		case ws := <-m.participant_signup:
			_, check_uid := m.participant_uid_to_user[ws.u.uid] // check if the user is in the match object
			_, check_mid := ws.u.mid_to_match[m.mid]            // check if the match is to the user object
			fmt.Println("MID JOIN:", m.mid)
			fmt.Println("UID JOIN:", ws.u.uid)

			// if the user is not in the match already and we are at capacity
			if (check_uid == false) && (len(m.participant_uid_to_user) > int(m.capacity)) {
				fmt.Println("the match is full")
				msg := &message{Name: ws.u.email, Message: "the match is full", Event: "failedMatchJoin", When: time.Now(), MatchID: m.mid}
				ws.incoming_message <- msg
			} else { // add the match socket to the room

				// if user is not in the match object, add the user and create a socket object
				m.mutex.Lock()
				if check_uid == false {
					m.participant_uid_to_user[ws.u.uid] = ws.u
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
						mmid_to_matchmaking.mutex.RLock()
						for _, j := range mmid_to_matchmaking.matchmaking {
							j.incoming_message <- msg
						}
						mmid_to_matchmaking.mutex.RUnlock()
					}()
				}

				fmt.Println("a mmsocket has joined the match")
				printAllMatchUserWS()
			}

		case ws := <-m.participant_signout:

			_, check_uid := m.participant_uid_to_user[ws.u.uid] // check if the user is in the match object
			_, check_mid := ws.u.mid_to_match[m.mid]            // check if the match is to the user object
			fmt.Println("MID LEAVE:", m.mid)
			fmt.Println("UID LEAVE:", ws.u.uid)

			// if user is in the match object, delete all sockets as well
			m.mutex.Lock()
			if check_uid == true {
				delete(m.participant_uid_to_user, ws.u.uid)
				delete(m.participant_uid_to_msid_to_match_socket, ws.u.uid)
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
					mmid_to_matchmaking.mutex.RLock()
					for _, j := range mmid_to_matchmaking.matchmaking {
						j.incoming_message <- msg
					}
					mmid_to_matchmaking.mutex.RUnlock()
				}()
			}

			fmt.Println("a mmsocket has left the match")
			printAllMatchUserWS()

		case ws := <-m.participant_join: // joining

			_, check_uid := m.participant_uid_to_user[ws.u.uid] // check if the user is in the match object
			_, check_mid := ws.u.mid_to_match[m.mid]            // check if the match is to the user object
			fmt.Println("MID JOIN:", m.mid)
			fmt.Println("UID JOIN:", ws.u.uid)

			// if the user is not in the match already and we are at capacity
			if (check_uid == false) && (len(m.participant_uid_to_user) > int(m.capacity)) {
				fmt.Println("the match is full")
				msg := &message{Name: ws.u.email, Message: "the match is full", Event: "failedMatchJoin", When: time.Now(), MatchID: m.mid}
				ws.incoming_message <- msg
			} else { // add the match socket to the room

				// if user is not in the match object, add the user and create a socket object
				m.mutex.Lock()
				if check_uid == false {
					m.participant_uid_to_user[ws.u.uid] = ws.u
					m.participant_uid_to_msid_to_match_socket[ws.u.uid] = make(map[uuid.UUID]*match_socket)
				}
				m.participant_uid_to_msid_to_match_socket[ws.u.uid][ws.msid] = ws
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
			}

		case msg := <-m.broadcast: // forward message to all clients

			m.mutex.Lock()
			m.message_logs = append(m.message_logs, msg)
			m.mutex.Unlock()

			fmt.Println("sending:", msg)

			for _, i := range m.participant_uid_to_msid_to_match_socket {
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
	mmid_to_matchmaking.mutex.RLock()
	for _, i := range mmid_to_matchmaking.matchmaking {
		select {
		case i.incoming_message <- msg:
		}
	}

	mmid_to_matchmaking.mutex.RUnlock()
}

func printAllMatchUserWS() {
	fmt.Println("printing everything")

	for i, j := range mid_to_match.match {
		fmt.Println("MD:", i)
		for k, _ := range j.participant_uid_to_user {
			fmt.Println("UID:", k)
		}
	}
}
