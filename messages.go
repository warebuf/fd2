package main

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"sync"
	"time"
)

func newRoom(name string) bool {

	fmt.Println("called 'newRoom'")

	roomname_to_rid.mutex.RLock()
	_, found2 := roomname_to_rid.name_to_rid[name]
	roomname_to_rid.mutex.RUnlock()

	if found2 {
		fmt.Println(name, "already exists, was not added to global struct")
		return false
	}

	// create a global room for users to chat in
	not_assigned := true
	for not_assigned {
		random_number := uuid.New()

		rid_to_room.mutex.RLock()
		_, found := rid_to_room.rooms[random_number]
		rid_to_room.mutex.RUnlock()
		if !found {

			temp := &room{

				mutex: sync.RWMutex{},

				rid:       random_number,
				room_name: name,

				broadcast: make(chan *message),
				join:      make(chan *socket),
				leave:     make(chan *socket),

				uid_to_sid_to_socket: make(map[uuid.UUID]map[uuid.UUID]*socket),
				uid_to_user:          make(map[uuid.UUID]*user),

				chat_logs: make([]*message, 0, 16),

				open: make(chan bool),
			}

			rid_to_room.mutex.Lock()
			rid_to_room.rooms[random_number] = temp
			rid_to_room.mutex.Unlock()
			roomname_to_rid.mutex.Lock()
			roomname_to_rid.name_to_rid[name] = random_number
			roomname_to_rid.mutex.Unlock()
			not_assigned = false
			fmt.Println("created RID:", random_number)
			go rid_to_room.rooms[random_number].run()
		}
	}

	return true
}

// creates a websocket connection
func (r *room) ServeHTTP(w http.ResponseWriter, req *http.Request) {

	log.Println("/creating a ws")

	session, err := store.Get(req, "session-name")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var upgrader = &websocket.Upgrader{ReadBufferSize: 1024, WriteBufferSize: 1024}
	sock, err := upgrader.Upgrade(w, req, nil)
	//sock.SetReadLimit(128000) //have to handle this read limit correctly
	if err != nil {
		log.Fatal("ServeHTTP:", err)
		return
	}

	fmt.Println("UUID:", session.Values["uid"])
	uid, _ := uuid.Parse(session.Values["uid"].(string)) // convert string type to UUID type

	uid_to_user.mutex.RLock()
	requesting_user := uid_to_user.users[uid]
	uid_to_user.mutex.RUnlock()

	not_assigned := true
	var random_sid uuid.UUID
	for not_assigned {
		random_sid = uuid.New()
		sid_to_sock.mutex.RLock()
		_, found := sid_to_sock.sid_to_sock[random_sid]
		sid_to_sock.mutex.RUnlock()
		if !found {
			not_assigned = false

			temp := &socket{
				mutex:            sync.RWMutex{},
				socket:           sock,
				sid:              random_sid,
				u:                requesting_user,
				r:                r,
				incoming_message: make(chan *message),
				open:             true,
			}
			r.join <- temp

			go uid_to_user.users[uid].write(temp)
			go uid_to_user.users[uid].read(temp)
		}
	}
}

// this function runs forever, but if we run as a Go routine, it will run in the background
// checks all 3 channels, runs the code if one hits
func (r *room) run() {
	for {
		select {
		case ws := <-r.join: // joining

			// add ws to global socket data structure
			sid_to_sock.mutex.Lock()
			sid_to_sock.sid_to_sock[ws.sid] = ws
			sid_to_sock.mutex.Unlock()

			// add ws to room object
			rid_to_room.mutex.RLock()
			r.mutex.RLock()
			_, check := r.uid_to_user[ws.u.uid]
			r.mutex.RUnlock()
			rid_to_room.mutex.RUnlock()
			rid_to_room.mutex.Lock()
			r.mutex.Lock()
			if check == false {
				r.uid_to_user[ws.u.uid] = ws.u
				r.uid_to_sid_to_socket[ws.u.uid] = make(map[uuid.UUID]*socket)
			}
			r.uid_to_sid_to_socket[ws.u.uid][ws.sid] = ws
			rid_to_room.mutex.Unlock()
			r.mutex.Unlock()

			// add ws to user object
			uid_to_user.mutex.RLock()
			ws.u.mutex.RLock()
			_, check = ws.u.rid_to_room[r.rid]
			uid_to_user.mutex.RUnlock()
			ws.u.mutex.RUnlock()
			uid_to_user.mutex.Lock()
			ws.u.mutex.Lock()
			if check == false {
				ws.u.rid_to_room[r.rid] = r
				ws.u.rid_to_sid_to_socket[r.rid] = make(map[uuid.UUID]*socket)
			}
			ws.u.rid_to_sid_to_socket[r.rid][ws.sid] = ws
			ws.u.mutex.Unlock()
			uid_to_user.mutex.Unlock()

			go func(m []*message, w *socket, c bool, email string) {

				for i, j := range m {
					w.incoming_message <- j
					fmt.Println(i, j.Message)
				}

				if check == false {
					msg := &message{Name: email, Message: "x entered the chat", Event: "entered", When: time.Now()}
					w.r.broadcast <- msg
				}

			}(r.chat_logs, ws, check, ws.u.email)

			fmt.Println("a socket has joined the room")
		case ws := <-r.leave: // leaving
			fmt.Println("a socket has left the room")

			check := true

			// remove rid from user object
			uid_to_user.mutex.Lock()
			ws.u.mutex.Lock()
			delete(ws.u.rid_to_sid_to_socket[ws.r.rid], ws.sid)
			if len(ws.u.rid_to_sid_to_socket[ws.r.rid]) == 0 {
				delete(ws.u.rid_to_sid_to_socket, ws.r.rid)
				delete(ws.u.rid_to_room, ws.r.rid)
				check = false
			}
			ws.u.mutex.Unlock()
			uid_to_user.mutex.Unlock()

			// remove uid from room object
			rid_to_room.mutex.Lock()
			ws.r.mutex.Lock()
			delete(ws.r.uid_to_sid_to_socket[ws.u.uid], ws.sid)
			if len(ws.r.uid_to_sid_to_socket[ws.u.uid]) == 0 {
				delete(ws.r.uid_to_sid_to_socket, ws.u.uid)
				delete(ws.r.uid_to_user, ws.u.uid)
			}
			ws.r.mutex.Unlock()
			rid_to_room.mutex.Unlock()

			// broadcast to all users of the room that the user has left
			go func(r *room, c bool, email string) {

				fmt.Println("left the chat entered")

				if check == false {
					msg := &message{Name: email, Message: "x left the chat", Event: "left", When: time.Now()}
					r.broadcast <- msg
				}

			}(r, check, ws.u.email)

			// remove/take care of socket object
			sid_to_sock.mutex.Lock()
			ws.u = nil
			ws.r = nil
			close(ws.incoming_message)
			ws.socket.Close()
			ws.open = false
			delete(sid_to_sock.sid_to_sock, ws.sid)
			sid_to_sock.mutex.Unlock()

		case msg := <-r.broadcast: // forward message to all clients

			r.mutex.Lock()
			r.chat_logs = append(r.chat_logs, msg)
			r.mutex.Unlock()

			fmt.Println(msg)

			for _, i := range r.uid_to_sid_to_socket {
				for _, j := range i {
					select {
					case j.incoming_message <- msg:
					}
				}
			}

		case <-r.open:
			fmt.Printf("deleting the room\n")

			// go through each user and delete their sockets
			uid_to_user.mutex.Lock()
			for uid, _ := range r.uid_to_sid_to_socket {
				uid_to_user.users[uid].mutex.Lock()
				delete(uid_to_user.users[uid].rid_to_sid_to_socket, r.rid)
				delete(uid_to_user.users[uid].rid_to_room, r.rid)
				uid_to_user.users[uid].mutex.Unlock()
			}
			uid_to_user.mutex.Unlock()

			// delete all sockets out of global utility map
			sid_to_sock.mutex.Lock()
			for _, socket_map := range r.uid_to_sid_to_socket {
				for sid, ws := range socket_map {
					ws.socket.Close()
					ws.open = false
					close(ws.incoming_message)
					delete(sid_to_sock.sid_to_sock, sid)
				}
			}
			sid_to_sock.mutex.Unlock()

			// delete room name out of global utility map
			roomname_to_rid.mutex.Lock()
			delete(roomname_to_rid.name_to_rid, r.room_name)
			roomname_to_rid.mutex.Unlock()

			rid_to_room.mutex.Lock()
			//close(r.broadcast)
			//close(r.join)
			//close(r.leave)
			//close(r.open)

			delete(rid_to_room.rooms, r.rid)
			rid_to_room.mutex.Unlock()

			return
		}
	}
}

// don't see why this needs to be running infinitely, can just call a goroutine when needed
// maybe having two goroutines using the WriteJSON function at the same time isn't the smartest
func (u *user) write(s *socket) {
	for msg := range s.incoming_message { // special trait of a channel, will block until something is in the channel or it is closed
		if err := s.socket.WriteJSON(msg); err != nil {
			fmt.Println("error writing to socket")
			break
		} else {
			fmt.Println("successful write")
		}
	}
}

func (u *user) read(s *socket) {
	defer func() {
		fmt.Printf("read socket closing: %+v\n", s)
		if s.open {
			s.r.leave <- s
			fmt.Println("client left the room", s)
		}
	}()

	for {
		var msg *message
		if err := s.socket.ReadJSON(&msg); err == nil {
			fmt.Printf("JSON: %+v\n", msg)

			msg.When = time.Now()
			msg.Name = u.email
			fmt.Println("message~", msg.Name, s.r.room_name, ": ", msg.Message)

			if msg.Event == "ping" {
				fmt.Println("ping!")
			} else if msg.Event == "hello" {
				s.r.broadcast <- msg
			}
		} else {
			fmt.Println("error reading from socket")
			break
		}
	}
}
