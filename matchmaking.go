package main

import (
	"fmt"
	"github.com/google/uuid"
	"sync"
	"time"
)

func ms_write(ms *match_socket) {
	for msg := range ms.incoming_message { // special trait of a channel, will block until something is in the channel or it is closed
		if err := ms.socket.WriteJSON(msg); err != nil {
			fmt.Println("error writing to socket")
			break
		} else {
			fmt.Println("successful write")
		}
	}
}

func ms_read(ms *match_socket) {
	defer func() {
		fmt.Printf("read socket closing: %+v\n", ms)
		if ms.open && (ms.m != nil) {
			ms.m.leave <- ms
			fmt.Println("client left the room", ms)
		}
	}()

	for {
		var msg *message
		if err := ms.socket.ReadJSON(&msg); err == nil {
			fmt.Printf("JSON: %+v\n", msg)

			msg.When = time.Now()
			msg.Name = ms.u.email
			fmt.Println(msg.When, "message~", msg.Name, ", Event: ", msg.Event, ", Message: ", msg.Message)

			if msg.Event == "ping" {
				fmt.Println("ping!")
			} else if msg.Event == "createMatch" {
				num := createMatch(ms, msg)
				mid_to_match.match[num].join <- ms
			}
		} else {
			fmt.Println("error reading from socket")
			break
		}
	}
}

func createMatch(ms *match_socket, msg *message) uuid.UUID {
	fmt.Println("called 'createMatch'")

	var random_number uuid.UUID
	// create a global room for users to chat in
	not_assigned := true
	for not_assigned {
		random_number = uuid.New()

		mid_to_match.mutex.RLock()
		_, found := mid_to_match.match[random_number]
		mid_to_match.mutex.RUnlock()
		if !found {

			temp := &match{

				mutex: sync.RWMutex{},

				mid: random_number,

				broadcast: make(chan *message),
				join:      make(chan *match_socket),
				leave:     make(chan *match_socket),

				uid_to_sid_to_match_socket: make(map[uuid.UUID]map[uuid.UUID]*match_socket),
				uid_to_user:                make(map[uuid.UUID]*user),

				message_logs: make([]*message, 0, 16),

				open: make(chan bool),
			}

			mid_to_match.mutex.Lock()
			mid_to_match.match[random_number] = temp
			mid_to_match.mutex.Unlock()
			not_assigned = false
			fmt.Println("created MID:", random_number)
			go mid_to_match.match[random_number].run()

		}
	}
	return random_number
}

func (m *match) run() {
	for {
		select {
		case ws := <-m.join: // joining

			// add ws to match object
			// check if user is already in the match
			rid_to_room.mutex.RLock()
			m.mutex.RLock()
			_, check := m.uid_to_user[ws.u.uid]
			m.mutex.RUnlock()
			rid_to_room.mutex.RUnlock()

			// if the user is not in the match, add the user as well as the web socket
			rid_to_room.mutex.Lock()
			m.mutex.Lock()
			if check == false {
				m.uid_to_user[ws.u.uid] = ws.u
				m.uid_to_sid_to_match_socket[ws.u.uid] = make(map[uuid.UUID]*match_socket)
			}
			m.uid_to_sid_to_match_socket[ws.u.uid][ws.msid] = ws
			rid_to_room.mutex.Unlock()
			m.mutex.Unlock()

			// add ws to user object
			uid_to_user.mutex.RLock()
			ws.u.mutex.RLock()
			_, check = ws.u.mid_to_match[m.mid]
			uid_to_user.mutex.RUnlock()
			ws.u.mutex.RUnlock()

			uid_to_user.mutex.Lock()
			ws.u.mutex.Lock()
			if check == false {
				fmt.Println(ws.u.mid_to_match)
				ws.u.mid_to_match[m.mid] = m
				ws.u.mid_to_msid_to_match_socket[m.mid] = make(map[uuid.UUID]*match_socket)
			}
			ws.u.mid_to_msid_to_match_socket[m.mid][ws.msid] = ws
			ws.u.mutex.Unlock()
			uid_to_user.mutex.Unlock()

			// if this is the first socket the user has opened for this room, send a join message
			if check == false {
				msg := &message{Name: ws.u.email, Message: "x entered the chat", Event: "joinedMatch", When: time.Now()}
				ws.m.broadcast <- msg
			}

			fmt.Println("a socket has joined the match")
		case msg := <-m.broadcast: // forward message to all clients

			m.mutex.Lock()
			m.message_logs = append(m.message_logs, msg)
			m.mutex.Unlock()

			fmt.Println(msg)

			for _, i := range m.uid_to_sid_to_match_socket {
				for _, j := range i {
					select {
					case j.incoming_message <- msg:
					}
				}
			}
		}
	}
}
