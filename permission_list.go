package main

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"strconv"
	"sync"
	"time"
)

type permission_list struct {
	mutex sync.RWMutex

	plid      uuid.UUID
	game_mode string
	sides     uint
	capacity  uint
	started   bool

	gamer_permission_signup  chan *permission_socket
	gamer_permission_list    map[uuid.UUID]*user
	gamer_permission_signout chan *permission_socket

	bot chan *sync.WaitGroup

	ended chan bool

	spectator_join                        chan *match_socket // a channel for clients wishing to join
	spectator_uid_to_msid_to_match_socket map[uuid.UUID]map[uuid.UUID]*match_socket
	spectator_uid_to_user                 map[uuid.UUID]*user
}

type permission_socket struct {
	socket *websocket.Conn

	pid uuid.UUID

	u *user
	//pl *permission_list

	incoming_message chan *pmessage // send is a channel on which messages are sent.

	open bool
}

type pmessage struct {
	Name    string
	Message string
	When    time.Time
	Event   string
	PLID    uuid.UUID
}

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
		fmt.Println("read psocket closing")

		pid_to_permissions.mutex.Lock()
		delete(pid_to_permissions.global, p.pid)
		pid_to_permissions.mutex.Unlock()

		if p.open {
			fmt.Println("client left waitroom (psocket)")
		}
	}()

	for {
		var msg *pmessage
		if err := p.socket.ReadJSON(&msg); err == nil {
			fmt.Printf("JSON: %+v\n", msg)

			msg.When = time.Now()
			msg.Name = p.u.email
			fmt.Println(msg.When, "pmessage~", msg.Name, ", Event: ", msg.Event, ", Message: ", msg.Message)

			// user requests to create match
			if msg.Event == "createPermissionList" {

				fmt.Println("received createPermissionList")

				perm_list := createPL(msg)
				go perm_list.run()
				psocketBroadcast(&pmessage{Event: "newPL", Message: perm_list.game_mode + strconv.Itoa(int(perm_list.capacity)), When: time.Now(), PLID: perm_list.plid}) // let everyone know there is a new room
				perm_list.gamer_permission_signup <- p

			} else if msg.Event == "participantJoin" { // user request to join match

				fmt.Println("received participantJoin")

				plid_to_permission_list.mutex.Lock()
				pl, exists := plid_to_permission_list.global[msg.PLID]
				plid_to_permission_list.mutex.Unlock()

				if exists == true {
					_, check_plid := pl.gamer_permission_list[p.u.uid]
					if check_plid == true {
						fmt.Println("participant ws is already is the match!")
					} else if pl.started == true {
						fmt.Println("match has already started, of course you can't join")
					} else if len(pl.gamer_permission_list) > int(pl.capacity) {
						fmt.Println("the match is full")
						p.incoming_message <- &pmessage{Name: p.u.email, Message: "the match is full", Event: "failedMatchJoin", When: time.Now(), PLID: pl.plid}
					} else {
						pl.gamer_permission_signup <- p
					}
				}

			} else if msg.Event == "participantLeave" {

				fmt.Println("received participantLeave")

				plid_to_permission_list.mutex.Lock()
				pl, exists := plid_to_permission_list.global[msg.PLID]
				plid_to_permission_list.mutex.Unlock()

				if exists {
					_, check_uid := pl.gamer_permission_list[p.u.uid]
					if check_uid == true {
						if pl.started == false {
							pl.gamer_permission_signout <- p
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

func createPL(msg *pmessage) *permission_list {
	fmt.Println("called 'createMatch'")

	var ans *permission_list

	not_assigned := true
	for not_assigned {
		random_number := uuid.New()

		plid_to_permission_list.mutex.RLock()
		_, found := plid_to_permission_list.global[random_number]
		plid_to_permission_list.mutex.RUnlock()
		if !found {

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

			ans = &permission_list{

				mutex: sync.RWMutex{},

				plid:      random_number,
				game_mode: gm,
				capacity:  num,
				sides:     sd,
				started:   false,

				gamer_permission_signup:  make(chan *permission_socket),
				gamer_permission_list:    make(map[uuid.UUID]*user),
				gamer_permission_signout: make(chan *permission_socket),

				bot: make(chan *sync.WaitGroup),

				ended: make(chan bool),

				spectator_join:                        make(chan *match_socket),
				spectator_uid_to_msid_to_match_socket: make(map[uuid.UUID]map[uuid.UUID]*match_socket),
				spectator_uid_to_user:                 make(map[uuid.UUID]*user),
			}

			fmt.Println("created PLID:", random_number)

			plid_to_permission_list.mutex.Lock()
			plid_to_permission_list.global[random_number] = ans
			plid_to_permission_list.mutex.Unlock()

			not_assigned = false
		}

	}

	return ans
}

func (pl *permission_list) run() {
	for {
		select {

		case ws := <-pl.gamer_permission_signup: // joining the PL
			fmt.Println("gamer_permission_signup", ws.u.uid)

			// if user is not in the PL object, add the user object
			// TODO: BUG, same user has different UIDs!!!, therefore the same user on different browsers can join twice
			_, check_uid := pl.gamer_permission_list[ws.u.uid]
			pl.mutex.Lock()
			if check_uid == false {
				pl.gamer_permission_list[ws.u.uid] = ws.u
			}
			pl.mutex.Unlock()

			// if this is the first join, send a message to everyone that he's joined
			if check_uid == false {
				msg := &pmessage{Name: ws.u.email, Event: "participantJoinSuccess", When: time.Now(), PLID: pl.plid}
				pid_to_permissions.mutex.RLock()
				for _, j := range pid_to_permissions.global {
					j.incoming_message <- msg
				}
				pid_to_permissions.mutex.RUnlock()
			}

			printAllMatchUserWS()
			continue

		case ws := <-pl.gamer_permission_signout: // leaving the waitroom for matchmaking

			fmt.Println("gamer_permission_signout", ws.u.uid)

			// if user is in the match object, delete all sockets as well
			_, check_uid := pl.gamer_permission_list[ws.u.uid] // check if the user is in the match object
			pl.mutex.Lock()
			if check_uid == true {
				delete(pl.gamer_permission_list, ws.u.uid)
			}
			pl.mutex.Unlock()

			// send a message to everyone that he's left
			if check_uid == true {
				msg := &pmessage{Name: ws.u.email, Message: "participantLeaveSuccess", Event: "participantLeaveSuccess", When: time.Now(), PLID: pl.plid}
				pid_to_permissions.mutex.RLock()
				for _, j := range pid_to_permissions.global {
					j.incoming_message <- msg
				}
				pid_to_permissions.mutex.RUnlock()

			}
			printAllMatchUserWS()
			continue

		case wg := <-pl.bot:
			not_assigned := true
			var random_number uuid.UUID
			for not_assigned {
				random_number = uuid.New()
				if _, found := uid_to_user.users[random_number]; !found {
					not_assigned = false
					fmt.Println("assigning bot UID:", random_number)

					user_object := &user{
						mutex: sync.RWMutex{},

						uid:          random_number,
						email:        "bot@bot.com",
						nickname:     "BOT",
						currency:     0,
						admin_status: false,
						bot_status:   true,

						rid_to_sid_to_socket: nil,
						rid_to_room:          nil,

						mid_to_msid_to_match_socket: nil,
						mid_to_match:                nil,
					}
					pl.gamer_permission_list[random_number] = user_object

					msg := &pmessage{Name: user_object.email, Message: "participantJoinSuccess", Event: "participantJoinSuccess", When: time.Now(), PLID: pl.plid}
					pid_to_permissions.mutex.RLock()
					for _, j := range pid_to_permissions.global {
						j.incoming_message <- msg
					}
					pid_to_permissions.mutex.RUnlock()
				}
			}
			wg.Done()

		case <-pl.ended:

			fmt.Println("<-pl.ended")
			delete(plid_to_permission_list.global, pl.plid)
			psocketBroadcast(&pmessage{Event: "removeMatch", Message: pl.plid.String()})
		}
	}
}

func psocketBroadcast(msg *pmessage) {
	fmt.Println("global broadcast")
	pid_to_permissions.mutex.RLock()
	for _, i := range pid_to_permissions.global {
		select {
		case i.incoming_message <- msg:
		}
	}

	pid_to_permissions.mutex.RUnlock()
}
