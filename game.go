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
			} else if msg.Event == "act" {
				fmt.Println("received act")

				// check if the match has started and the action is to a user's own bot
				if m.m.started == true && m.m.ended == false && m.m.uuid_to_team_int[m.u.uid].a == msg.Team_index && m.m.uuid_to_team_int[m.u.uid].b == msg.Client_index {
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

				// check if room is full and game hasn't started, if so, start the countdown start timer
				if m.m.started == false && len(m.m.gamer_uid_to_user) == int(m.m.capacity) {
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
						m.m.started = true
						m.m.start_ticker <- true
					}
				} else if m.m.started == true {
					m.m.sharepos(nil)
				}
			}

		} else {
			fmt.Println("error reading from socket")
			break
		}
	}
}
