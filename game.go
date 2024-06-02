package main

import (
	"fmt"
	"strconv"
	"time"
)

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
			fmt.Println("client left the room", m)
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
			if msg.Event == "createMatch" {
			} else if msg.Event == "taptap" {
				fmt.Println("got a tap tap")
			} else if msg.Event == "clockSyncResponse" {
				test, _ := strconv.ParseInt(msg.Message, 10, 64)
				m.user_time = time.UnixMilli(test)
				fmt.Println("dif", m.system_time.Sub(m.user_time))

				if len(m.m.gamer_uid_to_user) == int(m.m.capacity) {
					allset := true
					for i, j := range m.m.gamer_uid_to_msid_to_match_socket {
						if uid_to_user.users[i].bot_status == true {

						} else {
							for _, l := range j {
								fmt.Println(l.msid)
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
					if allset == true {
						m.m.start_ticker <- true
					}
				}
			}

		} else {
			fmt.Println("error reading from socket")
			break
		}
	}
}
