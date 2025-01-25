package main

import (
	"container/heap"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"math"
	"math/rand"
	"reflect"
	"strconv"
	"sync"
	"time"
)

import "github.com/warebuf/bezier-lib/bezier"

type Camp struct {
	HP    int
	XY    Coordinate
	Units [][]int // second array is of the form i, j, k for a *hero object in TCH

}

type cumulativeStats struct {
	DistanceTravelled float64
	DamageTaken       int
	DamageGiven       int
	AttacksMade       int
	LocationsReached  int
	CampsDestroyed    int
}

type hero struct {
	Bot  bool
	Name string

	Start       Coordinate
	Position    Coordinate
	Destination Coordinate
	Path        [][]float64

	DistanceLeft float64
	Angle        float64
	Status       int // 0 = requires destination, 1 = in route, 2 = attacking something

	H part
	L part
	R part
	B part

	HP  int
	ATK int
	AS  int
	SPD int

	EXP int

	Battle *Camp
	TTD    float64
}

type part struct {
	SECTION int // 0 is H, 1 is L, 2 is R, 3 is B
	SERIAL  int
	NAME    string

	HP  int
	DEF int
	DOG int

	ATK int
	ACC int
	CRT int

	CD     int
	CLU    int
	Weight int

	Use_current int
	Use_outof   int

	SPD int
}

type attack struct {
	Attacker []int      //i,j,k
	Defender [][]int    //list of i,j,k
	Damage   [][]string // list of damage strings - 0: H, 1: DMG, 2: H%, 3: L%, 4: R%, 5: B%, 6: RNG
	Type     string     //miss,
}

type Coordinate struct {
	X float64
	Y float64

	Type string // only types rn: camp, unit
	I    int
	J    int
	K    int
}

type match struct {
	mutex sync.RWMutex

	mid       uuid.UUID
	game_mode string
	sides     uint
	capacity  uint
	ended     bool

	// each client is represented as an int
	state_count      int
	team_client_hero [][][]*hero
	camps            []*Camp

	uuid_to_team_int map[uuid.UUID]pair
	turn             string // tells you turn #

	//bench map[string][]*part // team, user index to part

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

type message struct {
	Name    string
	Message string
	When    time.Time
	Event   string
	MatchID uuid.UUID

	Phase string

	// game state stuff
	TCH         [][][]string
	Turn        string
	Atk         []string
	UnitsOfTime string
	Camps       []string

	Route Coordinate
	Path  [][]float64

	// should add the user's TCH
	Team_index   int
	Client_index int
	Hero_index   int
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
			fmt.Println("Event:", msg.Event, ", Message: ", msg.Message)

			// user requests to create match
			if m.m.ended == true {
				continue
			} else if msg.Event == "act" {

				// check if the match has started and the action is to a user's own hero
				if m.m.phase == "TURN" && m.m.ended == false && m.m.uuid_to_team_int[m.u.uid].a == msg.Team_index && m.m.uuid_to_team_int[m.u.uid].b == msg.Client_index {
					// check if the hero is alive and at starting position
					if (m.m.team_client_hero[msg.Team_index][msg.Client_index][msg.Hero_index].HP > 0) &&
						((m.m.team_client_hero[msg.Team_index][msg.Client_index][msg.Hero_index].Status == 0) ||
							(sameCoordinates(m.m.team_client_hero[msg.Team_index][msg.Client_index][msg.Hero_index].Position, m.m.team_client_hero[msg.Team_index][msg.Client_index][msg.Hero_index].Start) &&
								(m.m.team_client_hero[msg.Team_index][msg.Client_index][msg.Hero_index].Status == 1))) {

						if (msg.Route.I < len(m.m.camps)) && (msg.Route.I >= 0) { // possibly check if camp is still up here

							if confirmValidPath(m.m.team_client_hero, m.m.camps, msg.Team_index, msg.Client_index, msg.Hero_index, msg.Path, msg.Route) {

								fmt.Println("msg.path, msg.route", msg.Path, msg.Route)
								fmt.Println("Type of myInt:", reflect.TypeOf(msg.Path))
								setPath(m.m.team_client_hero, msg.Team_index, msg.Client_index, msg.Hero_index, msg.Route, msg.Path)
								m.m.sharepos(nil, 0)

								// check if all heroes have commands (ignore bots). If all heroes have commands, calculate next position
								sim_check := true
								for i := 0; i < len(m.m.team_client_hero); i++ {
									for j := 0; j < len(m.m.team_client_hero[i]); j++ {
										for k := 0; k < len(m.m.team_client_hero[i][j]); k++ {
											if (m.m.team_client_hero[i][j][k].Bot == false) &&
												(m.m.team_client_hero[i][j][k].Status == 0) &&
												(m.m.team_client_hero[i][j][k].HP > 0) {
												sim_check = false
												break
											}
										}
									}
								}
								if sim_check == true {
									fmt.Println("simulate")
									m.m.simulate <- true
								}
							}

						} else {
							fmt.Printf("this route does not exist")
						}

					}
				}

			} else if msg.Event == "clockSyncResponse" {
				fmt.Println("received clockSyncResponse")
				test, _ := strconv.ParseInt(msg.Message, 10, 64)
				m.user_time = time.UnixMilli(test)
				fmt.Println("dif", m.system_time.Sub(m.user_time))

				msg := &message{Event: "ticker_start", When: time.Now(), MatchID: m.m.mid}
				msg.Message = m.m.next_time.Add(m.user_time.Sub(m.system_time)).UTC().String()
				m.incoming_message <- msg
				msg2 := &message{Event: "update_phase", Phase: m.m.phase, MatchID: m.m.mid}
				m.incoming_message <- msg2

				if m.m.phase == "TURN" {
					m.m.sharepos(nil, 0)
				} else if m.m.phase == "CHARACTER SELECTION" {
					m.m.sharebench()
					m.m.sharepos(nil, 0)
				}
				m.incoming_message <- &message{Event: "assignInt", Message: m.m.uuid_to_team_int[m.u.uid].ab, MatchID: m.m.mid}

			} else if msg.Event == "endCharSel" {
				fmt.Println("recieved endCharSel")

				if m.m.char_sel_done[m.u.uid] == false {
					m.m.char_sel_done[m.u.uid] = true
					msg := &message{Event: "sel_finished", Name: m.u.email, Message: m.m.uuid_to_team_int[m.u.uid].ab, When: time.Now(), MatchID: m.m.mid}
					m.mutex.Lock()
					m.m.message_logs = append(m.m.message_logs, msg)
					m.mutex.Unlock()
					for _, i := range m.m.gamer_uid_to_msid_to_match_socket {
						for _, j := range i {
							select {
							case j.incoming_message <- msg:
							}
						}
					}

					for _, i := range m.m.spectator_uid_to_msid_to_match_socket {
						for _, j := range i {
							select {
							case j.incoming_message <- msg:
							}
						}
					}
				}

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
			} else if msg.Event == "resign" {

				m.m.ended = true
				m.m.ticker.Stop()

				msg := &message{Event: "game_over"}

				for _, i := range m.m.gamer_uid_to_msid_to_match_socket {
					for _, j := range i {
						select {
						case j.incoming_message <- msg:
						}
					}
				}

				for _, i := range m.m.spectator_uid_to_msid_to_match_socket {
					for _, j := range i {
						select {
						case j.incoming_message <- msg:
						}
					}
				}
			} else {
				fmt.Println("cannot swap incompatible body parts")
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
	_, found := mid_to_match.match[pl.plid] // USE THE SAME PLID FOR MID
	mid_to_match.mutex.RUnlock()
	if !found {

		ans = &match{

			mutex: sync.RWMutex{},

			mid:       pl.plid,
			game_mode: pl.game_mode,
			capacity:  pl.capacity,
			sides:     pl.sides,
			ended:     false,

			state_count:      0,
			team_client_hero: make([][][]*hero, 0, 0),
			uuid_to_team_int: make(map[uuid.UUID]pair),

			camps: make([]*Camp, 0, 0),

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

			//bench: make(map[string][]*part),

			char_sel_done:   make(map[uuid.UUID]bool),
			char_sel_ticker: make(chan bool),

			message_logs: make([]*message, 0, 16),

			simulate: make(chan bool),
		}

		// add bots to gamer_uid_to_user
		for _, j := range pl.spectator_uid_to_user {
			if j.bot_status == true {
				ans.bot_join <- j
			}
		}

		// creating camps
		random_count := 5 + rand.Intn(5)
		for i := 0; i < random_count; i++ {
			random_x := float64(rand.Intn(40))
			random_y := float64(rand.Intn(40))
			c := Coordinate{X: 50.0 + (random_x * 10.0), Y: 50.0 + (random_y * 10.0), Type: "camp", I: i, J: -1, K: -1}
			ans.camps = append(ans.camps, &Camp{HP: rand.Intn(10)*100 + 100, XY: c, Units: make([][]int, 0)})
		}

		// set up the game state by: (1) creating each team, (2) assigning clients to each team, (3) creating the basic hero unit for each client
		// (1) creating each team
		for i := 0; i < int(ans.sides); i++ {
			ans.team_client_hero = append(ans.team_client_hero, make([][]*hero, 0))
		}

		// (2) and (3) assigning clients and heroes to each team
		team_int := 0
		client_int := 0

		for i, j := range pl.gamer_permission_list {

			if j.bot_status == false {
				ans.char_sel_done[i] = false // setting up a map of user to a boolean to check/track the state of all clients
			} else {
				ans.char_sel_done[i] = true
			}

			ans.uuid_to_team_int[i] = pair{team_int, client_int, strconv.Itoa(team_int) + ";" + strconv.Itoa(client_int)} // assigning a pair of ints (team, pos in team) to each client
			ans.team_client_hero[team_int] = append(ans.team_client_hero[team_int], make([]*hero, 0, 5))

			names := []string{
				"EYE OF GOD", "BLUE EYES W.", "PIKACHU", "MEGAN FOX", "MUMEN RIDER",
				"RAINMAN", "CARMACK", "CAGED", "JINZO", "LOADED GUN CONJECT.",
				"TERRACOTTA", "DEATH BEAM", "ONE EDGE BLADE", "DEATHBLADE", "GUINSOO",
				"SOUND OF THE RAIN", "CAT TAIL", "HOG WILD", "YEAR ZERO",
			}

			for y := 0; y < 3; y++ {
				temp_coordinate := Coordinate{X: float64(team_int * 500), Y: 500 - (float64(y) * 100), Type: "unit", I: team_int, J: client_int, K: y}
				temp := &hero{
					Bot:  j.bot_status,
					Name: names[rand.Intn(len(names))],

					Start:       temp_coordinate,
					Position:    temp_coordinate,
					Destination: temp_coordinate,
					Path:        make([][]float64, 0),

					DistanceLeft: getDistanceTwoPoints(temp_coordinate, temp_coordinate),
					Angle:        getTheta(temp_coordinate, temp_coordinate),

					Status: 0,

					HP:  rand.Intn(100) + 1,
					ATK: rand.Intn(100) + 1,
					AS:  rand.Intn(100) + 1,
					SPD: rand.Intn(100) + 1,

					Battle: nil,
					TTD:    100.0,
				}
				ans.team_client_hero[team_int][client_int] = append(ans.team_client_hero[team_int][client_int], temp)
			}

			//ans.bench[ans.uuid_to_team_int[i].ab] = make([]*part, 0)

			/*
				for y := 0; y < 5; y++ {
						pick := rand.Intn(4)
						if pick == 0 {
							temp := rand.Intn(len(namesh))
							p := &part{
								SECTION:     0,
								SERIAL:      temp,
								NAME:        namesh[temp],
								HP:          rand.Intn(99) + 1,
								DOG:         rand.Intn(99) + 1,
								ATK:         rand.Intn(99) + 1,
								DEF:         rand.Intn(99) + 1,
								ACC:         rand.Intn(99) + 1,
								CRT:         rand.Intn(99) + 1,
								CD:          rand.Intn(99) + 1,
								CLU:         rand.Intn(99) + 1,
								Use_current: 0,
								Use_outof:   0,
								Weight:      rand.Intn(99) + 1,
							}
							ans.bench[ans.uuid_to_team_int[i].ab] = append(ans.bench[ans.uuid_to_team_int[i].ab], p)
						} else if pick == 1 {
							temp := rand.Intn(len(namesl))
							p := &part{
								SECTION:     1,
								SERIAL:      temp,
								NAME:        namesl[temp],
								HP:          rand.Intn(99) + 1,
								DOG:         rand.Intn(99) + 1,
								ATK:         rand.Intn(99) + 1,
								DEF:         rand.Intn(99) + 1,
								ACC:         rand.Intn(99) + 1,
								CRT:         rand.Intn(99) + 1,
								CD:          rand.Intn(99) + 1,
								CLU:         rand.Intn(99) + 1,
								Use_current: 0,
								Use_outof:   0,
								Weight:      rand.Intn(99) + 1,
							}
							ans.bench[ans.uuid_to_team_int[i].ab] = append(ans.bench[ans.uuid_to_team_int[i].ab], p)
						} else if pick == 2 {
							temp := rand.Intn(len(namesr))
							p := &part{
								SECTION:     2,
								SERIAL:      temp,
								NAME:        namesr[temp],
								HP:          rand.Intn(99) + 1,
								DOG:         rand.Intn(99) + 1,
								ATK:         rand.Intn(99) + 1,
								DEF:         rand.Intn(99) + 1,
								ACC:         rand.Intn(99) + 1,
								CRT:         rand.Intn(99) + 1,
								CD:          rand.Intn(99) + 1,
								CLU:         rand.Intn(99) + 1,
								Use_current: 0,
								Use_outof:   0,
								Weight:      rand.Intn(99) + 1,
							}
							ans.bench[ans.uuid_to_team_int[i].ab] = append(ans.bench[ans.uuid_to_team_int[i].ab], p)
						} else {
							temp := rand.Intn(len(namesb))
							p := &part{
								SECTION:     3,
								SERIAL:      temp,
								NAME:        namesb[temp],
								HP:          rand.Intn(99) + 1,
								ATK:         rand.Intn(99) + 1,
								DEF:         rand.Intn(99) + 1,
								ACC:         rand.Intn(99) + 1,
								CRT:         rand.Intn(99) + 1,
								CD:          rand.Intn(99) + 1,
								CLU:         rand.Intn(99) + 1,
								SPD:         rand.Intn(99) + 1,
								DOG:         rand.Intn(99) + 1,
								Use_current: 0,
								Use_outof:   0,
								Weight:      rand.Intn(99) + 1,
							}
							ans.bench[ans.uuid_to_team_int[i].ab] = append(ans.bench[ans.uuid_to_team_int[i].ab], p)
						}
					}
			*/

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

			// send the map to the new user
			fmt.Println("sent map")
			ws.incoming_message <- &message{Event: "map", Message: mapListJSON[0]}

			// send message history to new user
			for _, j := range m.message_logs {
				ws.incoming_message <- j
			}

			if m.ended == true {
				ws.incoming_message <- &message{Event: "game_over"}
			}

			// communicate to all participants - a new user has joined ( only unique users, not unique sockets )
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
				plid_to_permission_list.global[m.mid].ended <- true
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

			// send to all match sockets that the bot has entered the match
			msg := &message{Name: u.email, Event: "entered", When: time.Now()}
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

			continue

		case u := <-m.bot_leave:

			m.mutex.Lock()
			delete(m.gamer_uid_to_user, u.uid)
			//delete(m.gamer_uid_to_msid_to_match_socket, u.uid)
			m.mutex.Unlock()

			msg := &message{Name: u.email, Event: "left", When: time.Now()}
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
			fmt.Println("bot has left the game")
			fmt.Println("removed a bot")
			continue

		case <-m.ticker.C: // ticker goes off

			fmt.Println("m.Ticker.C")

			if m.phase == "CHARACTER SELECTION" {
				m.start_ticker <- true
			} else if m.phase == "TURN" {

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

			m.sharebench()
			m.sharepos(nil, 0)
			m.shareint()

		case <-m.start_ticker:

			fmt.Println("/startTicker")

			m.phase = "TURN"
			m.turn = "TURN 0"
			msg := &message{Event: "ticker_start", When: time.Now(), MatchID: m.mid}
			msg2 := &message{Event: "update_phase", Phase: m.phase, MatchID: m.mid}

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

			m.sharepos(nil, 0)

			continue

		case msg := <-m.broadcast: // TODO: I would get rid of this, replace it with a function roomBroadcast

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
						if (m.team_client_hero[i][j][k].Bot == true) && (m.team_client_hero[i][j][k].Status == 0) {
							fmt.Println("bot action", i, j, k)
							p, pint := determineBotAction(m.team_client_hero[i][j][k], m.camps)
							if len(p) > 0 {
								// figure out how to end the game here
							} else {
								setPath(m.team_client_hero, i, j, k, pint, p)
								gave_bots_acts = true
							}

						}
					}
				}
			}
			if gave_bots_acts == true {
				m.sharepos(nil, 0)
			}

			// calculate min unit of time to action
			min_units := 999999.99
			for i := 0; i < len(m.team_client_hero); i++ {
				for j := 0; j < len(m.team_client_hero[i]); j++ {
					for k := 0; k < len(m.team_client_hero[i][j]); k++ {
						if (m.team_client_hero[i][j][k].HP > 0) && (m.team_client_hero[i][j][k].Status == 1) && (m.team_client_hero[i][j][k].SPD != 0) {
							units_of_time := toFixed(m.team_client_hero[i][j][k].DistanceLeft/float64(m.team_client_hero[i][j][k].SPD), 3)
							if units_of_time < min_units {
								min_units = units_of_time
							}
						}
					}
				}
			}

			// calculate time until damage
			for i := 0; i < len(m.camps); i++ {
				if (m.camps[i].HP > 0) && (len(m.camps[i].Units) > 0) {
					for j := 0; j < len(m.camps[i].Units); j++ {
						unitOfInterest := m.team_client_hero[m.camps[i].Units[j][0]][m.camps[i].Units[j][1]][m.camps[i].Units[j][2]]
						if unitOfInterest.HP > 0 {
							ttd := unitOfInterest.TTD / float64(unitOfInterest.AS)
							if ttd < min_units {
								min_units = ttd
							}
						}
					}
				}
			}

			fmt.Println("min", min_units)

			has_user_cmd := false

			// calculate TTD for each unit, save the units that are going to attack next
			atk_list := make([]*attack, 0)
			for i := 0; i < len(m.camps); i++ {
				if m.camps[i].HP > 0 {
					for j := 0; j < len(m.camps[i].Units); j++ {
						unitOfInterest := m.team_client_hero[m.camps[i].Units[j][0]][m.camps[i].Units[j][1]][m.camps[i].Units[j][2]]
						if unitOfInterest.HP > 0 {
							unitOfInterest.TTD = unitOfInterest.TTD - (min_units * float64(unitOfInterest.AS))
							if unitOfInterest.TTD <= 0.1 {
								unitOfInterest.TTD = 0

								a := &attack{
									Attacker: make([]int, 3),
									Defender: make([][]int, 0),
									Damage:   make([][]string, 0),
									Type:     "hit",
								}
								a.Attacker[0] = unitOfInterest.Position.I
								a.Attacker[1] = unitOfInterest.Position.J
								a.Attacker[2] = unitOfInterest.Position.K
								atk_list = append(atk_list, a)
							}
						}
					}
				}
			}

			// move all unit positions
			for i := 0; i < len(m.team_client_hero); i++ {
				for j := 0; j < len(m.team_client_hero[i]); j++ {
					for k := 0; k < len(m.team_client_hero[i][j]); k++ {
						if (m.team_client_hero[i][j][k].HP > 0) && (m.team_client_hero[i][j][k].Status == 1) {

							distanceTravelled := min_units * float64(m.team_client_hero[i][j][k].SPD)
							newDistanceLeft := m.team_client_hero[i][j][k].DistanceLeft - distanceTravelled

							if newDistanceLeft <= 0.1 {
								m.team_client_hero[i][j][k].DistanceLeft = 0
								if m.camps[m.team_client_hero[i][j][k].Destination.I].HP > 0 {
									m.team_client_hero[i][j][k].Status = 2
									m.team_client_hero[i][j][k].Battle = m.camps[m.team_client_hero[i][j][k].Destination.I]
									appending := []int{i, j, k}
									m.team_client_hero[i][j][k].Battle.Units = append(m.team_client_hero[i][j][k].Battle.Units, appending)
								} else {
									m.team_client_hero[i][j][k].Status = 0
									if m.team_client_hero[i][j][k].Bot == false {
										has_user_cmd = true
									}
								}
								m.team_client_hero[i][j][k].Position = m.team_client_hero[i][j][k].Destination
								m.team_client_hero[i][j][k].Position.Type = "unit"
								m.team_client_hero[i][j][k].Position.I = i
								m.team_client_hero[i][j][k].Position.J = j
								m.team_client_hero[i][j][k].Position.K = k

							} else {
								m.team_client_hero[i][j][k].DistanceLeft = newDistanceLeft
								travelX, travelY := getXYComponent(m.team_client_hero[i][j][k].Angle, distanceTravelled)
								m.team_client_hero[i][j][k].Position.X = travelX + m.team_client_hero[i][j][k].Position.X
								m.team_client_hero[i][j][k].Position.Y = travelY + m.team_client_hero[i][j][k].Position.Y
							}
						}
					}
				}
			}

			// simulate all attacks
			for i := 0; i < len(atk_list); i++ {
				unitOfInterest := m.team_client_hero[atk_list[i].Attacker[0]][atk_list[i].Attacker[1]][atk_list[i].Attacker[2]]
				campOfInterest := unitOfInterest.Battle
				if campOfInterest.HP > 0 {
					campOfInterest.HP -= unitOfInterest.ATK
					if campOfInterest.HP <= 0 {
						campOfInterest.HP = 0
						unitOfInterest.EXP += 25
					} else {
						unitOfInterest.EXP += 10
					}
				}

				if m.team_client_hero[atk_list[i].Attacker[0]][atk_list[i].Attacker[1]][atk_list[i].Attacker[2]].Start == m.team_client_hero[atk_list[i].Attacker[0]][atk_list[i].Attacker[1]][atk_list[i].Attacker[2]].Destination {
					m.ended = true
				}
			}

			num, _ := strconv.Atoi(m.turn[5:])
			if len(atk_list) > 0 {
				m.turn = "TURN " + strconv.Itoa(num+1)
				m.sharepos(atk_list, min_units)
			} else {
				m.turn = "TURN " + strconv.Itoa(num+len(atk_list))
				m.sharepos(nil, 0)
			}

			if game_over_check(m.team_client_hero) {
				has_user_cmd = true
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

			// determine if camp is dead, reset camps
			for i := 0; i < len(atk_list); i++ {
				unitOfInterest := m.team_client_hero[atk_list[i].Attacker[0]][atk_list[i].Attacker[1]][atk_list[i].Attacker[2]]
				campOfInterest := unitOfInterest.Battle
				if campOfInterest == nil {
					continue
				} else if campOfInterest.HP <= 0 {
					for j := 0; j < len(campOfInterest.Units); j++ {
						unitOfInterest2 := m.team_client_hero[campOfInterest.Units[j][0]][campOfInterest.Units[j][1]][campOfInterest.Units[j][2]]
						unitOfInterest2.TTD = 100
						unitOfInterest2.Status = 0
						unitOfInterest2.Battle = nil
						if unitOfInterest2.Bot == false {
							has_user_cmd = true
						}
					}
					campOfInterest.Units = make([][]int, 0)
				} else if unitOfInterest.TTD <= 0 {
					unitOfInterest.TTD = 100
				}
			}
			if len(atk_list) > 0 {
				m.sharepos(nil, 0)
			}

			// send new timer to everyone
			timer1_length := time.Second * 120
			timer2_length := time.Second * 121

			init_time := time.Now().Add(timer1_length)
			msg := &message{Event: "ticker_start", When: time.Now(), MatchID: m.mid}
			m.ticker = time.NewTicker(timer2_length)

			// send ticker to everyone
			m.mutex.Lock()
			m.message_logs = append(m.message_logs, msg)
			m.mutex.Unlock()

			fmt.Println("sending:", msg)

			for _, i := range m.gamer_uid_to_msid_to_match_socket {
				for _, j := range i {
					msg.Message = init_time.Add(j.user_time.Sub(j.system_time)).UTC().String()
					select {
					case j.incoming_message <- msg:
					}
				}
			}

			for _, i := range m.spectator_uid_to_msid_to_match_socket {
				for _, j := range i {
					msg.Message = init_time.Add(j.system_time.Sub(j.user_time)).String()
					select {
					case j.incoming_message <- msg:
					}
				}
			}

			m.state_count += 1
			if (has_user_cmd == false) && (m.state_count < 1000) {
				fmt.Println("concurrent SIMULATE called", m.state_count)
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

func (m *match) sharepos(a []*attack, utime float64) {
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

	utime_conv := ""
	if utime != 0 {
		utime_conv = strconv.FormatFloat(utime, 'E', 3, 64)
	}

	// convert game state to JSON
	temp := make([][][]string, len(m.team_client_hero))
	for i := 0; i < len(m.team_client_hero); i++ {
		temp[i] = make([][]string, len(m.team_client_hero[i]))
		for j := 0; j < len(m.team_client_hero[i]); j++ {
			temp[i][j] = make([]string, len(m.team_client_hero[i][j]))
			for k := 0; k < len(m.team_client_hero[i][j]); k++ {
				marshalled, _ := json.Marshal(m.team_client_hero[i][j][k])
				temp[i][j][k] = string(marshalled)
			}
		}
	}

	tempCamp := make([]string, 0)
	for i := 0; i < len(m.camps); i++ {
		marshalled, _ := json.Marshal(m.camps[i])
		tempCamp = append(tempCamp, string(marshalled))
	}

	// send updated positions to everyone
	for k, i := range m.gamer_uid_to_msid_to_match_socket {
		for _, j := range i {
			select {
			case j.incoming_message <- &message{Event: "game_state", TCH: temp, Atk: atk_temp, Message: m.uuid_to_team_int[k].ab, Turn: m.turn, UnitsOfTime: utime_conv, Camps: tempCamp, When: time.Now(), MatchID: m.mid}:

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

func game_over_check(state [][][]*hero) bool {

	// check all teams, if there are two+ teams with alive heroes, then it is not game over
	counter := 0
	for i := 0; i < len(state); i++ {
		at_least_one_alive := false
		for j := 0; j < len(state[i]); j++ {
			for k := 0; k < len(state[i][j]); k++ {
				if state[i][j][k].HP > 0 {
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

func (m *match) sharebench() {

	/*
		fmt.Println("sharebench")

			marshalled, _ := json.Marshal(m.bench)

			// send updated positions to everyone
			for _, i := range m.gamer_uid_to_msid_to_match_socket {
				for _, j := range i {
					select {
					case j.incoming_message <- &message{Event: "bench", Message: string(marshalled), MatchID: m.mid}:
					}
				}
			}
			for _, i := range m.spectator_uid_to_msid_to_match_socket {
				for _, j := range i {
					select {
					case j.incoming_message <- &message{Event: "bench", Message: string(marshalled), MatchID: m.mid}:
					}
				}
			}
	*/

}

func (m *match) shareint() {
	fmt.Println("shareint")

	for k, i := range m.gamer_uid_to_msid_to_match_socket {
		for _, j := range i {
			select {
			case j.incoming_message <- &message{Event: "assignInt", Message: m.uuid_to_team_int[k].ab, MatchID: m.mid}:
			}
		}
	}
}

func getDistanceTwoPoints(c1 Coordinate, c2 Coordinate) float64 {
	y_dist := c2.Y - c1.Y
	x_dist := c2.X - c1.X
	return toFixed(math.Sqrt((y_dist*y_dist)+(x_dist*x_dist)), 3)
}
func getDistanceArray(c1 Coordinate, c2 []float64) float64 {
	y_dist := c2[1] - c1.Y
	x_dist := c2[0] - c1.X
	return toFixed(math.Sqrt((y_dist*y_dist)+(x_dist*x_dist)), 3)
}

func getXYComponent(theta float64, r float64) (float64, float64) {
	return toFixed(r*math.Cos(theta), 3), toFixed(r*math.Sin(theta), 3)
}

func getTheta(c1 Coordinate, c2 Coordinate) float64 {
	y := c2.Y - c1.Y
	x := c2.X - c1.X
	theta := math.Atan2(y, x)
	if theta < 0 {
		theta = theta + (2 * math.Pi)
	}
	return toFixed(theta, 3)
}
func getThetaArray(c1 Coordinate, c2 []float64) float64 {
	y := c2[1] - c1.Y
	x := c2[0] - c1.X
	theta := math.Atan2(y, x)
	if theta < 0 {
		theta = theta + (2 * math.Pi)
	}
	return toFixed(theta, 3)
}

func sameCoordinates(a Coordinate, b Coordinate) bool {
	if (a.X == b.X) && (a.Y == b.Y) {
		return true
	}
	return false
}

func setPath(state [][][]*hero, i int, j int, k int, dest Coordinate, p [][]float64) {
	state[i][j][k].Start = state[i][j][k].Position
	state[i][j][k].Destination = dest
	state[i][j][k].Path = p

	dist := 0.0
	for x := 0; x < len(p); x++ {
		dist += getDistanceArray(state[i][j][k].Start, p[x])
	}

	fmt.Println("p", len(p), p)
	state[i][j][k].DistanceLeft = dist
	state[i][j][k].Angle = getThetaArray(state[i][j][k].Start, p[1])
	state[i][j][k].Status = 1
}

func confirmValidPath(state [][][]*hero, camps []*Camp, i int, j int, k int, path [][]float64, dest Coordinate) bool {

	// path of 1... wth???
	if len(path) <= 1 {
		fmt.Println("returned path of 1")
		return false
	}

	// check if the start of the path is the same as the unit's coordinate
	if (state[i][j][k].Position.X != path[0][0]) || (state[i][j][k].Position.Y != path[0][1]) {
		fmt.Println("unit position is same as start position")
		return false
	}

	// check if the end of the path is a valid destination
	if dest.I >= 0 && dest.I < len(camps) && dest.J < 0 && dest.K < 0 {
		if (camps[dest.I].XY.X != path[len(path)-1][0]) || (camps[dest.I].XY.Y != path[len(path)-1][1]) {
			fmt.Println("path destination does not match camp coordinate")
			return false
		}
	} else {
		fmt.Println("camp index does not exit")
		return false
	}

	start := time.Now()
	// check if any of the line segments cross an obstacle
	for i := 0; i < len(path)-1; i++ {
		line_segment := []float64{path[i][0], path[i][1], path[i+1][0], path[i+1][1]}
		for j := 0; j < len(mapList[0].Obstacles); j++ {
			if mapList[0].Obstacle_type[j] == "BEZIER" {
				if bezier.IntersectBezierLine(mapList[0].Obstacles[j], line_segment) {
					fmt.Println("line segment intersects obstacle1")
					return false
				}
			}
		}
	}
	elapsed := time.Since(start)

	fmt.Println("isvalidpath", elapsed)
	return true
}

func determineBotAction(b *hero, camps []*Camp) ([][]float64, Coordinate) {

	// for now just pick a random camp, don't care about HP or pathfinding
	// if all camps are inaccessible... maybe end the game?

	checked := make([]int, len(camps))
	count := 0

	for count < len(camps) {
		rng := rand.Intn(len(camps))
		fmt.Println("rng", rng, checked)
		if checked[rng] == 0 {
			checked[rng] = 1
			count++

			ans := findPath(b.Position, camps[rng].XY)

			if len(ans) > 0 {
				return ans, camps[rng].XY
			} else {
				continue
			}
		} else {
			continue
		}
	}

	return [][]float64{}, Coordinate{}
}

func findPath(start Coordinate, dest Coordinate) [][]float64 {

	fmt.Println("MADE IT TO FINDPATH", start, dest)

	startTime := time.Now()

	// check if direct path from starting node to target is available
	direct_path := true
	line_segment := []float64{start.X, start.Y, dest.X, dest.Y}
	for i := 0; i < len(mapList[0].Obstacles); i++ {
		if mapList[0].Obstacle_type[i] == "BEZIER" {
			if bezier.IntersectBezierLine(mapList[0].Obstacles[i], line_segment) {
				fmt.Println("line segment intersects obstacle2")
				direct_path = false
				break
			}
		}
	}

	fmt.Println("MADE IT TO FINDPATH1", direct_path)
	if direct_path == true {
		ans := [][]float64{{start.X, start.Y}, {dest.X, dest.Y}}
		return ans
	}

	// create destination node, if no edges exist, return empty
	dest_edges, dest_count := returnEdgesArray(dest)
	fmt.Println("MADE IT TO FINDPATH2", dest_edges, dest_count, dest_count == 0)
	if dest_count == 0 {

		return [][]float64{}
	}

	fmt.Println("MADE IT TO FINDPATH3")
	// create all edges for starting node, if no edges exist, return empty
	start_edges, start_dist := returnEdgesList(start)
	if len(start_dist) == 0 {
		return [][]float64{}
	}

	fmt.Println("MADE IT TO FINDPATH4")
	// aStar
	ans := aStar(start, dest, start_edges, start_dist, dest_edges)
	fmt.Println("MADE IT TO FINDPATH5", ans)
	elapsed := time.Since(startTime)
	milliseconds := elapsed.Milliseconds()
	fmt.Printf("findPath took %d ms\n", milliseconds)

	return ans
}

func returnEdgesArray(c Coordinate) ([]float64, int) {

	edges := make([]float64, len(mapList[0].Intersections), len(mapList[0].Intersections))
	count := 0
	for i := 0; i < len(edges); i++ {
		edges[i] = -1
		line_seg := []float64{c.X, c.Y, mapList[0].Intersections[i][0], mapList[0].Intersections[i][1]}
		blocked := false
		for j := 0; j < len(mapList[0].Obstacles); j++ {
			if mapList[0].Obstacle_type[j] == "BEZIER" {
				if bezier.IntersectBezierLine(mapList[0].Obstacles[j], line_seg) {
					fmt.Println("line segment intersects obstacle3")
					blocked = true
					break
				}
			}
		}
		if !blocked {
			edges[i] = getDistanceArray(c, mapList[0].Intersections[i])
			count++
		}
	}
	return edges, count
}

func returnEdgesList(c Coordinate) ([]int, []float64) {
	edges := make([]int, 0)
	dist := make([]float64, 0)
	for i := 0; i < len(mapList[0].Intersections); i++ {
		line_seg := []float64{c.X, c.Y, mapList[0].Intersections[i][0], mapList[0].Intersections[i][1]}
		blocked := false
		for j := 0; j < len(mapList[0].Obstacles); j++ {
			if mapList[0].Obstacle_type[j] == "BEZIER" {
				if bezier.IntersectBezierLine(mapList[0].Obstacles[j], line_seg) {
					fmt.Println("line segment intersects obstacle4")
					blocked = true
					break
				}
			}
		}
		if !blocked {
			edges = append(edges, i)
			dist = append(dist, getDistanceArray(c, mapList[0].Intersections[i]))
		}
	}
	return edges, dist
}

type aStarNode struct {
	XY         Coordinate
	Edges      *[]int
	EdgeLength *[]float64

	EdgeIndex int

	G      float64 // Cost from start to this node
	H      float64
	F      float64
	Parent *aStarNode // Pointer to the parent node for path reconstruction
	Index  int        // Index in the heap
}

type PriorityQueue []*aStarNode

func (pq PriorityQueue) Len() int { return len(pq) }
func (pq PriorityQueue) Less(i, j int) bool {
	return pq[i].F < pq[j].F
}
func (pq PriorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
	pq[i].Index = i
	pq[j].Index = j
}
func (pq *PriorityQueue) Push(x interface{}) {
	node := x.(*aStarNode)
	node.Index = len(*pq)
	*pq = append(*pq, node)
}
func (pq *PriorityQueue) Pop() interface{} {
	old := *pq
	n := len(old)
	node := old[n-1]
	node.Index = -1
	*pq = old[0 : n-1]
	return node
}
func (pq *PriorityQueue) update(node *aStarNode, gCost float64, fCost float64) {
	node.G = gCost
	node.F = fCost
	heap.Fix(pq, node.Index)
}

func aStar(s Coordinate, g Coordinate, sEdges []int, sDist []float64, eEdges []float64) [][]float64 {

	fmt.Println(1, s, g)
	fmt.Println(2, sEdges)
	fmt.Println(3, sDist)
	fmt.Println(4, eEdges)

	open := &PriorityQueue{}
	track := make(map[int]*aStarNode)
	closed := make(map[Coordinate]bool)

	heap.Init(open)

	startNode := &aStarNode{XY: s, G: 0, F: 0, Parent: nil, Edges: &sEdges, EdgeLength: &sDist, EdgeIndex: -1}
	startNode.F = euclideanDistance(s.X, s.Y, g.X, g.Y)
	heap.Push(open, startNode)

	fmt.Println(5, open)
	fmt.Println(6, startNode)

	for open.Len() > 0 {
		current := heap.Pop(open).(*aStarNode)

		fmt.Println(7, current)

		if euclideanDistance(current.XY.X, current.XY.Y, g.X, g.Y) < 1e-3 {

			fmt.Println("finished")

			path := [][]float64{}
			for node := current; node != nil; node = node.Parent {
				path = append(path, []float64{node.XY.X, node.XY.Y})
			}
			// reverse the path
			for i, j := 0, len(path)-1; i < j; i, j = i+1, j-1 {
				path[i], path[j] = path[j], path[i]
			}
			return path
		}

		closed[current.XY] = true
		fmt.Println(8, closed)

		for i := 0; i < len(*current.Edges); i++ {
			neighborXY := mapList[0].IntersectionsC[(*current.Edges)[i]]

			fmt.Println(9, i, neighborXY)

			_, found := closed[neighborXY]

			fmt.Println(10, found)
			if found {
				continue
			} else {
				trackedNode, tracked := track[(*current.Edges)[i]]

				fmt.Println(11, trackedNode, tracked)
				if tracked {
					tentativeG := current.G + (*current.EdgeLength)[i]

					fmt.Println(12, tentativeG)

					if tentativeG < trackedNode.G {
						trackedNode.G = tentativeG
						trackedNode.H = euclideanDistance(neighborXY.X, neighborXY.Y, g.X, g.Y)
						trackedNode.F = trackedNode.G + trackedNode.G
						trackedNode.Parent = current

						fmt.Println(13, trackedNode)
					}
				} else {
					tempG := current.G + (*current.EdgeLength)[i]
					tempH := euclideanDistance(neighborXY.X, neighborXY.Y, g.X, g.Y)
					newNode := &aStarNode{
						XY:         neighborXY,
						G:          tempG,
						H:          tempH,
						F:          tempG + tempH,
						Parent:     current,
						Edges:      &mapList[0].Intersections_edges_int[(*current.Edges)[i]],
						EdgeLength: &mapList[0].Intersections_edgelength[(*current.Edges)[i]],
						EdgeIndex:  (*current.Edges)[i],
					}
					track[newNode.EdgeIndex] = newNode
					open.Push(newNode)

					fmt.Println(14, newNode.G, newNode.H, newNode.F)

				}
			}
		}
		if (current.EdgeIndex >= 0) && (eEdges[current.EdgeIndex] > 0) {

			fmt.Println(17, eEdges[current.EdgeIndex])

			trackedNode, tracked := track[-2]

			fmt.Println(18, trackedNode, tracked)

			if tracked {
				tentativeG := current.G + eEdges[current.EdgeIndex]
				if tentativeG < trackedNode.G {
					trackedNode.G = tentativeG
					trackedNode.H = 0
					trackedNode.F = trackedNode.G
					trackedNode.Parent = current
				}
				fmt.Println(19, tentativeG, trackedNode, tracked)
			} else {
				tempG := current.G + eEdges[current.EdgeIndex]
				newNode := &aStarNode{
					XY:         g,
					G:          tempG,
					H:          0,
					F:          tempG,
					Parent:     current,
					Edges:      nil,
					EdgeLength: nil,
					EdgeIndex:  -2,
				}
				track[newNode.EdgeIndex] = newNode
				open.Push(newNode)

				fmt.Println(20, newNode.G, newNode.H, newNode.F)
			}
		}

		heap.Init(open)
	}

	return [][]float64{}

}

func euclideanDistance(x1 float64, y1 float64, x2 float64, y2 float64) float64 {
	dx := x2 - x1
	dy := y2 - y1
	return math.Sqrt(dx*dx + dy*dy)
}
