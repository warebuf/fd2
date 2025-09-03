package main

import (
	"container/heap"
	"container/list"
	"encoding/json"
	"log"

	"fmt"
	"math"
	"math/rand"
	"strconv"
	"sync"
	"time"

	"fd/libs/bezier"
	"fd/libs/heroes"
	"fd/libs/logger"
	"fd/libs/maps"
	"fd/libs/pathfinding"
	"fd/libs/prediction"
	"fd/libs/util"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type attack struct {
	Attacker [3]int //i,j,k
	Defender []int  //i or i,j,k
	Damage   int
	Type     string //miss, hit
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
	SPDTable     [][][]int
	//heroGraph   [][][]*pathfinding.DLiteGraph

	camps      []*heroes.Camp
	accessible []bool

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

	logger *logger.Logger
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
	SPDTable    [][][]int
	Turn        string
	UnitsOfTime float64
	Atk         string
	Camps       []string

	DestinationType  int
	DestinationIndex []int

	UnitIndex [3]int
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
			if m.m.ended {
				continue
			} else if msg.Event == "act" {

				// check if the match has started and the action is to a user's own hero
				if m.m.phase == "TURN" && !m.m.ended && m.m.uuid_to_team_int[m.u.uid].a == msg.UnitIndex[0] && m.m.uuid_to_team_int[m.u.uid].b == msg.UnitIndex[1] {
					// check if the hero is alive and at starting position
					if (m.m.heroes[msg.UnitIndex[0]][msg.UnitIndex[1]][msg.UnitIndex[2]].HP > 0) &&
						((m.m.heroes[msg.UnitIndex[0]][msg.UnitIndex[1]][msg.UnitIndex[2]].Status == 0) ||
							((m.m.heroes[msg.UnitIndex[0]][msg.UnitIndex[1]][msg.UnitIndex[2]].Position == m.m.heroes[msg.UnitIndex[0]][msg.UnitIndex[1]][msg.UnitIndex[2]].Start) &&
								(m.m.heroes[msg.UnitIndex[0]][msg.UnitIndex[1]][msg.UnitIndex[2]].Status == 1))) {

						canSetPath := false
						oldTarget := []int{}
						if msg.DestinationType == 0 {
							if confirmValidPath(m.m.heroes, m.m.camps, msg.UnitIndex[:], msg.DestinationIndex, msg.DestinationType) {
								m.m.heroes[msg.UnitIndex[0]][msg.UnitIndex[1]][msg.UnitIndex[2]].DestinationIndex = []int{msg.DestinationIndex[0]}
								m.m.heroes[msg.UnitIndex[0]][msg.UnitIndex[1]][msg.UnitIndex[2]].DynamicCircle = []float64{}
								m.m.heroes[msg.UnitIndex[0]][msg.UnitIndex[1]][msg.UnitIndex[2]].LastMileRadius = 25
								canSetPath = true
							}
						} else if msg.DestinationType == 1 {
							if confirmValidPath(m.m.heroes, m.m.camps, msg.UnitIndex[:], msg.DestinationIndex, msg.DestinationType) {
								oldTarget = m.m.heroes[msg.UnitIndex[0]][msg.UnitIndex[1]][msg.UnitIndex[2]].DestinationIndex
								var dynamicCircle [2]float64
								copy(dynamicCircle[:], m.m.heroes[msg.DestinationIndex[0]][msg.DestinationIndex[1]][msg.DestinationIndex[2]].Position[:])
								m.m.heroes[msg.UnitIndex[0]][msg.UnitIndex[1]][msg.UnitIndex[2]].DynamicCircle = dynamicCircle[:]
								m.m.heroes[msg.UnitIndex[0]][msg.UnitIndex[1]][msg.UnitIndex[2]].LastMileRadius = 25
								canSetPath = true
							}
						} else {
							fmt.Printf("this route does not exist")
						}

						// setting the path
						if canSetPath {

							m.m.logger.Write("user: " + util.IntsToString(msg.UnitIndex[:]) + "->" + util.IntsToString(msg.DestinationIndex))

							fmt.Println("calling setPath from m_read", msg.UnitIndex)

							var dest [2]float64
							if msg.DestinationType == 0 {
								dest = m.m.camps[msg.DestinationIndex[0]].Position
							} else if msg.DestinationType == 1 {
								dest = m.m.heroes[msg.DestinationIndex[0]][msg.DestinationIndex[1]][msg.DestinationIndex[2]].Position
							}

							AStarGraph := pathfinding.AStarInit(mapList[0], m.m.heroes[msg.UnitIndex[0]][msg.UnitIndex[1]][msg.UnitIndex[2]].Position, dest)
							path := AStarGraph.ComputeShortestPath()
							if len(path) > 0 {
								setPath(m.m.heroes[msg.UnitIndex[0]][msg.UnitIndex[1]][msg.UnitIndex[2]], dest, path, msg.DestinationType, msg.DestinationIndex)
							} else {
								canSetPath = false
							}

						}

						if canSetPath {

							updateDependencyGraph(m.m.heroes[msg.UnitIndex[0]][msg.UnitIndex[1]][msg.UnitIndex[2]], oldTarget, m.m.heroes[msg.UnitIndex[0]][msg.UnitIndex[1]][msg.UnitIndex[2]].DestinationIndex, m.m.dependencies)

							// checking the hunter's distance from the target to decide the pathfinding algorithm
							hardUpdate := make(map[[3]int]int)
							if len(msg.DestinationIndex) <= 1 {
								hardUpdate[msg.UnitIndex] = 0
							} else {
								hunter := m.m.heroes[msg.UnitIndex[0]][msg.UnitIndex[1]][msg.UnitIndex[2]]
								targetIdx := m.m.heroes[msg.UnitIndex[0]][msg.UnitIndex[1]][msg.UnitIndex[2]].DestinationIndex
								target := m.m.heroes[targetIdx[0]][targetIdx[1]][targetIdx[2]]
								dist := maps.EuclideanDistance2(hunter.Position, target.Position)

								if dist <= 1 {
									hardUpdate[msg.UnitIndex] = 2
								} else if dist <= 25 {
									hardUpdate[msg.UnitIndex] = 1
								} else {
									hardUpdate[msg.UnitIndex] = 0
								}
							}

							// running the pathfinding algorithm and updating dependencies
							points := [][3]int{msg.UnitIndex}
							cycle, strand := prediction.GetCyclesAndStrands(m.m.heroes, points, hardUpdate)
							updateDependencies(m.m.heroes, cycle, strand, m.m.dependencies, hardUpdate) // the way that updateDependencies works right now, it updates the unit's path twice (in setPath and updateDependencies)

							m.m.sharepos(nil, 0)

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
							j.incoming_message <- msg
						}
					}

					for _, i := range m.m.spectator_uid_to_msid_to_match_socket {
						for _, j := range i {
							j.incoming_message <- msg
						}
					}
				}

				all_done := true
				for _, j := range m.m.char_sel_done {
					if !j {
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
						j.incoming_message <- msg
					}
				}

				for _, i := range m.m.spectator_uid_to_msid_to_match_socket {
					for _, j := range i {
						j.incoming_message <- msg
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

		// logging stuff
		logFilename := fmt.Sprintf("matchlog_%s.txt", time.Now().Format("20060102_150405"))
		logger, err := logger.NewLogger(logFilename)
		if err != nil {
			log.Fatal(err)
		}

		ans = &match{

			mutex: sync.RWMutex{},

			mid:       pl.plid,
			game_mode: pl.game_mode,
			capacity:  pl.capacity,
			sides:     pl.sides,
			ended:     false,

			state_count: 0,
			heroes:      make([][][]*heroes.Hero, 0),
			//heroGraph:   make([][][]*pathfinding.DLiteGraph, 0, 0),
			dependencies: make(map[[3]int][][3]int),

			uuid_to_team_int: make(map[uuid.UUID]pair),

			camps: make([]*heroes.Camp, 0),

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

			logger: logger,
		}

		// add bots to gamer_uid_to_user
		for _, j := range pl.spectator_uid_to_user {
			if j.bot_status {
				ans.bot_join <- j
			}
		}

		// creating camps
		hardcodedCamps := [][2]float64{
			{190, 420},
			{350, 430},
			{330, 410},
			{240, 110},
			{350, 420},
			{220, 360},
			{280, 370},
			{120, 370},
			{320, 390},
			{180, 100},
			{410, 70},
			{430, 370},
		}
		hardcodedHP := []int{
			70, 30, 70, 66, 100, 32, 5, 36, 66, 54, 84, 25,
		}
		for i := 0; i < len(hardcodedCamps); i++ {
			ans.camps = append(ans.camps, &heroes.Camp{Index: [1]int{i}, MaxHP: hardcodedHP[i], HP: hardcodedHP[i], Position: hardcodedCamps[i], Units: make([][3]int, 0)})
		}

		// random_count := 10 + rand.Intn(5)
		// for i := 0; i < random_count; i++ {
		// 	random_x := float64(rand.Intn(40))
		// 	random_y := float64(rand.Intn(40))
		// 	c := [2]float64{50.0 + (random_x * 10.0), 50.0 + (random_y * 10.0)}
		// 	HP := rand.Intn(100) + 1
		// 	//HP := 1
		// 	ans.camps = append(ans.camps, &heroes.Camp{Index: [1]int{i}, MaxHP: HP, HP: HP, Position: c, Units: make([][3]int, 0)})
		// }

		// check for camp accessiblity
		accessible := make([]bool, len(ans.camps))
		for i := 0; i < len(ans.camps); i++ {
			s := [2]float64{0, 0}
			e := ans.camps[i].Position
			path := findPath(s, e)
			if len(path) > 0 {
				accessible[i] = true
			}
		}
		ans.accessible = accessible

		// set up the game state by: (1) creating each team, (2) assigning clients to each team, (3) creating the basic hero unit for each client
		// (1) creating each team
		for i := 0; i < int(ans.sides); i++ {
			ans.heroes = append(ans.heroes, make([][]*heroes.Hero, 0))
			//ans.heroGraph = append(ans.heroGraph, make([][]*pathfinding.DLiteGraph, 0))
		}

		// (2) and (3) assigning clients and heroes to each team
		team_int := 0
		client_int := 0

		SPDHardcode := []int{72, 39, 62, 99, 5, 72}
		DEFHardcode := []int{29, 54, 29, 45, 38, 9}

		a := 0

		for i, j := range pl.gamer_permission_list {

			if !j.bot_status {
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
				//stats := heroes.RandomBudget(100, 4)
				scale := []int{rand.Intn(10) * 10, rand.Intn(10) * 10, rand.Intn(10) * 10, rand.Intn(10) * 10, rand.Intn(10) * 10}
				temp_coordinate := [2]float64{float64(team_int * 500), 500 - (float64(y) * 100)}
				temp := &heroes.Hero{
					Bot:   j.bot_status,
					Name:  names[rand.Intn(len(names))],
					Index: [3]int{team_int, client_int, y},

					Start:        temp_coordinate,
					Position:     temp_coordinate,
					Destination:  temp_coordinate,
					Path:         make([][2]float64, 0),
					Path_Segment: 0,

					DestinationIndex: []int{},
					DynamicCircle:    []float64{},

					LastMile:       0,
					LastMileRadius: 25,

					DistanceLeft: maps.EuclideanDistance2(temp_coordinate, temp_coordinate),
					Status:       0,

					HP:    100,
					ATK:   50, //stats[0],
					DEF:   DEFHardcode[a],
					AS:    75,             //stats[2],
					SPD:   SPDHardcode[a], //stats[3], rand.Intn(100), SPDHardcode[a]
					Scale: scale,

					Battle: nil,
					TTD:    100.0,
				}
				a += 1
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

		// create speed table
		SPDTable := [][][]int{}
		for i := 0; i < len(ans.heroes); i++ {
			subside := [][]int{}
			for j := 0; j < len(ans.heroes[i]); j++ {
				subteam := []int{}
				for k := 0; k < len(ans.heroes[i][j]); k++ {
					subteam = append(subteam, ans.heroes[i][j][k].SPD)
				}
				subside = append(subside, subteam)
			}
			SPDTable = append(SPDTable, subside)
		}
		ans.SPDTable = SPDTable

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
			if !check_uid {
				m.gamer_uid_to_user[ws.u.uid] = ws.u
				m.gamer_uid_to_msid_to_match_socket[ws.u.uid] = make(map[uuid.UUID]*match_socket)
			}
			m.gamer_uid_to_msid_to_match_socket[ws.u.uid][ws.msid] = ws
			m.mutex.Unlock()

			// if match is not in the user object, add the match and create a match_socket object
			ws.u.mutex.Lock()
			if !check_mid {
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

			if m.ended {
				ws.incoming_message <- &message{Event: "game_over"}
			}

			// communicate to all participants - a new user has joined ( only unique users, not unique sockets )
			if !check_uid {
				msg := &message{Name: ws.u.email, Message: "x entered the chat", Event: "entered", When: time.Now()}

				m.mutex.Lock()
				m.message_logs = append(m.message_logs, msg)
				m.mutex.Unlock()

				fmt.Println("sending:", msg)

				for _, i := range m.gamer_uid_to_msid_to_match_socket {
					for _, j := range i {
						j.incoming_message <- msg
					}
				}

				for _, i := range m.spectator_uid_to_msid_to_match_socket {
					for _, j := range i {
						j.incoming_message <- msg
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
						j.incoming_message <- msg
					}
				}

				for _, i := range m.spectator_uid_to_msid_to_match_socket {
					for _, j := range i {
						j.incoming_message <- msg
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
			if m.ended && len(m.gamer_uid_to_msid_to_match_socket) == 0 {
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
					j.incoming_message <- msg
				}
			}
			for _, i := range m.spectator_uid_to_msid_to_match_socket {
				for _, j := range i {
					j.incoming_message <- msg
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
					j.incoming_message <- msg
				}
			}
			for _, i := range m.spectator_uid_to_msid_to_match_socket {
				for _, j := range i {
					j.incoming_message <- msg
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
				m.ticker = time.NewTicker(time.Until(m.next_time)) //will tick in 30 s

				// send ticker to everyone
				m.mutex.Lock()
				m.message_logs = append(m.message_logs, msg)
				m.mutex.Unlock()

				fmt.Println("sending:", msg)

				for _, i := range m.gamer_uid_to_msid_to_match_socket {
					for _, j := range i {
						msg.Message = m.next_time.Add(j.user_time.Sub(j.system_time)).UTC().String()
						j.incoming_message <- msg
					}
				}

				for _, i := range m.spectator_uid_to_msid_to_match_socket {
					for _, j := range i {
						msg.Message = m.next_time.Add(j.system_time.Sub(j.user_time)).String()
						j.incoming_message <- msg
					}
				}
			}

		case <-m.char_sel_ticker:

			m.phase = "CHARACTER SELECTION"
			msg := &message{Event: "ticker_start", When: time.Now(), MatchID: m.mid}
			msg2 := &message{Event: "update_phase", Phase: m.phase, MatchID: m.mid}
			m.next_time = time.Now().Add(600 * time.Second)
			m.ticker = time.NewTicker(time.Until(m.next_time))

			// send ticker to everyone
			fmt.Println("sending:", msg)

			for _, i := range m.gamer_uid_to_msid_to_match_socket {
				for _, j := range i {
					msg.Message = m.next_time.Add(j.user_time.Sub(j.system_time)).UTC().String()
					j.incoming_message <- msg
					j.incoming_message <- msg2
				}
			}

			for _, i := range m.spectator_uid_to_msid_to_match_socket {
				for _, j := range i {
					msg.Message = m.next_time.Add(j.system_time.Sub(j.user_time)).String()
					j.incoming_message <- msg
					j.incoming_message <- msg2
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
			m.ticker = time.NewTicker(time.Until(m.next_time)) // will tick in 30 s

			// send ticker to everyone
			m.mutex.Lock()
			m.message_logs = append(m.message_logs, msg)
			m.mutex.Unlock()

			fmt.Println("sending:", msg)

			for _, i := range m.gamer_uid_to_msid_to_match_socket {
				for _, j := range i {
					msg.Message = m.next_time.Add(j.user_time.Sub(j.system_time)).UTC().String()
					j.incoming_message <- msg
					j.incoming_message <- msg2
				}
			}

			for _, i := range m.spectator_uid_to_msid_to_match_socket {
				for _, j := range i {
					msg.Message = m.next_time.Add(j.system_time.Sub(j.user_time)).String()
					j.incoming_message <- msg
					j.incoming_message <- msg2
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
					j.incoming_message <- msg
				}
			}

			for _, i := range m.spectator_uid_to_msid_to_match_socket {
				for _, j := range i {
					j.incoming_message <- msg
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
						if m.heroes[i][j][k].Bot && (m.heroes[i][j][k].Status == 0) {
							path, dest, idx := determineBotAction(m.heroes[i][j][k], m.heroes, m.camps, m.accessible, false)
							destinationType := 0
							if len(idx) == 3 {
								destinationType = 1
							}
							if len(path) == 0 || len(idx) == 0 {
								// figure out how to end the game here
							} else {

								m.logger.Write("bot: " + strconv.Itoa(i) + "," + strconv.Itoa(j) + "," + strconv.Itoa(k) + "->" + util.IntsToString(idx))

								setPath(m.heroes[i][j][k], dest, path, destinationType, idx)

								hardUpdate := make(map[[3]int]int)
								if len(idx) == 1 {
									updateDependencyGraph(m.heroes[i][j][k], m.heroes[i][j][k].DestinationIndex, []int{}, m.dependencies)
									hardUpdate[[3]int{i, j, k}] = 0
								} else {
									updateDependencyGraph(m.heroes[i][j][k], m.heroes[i][j][k].DestinationIndex, idx, m.dependencies)

									// get hunter distance from target and determine algorithm
									hunter := m.heroes[i][j][k]
									targetIdx := m.heroes[i][j][k].DestinationIndex
									target := m.heroes[targetIdx[0]][targetIdx[1]][targetIdx[2]]
									dist := maps.EuclideanDistance2(hunter.Position, target.Position)

									if dist <= 1 {
										hardUpdate[[3]int{i, j, k}] = 2
									} else if dist <= 25 {
										hardUpdate[[3]int{i, j, k}] = 1
									} else {
										hardUpdate[[3]int{i, j, k}] = 0
									}
								}

								points := [][3]int{{i, j, k}}
								cycle, strand := prediction.GetCyclesAndStrands(m.heroes, points, hardUpdate)
								updateDependencies(m.heroes, cycle, strand, m.dependencies, hardUpdate)

								gave_bots_acts = true
							}

						}
					}
				}
			}
			if gave_bots_acts {
				m.sharepos(nil, 0)
			}

			m.SPDTable = heroes.SpeedTable(m.heroes, m.dependencies)

			// calculate min unit of time to action
			min_units := 999999.99
			for i := 0; (min_units > 0) && i < len(m.heroes); i++ {
				for j := 0; (min_units > 0) && j < len(m.heroes[i]); j++ {
					for k := 0; (min_units > 0) && k < len(m.heroes[i][j]); k++ {

						// if the unit is ready to attack, set min_units to 0
						if (m.heroes[i][j][k].HP > 0) && (m.heroes[i][j][k].Status == 1) && (m.heroes[i][j][k].LastMile == 2) && (m.heroes[i][j][k].TTD <= 0) {
							min_units = 0
							continue
						}

						// calculate min units of time to destiation
						if (m.heroes[i][j][k].HP > 0) && (m.heroes[i][j][k].Status == 1) && (m.SPDTable[i][j][k] != 0) && (m.heroes[i][j][k].LastMile != 2) && (m.heroes[i][j][k].LastMile != 3) {
							unitsOfTime := toFixed(m.heroes[i][j][k].DistanceLeft/float64(m.SPDTable[i][j][k]), 10)
							if unitsOfTime < min_units {
								min_units = unitsOfTime
							}
						}

						// calculate min units of time to attack reset
						if (m.heroes[i][j][k].HP > 0) && (m.heroes[i][j][k].AS != 0) && (m.heroes[i][j][k].TTD > 0) {
							unitsOfTime := toFixed(m.heroes[i][j][k].TTD/float64(m.heroes[i][j][k].AS), 10)
							if unitsOfTime < min_units {
								min_units = unitsOfTime
							}
						}

					}
				}
			}

			fmt.Println("min", min_units)

			recompute := [][3]int{}
			lastMile := [][3]int{}
			pathDup := [][3]int{}

			for i := 0; (min_units > 0) && i < len(m.heroes); i++ {
				for j := 0; (min_units > 0) && j < len(m.heroes[i]); j++ {
					for k := 0; (min_units > 0) && k < len(m.heroes[i][j]); k++ {

						if m.heroes[i][j][k].HP > 0 && len(m.heroes[i][j][k].DestinationIndex) > 1 {

							hunter := m.heroes[i][j][k]
							hunterSPD := m.SPDTable[i][j][k]
							target := m.heroes[m.heroes[i][j][k].DestinationIndex[0]][m.heroes[i][j][k].DestinationIndex[1]][m.heroes[i][j][k].DestinationIndex[2]]
							targetSPD := m.SPDTable[m.heroes[i][j][k].DestinationIndex[0]][m.heroes[i][j][k].DestinationIndex[1]][m.heroes[i][j][k].DestinationIndex[2]]

							var prox float64
							var entry string
							if hunter.LastMile == 0 {

								// radius check
								radiusCheck := maps.EuclideanDistance2(hunter.Position, target.Position)
								if util.SigDig3(radiusCheck) < hunter.LastMileRadius {
									fmt.Println("weird behavior!!!", i, j, k, radiusCheck, hunter.LastMileRadius)
									panic("should not be less than hunter.LastMileRadius")
								}

								prox, entry = prediction.ProximityCheck(hunter.Path, target.Path, hunter.DistanceLeft, target.DistanceLeft, float64(hunterSPD), float64(targetSPD), hunter.LastMileRadius, "enter")

								if prox >= 0 && prox-min_units <= 1e-6 && entry == "entering" {
									fmt.Println(i, j, k, "replaced min", min_units, "w/ prox (lm0)", prox, entry)
									absDiff := math.Abs(prox - min_units)
									if absDiff <= 1e-6 {
										lastMile = append(lastMile, [3]int{i, j, k})
										min_units = math.Max(min_units, prox)
									} else {
										lastMile = [][3]int{{i, j, k}}
										recompute = [][3]int{}
										min_units = prox
									}
								}

								targetTravelled := float64(targetSPD) * min_units
								if target.DistanceLeft <= 0 {
									targetTravelled = 0
								}
								distanceTravelled := float64(hunterSPD) * min_units

								ans := prediction.CircleIntersectPath(
									target.Path,
									target.Position[0], target.Position[1],
									target.DistanceLeft, target.DistanceLeft-targetTravelled,
									m.heroes[i][j][k].DynamicCircle,
									m.heroes[i][j][k].DistanceLeft, (m.heroes[i][j][k].DistanceLeft - distanceTravelled),
									hunterSPD, targetSPD,
									3.0,
								)
								fmt.Println(i, j, k, "intersect found for unit", m.heroes[i][j][k].DestinationIndex[0], m.heroes[i][j][k].DestinationIndex[1], m.heroes[i][j][k].DestinationIndex[2], "at tick:", ans)

								if ans >= 0 {
									fmt.Println("replaced min ", min_units, "w/ recompute", ans, ans-min_units)
									absDiff := math.Abs(ans - min_units)
									if absDiff <= 1e-6 {
										recompute = append(recompute, [3]int{i, j, k})
										min_units = math.Max(min_units, ans)
									} else if ans-min_units <= 0 {
										recompute = [][3]int{{i, j, k}}
										lastMile = [][3]int{}
										pathDup = [][3]int{}
										min_units = ans
									}
								}

							} else if hunter.LastMile == 1 {

								// radius check
								radiusCheck := maps.EuclideanDistance2(hunter.Position, target.Position)
								if util.SigDig3(radiusCheck) > hunter.LastMileRadius {
									fmt.Println("weird behavior!!!", i, j, k, radiusCheck, hunter.LastMileRadius)
									panic("should not exceed hunter.LastMileRadius (1)")
								}

								prox, entry = prediction.ProximityCheck(hunter.Path, target.Path, hunter.DistanceLeft, target.DistanceLeft, float64(hunterSPD), float64(targetSPD), hunter.LastMileRadius, "exit")

								if prox >= 0 && prox-min_units <= 1e-6 && (entry == "exiting" || entry == "passing") {

									fmt.Println("replace min w/ recompute", min_units, prox, entry)
									absDiff := math.Abs(prox - min_units)
									if absDiff <= 1e-6 {
										recompute = append(recompute, [3]int{i, j, k})
										min_units = math.Max(min_units, prox)
									} else {
										recompute = [][3]int{{i, j, k}}
										lastMile = [][3]int{}
										pathDup = [][3]int{}
										min_units = prox
									}
								}

							} else if hunter.LastMile == 2 {

								dynamicDist := getDistFromDynamic(m.heroes, hunter)
								fmt.Println("dynamicDist", hunter.Index, target.Index, dynamicDist, hunter.Position, target.Position)

								radiusCheck := maps.EuclideanDistance2(hunter.Position, target.Position)
								if util.SigDig3(radiusCheck) > 1 {
									fmt.Println("weird behavior2!!!", i, j, k, radiusCheck)
									panic("should not exceed 1 radius")
								}

								prox, entry = prediction.ProximityCheck(hunter.Path, target.Path, hunter.DistanceLeft, target.DistanceLeft, float64(hunterSPD), float64(targetSPD), 1, "exit")

								if prox >= 0 && prox-min_units <= 1e-6 && entry == "exiting" {
									fmt.Println(i, j, k, "replaced min", min_units, "w/ prox (lm2)", prox, entry, prox-min_units)
									absDiff := math.Abs(prox - min_units)
									if absDiff <= 1e-6 {
										lastMile = append(lastMile, [3]int{i, j, k})
										min_units = math.Max(min_units, prox)
									} else {
										recompute = [][3]int{}
										lastMile = [][3]int{{i, j, k}}
										pathDup = [][3]int{}
										min_units = prox
									}
								}

							} else if hunter.LastMile == 3 {

								intersect := prediction.MeetsPoint(target.Path, hunter.Position, float64(targetSPD))
								if intersect >= 0 && intersect-min_units <= 1e-6 {
									fmt.Println(i, j, k, "replaced min", min_units, "w/ int (lm3)", intersect, intersect-min_units)
									absDiff := math.Abs(intersect - min_units)
									if absDiff <= 1e-6 {
										pathDup = append(pathDup, [3]int{i, j, k})
										min_units = math.Max(min_units, intersect)
									} else {
										recompute = [][3]int{}
										lastMile = [][3]int{}
										pathDup = [][3]int{{i, j, k}}
										min_units = intersect
									}
								}

							}

						}

					}
				}
			}

			fmt.Println("FINAL min", min_units)

			has_user_cmd := false

			printState(m.heroes, m.SPDTable, "before")

			// move all unit positions
			movementChange := 0.0
			for i := 0; i < len(m.heroes); i++ {
				for j := 0; j < len(m.heroes[i]); j++ {
					for k := 0; k < len(m.heroes[i][j]); k++ {
						if (m.heroes[i][j][k].HP > 0) && (m.heroes[i][j][k].Status == 1) {

							if m.heroes[i][j][k].LastMile == 3 {
								continue
							}

							toMove := min_units * float64(m.SPDTable[i][j][k])
							pos_x := m.heroes[i][j][k].Position[0]
							pos_y := m.heroes[i][j][k].Position[1]

							startPos := m.heroes[i][j][k].Position

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
										m.heroes[i][j][k].Position[0] = pos_x
										m.heroes[i][j][k].Position[1] = pos_y
										m.heroes[i][j][k].Path_Segment++
									} else {
										travelX, travelY := getLineSegmentTravelled(pos_x, pos_y, m.heroes[i][j][k].Path[m.heroes[i][j][k].Path_Segment+1], toMove)
										m.heroes[i][j][k].Position[0], m.heroes[i][j][k].Position[1] = travelX, travelY
									}
								}
							}

							// move the unit
							m.heroes[i][j][k].DistanceLeft -= min_units * float64(m.SPDTable[i][j][k])

							if m.heroes[i][j][k].DistanceLeft <= 1e-6 {

								m.heroes[i][j][k].DistanceLeft = 0

								if m.heroes[i][j][k].LastMile == 1 {

									pathDup = append(pathDup, [3]int{i, j, k})
									m.heroes[i][j][k].Path_Segment = len(m.heroes[i][j][k].Path) - 1

								} else if m.heroes[i][j][k].LastMile == 2 {

									m.heroes[i][j][k].Position = m.heroes[i][j][k].Destination
									m.heroes[i][j][k].Path_Segment = len(m.heroes[i][j][k].Path) - 1

								} else {

									if m.heroes[i][j][k].DestinationType == 0 && m.camps[m.heroes[i][j][k].DestinationIndex[0]].HP > 0 {
										m.heroes[i][j][k].Status = 2
									} else {
										m.heroes[i][j][k].Status = 0
										if !m.heroes[i][j][k].Bot {
											has_user_cmd = true
										}
									}
									m.heroes[i][j][k].Position = m.heroes[i][j][k].Destination
									m.heroes[i][j][k].Path_Segment = len(m.heroes[i][j][k].Path) - 1

								}

							}

							endPos := m.heroes[i][j][k].Position
							movementChange += maps.EuclideanDistance2(startPos, endPos)

						}
					}
				}
			}

			//printState(m.heroes, m.SPDTable, "after")

			// simulate all cooldowns
			cooldownChange := 0.0
			for i := 0; i < len(m.heroes); i++ {
				for j := 0; j < len(m.heroes[i]); j++ {
					for k := 0; k < len(m.heroes[i][j]); k++ {

						if m.heroes[i][j][k].HP > 0 {
							cooldown := min_units * float64(m.heroes[i][j][k].AS)
							m.heroes[i][j][k].TTD = m.heroes[i][j][k].TTD - cooldown
							cooldownChange += cooldown

							if m.heroes[i][j][k].TTD <= 1e-6 {
								m.heroes[i][j][k].TTD = 0
							}

						}
					}
				}
			}

			if movementChange > 0 || cooldownChange > 0 {
				m.sharepos(nil, 0)
			}

			// simulate attacks
			dead := make([][]int, 0)
			simulatedAttack := false
			for i := 0; i < len(m.heroes); i++ {
				for j := 0; j < len(m.heroes[i]); j++ {
					for k := 0; k < len(m.heroes[i][j]); k++ {

						if m.heroes[i][j][k].HP > 0 && m.heroes[i][j][k].TTD <= 0 {
							if (m.heroes[i][j][k].Status == 1 && (m.heroes[i][j][k].LastMile == 2 || m.heroes[i][j][k].LastMile == 3)) ||
								(m.heroes[i][j][k].Status == 2) {

								if m.heroes[i][j][k].DestinationType == 0 {

									enemyCamp := m.camps[m.heroes[i][j][k].DestinationIndex[0]]
									if enemyCamp.HP <= 0 { // camp is dead, do not continue
										continue
									}

									// calculate DMG & subtract HP
									ATKDMG := m.heroes[i][j][k].ATK
									if enemyCamp.HP < ATKDMG {
										ATKDMG = enemyCamp.HP
									}
									enemyCamp.HP -= ATKDMG
									m.heroes[i][j][k].TTD = 100
									simulatedAttack = true
									if enemyCamp.HP <= 0 {
										dead = append(dead, m.heroes[i][j][k].DestinationIndex)
									}

									// sharepos w/ attack
									ATK := &attack{Attacker: [3]int{i, j, k}, Defender: m.heroes[i][j][k].DestinationIndex, Damage: ATKDMG, Type: "hit"}
									m.sharepos(ATK, 0)

									// calculate EXP to add
									EXPGain := ATKDMG
									if enemyCamp.HP <= 0 {
										EXPGain += 100
									}
									addEXP(m.heroes[i][j][k], EXPGain)

									// sharepos - EXP
									m.sharepos(nil, 0)

								} else if m.heroes[i][j][k].DestinationType == 1 {

									enemyUnit := m.heroes[m.heroes[i][j][k].DestinationIndex[0]][m.heroes[i][j][k].DestinationIndex[1]][m.heroes[i][j][k].DestinationIndex[2]]
									if enemyUnit.HP <= 0 { // unit is dead, do not continue
										continue
									}

									// calculate DMG & subtract HP
									ATKDMG := m.heroes[i][j][k].ATK - enemyUnit.DEF
									DMGToAttacker := 0
									DMGToDefender := 0
									if ATKDMG > 0 {
										DMGToDefender = min(enemyUnit.HP, ATKDMG)
									} else if ATKDMG < 0 {
										DMGToAttacker = min(m.heroes[i][j][k].HP, -ATKDMG)
									}
									m.heroes[i][j][k].HP -= DMGToAttacker
									enemyUnit.HP -= DMGToDefender
									m.heroes[i][j][k].TTD = 100
									simulatedAttack = true
									if enemyUnit.HP <= 0 {
										dead = append(dead, m.heroes[i][j][k].DestinationIndex)
									}

									// sharepos w/ attack
									ATK := &attack{Attacker: [3]int{i, j, k}, Defender: m.heroes[i][j][k].DestinationIndex, Damage: ATKDMG, Type: "hit"}
									m.sharepos(ATK, 0)

									// TODO: have to play around with how much EXP gain for defender and attacker
									attackerEXPGain := max(m.heroes[i][j][k].ATK, enemyUnit.DEF)
									if enemyUnit.HP <= 0 {
										attackerEXPGain += 100
									}
									addEXP(m.heroes[i][j][k], attackerEXPGain)

									defenderEXPGain := max(m.heroes[i][j][k].ATK, enemyUnit.DEF)
									if m.heroes[i][j][k].HP <= 0 {
										defenderEXPGain += 100
									}
									addEXP(m.heroes[i][j][k], defenderEXPGain)

									m.sharepos(nil, 0)

								}

							}
						}

					}
				}
			}

			// increment the turn
			num, _ := strconv.Atoi(m.turn[5:])
			if simulatedAttack {
				m.turn = "TURN " + strconv.Itoa(num+1)
				m.sharepos(nil, min_units)
			}

			if !simulatedAttack && movementChange <= 0 && cooldownChange <= 0 {
				fmt.Println("no more movement, CD, or attacks")
			}

			// update paths
			allToUpdate := append(append(recompute, lastMile...), pathDup...)
			hardUpdate := make(map[[3]int]int)
			for i := 0; i < len(recompute); i++ {
				hardUpdate[[3]int(recompute[i])] = 0
			}
			for i := 0; i < len(lastMile); i++ {
				hardUpdate[[3]int(lastMile[i])] = 1
			}
			for i := 0; i < len(pathDup); i++ {
				hardUpdate[[3]int(pathDup[i])] = 2
			}
			cycles, strands := prediction.GetCyclesAndStrands(m.heroes, allToUpdate, hardUpdate)
			updateDependencies(m.heroes, cycles, strands, m.dependencies, hardUpdate)

			if len(recompute) > 0 || len(lastMile) > 0 || len(pathDup) > 0 {
				m.sharepos(nil, 0)
			}

			//printState(m.heroes, m.SPDTable, "after2")

			if game_over_check(m.heroes) {
				m.logger.Close()
				has_user_cmd = true
				m.ended = true
				m.ticker.Stop()
				msg := &message{Event: "game_over"}

				for _, i := range m.gamer_uid_to_msid_to_match_socket {
					for _, j := range i {
						j.incoming_message <- msg
					}
				}

				for _, i := range m.spectator_uid_to_msid_to_match_socket {
					for _, j := range i {
						j.incoming_message <- msg
					}
				}
			}

			// determine if camp/unit is dead, reset hunters
			if len(dead) > 0 {

				for i := 0; i < len(m.heroes); i++ {
					for j := 0; j < len(m.heroes[i]); j++ {
						for k := 0; k < len(m.heroes[i][j]); k++ {

							// iterate through dead, if unit's destination is equal to dead, then set to Status 0 (waiting for new input)
							for l := 0; l < len(dead); l++ {

								if util.EqualSlices(dead[l], m.heroes[i][j][k].DestinationIndex) {

									m.heroes[i][j][k].Status = 0
									if !m.heroes[i][j][k].Bot {
										has_user_cmd = true
									}

								}

							}

						}
					}
				}

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
					j.incoming_message <- msg
				}
			}

			for _, i := range m.spectator_uid_to_msid_to_match_socket {
				for _, j := range i {
					msg.Message = init_time.Add(j.system_time.Sub(j.user_time)).String()
					j.incoming_message <- msg
				}
			}

			m.state_count += 1

			if false && !has_user_cmd {
				fmt.Println("concurrent SIMULATE called", m.state_count)
				go func() { m.simulate <- true }()
			}

			if !has_user_cmd && (m.state_count < 8000) {
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

func printState(state [][][]*heroes.Hero, SPDTable [][][]int, period string) {
	for i := 0; i < len(state); i++ {
		for j := 0; j < len(state[i]); j++ {
			for k := 0; k < len(state[i][j]); k++ {
				if len(state[i][j][k].DestinationIndex) > 1 {
					fmt.Println(period, i, j, k, "Status", state[i][j][k].Status, "LM:", state[i][j][k].LastMile, "DestinationIndex", state[i][j][k].DestinationIndex, "DynamicDist:", getDistFromDynamic(state, state[i][j][k]), "Dist:", state[i][j][k].DistanceLeft, "Pos:", state[i][j][k].Position, "SPD", state[i][j][k].SPD, "cSPD", SPDTable[i][j][k], "DynamicCircle", state[i][j][k].DynamicCircle)
				} else {
					fmt.Println(period, i, j, k, "Status", state[i][j][k].Status, "LM:", state[i][j][k].LastMile, "DestinationIndex", state[i][j][k].DestinationIndex, "Dist:", state[i][j][k].DistanceLeft, "Pos:", state[i][j][k].Position, "SPD", state[i][j][k].SPD, "cSPD", SPDTable[i][j][k], "DynamicCircle", state[i][j][k].DynamicCircle)
				}
			}
		}
	}
}

func (m *match) sharepos(a *attack, t float64) {
	fmt.Println("*****************************sharepos*****************************", a)

	atk_temp := ""
	if a != nil {
		marshalled, _ := json.Marshal(a)
		atk_temp = string(marshalled)
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
			j.incoming_message <- &message{Event: "game_state", TCH: temp, Atk: atk_temp, SPDTable: m.SPDTable, Message: m.uuid_to_team_int[k].ab, Turn: m.turn, UnitsOfTime: t, Camps: tempCamp, When: time.Now(), MatchID: m.mid}
		}
	}
	for _, i := range m.spectator_uid_to_msid_to_match_socket {
		for _, j := range i {
			j.incoming_message <- &message{Event: "game_state", TCH: temp, Atk: atk_temp, When: time.Now(), MatchID: m.mid}
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
			if at_least_one_alive {
				break
			}
		}
		if at_least_one_alive {
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
			j.incoming_message <- &message{Event: "assignInt", Message: m.uuid_to_team_int[k].ab, MatchID: m.mid}
		}
	}
}

func getDistFromDynamic(state [][][]*heroes.Hero, unit *heroes.Hero) float64 {

	if len(unit.DestinationIndex) > 1 {

		hunter := unit
		target := state[hunter.DestinationIndex[0]][hunter.DestinationIndex[1]][hunter.DestinationIndex[2]]
		x_dist := target.Position[0] - hunter.Position[0]
		y_dist := target.Position[1] - hunter.Position[1]
		return toFixed(math.Sqrt((y_dist*y_dist)+(x_dist*x_dist)), 10)
	}
	return 0.0

}

func getLineSegmentTravelled(refX float64, refY float64, dest [2]float64, distanceTravelled float64) (float64, float64) {
	if distanceTravelled == 0 {
		return refX, refY
	}
	lineX := dest[0] - refX
	lineY := dest[1] - refY
	lineDist := math.Sqrt((lineX * lineX) + (lineY * lineY)) // calculate remaining segment distance
	percent := (lineDist + distanceTravelled) / lineDist

	return refX + (percent * lineX), refY + (percent * lineY)
}

func setPath(unit *heroes.Hero, dest [2]float64, p [][2]float64, destinationType int, destinationIndex []int) {

	fmt.Println("?!?!?!?!?!?!?!?!?!?!", unit.Start, unit.Position)

	// if the unit's position does not equal the start of the path, recalculate the path
	if (unit.Position[0] != p[0][0]) || (unit.Position[1] != p[0][1]) {

		dist := maps.EuclideanDistance2(unit.Position, p[0])

		if dist < 1e-8 {
			unit.Position = p[0]
		} else if dist <= 1 {
			// to me the best option here would be to calculate A* from unit.Position to the destination
			AStarGraph := pathfinding.AStarInit(mapList[0], unit.Position, dest)
			p = AStarGraph.ComputeShortestPath()
		} else {
			// seems like uses start not position
			// think it has to do with concurrency, may have to add a mutex (perhaps for each player)
			panic("Path and unit position exceed 1 dist!")
		}

	}

	unit.Start = unit.Position
	unit.Destination = dest
	unit.Path = p
	unit.Path_Segment = 0

	unit.DestinationType = destinationType
	unit.DestinationIndex = destinationIndex

	thetas := make([]float64, len(p))
	dist := 0.0
	for x := 0; x+1 < len(p); x++ {
		dist += maps.EuclideanDistance2(p[x], p[x+1])
		thetas[x] = getTheta(p[x], p[x+1])
		if x+1 == len(p)-1 {
			thetas[x+1] = thetas[x]
		}
	}
	unit.Thetas = thetas

	unit.DistanceLeft = dist
	unit.Status = 1
}

// pathType can be 0 (camp) or 1 (unit),
func confirmValidPath(state [][][]*heroes.Hero, camps []*heroes.Camp, index []int, destinationIndex []int, destinationType int) bool {

	// destinationIndex of 0
	if len(destinationIndex) == 0 {
		fmt.Println("returned destinationIndex of 0")
		return false
	}

	if len(index) == len(destinationIndex) {
		different := false
		for i := range index {
			if index[i] != destinationIndex[i] {
				different = true
				break
			}
		}
		if !different {
			fmt.Println("cannot path to itself")
			return false
		}
	}

	// check if the end of the path is a valid destination
	if destinationType == 0 {
		if len(destinationIndex) == 1 &&
			destinationIndex[0] >= 0 && destinationIndex[0] < len(camps) &&
			camps[destinationIndex[0]].HP > 0 {
		} else {
			fmt.Println("camp index does not exist")
			return false
		}
	} else if destinationType == 1 {
		if len(destinationIndex) == 3 &&
			destinationIndex[0] >= 0 && destinationIndex[0] < len(state) &&
			destinationIndex[1] >= 0 && destinationIndex[1] < len(state[destinationIndex[0]]) &&
			destinationIndex[2] >= 0 && destinationIndex[2] < len(state[destinationIndex[0]][destinationIndex[1]]) &&
			state[destinationIndex[0]][destinationIndex[1]][destinationIndex[2]].HP > 0 {

			if index[0] == destinationIndex[0] {
				fmt.Println("cannot path to unit of same team")
				return false
			}

		} else {
			fmt.Println("destination index does not exist")
			return false
		}
	}

	// TODO: shouldn't I be calling A* here?

	return true
}

func determineBotAction(b *heroes.Hero, state [][][]*heroes.Hero, camps []*heroes.Camp, accessible []bool, campOnly bool) ([][2]float64, [2]float64, []int) {

	// get an array of valid camps
	validCamps := []int{}
	for i := 0; i < len(camps); i++ {
		if accessible[i] && camps[i].HP > 0 {
			validCamps = append(validCamps, i)
		}
	}

	// also append the array of valid units
	validUnits := [][3]int{}
	for i := 0; i < len(state); i++ {
		if i != b.Index[0] {
			for j := 0; j < len(state[i]); j++ {
				for k := 0; k < len(state[i][j]); k++ {
					if state[i][j][k].HP > 0 { // no accessibility check for now
						validUnits = append(validUnits, [3]int{i, j, k})
					}
				}
			}
		}
	}

	if campOnly && len(validCamps) > 0 {
		rng := rand.Intn(len(validCamps))
		ans := findPath(b.Position, camps[validCamps[rng]].Position)
		return ans, camps[validCamps[rng]].Position, []int{validCamps[rng]}
	}

	if !campOnly && len(validCamps)+len(validUnits) > 0 {
		rng := rand.Intn(len(validCamps) + len(validUnits))

		if rng < len(validCamps) {
			ans := findPath(b.Position, camps[validCamps[rng]].Position)
			return ans, camps[validCamps[rng]].Position, []int{validCamps[rng]}
		} else {
			unitRNG := rng - len(validCamps)
			target := state[validUnits[unitRNG][0]][validUnits[unitRNG][1]][validUnits[unitRNG][2]]
			ans := findPath(b.Position, target.Position)
			return ans, target.Position, validUnits[unitRNG][:]
		}

	}

	return [][2]float64{}, [2]float64{}, []int{}
}

func findPath(start [2]float64, dest [2]float64) [][2]float64 {

	startTime := time.Now()

	// check if direct path from starting node to target is available
	direct_path := true
	line_segment := []float64{start[0], start[1], dest[0], dest[1]}
	for i := 0; i < len(mapList[0].Obstacles); i++ {
		if mapList[0].Obstacle_type[i] == "BEZIER" {
			if bezier.IntersectBezierLine(mapList[0].Obstacles[i], line_segment) {
				direct_path = false
				break
			}
		}
	}

	if direct_path {
		return [][2]float64{start, dest}
	}

	//ans := aStar(start, dest, start_edges, start_dist, dest_edges)
	AStarGraph := pathfinding.AStarInit(mapList[0], start, dest)
	ans := AStarGraph.ComputeShortestPath()

	elapsed := time.Since(startTime)
	milliseconds := elapsed.Milliseconds()
	fmt.Printf("findPath took %d ms\n", milliseconds)

	return ans
}

func getEdges(c1 [2]float64, LPA *pathfinding.LPAGraph) ([][2]float64, []float64) {
	edges := make([][2]float64, 0)
	dist := make([]float64, 0)

	for key := range LPA.Nodes {
		line_seg := []float64{key[0], key[1], c1[0], c1[1]}
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
			dist = append(dist, maps.EuclideanDistance2(c1, key))
		}
	}

	return edges, dist
}

func getTheta(c1 [2]float64, c2 [2]float64) float64 {
	y := c2[1] - c1[1]
	x := c2[0] - c1[0]
	theta := math.Atan2(y, x)
	if theta < 0 {
		theta = theta + (2 * math.Pi)
	}
	return toFixed(theta, 3)
}

func LPA(start [2]float64, end [2]float64) {
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

func DStarLiteInit(start [2]float64, end [2]float64) *pathfinding.DLiteGraph {

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

	dist := maps.EuclideanDistance2(hunter.Position, target.Position)

	if dist > 1 { // this should break the program
		panic("Collision should not happen! Distance > 1!")
	}

	end := target.Path[len(target.Path)-1]
	newPath := prediction.GetSubPath(target.Path, target.DistanceLeft, 0)
	setPath(hunter, end, newPath, hunter.DestinationType, hunter.DestinationIndex)
	hunter.LastMile = 2
}

func setRecompute(state [][][]*heroes.Hero, idx [3]int) {
	fmt.Println("recompute[x]", idx)
	hunter := state[idx[0]][idx[1]][idx[2]]

	if len(hunter.DestinationIndex) <= 1 {
		hunter.LastMile = 0
		return
	}

	target := state[hunter.DestinationIndex[0]][hunter.DestinationIndex[1]][hunter.DestinationIndex[2]]
	var dynamicCircle [2]float64
	copy(dynamicCircle[:], target.Position[:])
	hunter.DynamicCircle = dynamicCircle[:]
	if hunter.LastMile == 1 {
		hunter.LastMileRadius = 25
	}

	AStarGraph := pathfinding.AStarInit(mapList[0], hunter.Position, target.Position)
	path := AStarGraph.ComputeShortestPath()
	setPath(state[idx[0]][idx[1]][idx[2]], path[len(path)-1], path, hunter.DestinationType, hunter.DestinationIndex)
	hunter.LastMile = 0
}

func setLastMile(state [][][]*heroes.Hero, idx [3]int) {
	fmt.Println("lastMile[x]", idx)

	// calculate the last mile path
	hunter := state[idx[0]][idx[1]][idx[2]]
	target := state[hunter.DestinationIndex[0]][hunter.DestinationIndex[1]][hunter.DestinationIndex[2]]
	targetSPD := heroes.SpeedLookup(state, target)
	path := prediction.LastMilePrediction(mapList[0], target.Path, target.DistanceLeft, targetSPD, hunter.Position, hunter.SPD)

	// check if the last mile path doesn't backtrack (causing a corner case)
	hunterStart := path[0]
	hunterEnd := path[0]
	if len(path) > 1 {
		hunterEnd = path[1]
	}
	targetStart := target.Position
	targetEnd := target.Position
	if target.Path_Segment+1 < len(target.Path) {
		targetEnd = target.Path[target.Path_Segment+1]
	}
	gainingGround := prediction.GainRadius(hunterStart, hunterEnd, targetStart, targetEnd, hunter.SPD, targetSPD)
	distRadius := maps.EuclideanDistance2(hunter.Position, target.Position)

	if target.LastMileRadius-2e-4 < 0 {
		panic("target.LastMileRadius-2e-4 is <0")
	}

	if !gainingGround && hunter.LastMile == 0 && distRadius >= target.LastMileRadius-2e-4 { // we check 2e-4 because the unit is stopped at LastMileRadius-1e-4, we do slightly inside because we have to care about floating point error

		hunter.LastMileRadius = hunter.LastMileRadius * 0.75
		var dynamicCircle [2]float64
		copy(dynamicCircle[:], target.Position[:])
		hunter.DynamicCircle = dynamicCircle[:]
		AStarGraph := pathfinding.AStarInit(mapList[0], hunter.Position, target.Position)
		path := AStarGraph.ComputeShortestPath()
		setPath(state[idx[0]][idx[1]][idx[2]], path[len(path)-1], path, hunter.DestinationType, hunter.DestinationIndex)
		hunter.LastMile = 0

	} else {
		hunter.LastMile = 1
		hunter.DynamicCircle = []float64{}
		hunter.LastMileRadius = 25
		setPath(state[idx[0]][idx[1]][idx[2]], path[len(path)-1], path, hunter.DestinationType, hunter.DestinationIndex)
	}
}

func setPathDup(state [][][]*heroes.Hero, idx [3]int) {
	hunter := state[idx[0]][idx[1]][idx[2]]
	target := state[hunter.DestinationIndex[0]][hunter.DestinationIndex[1]][hunter.DestinationIndex[2]]
	hunter.DynamicCircle = []float64{}

	// if the hunter and target are > 1 dist, then either
	// 1) the hunter has arrived early -> set LastMile to 3 and wait for collision
	// 2) the hunter has a slight rounding error, set LastMile to 1
	dist := getDistFromDynamic(state, hunter)
	if dist > 1 {

		if prediction.IsPointOnPath(target.Path, hunter.Position) { // set LM3
			path := [][2]float64{hunter.Position}
			setPath(hunter, hunter.Position, path, hunter.DestinationType, hunter.DestinationIndex)
			hunter.LastMile = 3
		} else { // set to LM1
			setLastMile(state, hunter.Index)
		}

	} else {
		PathDuplication(hunter, target)
	}
}

func updateDependencyGraph(hunter *heroes.Hero, oldTarget []int, newTarget []int, dependencies map[[3]int][][3]int) {

	hunterIndex := [3]int{hunter.Index[0], hunter.Index[1], hunter.Index[2]}

	// remove from old array
	if len(oldTarget) == 3 {
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
	if len(newTarget) == 3 {
		newKey := [3]int{newTarget[0], newTarget[1], newTarget[2]}
		dependencies[newKey] = append(dependencies[newKey], hunterIndex)
	}

}

func updateStrand(state [][][]*heroes.Hero, point [3]int, graph map[[3]int][][3]int, updated map[[3]int]struct{}, hardUpdate map[[3]int]int) {

	// if we have already updated this strand, skip
	if _, found := updated[point]; found {
		return
	}

	// add the current strand to the queue
	queue := list.New()
	queue.PushBack(point)

	for queue.Len() > 0 {

		current := queue.Front().Value.([3]int)
		queue.Remove(queue.Front())
		currentHunter := state[current[0]][current[1]][current[2]]

		// if we have already updated this, then skip
		if _, found := updated[current]; found {
			continue
		} else {
			updated[current] = struct{}{}
		}

		// if current HP < 0, skip
		if state[current[0]][current[1]][current[2]].HP <= 0 {
			continue
		}

		// update the current node
		notRealUpdate := false // if the LastMile == 0 and is not a hard update, it doesn't update the path, so we don't need to continue

		if status, ok := hardUpdate[current]; ok {

			if status == 0 {
				setRecompute(state, currentHunter.Index)
			} else if status == 1 {
				setLastMile(state, currentHunter.Index)
			} else if status == 2 {
				setPathDup(state, currentHunter.Index)
			}

		} else if currentHunter.DestinationType == 0 {
		} else if currentHunter.LastMile == 0 {
			notRealUpdate = true
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

		// add the current node's dependencies to the queue
		if !notRealUpdate {
			if val, ok := graph[current]; ok {
				for i := 0; i < len(val); i++ {
					queue.PushBack(val[i])
				}
			}
		}

	}

}

func updateDependencies(state [][][]*heroes.Hero, cycle [][3]int, strands [][3]int, graph map[[3]int][][3]int, hardUpdate map[[3]int]int) {

	updated, dependents := updatedCycles(state, cycle, graph, hardUpdate)

	// update cycle dependencies
	for node := range dependents {
		if _, found := updated[node]; !found {
			updateStrand(state, node, graph, updated, hardUpdate)
		}
	}

	// update strands
	for i := 0; i < len(strands); i++ {
		if _, found := updated[strands[i]]; !found {
			updateStrand(state, strands[i], graph, updated, hardUpdate)
		}
	}

}

// could do frechet, or mirror major points, but just going to do simple major point comparison
func comparePathSimilarity(p1 [][2]float64, p2 [][2]float64) bool {

	if len(p1) != len(p2) {
		return false
	}

	it := 0

	for {

		if it >= len(p1) {
			break
		}

		compare := maps.EuclideanDistance2(p1[it], p2[it])
		if math.Abs(compare) > 0.05 { // instead of 1e-6 because it's too precise, we are just looking for near steady state
			return false
		}

		it += 1
	}

	return true

}

// make sure to handle LastMile==3 case
func updatedCycles(state [][][]*heroes.Hero, cyclePoints [][3]int, graph map[[3]int][][3]int, hardUpdate map[[3]int]int) (map[[3]int]struct{}, map[[3]int]struct{}) {

	updated := make(map[[3]int]struct{})
	dependents := make(map[[3]int]struct{})

	completed := make(map[[3]int]bool)
	for i := 0; i < len(cyclePoints); i++ {
		completed[cyclePoints[i]] = false
	}

	for i := 0; i < len(cyclePoints); i++ {

		if completed[cyclePoints[i]] {
			continue
		}

		// identify the cycle
		start := cyclePoints[i]
		current := start
		cycle := [][3]int{}
		for {
			cycle = append(cycle, current)
			if _, found := completed[current]; found {
				completed[current] = true
			}
			next := state[current[0]][current[1]][current[2]].DestinationIndex
			current = [3]int{next[0], next[1], next[2]}
			if current == start {
				break
			}
		}

		it := 0
		first := true

		iterationsCount := 0

		for {

			// iterate backwards, solve path for current, check for change
			currentUnit := state[cycle[it][0]][cycle[it][1]][cycle[it][2]]
			previousPath := currentUnit.Path

			// add the current node and dependents to our updated/dependent map
			if _, exists := updated[cycle[it]]; !exists {
				updated[cycle[it]] = struct{}{}
				dependentList := graph[cycle[it]]
				for j := 0; j < len(dependentList); j++ {
					dependents[dependentList[j]] = struct{}{}
				}
			}

			if status, ok := hardUpdate[cycle[it]]; ok {
				if status == 0 {
					setRecompute(state, cycle[it])
				} else if status == 1 {
					setLastMile(state, cycle[it])
				} else if status == 2 {
					setPathDup(state, cycle[it])
				}
			} else if currentUnit.LastMile == 0 {

			} else if currentUnit.LastMile == 1 {
				setLastMile(state, cycle[it])
			} else if currentUnit.LastMile == 2 {
				setPathDup(state, cycle[it])
			} else if currentUnit.LastMile == 3 {
				dist := getDistFromDynamic(state, currentUnit)
				if dist <= 1 {
					setPathDup(state, cycle[it])
				} else if dist <= 25 {
					setLastMile(state, cycle[it])
				} else {
					setRecompute(state, cycle[it])
				}
			}

			// check if we hit a steady state
			newPath := currentUnit.Path
			if !first && comparePathSimilarity(previousPath, newPath) {
				break
			}
			first = false

			if it == 0 {
				iterationsCount += 1
			}
			it -= 1
			if it < 0 {
				it = len(cycle) - 1
			}

			if iterationsCount >= 10 { // we encountered an oscillation and will not hit steady-state, let's brute force this
				bruteForceCycleSolver(state, cycle)
			}
		}

	}

	return updated, dependents

}

func bruteForceCycleSolver(state [][][]*heroes.Hero, cycle [][3]int) {

	selectedUnit := [3]int{-1, -1, -1}
	broken := false
	for i := 0; i < len(state) && !broken; i++ {
		for j := 0; j < len(state[i]) && !broken; j++ {
			for k := 0; k < len(state[i][j]) && !broken; k++ {
				if state[i][j][k].HP > 0 && state[i][j][k].Status == 1 && state[i][j][k].LastMile == 1 {
					selectedUnit = [3]int{i, j, k}
					broken = true
				}
			}
		}
	}

	if selectedUnit[0] < 0 || selectedUnit[1] < 0 || selectedUnit[2] < 0 {
		panic("this should not happen, there should be at least 1 LM1 unit right?")
	}

	// take the endpoint this unit is going to and path all units in the cycle to that endpoint
	selectedPath := state[selectedUnit[0]][selectedUnit[1]][selectedUnit[2]].Path
	selectedEndpoint := selectedPath[len(selectedPath)-1]

	for i := 0; i < len(cycle); i++ {
		current := state[cycle[i][0]][cycle[i][1]][cycle[i][2]]

		if current.HP > 0 && current.Status == 1 && (current.LastMile == 1 || current.LastMile == 3) {
			AStarGraph := pathfinding.AStarInit(mapList[0], current.Position, selectedEndpoint)
			path := AStarGraph.ComputeShortestPath()
			setPath(current, selectedEndpoint, path, current.DestinationType, current.DestinationIndex)
			current.LastMileRadius = 25
		}
	}

	for i := 0; i < len(cycle); i++ {
		current := state[cycle[i][0]][cycle[i][1]][cycle[i][2]]

		if current.HP > 0 && current.Status == 1 && current.LastMile == 2 {
			setPathDup(state, cycle[i])
		}
	}

}

func addEXP(unit *heroes.Hero, EXP int) {

	startEXP := unit.EXP
	unit.EXP += EXP

	levels := int(math.Floor(float64(unit.EXP)/100.0) - math.Floor(float64(startEXP)/100.0))
	for i := 0; i < levels; i++ {
		continue
		heroes.LevelUp(unit, 20)
	}

}
