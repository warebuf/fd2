package main

import (
	"container/heap"
	"container/list"
	"encoding/json"

	"fmt"
	"math"
	"math/rand"
	"strconv"
	"sync"
	"time"

	"fd/libs/bezier"
	"fd/libs/heroes"
	"fd/libs/maps"
	"fd/libs/pathfinding"
	"fd/libs/prediction"
	"fd/libs/util"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type attack struct {
	Attacker []int      //i,j,k
	Defender [][]int    //list of i,j,k
	Damage   [][]string // list of damage strings - 0: H, 1: DMG, 2: H%, 3: L%, 4: R%, 5: B%, 6: RNG
	Type     string     //miss,
}

type match struct {
	mutex sync.RWMutex

	mid       uuid.UUID
	game_mode string
	sides     uint
	capacity  uint
	ended     bool

	// each client is represented as an int
	state_count  int
	heroes       [][][]*heroes.Hero
	dependencies map[[3]int][][3]int
	//heroGraph   [][][]*pathfinding.DLiteGraph
	camps []*heroes.Camp

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

	Route maps.Coordinate
	Path  [][]float64

	IndexI int
	IndexJ int
	IndexK int
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
				if m.m.phase == "TURN" && m.m.ended == false && m.m.uuid_to_team_int[m.u.uid].a == msg.IndexI && m.m.uuid_to_team_int[m.u.uid].b == msg.IndexJ {
					// check if the hero is alive and at starting position
					if (m.m.heroes[msg.IndexI][msg.IndexJ][msg.IndexK].HP > 0) &&
						((m.m.heroes[msg.IndexI][msg.IndexJ][msg.IndexK].Status == 0) ||
							(sameCoordinates(m.m.heroes[msg.IndexI][msg.IndexJ][msg.IndexK].Position, m.m.heroes[msg.IndexI][msg.IndexJ][msg.IndexK].Start) &&
								(m.m.heroes[msg.IndexI][msg.IndexJ][msg.IndexK].Status == 1))) {

						canSetPath := false
						oldTarget := []int{}
						if msg.Route.I >= 0 &&
							msg.Route.I < len(m.m.heroes) &&
							msg.Route.J >= 0 &&
							msg.Route.J < len(m.m.heroes[msg.Route.I]) &&
							msg.Route.K >= 0 &&
							msg.Route.K < len(m.m.heroes[msg.Route.I][msg.Route.J]) &&
							m.m.heroes[msg.Route.I][msg.Route.J][msg.Route.K].HP > 0 {

							if confirmValidPath(m.m.heroes, m.m.camps, msg.IndexI, msg.IndexJ, msg.IndexK, msg.Path, msg.Route, 1) {
								oldTarget = m.m.heroes[msg.IndexI][msg.IndexJ][msg.IndexK].DynamicDestination
								m.m.heroes[msg.IndexI][msg.IndexJ][msg.IndexK].DynamicDestination = []int{msg.Route.I, msg.Route.J, msg.Route.K}
								m.m.heroes[msg.IndexI][msg.IndexJ][msg.IndexK].DynamicCircle = []float64{m.m.heroes[msg.Route.I][msg.Route.J][msg.Route.K].Position.X, m.m.heroes[msg.Route.I][msg.Route.J][msg.Route.K].Position.Y}
								m.m.heroes[msg.IndexI][msg.IndexJ][msg.IndexK].LastMileRadius = 25
								canSetPath = true
							}

						} else if (msg.Route.I < len(m.m.camps)) && (msg.Route.I >= 0) { // possibly check if camp is still up here

							if confirmValidPath(m.m.heroes, m.m.camps, msg.IndexI, msg.IndexJ, msg.IndexK, msg.Path, msg.Route, 0) {
								oldTarget = m.m.heroes[msg.IndexI][msg.IndexJ][msg.IndexK].DynamicDestination
								m.m.heroes[msg.IndexI][msg.IndexJ][msg.IndexK].DynamicDestination = []int{}
								m.m.heroes[msg.IndexI][msg.IndexJ][msg.IndexK].DynamicCircle = []float64{}
								m.m.heroes[msg.IndexI][msg.IndexJ][msg.IndexK].LastMileRadius = 25
								canSetPath = true
							}

						} else {
							fmt.Printf("this route does not exist")
						}

						if canSetPath {

							//startLoc := pathfinding.Location{X: msg.Path[0][0], Y: msg.Path[0][1]}
							//endLoc := pathfinding.Location{X: msg.Path[len(msg.Path)-1][0], Y: msg.Path[len(msg.Path)-1][1]}
							//m.m.heroGraph[msg.IndexI][msg.IndexJ][msg.IndexK] = DStarLiteInit(startLoc, endLoc)
							fmt.Println("calling setPath from m_read", msg.IndexI, msg.IndexJ, msg.IndexK)
							setPath(m.m.heroes[msg.IndexI][msg.IndexJ][msg.IndexK], msg.Route, msg.Path)
							updateDependencyGraph(m.m.heroes[msg.IndexI][msg.IndexJ][msg.IndexK], oldTarget, m.m.heroes[msg.IndexI][msg.IndexJ][msg.IndexK].DynamicDestination, m.m.dependencies)
							updateDependencies(m.m.heroes, m.m.heroes[msg.IndexI][msg.IndexJ][msg.IndexK], m.m.dependencies)
							//checkShadowedUnits(m.m.heroes, m.m.heroes[msg.IndexI][msg.IndexJ][msg.IndexK])

							m.m.sharepos(nil, 0)
							//m.m.printState()

							// check if all heroes have commands (ignore bots). If all heroes have commands, calculate next position
							sim_check := true
							for i := 0; i < len(m.m.heroes); i++ {
								for j := 0; j < len(m.m.heroes[i]); j++ {
									for k := 0; k < len(m.m.heroes[i][j]); k++ {
										if (!m.m.heroes[i][j][k].Bot) &&
											(m.m.heroes[i][j][k].Status == 0) &&
											(m.m.heroes[i][j][k].HP > 0) {
											sim_check = false
											break
										}
									}
								}
							}
							if sim_check {
								fmt.Println("simulate")
								m.m.simulate <- true
							}
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
					m.m.sharepos(nil, 0)
				}
				m.incoming_message <- &message{Event: "assignInt", Message: m.m.uuid_to_team_int[m.u.uid].ab, MatchID: m.m.mid}

			} else if msg.Event == "endCharSel" {
				fmt.Println("recieved endCharSel")

				if !m.m.char_sel_done[m.u.uid] {
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

			state_count: 0,
			heroes:      make([][][]*heroes.Hero, 0, 0),
			//heroGraph:   make([][][]*pathfinding.DLiteGraph, 0, 0),
			dependencies: make(map[[3]int][][3]int),

			uuid_to_team_int: make(map[uuid.UUID]pair),

			camps: make([]*heroes.Camp, 0, 0),

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

		// add bots to gamer_uid_to_user
		for _, j := range pl.spectator_uid_to_user {
			if j.bot_status {
				ans.bot_join <- j
			}
		}

		// creating camps
		random_count := 10 + rand.Intn(5)
		for i := 0; i < random_count; i++ {
			random_x := float64(rand.Intn(40))
			random_y := float64(rand.Intn(40))
			c := maps.Coordinate{X: 50.0 + (random_x * 10.0), Y: 50.0 + (random_y * 10.0), Type: "camp", I: i, J: -1, K: -1}
			HP := rand.Intn(100) + 1
			//HP := 1
			ans.camps = append(ans.camps, &heroes.Camp{MaxHP: HP, HP: HP, XY: c, Units: make([][]int, 0)})
		}
		// c := maps.Coordinate{X: 265, Y: 80, Type: "camp", I: random_count, J: -1, K: -1}
		// ans.camps = append(ans.camps, &Camp{MaxHP: 10, HP: 10, XY: c, Units: make([][]int, 0)})

		// set up the game state by: (1) creating each team, (2) assigning clients to each team, (3) creating the basic hero unit for each client
		// (1) creating each team
		for i := 0; i < int(ans.sides); i++ {
			ans.heroes = append(ans.heroes, make([][]*heroes.Hero, 0))
			//ans.heroGraph = append(ans.heroGraph, make([][]*pathfinding.DLiteGraph, 0))
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
			ans.heroes[team_int] = append(ans.heroes[team_int], make([]*heroes.Hero, 0, 5))
			//ans.heroGraph[team_int] = append(ans.heroGraph[team_int], make([]*pathfinding.DLiteGraph, 0, 5))

			names := []string{
				"EYE OF GOD", "BLUE EYES WHITE DRAGON", "PIKACHU", "MEGAN FOX", "MUMEN RIDER",
				"RAINMAN", "CARMACK", "CAGED", "JINZO", "LOADED GUN CONJECT.",
				"TERRACOTTA", "DEATH BEAM", "ONE EDGE BLADE", "DEATHBLADE", "GUINSOO",
				"SOUND OF THE RAIN", "CAT TAIL", "HOG WILD", "YEAR ZERO",
			}

			for y := 0; y < 3; y++ {
				stats := heroes.RandomBudget(100, 4)
				scale := []int{rand.Intn(10) * 10, rand.Intn(10) * 10, rand.Intn(10) * 10, rand.Intn(10) * 10, rand.Intn(10) * 10}
				temp_coordinate := maps.Coordinate{X: float64(team_int * 500), Y: 500 - (float64(y) * 100), Type: "unit", I: team_int, J: client_int, K: y}
				temp := &heroes.Hero{
					Bot:   j.bot_status,
					Name:  names[rand.Intn(len(names))],
					Index: []int{team_int, client_int, y},

					Start:        temp_coordinate,
					Position:     temp_coordinate,
					Destination:  temp_coordinate,
					Path:         make([][]float64, 0),
					Path_Segment: 0,

					DynamicDestination: []int{},
					DynamicCircle:      []float64{},

					LastMile:       0,
					LastMileRadius: 25,

					DistanceLeft: getDistanceTwoPoints(temp_coordinate, temp_coordinate),
					Status:       0,

					HP:    100,
					ATK:   50, //stats[0],
					DEF:   stats[1],
					AS:    75,             //stats[2],
					SPD:   rand.Intn(100), //stats[3],
					Scale: scale,

					Battle: nil,
					TTD:    100.0,
				}
				temp.MaxHP = temp.HP
				ans.heroes[team_int][client_int] = append(ans.heroes[team_int][client_int], temp)
				//ans.heroGraph[team_int][client_int] = append(ans.heroGraph[team_int][client_int], nil)

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

			if len(ans.heroes) > team_int {
				client_int = len(ans.heroes[team_int])
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
			for i := 0; i < len(m.heroes); i++ {
				for j := 0; j < len(m.heroes[i]); j++ {
					for k := 0; k < len(m.heroes[i][j]); k++ {
						if (m.heroes[i][j][k].Bot == true) && (m.heroes[i][j][k].Status == 0) {
							p, pint := determineBotAction(m.heroes[i][j][k], m.camps)
							if len(p) == 0 {
								// figure out how to end the game here
							} else {
								fmt.Println("calling setPath from gave_bot_acts", i, j, k)
								setPath(m.heroes[i][j][k], pint, p)
								updateDependencyGraph(m.heroes[i][j][k], m.heroes[i][j][k].DynamicDestination, []int{}, m.dependencies)
								updateDependencies(m.heroes, m.heroes[i][j][k], m.dependencies)
								//checkShadowedUnits(m.heroes, m.heroes[i][j][k])
								gave_bots_acts = true
							}

						}
					}
				}
			}
			if gave_bots_acts {
				m.sharepos(nil, 0)
			}

			// calculate min unit of time to action
			min_units := 999999.99
			for i := 0; i < len(m.heroes); i++ {
				for j := 0; j < len(m.heroes[i]); j++ {
					for k := 0; k < len(m.heroes[i][j]); k++ {
						if (m.heroes[i][j][k].HP > 0) && (m.heroes[i][j][k].Status == 1) && (m.heroes[i][j][k].SPD != 0) && (m.heroes[i][j][k].LastMile != 2) {
							units_of_time := toFixed(m.heroes[i][j][k].DistanceLeft/float64(m.heroes[i][j][k].SPD), 10)
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
						unitOfInterest := m.heroes[m.camps[i].Units[j][0]][m.camps[i].Units[j][1]][m.camps[i].Units[j][2]]
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

			recompute := [][]int{}
			lastMile := [][]int{}
			pathDup := [][]int{}

			for i := 0; i < len(m.heroes); i++ {
				for j := 0; j < len(m.heroes[i]); j++ {
					for k := 0; k < len(m.heroes[i][j]); k++ {

						if m.heroes[i][j][k].HP > 0 && len(m.heroes[i][j][k].DynamicDestination) > 0 {

							hunter := m.heroes[i][j][k]
							target := m.heroes[m.heroes[i][j][k].DynamicDestination[0]][m.heroes[i][j][k].DynamicDestination[1]][m.heroes[i][j][k].DynamicDestination[2]]

							var prox float64
							var entry string
							if hunter.LastMile == 0 {

								// radius check
								radiusCheck := getDistanceTwoPoints(hunter.Position, target.Position)
								if util.SigDig3(radiusCheck) < hunter.LastMileRadius {
									fmt.Println("weird behavior!!!", i, j, k, radiusCheck, hunter.LastMileRadius)
									panic("should not be less than hunter.LastMileRadius")
								}

								prox, entry = prediction.ProximityCheck(hunter.Path, target.Path, hunter.DistanceLeft, target.DistanceLeft, float64(hunter.SPD), float64(target.SPD), hunter.LastMileRadius, "enter")
								prox2, entry2 := prediction.ProximityCheck(hunter.Path, target.Path, hunter.DistanceLeft, target.DistanceLeft, float64(hunter.SPD), float64(target.SPD), hunter.LastMileRadius, "test")

								fmt.Println(i, j, k, "LM0", prox, min_units, entry, prox-min_units)
								fmt.Println(i, j, k, "LM0-2", prox2, min_units, entry2)

								if prox >= 0 && prox-min_units <= 1e-6 && entry == "entering" {
									fmt.Println(i, j, k, "replaced min", min_units, "w/ prox (lm0)", prox, entry)
									absDiff := math.Abs(prox - min_units)
									if absDiff <= 1e-6 {
										lastMile = append(lastMile, []int{i, j, k})
										min_units = math.Max(min_units, prox)
									} else {
										lastMile = [][]int{[]int{i, j, k}}
										recompute = [][]int{}
										min_units = prox
									}
								}

								targetTravelled := float64(target.SPD) * min_units
								if target.DistanceLeft <= 0 {
									targetTravelled = 0
								}
								distanceTravelled := float64(m.heroes[i][j][k].SPD) * min_units

								ans := prediction.CircleIntersectPath(
									target.Path,
									target.Position.X, target.Position.Y,
									target.DistanceLeft, target.DistanceLeft-targetTravelled,
									m.heroes[i][j][k].DynamicCircle,
									m.heroes[i][j][k].DistanceLeft, (m.heroes[i][j][k].DistanceLeft - distanceTravelled),
									m.heroes[i][j][k].SPD, target.SPD,
									3.0,
								)
								fmt.Println(i, j, k, "intersect found for unit", m.heroes[i][j][k].DynamicDestination[0], m.heroes[i][j][k].DynamicDestination[1], m.heroes[i][j][k].DynamicDestination[2], "at tick:", ans)

								if ans >= 0 {
									fmt.Println("replaced min ", min_units, "w/ recompute", ans, ans-min_units)
									absDiff := math.Abs(ans - min_units)
									if absDiff <= 1e-6 {
										recompute = append(recompute, []int{i, j, k})
										min_units = math.Max(min_units, ans)
									} else if ans-min_units <= 0 {
										recompute = [][]int{[]int{i, j, k}}
										lastMile = [][]int{}
										pathDup = [][]int{}
										min_units = ans
									}
								}

							} else if hunter.LastMile == 1 {

								// radius check
								radiusCheck := getDistanceTwoPoints(hunter.Position, target.Position)
								if util.SigDig3(radiusCheck) > hunter.LastMileRadius {
									fmt.Println("weird behavior!!!1", i, j, k, radiusCheck, hunter.LastMileRadius)
									panic("should not exceed hunter.LastMileRadius (1)")
								}

								prox, entry = prediction.ProximityCheck(hunter.Path, target.Path, hunter.DistanceLeft, target.DistanceLeft, float64(hunter.SPD), float64(target.SPD), hunter.LastMileRadius, "exit")
								prox2, entry2 := prediction.ProximityCheck(hunter.Path, target.Path, hunter.DistanceLeft, target.DistanceLeft, float64(hunter.SPD), float64(target.SPD), hunter.LastMileRadius, "test")

								fmt.Println(i, j, k, "prox (LM1)", prox, entry, min_units, prox-min_units)
								fmt.Println(i, j, k, "prox2 (LM1)", prox2, entry2, min_units, prox-min_units)

								if prox >= 0 && prox-min_units <= 1e-6 && (entry == "exiting" || entry == "passing") {

									fmt.Println("replace min w/ recompute", min_units, prox, entry)
									absDiff := math.Abs(prox - min_units)
									if absDiff <= 1e-6 {
										recompute = append(recompute, []int{i, j, k})
										min_units = math.Max(min_units, prox)
									} else {
										recompute = [][]int{[]int{i, j, k}}
										lastMile = [][]int{}
										pathDup = [][]int{}
										min_units = prox
									}
								}

							} else if hunter.LastMile == 2 {

								radiusCheck := getDistanceTwoPoints(hunter.Position, target.Position)
								if util.SigDig3(radiusCheck) > 1 {
									fmt.Println("weird behavior2!!!", i, j, k, radiusCheck)
									panic("should not exceed 1 radius")
								}

								hunterSPD := hunter.SPD
								targetSPD := target.SPD
								if hunterSPD > targetSPD {
									hunterSPD = targetSPD
								}

								prox, entry = prediction.ProximityCheck(hunter.Path, target.Path, hunter.DistanceLeft, target.DistanceLeft, float64(hunterSPD), float64(targetSPD), 1, "exit")
								prox2, entry2 := prediction.ProximityCheck(hunter.Path, target.Path, hunter.DistanceLeft, target.DistanceLeft, float64(hunterSPD), float64(targetSPD), 1, "test")

								fmt.Println(i, j, k, "prox (LM2)", prox, min_units, entry, prox-min_units)
								fmt.Println(i, j, k, "prox2 (LM2)", prox2, min_units, entry2)

								if prox >= 0 && prox-min_units <= 1e-6 && entry == "exiting" {
									fmt.Println(i, j, k, m.heroes[i][j][k].DynamicDestination)
									fmt.Println(i, j, k, "replaced min", min_units, "w/ prox (lm2)", prox, entry, prox-min_units)
									absDiff := math.Abs(prox - min_units)
									if absDiff <= 1e-6 {
										lastMile = append(lastMile, []int{i, j, k})
										min_units = math.Max(min_units, prox)
									} else {
										recompute = [][]int{}
										lastMile = [][]int{[]int{i, j, k}}
										pathDup = [][]int{}
										min_units = prox
									}
								}

							} else if hunter.LastMile == 3 {

								pos := []float64{hunter.Position.X, hunter.Position.Y}
								intersect := prediction.MeetsPoint(target.Path, pos, float64(target.SPD))
								fmt.Println("found int", hunter.Index, intersect)
								if intersect >= 0 && intersect-min_units <= 1e-6 {
									fmt.Println(i, j, k, "replaced min", min_units, "w/ int (lm3)", intersect, intersect-min_units)
									absDiff := math.Abs(intersect - min_units)
									if absDiff <= 1e-6 {
										pathDup = append(pathDup, []int{i, j, k})
										min_units = math.Max(min_units, intersect)
									} else {
										recompute = [][]int{}
										lastMile = [][]int{}
										pathDup = [][]int{[]int{i, j, k}}
										min_units = intersect
									}

									panic("I just want to see it work 1 time")
								}

							}

						}

					}
				}
			}

			fmt.Println("FINAL min", min_units)

			has_user_cmd := false

			// calculate TTD for each unit, save the units that are going to attack next
			atk_list := make([]*attack, 0)
			for i := 0; i < len(m.camps); i++ {
				if m.camps[i].HP > 0 {
					for j := 0; j < len(m.camps[i].Units); j++ {
						unitOfInterest := m.heroes[m.camps[i].Units[j][0]][m.camps[i].Units[j][1]][m.camps[i].Units[j][2]]
						if unitOfInterest.HP > 0 {
							unitOfInterest.TTD = unitOfInterest.TTD - (min_units * float64(unitOfInterest.AS))
							if unitOfInterest.TTD <= 1e-6 {
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

			// quick check
			for i := 0; i < len(m.heroes); i++ {
				for j := 0; j < len(m.heroes[i]); j++ {
					for k := 0; k < len(m.heroes[i][j]); k++ {
						if len(m.heroes[i][j][k].DynamicDestination) > 0 {
							fmt.Println("before", i, j, k, "LM:", m.heroes[i][j][k].LastMile, "DynamicDist:", getDistFromDynamic(m.heroes, m.heroes[i][j][k]), "Dist:", m.heroes[i][j][k].DistanceLeft, "Pos:", m.heroes[i][j][k].Position)
						} else {
							fmt.Println("before", i, j, k, "LM:", m.heroes[i][j][k].LastMile, "Dist:", m.heroes[i][j][k].DistanceLeft, "Pos:", m.heroes[i][j][k].Position)
						}
					}
				}
			}

			// move all unit positions
			for i := 0; i < len(m.heroes); i++ {
				for j := 0; j < len(m.heroes[i]); j++ {
					for k := 0; k < len(m.heroes[i][j]); k++ {
						if (m.heroes[i][j][k].HP > 0) && (m.heroes[i][j][k].Status == 1) {

							SPD := float64(m.heroes[i][j][k].SPD)

							if m.heroes[i][j][k].LastMile == 2 {
								hunter := m.heroes[i][j][k]
								target := m.heroes[hunter.DynamicDestination[0]][hunter.DynamicDestination[1]][hunter.DynamicDestination[2]]
								if hunter.SPD > target.SPD {
									SPD = float64(target.SPD)
								}
							} else if m.heroes[i][j][k].LastMile == 3 {
								continue
							}

							toMove := min_units * SPD
							pos_x := m.heroes[i][j][k].Position.X
							pos_y := m.heroes[i][j][k].Position.Y

							// figure out what line segment we end up at by incrementing Path_Segment and decreasing toMove
							// when toMove is negative, it indicates how much distance we have left in that line segment
							for toMove > 0 {

								if m.heroes[i][j][k].Path_Segment+1 >= len(m.heroes[i][j][k].Path) {
									toMove = 0
								} else {
									seg_x := m.heroes[i][j][k].Path[m.heroes[i][j][k].Path_Segment+1][0] - pos_x
									seg_y := m.heroes[i][j][k].Path[m.heroes[i][j][k].Path_Segment+1][1] - pos_y
									segment_distance := math.Sqrt((seg_x * seg_x) + (seg_y * seg_y)) // calculate remaining segment distance
									toMove -= segment_distance
									if toMove >= 0 {
										pos_x = m.heroes[i][j][k].Path[m.heroes[i][j][k].Path_Segment+1][0]
										pos_y = m.heroes[i][j][k].Path[m.heroes[i][j][k].Path_Segment+1][1]
										m.heroes[i][j][k].Position.X = pos_x
										m.heroes[i][j][k].Position.Y = pos_y
										m.heroes[i][j][k].Path_Segment++
									} else {
										travelX, travelY := getLineSegmentTravelled(pos_x, pos_y, m.heroes[i][j][k].Path[m.heroes[i][j][k].Path_Segment+1], toMove)
										m.heroes[i][j][k].Position.X, m.heroes[i][j][k].Position.Y = travelX, travelY
									}
								}
							}

							// move the unit
							m.heroes[i][j][k].DistanceLeft -= min_units * SPD

							if m.heroes[i][j][k].DistanceLeft <= 1e-6 {

								m.heroes[i][j][k].DistanceLeft = 0

								if m.heroes[i][j][k].LastMile == 1 {

									pathDup = append(pathDup, []int{i, j, k})
									m.heroes[i][j][k].Path_Segment = len(m.heroes[i][j][k].Path) - 1

								} else if m.heroes[i][j][k].LastMile == 2 {

									m.heroes[i][j][k].Position = m.heroes[i][j][k].Destination
									m.heroes[i][j][k].Path_Segment = len(m.heroes[i][j][k].Path) - 1

								} else {

									if m.camps[m.heroes[i][j][k].Destination.I].HP > 0 {
										m.heroes[i][j][k].Status = 2
										m.heroes[i][j][k].Battle = m.camps[m.heroes[i][j][k].Destination.I]
										m.heroes[i][j][k].Battle.Units = append(m.heroes[i][j][k].Battle.Units, []int{i, j, k})
									} else {
										m.heroes[i][j][k].Status = 0
										if !m.heroes[i][j][k].Bot {
											has_user_cmd = true
										}
									}
									m.heroes[i][j][k].Position = m.heroes[i][j][k].Destination
									m.heroes[i][j][k].Position.Type = "unit"
									m.heroes[i][j][k].Position.I = i
									m.heroes[i][j][k].Position.J = j
									m.heroes[i][j][k].Position.K = k
									m.heroes[i][j][k].Path_Segment = len(m.heroes[i][j][k].Path) - 1

								}

							}
						}
					}
				}
			}

			// quick check
			for i := 0; i < len(m.heroes); i++ {
				for j := 0; j < len(m.heroes[i]); j++ {
					for k := 0; k < len(m.heroes[i][j]); k++ {
						if len(m.heroes[i][j][k].DynamicDestination) > 0 {
							fmt.Println("after1", i, j, k, "LM:", m.heroes[i][j][k].LastMile, "DynamicDist:", getDistFromDynamic(m.heroes, m.heroes[i][j][k]), "Dist:", m.heroes[i][j][k].DistanceLeft, "Pos:", m.heroes[i][j][k].Position)
						} else {
							fmt.Println("after1", i, j, k, "LM:", m.heroes[i][j][k].LastMile, "Dist:", m.heroes[i][j][k].DistanceLeft, "Pos:", m.heroes[i][j][k].Position)
						}
					}
				}
			}

			// simulate all attacks
			levelUp := [][]int{}
			levelCount := []int{}
			for i := 0; i < len(atk_list); i++ {
				unitOfInterest := m.heroes[atk_list[i].Attacker[0]][atk_list[i].Attacker[1]][atk_list[i].Attacker[2]]
				campOfInterest := unitOfInterest.Battle
				if campOfInterest.HP > 0 {
					calculatedEXP := 0
					startEXP := unitOfInterest.EXP
					fmt.Println(unitOfInterest.ATK, campOfInterest.HP)
					if unitOfInterest.ATK <= campOfInterest.HP {
						calculatedEXP = unitOfInterest.ATK
					} else {
						calculatedEXP = campOfInterest.HP
					}
					campOfInterest.HP -= unitOfInterest.ATK
					if campOfInterest.HP <= 0 {
						campOfInterest.HP = 0
						unitOfInterest.EXP += calculatedEXP + 100
					} else {
						unitOfInterest.EXP += calculatedEXP
					}
					fmt.Println("startEXP", startEXP)
					fmt.Println("endEXP", unitOfInterest.EXP)
					levels := math.Floor(float64(unitOfInterest.EXP)/100.0) - math.Floor(float64(startEXP)/100.0)
					if levels > 0 {
						levelUp = append(levelUp, []int{atk_list[i].Attacker[0], atk_list[i].Attacker[1], atk_list[i].Attacker[2]})
						levelCount = append(levelCount, int(levels))
					}
				}

				// TODO: not sure why this is here TBH
				if m.heroes[atk_list[i].Attacker[0]][atk_list[i].Attacker[1]][atk_list[i].Attacker[2]].Start == m.heroes[atk_list[i].Attacker[0]][atk_list[i].Attacker[1]][atk_list[i].Attacker[2]].Destination {
					m.ended = true
				}
			}

			// quick check
			for i := 0; i < len(m.heroes); i++ {
				for j := 0; j < len(m.heroes[i]); j++ {
					for k := 0; k < len(m.heroes[i][j]); k++ {
						if len(m.heroes[i][j][k].DynamicDestination) > 0 {
							fmt.Println("after2", i, j, k, "LM:", m.heroes[i][j][k].LastMile, "DynamicDist:", getDistFromDynamic(m.heroes, m.heroes[i][j][k]), "Dist:", m.heroes[i][j][k].DistanceLeft, "Pos:", m.heroes[i][j][k].Position)
						} else {
							fmt.Println("after2", i, j, k, "LM:", m.heroes[i][j][k].LastMile, "Dist:", m.heroes[i][j][k].DistanceLeft, "Pos:", m.heroes[i][j][k].Position)
						}
					}
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

			// handle level up
			for i := 0; i < len(levelUp); i++ {

				fmt.Println("levelUp[i]", levelUp[i])
				fmt.Println("levelCount[i]", levelCount[i])

				for j := 0; j < levelCount[i]; j++ {
					focus := m.heroes[levelUp[i][0]][levelUp[i][1]][levelUp[i][2]]
					additions := heroes.LevelUp(focus.Scale, 20)
					focus.MaxHP += additions[0]
					focus.HP += additions[0]
					focus.ATK += additions[1]
					focus.AS += additions[2]
					focus.DEF += additions[3]
					focus.SPD += additions[4]
					m.sharepos(nil, 0)
				}

			}

			// recalculate dynamic destinations (LM0)
			for x := 0; x < len(recompute); x++ {
				setRecompute(m.heroes, recompute[x])
				updateDependencies(m.heroes, m.heroes[recompute[x][0]][recompute[x][1]][recompute[x][2]], m.dependencies)
			}
			// calculate last mile paths (LM1)
			for x := 0; x < len(lastMile); x++ {
				setLastMile(m.heroes, lastMile[x])
				updateDependencies(m.heroes, m.heroes[lastMile[x][0]][lastMile[x][1]][lastMile[x][2]], m.dependencies)
			}
			// calculate collision path duplication routes (LM2)
			for x := 0; x < len(pathDup); x++ {
				setPathDup(m.heroes, pathDup[x])
				updateDependencies(m.heroes, m.heroes[pathDup[x][0]][pathDup[x][1]][pathDup[x][2]], m.dependencies)
			}

			if len(recompute) > 0 || len(lastMile) > 0 || len(pathDup) > 0 {
				m.sharepos(nil, 0)
			}

			if game_over_check(m.heroes) {
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
				unitOfInterest := m.heroes[atk_list[i].Attacker[0]][atk_list[i].Attacker[1]][atk_list[i].Attacker[2]]
				campOfInterest := unitOfInterest.Battle
				if campOfInterest == nil {
					continue
				} else if campOfInterest.HP <= 0 {
					for j := 0; j < len(campOfInterest.Units); j++ {
						unitOfInterest2 := m.heroes[campOfInterest.Units[j][0]][campOfInterest.Units[j][1]][campOfInterest.Units[j][2]]
						unitOfInterest2.TTD = 100
						unitOfInterest2.Status = 0
						unitOfInterest2.Battle = nil
						if !unitOfInterest2.Bot {
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

			//m.printState()

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

			if false && (has_user_cmd == false) {
				fmt.Println("concurrent SIMULATE called", m.state_count)
				go func() { m.simulate <- true }()
			}

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
	temp := make([][][]string, len(m.heroes))
	for i := 0; i < len(m.heroes); i++ {
		temp[i] = make([][]string, len(m.heroes[i]))
		for j := 0; j < len(m.heroes[i]); j++ {
			temp[i][j] = make([]string, len(m.heroes[i][j]))
			for k := 0; k < len(m.heroes[i][j]); k++ {
				marshalled, _ := json.Marshal(m.heroes[i][j][k])
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

func game_over_check(state [][][]*heroes.Hero) bool {

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

func getDistanceTwoPoints(c1 maps.Coordinate, c2 maps.Coordinate) float64 {
	y_dist := c2.Y - c1.Y
	x_dist := c2.X - c1.X
	return toFixed(math.Sqrt((y_dist*y_dist)+(x_dist*x_dist)), 10)
}

func getDistanceArray(c1 []float64, c2 []float64) float64 {
	y_dist := c2[1] - c1[1]
	x_dist := c2[0] - c1[0]
	return toFixed(math.Sqrt((y_dist*y_dist)+(x_dist*x_dist)), 10)
}

func getDistanceLocation(c1 pathfinding.Location, c2 pathfinding.Location) float64 {
	y_dist := c2.Y - c1.Y
	x_dist := c2.X - c1.X
	return toFixed(math.Sqrt((y_dist*y_dist)+(x_dist*x_dist)), 10)
}

func getDistFromDynamic(state [][][]*heroes.Hero, unit *heroes.Hero) float64 {

	if len(unit.DynamicDestination) > 0 {

		hunter := unit
		target := state[hunter.DynamicDestination[0]][hunter.DynamicDestination[1]][hunter.DynamicDestination[2]]
		x_dist := target.Position.X - hunter.Position.X
		y_dist := target.Position.Y - hunter.Position.Y
		return toFixed(math.Sqrt((y_dist*y_dist)+(x_dist*x_dist)), 10)
	}
	return 0.0

}

func getLineSegmentTravelled(refX float64, refY float64, dest []float64, distanceTravelled float64) (float64, float64) {
	if distanceTravelled == 0 {
		return refX, refY
	}
	lineX := dest[0] - refX
	lineY := dest[1] - refY
	lineDist := math.Sqrt((lineX * lineX) + (lineY * lineY)) // calculate remaining segment distance
	percent := (lineDist + distanceTravelled) / lineDist

	return refX + (percent * lineX), refY + (percent * lineY)
}

func sameCoordinates(a maps.Coordinate, b maps.Coordinate) bool {
	if (a.X == b.X) && (a.Y == b.Y) {
		return true
	}
	return false
}

func setPath(unit *heroes.Hero, dest maps.Coordinate, p [][]float64) {

	if (unit.Position.X != p[0][0]) || (unit.Position.Y != p[0][1]) {

		dist := maps.EuclideanDistance(unit.Position.X, unit.Position.Y, p[0][0], p[0][1])
		fmt.Println("X:", unit.Position.X, unit.Position.Y, p[0])
		fmt.Println("dist", dist)

		if dist < 1e-8 {
			unit.Position.X = p[0][0]
			unit.Position.Y = p[0][1]

		} else if dist <= 1 {

			// to me the best option here would be to calculate A* from unit.Position to the destination
			s := pathfinding.Location{X: unit.Position.X, Y: unit.Position.Y}
			e := pathfinding.Location{X: dest.X, Y: dest.Y}
			AStarGraph := pathfinding.AStarInit(mapList[0], s, e)
			p = AStarGraph.ComputeShortestPath()

		} else {
			panic("Path and unit position exceed 1 dist!")
		}

	}

	unit.Start = unit.Position
	unit.Destination = dest
	unit.Path = p
	unit.Path_Segment = 0

	thetas := make([]float64, len(p), len(p))
	dist := 0.0
	for x := 0; x+1 < len(p); x++ {
		dist += getDistanceArray(p[x], p[x+1])
		thetas[x] = getTheta(p[x], p[x+1])
		if x+1 == len(p)-1 {
			thetas[x+1] = thetas[x]
		}
	}
	unit.Thetas = thetas

	fmt.Println("setPath", len(p), p)
	unit.DistanceLeft = dist
	unit.Status = 1
}

// pathType can be 0 (camp) or 1 (unit),
func confirmValidPath(state [][][]*heroes.Hero, camps []*heroes.Camp, i int, j int, k int, path [][]float64, dest maps.Coordinate, pathType int) bool {

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

	if pathType == 0 {
		if dest.I >= 0 && dest.I < len(camps) && dest.J < 0 && dest.K < 0 {
			if (camps[dest.I].XY.X != path[len(path)-1][0]) || (camps[dest.I].XY.Y != path[len(path)-1][1]) {
				fmt.Println("path destination does not match camp coordinate")
				return false
			}
		} else {
			fmt.Println("camp index does not exit")
			return false
		}
	} else if pathType == 1 {
		if dest.I >= 0 && dest.I < len(state) && dest.J >= 0 && dest.J < len(state[dest.I]) && dest.K >= 0 && dest.K < len(state[dest.I][dest.J]) {
			if (state[dest.I][dest.J][dest.K].Position.X != path[len(path)-1][0]) || (state[dest.I][dest.J][dest.K].Position.Y != path[len(path)-1][1]) {
				fmt.Println("path destination does not match unit coordinate")
				return false
			}
		} else {
			fmt.Println("unit index does not exit")
			return false
		}
	}

	// check if any of the line segments cross an obstacle
	for i := 0; i < len(path)-1; i++ {
		line_segment := []float64{path[i][0], path[i][1], path[i+1][0], path[i+1][1]}
		for j := 0; j < len(mapList[0].Obstacles); j++ {
			if mapList[0].Obstacle_type[j] == "BEZIER" {
				if bezier.IntersectBezierLine(mapList[0].Obstacles[j], line_segment) {

					fmt.Println("line segment intersects obstacle", mapList[0].Obstacles[j], line_segment)
					return false
				}
			}
		}
	}

	return true
}

func determineBotAction(b *heroes.Hero, camps []*heroes.Camp) ([][]float64, maps.Coordinate) {

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

			if camps[rng].HP <= 0 {
				continue
			}

			ans := findPath(b.Position, camps[rng].XY)

			if len(ans) > 0 {

				//s := pathfinding.Location{X: b.Position.X, Y: b.Position.Y}
				//e := pathfinding.Location{X: camps[rng].XY.X, Y: camps[rng].XY.Y}
				//dGraph := DStarLiteInit(s, e)

				return ans, camps[rng].XY
			} else {
				continue
			}
		} else {
			continue
		}
	}

	return [][]float64{}, maps.Coordinate{}
}

func findPath(start maps.Coordinate, dest maps.Coordinate) [][]float64 {

	startTime := time.Now()

	// check if direct path from starting node to target is available
	direct_path := true
	line_segment := []float64{start.X, start.Y, dest.X, dest.Y}
	for i := 0; i < len(mapList[0].Obstacles); i++ {
		if mapList[0].Obstacle_type[i] == "BEZIER" {
			if bezier.IntersectBezierLine(mapList[0].Obstacles[i], line_segment) {
				direct_path = false
				break
			}
		}
	}

	if direct_path == true {
		ans := [][]float64{{start.X, start.Y}, {dest.X, dest.Y}}
		return ans
	}

	//ans := aStar(start, dest, start_edges, start_dist, dest_edges)
	s := pathfinding.Location{X: start.X, Y: start.Y}
	e := pathfinding.Location{X: dest.X, Y: dest.Y}
	AStarGraph := pathfinding.AStarInit(mapList[0], s, e)
	ans := AStarGraph.ComputeShortestPath()

	elapsed := time.Since(startTime)
	milliseconds := elapsed.Milliseconds()
	fmt.Printf("findPath took %d ms\n", milliseconds)

	return ans
}

func getEdges(c1 pathfinding.Location, LPA *pathfinding.LPAGraph) ([]pathfinding.Location, []float64) {
	edges := make([]pathfinding.Location, 0)
	dist := make([]float64, 0)

	for key, _ := range LPA.Nodes {
		line_seg := []float64{key.X, key.Y, c1.X, c1.Y}
		blocked := false
		for j := 0; j < len(mapList[0].Obstacles); j++ {
			if mapList[0].Obstacle_type[j] == "BEZIER" {
				if bezier.IntersectBezierLine(mapList[0].Obstacles[j], line_seg) {
					blocked = true
					break
				}
			}
		}
		if !blocked {
			edges = append(edges, key)
			dist = append(dist, getDistanceLocation(c1, key))
		}
	}

	return edges, dist
}

func getTheta(c1 []float64, c2 []float64) float64 {
	y := c2[1] - c1[1]
	x := c2[0] - c1[0]
	theta := math.Atan2(y, x)
	if theta < 0 {
		theta = theta + (2 * math.Pi)
	}
	return toFixed(theta, 3)
}

func LPA(start pathfinding.Location, end pathfinding.Location) {
	fmt.Println("ENTERED LPA* -------------------------------------------------------------------------------------")

	LPA := pathfinding.LPAInit(mapList[0], start, end)

	if _, exists := LPA.Nodes[start]; !exists {
		fmt.Println("LPA start node not found", start)
		edgeIndex, edgeCost := getEdges(start, LPA)
		LPA.AddNode(start)
		for i := 0; i < len(edgeIndex); i++ {
			LPA.AddEdge(start, edgeIndex[i], edgeCost[i])
		}
	}

	if _, exists := LPA.Nodes[end]; !exists {
		fmt.Println("LPA end node not found", end)
		edgeIndex, edgeCost := getEdges(end, LPA)
		LPA.AddNode(end)
		for i := 0; i < len(edgeIndex); i++ {
			LPA.AddEdge(end, edgeIndex[i], edgeCost[i])
		}
	}

	LPA.Nodes[start][1] = 0
	heap.Push(LPA.PriorityQueue, LPA.CalculateKey(start))

	LPA.ComputeShortestPath()
	LPA.ReturnShortestPath()

	// remove start and end nodes

}

func DStarLiteInit(start pathfinding.Location, end pathfinding.Location) *pathfinding.DLiteGraph {

	DStar := pathfinding.DLiteInit(mapList[0], start, end)

	if _, exists := DStar.Nodes[start]; !exists {
		edgeIndex, edgeCost := DStar.GetEdges(mapList[0], start)
		DStar.AddNode(start)
		for i := 0; i < len(edgeIndex); i++ {
			DStar.AddEdge(start, edgeIndex[i], edgeCost[i])
		}
	}

	if _, exists := DStar.Nodes[end]; !exists {
		edgeIndex, edgeCost := DStar.GetEdges(mapList[0], end)
		DStar.AddNode(end)
		for i := 0; i < len(edgeIndex); i++ {
			DStar.AddEdge(end, edgeIndex[i], edgeCost[i])
		}
	}

	DStar.Nodes[end][1] = 0
	heap.Push(DStar.PriorityQueue, DStar.CalculateKey(end))
	s := time.Now()
	DStar.ComputeShortestPath()
	elapsed := time.Since(s)
	fmt.Println("D* runtime", elapsed)
	DStar.ReturnShortestPath()

	return DStar
}

func PathDuplication(hunter *heroes.Hero, target *heroes.Hero) {

	dist := maps.EuclideanDistance(hunter.Position.X, hunter.Position.Y, target.Position.X, target.Position.Y)

	fmt.Println("dist", dist)

	if dist > 1 { // this should break the program
		fmt.Println("hunter index", hunter.Index)
		fmt.Println("hunter.Position", hunter.Position)
		fmt.Println("target.Position", target.Position)
		fmt.Println("hunter.Path", hunter.Path)
		fmt.Println("target.Path", target.Path)
		fmt.Println("hunter.DistanceLeft", hunter.DistanceLeft)
		fmt.Println("target.DistanceLeft", target.DistanceLeft)

		panic("Collision should not happen! Distance > 1!")
	}

	end := maps.Coordinate{X: target.Path[len(target.Path)-1][0], Y: target.Path[len(target.Path)-1][1], Type: "unit", I: target.Index[0], J: target.Index[1], K: target.Index[2]}
	newPath := prediction.GetSubPath(target.Path, target.DistanceLeft, 0)
	fmt.Println("hunter.Position", hunter.Position)
	fmt.Println("PathDuplication newPath:", newPath)
	fmt.Println("calling setPath from PathDuplication", hunter.Index)
	setPath(hunter, end, newPath)
	hunter.LastMile = 2
}

func checkShadowedUnits(state [][][]*heroes.Hero, source *heroes.Hero) {

	for i := 0; i < len(state); i++ {
		for j := 0; j < len(state[i]); j++ {
			for k := 0; k < len(state[i][j]); k++ {

				if len(state[i][j][k].DynamicDestination) > 0 &&
					state[i][j][k].DynamicDestination[0] == source.Index[0] &&
					state[i][j][k].DynamicDestination[1] == source.Index[1] &&
					state[i][j][k].DynamicDestination[2] == source.Index[2] {

					if state[i][j][k].LastMile == 0 {
						// set new center point and calculate A*

						state[i][j][k].DynamicCircle = []float64{source.Position.X, source.Position.Y}

						s := pathfinding.Location{X: state[i][j][k].Position.X, Y: state[i][j][k].Position.Y}
						e := pathfinding.Location{X: source.Position.X, Y: source.Position.Y}
						AStarGraph := pathfinding.AStarInit(mapList[0], s, e)
						path := AStarGraph.ComputeShortestPath()
						end := maps.Coordinate{X: path[len(path)-1][0], Y: path[len(path)-1][1], Type: "unit", I: source.Index[0], J: source.Index[1], K: source.Index[2]}
						fmt.Println("calling setPath from checkShadowedUnits0", i, j, k)
						setPath(state[i][j][k], end, path)

					} else if state[i][j][k].LastMile == 1 {
						// calculate new shadow
						hunter := state[i][j][k]
						target := state[hunter.DynamicDestination[0]][hunter.DynamicDestination[1]][hunter.DynamicDestination[2]]
						hunter.DynamicCircle = []float64{}
						hunterPos := []float64{hunter.Position.X, hunter.Position.Y}
						path := prediction.LastMilePrediction(mapList[0], target.Path, target.DistanceLeft, target.SPD, hunterPos, hunter.SPD)
						end := maps.Coordinate{X: path[len(path)-1][0], Y: path[len(path)-1][1], Type: "unit", I: target.Index[0], J: target.Index[1], K: target.Index[2]}
						fmt.Println("calling setPath from checkShadowedUnits1", i, j, k)
						setPath(hunter, end, path)

					} else if state[i][j][k].LastMile == 2 {
						// copy path
						hunter := state[i][j][k]
						target := state[hunter.DynamicDestination[0]][hunter.DynamicDestination[1]][hunter.DynamicDestination[2]]
						hunter.DynamicCircle = []float64{}
						fmt.Println("calling PathDuplication from checkShadowedUnits2", i, j, k)
						PathDuplication(hunter, target)
					}

				}

			}
		}
	}
}

func setRecompute(state [][][]*heroes.Hero, idx []int) {
	fmt.Println("recompute[x]", idx)
	hunter := state[idx[0]][idx[1]][idx[2]]
	target := state[hunter.DynamicDestination[0]][hunter.DynamicDestination[1]][hunter.DynamicDestination[2]]
	hunter.DynamicCircle = []float64{target.Position.X, target.Position.Y}
	if hunter.LastMile == 1 {
		hunter.LastMileRadius = 25
	}

	s := pathfinding.Location{X: hunter.Position.X, Y: hunter.Position.Y}
	e := pathfinding.Location{X: target.Position.X, Y: target.Position.Y}
	AStarGraph := pathfinding.AStarInit(mapList[0], s, e)
	path := AStarGraph.ComputeShortestPath()
	end := maps.Coordinate{X: path[len(path)-1][0], Y: path[len(path)-1][1], Type: "unit", I: idx[0], J: idx[1], K: idx[2]}
	fmt.Println("calling setPath from dynamic destinations", idx)
	setPath(state[idx[0]][idx[1]][idx[2]], end, path)
	hunter.LastMile = 0
}

func setLastMile(state [][][]*heroes.Hero, idx []int) {
	fmt.Println("lastMile[x]", idx)

	// calculate the last mile path
	hunter := state[idx[0]][idx[1]][idx[2]]
	target := state[hunter.DynamicDestination[0]][hunter.DynamicDestination[1]][hunter.DynamicDestination[2]]
	hunterPos := []float64{hunter.Position.X, hunter.Position.Y}
	path := prediction.LastMilePrediction(mapList[0], target.Path, target.DistanceLeft, target.SPD, hunterPos, hunter.SPD)

	// check if the last mile path doesn't backtrack (causing a corner case)
	hunterStart := path[0]
	hunterEnd := path[0]
	if len(path) > 1 {
		hunterEnd = path[1]
	}
	targetStart := []float64{target.Position.X, target.Position.Y}
	targetEnd := []float64{target.Position.X, target.Position.Y}
	if target.Path_Segment+1 < len(target.Path) {
		targetEnd = target.Path[target.Path_Segment+1]
	}
	gainingGround := prediction.GainRadius(hunterStart, hunterEnd, targetStart, targetEnd, hunter.SPD, target.SPD)

	if !gainingGround && hunter.LastMile == 0 {

		fmt.Println("hunter.path", path)
		fmt.Println("target.path", target.Path, target.Path_Segment)

		hunter.LastMileRadius = hunter.LastMileRadius * 0.75
		hunter.DynamicCircle = []float64{target.Position.X, target.Position.Y}
		s := pathfinding.Location{X: hunter.Position.X, Y: hunter.Position.Y}
		e := pathfinding.Location{X: target.Position.X, Y: target.Position.Y}
		AStarGraph := pathfinding.AStarInit(mapList[0], s, e)
		path := AStarGraph.ComputeShortestPath()
		end := maps.Coordinate{X: path[len(path)-1][0], Y: path[len(path)-1][1], Type: "unit", I: idx[0], J: idx[1], K: idx[2]}
		fmt.Println("calling setPath from gainingGround", idx)
		setPath(state[idx[0]][idx[1]][idx[2]], end, path)
		hunter.LastMile = 0

	} else {
		hunter.LastMile = 1
		hunter.DynamicCircle = []float64{}
		hunter.LastMileRadius = 25
		end := maps.Coordinate{X: path[len(path)-1][0], Y: path[len(path)-1][1], Type: "unit", I: idx[0], J: idx[1], K: idx[2]}
		fmt.Println("calling setPath from last_mile", idx)
		setPath(state[idx[0]][idx[1]][idx[2]], end, path)
	}
}

func setPathDup(state [][][]*heroes.Hero, idx []int) {
	hunter := state[idx[0]][idx[1]][idx[2]]
	target := state[hunter.DynamicDestination[0]][hunter.DynamicDestination[1]][hunter.DynamicDestination[2]]
	hunter.DynamicCircle = []float64{}

	// if the hunter and target are > 1 dist, then the hunter has arrived early, set the LastMile to 3 and wait for collision
	dist := getDistFromDynamic(state, hunter)
	fmt.Println(idx, "dist", dist)
	if dist > 1 {
		fmt.Println("dist is > 1", hunter.Index, hunter.Path, hunter.DistanceLeft, hunter.Position, hunter.DynamicCircle)
		fmt.Println("round2", hunter.LastMile, hunter.LastMileRadius)
		fmt.Println("round3", hunter.Battle, hunter.Destination, hunter.DynamicDestination)
		p := [][]float64{[]float64{hunter.Position.X, hunter.Position.Y}}
		setPath(hunter, hunter.Position, p)

		fmt.Println("post setPath", hunter.Index, hunter.Path, hunter.DistanceLeft, hunter.Position, hunter.DynamicCircle)
		fmt.Println("round2", hunter.LastMile, hunter.LastMileRadius)
		fmt.Println("round3", hunter.Battle, hunter.Destination, hunter.DynamicDestination)

		hunter.LastMile = 3
	} else {
		fmt.Println("calling PathDuplication from pathDup")
		PathDuplication(hunter, target)
	}
}

func updateDependencyGraph(hunter *heroes.Hero, oldTarget []int, newTarget []int, dependencies map[[3]int][][3]int) {

	hunterIndex := [3]int{hunter.Index[0], hunter.Index[1], hunter.Index[2]}

	// remove from old array
	if len(oldTarget) > 0 {
		oldKey := [3]int{oldTarget[0], oldTarget[1], oldTarget[2]}
		updatedArray := [][3]int{}
		for i := 0; i < len(dependencies[oldKey]); i++ {
			if dependencies[oldKey][i] != hunterIndex {
				updatedArray = append(updatedArray, dependencies[oldKey][i])
			}
		}
		dependencies[oldKey] = updatedArray
	}

	// update to new array
	if len(newTarget) > 0 {
		newKey := [3]int{newTarget[0], newTarget[1], newTarget[2]}
		dependencies[newKey] = append(dependencies[newKey], hunterIndex)
	}

}

func updateDependencies(state [][][]*heroes.Hero, parent *heroes.Hero, graph map[[3]int][][3]int) {

	key := [3]int{parent.Index[0], parent.Index[1], parent.Index[2]}
	val, ok := graph[key]
	if !ok || len(val) == 0 {
		return
	}

	queue := list.New()
	for i := 0; i < len(val); i++ {
		queue.PushBack(val[i])
	}

	for queue.Len() > 0 {

		current := queue.Front().Value.([3]int)
		currentHunter := state[current[0]][current[1]][current[2]]

		if currentHunter.Status != 1 {
			panic("why are we updating a dependency that is not a follower")
		}

		if currentHunter.LastMile == 0 {

		} else if currentHunter.LastMile == 1 {
			setLastMile(state, currentHunter.Index)
		} else if currentHunter.LastMile == 2 {
			setPathDup(state, currentHunter.Index)
		} else if currentHunter.LastMile == 3 {
			dist := getDistFromDynamic(state, currentHunter)
			if dist <= 1 {
				setPathDup(state, currentHunter.Index)
			} else if dist <= 25 {
				setLastMile(state, currentHunter.Index)
			} else {
				setRecompute(state, currentHunter.Index)
			}
		}

		if val, ok := graph[current]; ok {
			for i := 0; i < len(val); i++ {
				queue.PushBack(val[i])
			}
		}
		queue.Remove(queue.Front())
	}

}


