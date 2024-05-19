package main

import (
	"log"
	"net/http"
	"net/url"
	//"os"
	"strings"

	"fmt"
	"github.com/google/uuid" // https://github.com/google/uuid
	"github.com/gorilla/sessions"
	"github.com/gorilla/websocket"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/markbates/goth/providers/google"
	"html/template"
	"path/filepath"
	"sync"
	"time"
)

type user struct {
	mutex sync.RWMutex

	uid uuid.UUID

	email        string
	nickname     string
	currency     int
	admin_status bool

	rid_to_sid_to_socket map[uuid.UUID]map[uuid.UUID]*socket
	rid_to_room          map[uuid.UUID]*room
}

type room struct {
	mutex sync.RWMutex

	rid       uuid.UUID
	room_name string

	broadcast chan *message // a channel is a thread-safe queue, incoming messages
	join      chan *socket  // a channel for clients wishing to join
	leave     chan *socket  // a channel for clients wishing to leave

	uid_to_sid_to_socket map[uuid.UUID]map[uuid.UUID]*socket
	uid_to_user          map[uuid.UUID]*user

	chat_logs []*message

	open chan bool
}

type socket struct {
	mutex sync.RWMutex

	socket *websocket.Conn

	sid uuid.UUID

	u *user
	r *room

	incoming_message chan *message // send is a channel on which messages are sent.

	open bool
}

type message struct {
	Name    string
	Message string
	When    time.Time
	Event   string
}

// Global Variables - Utility
var store *sessions.CookieStore // used to create and decrypt cookies
var port string

// GLOBAL DATA STRUCTURES
type userbase struct {
	users map[uuid.UUID]*user
	mutex sync.RWMutex
}

type roombase struct {
	rooms map[uuid.UUID]*room
	mutex sync.RWMutex
}

// QUICK LOOK UP
type rname_to_rid struct {
	name_to_rid map[string]uuid.UUID
	mutex       sync.RWMutex
}
type sid_to_socket struct {
	sid_to_sock map[uuid.UUID]*socket
	mutex       sync.RWMutex
}

var uid_to_user userbase
var rid_to_room roombase
var roomname_to_rid rname_to_rid
var sid_to_sock sid_to_socket

func main() {

	// making GLOBAL data structures
	uid_to_user.users = make(map[uuid.UUID]*user)
	uid_to_user.mutex = sync.RWMutex{}
	rid_to_room.rooms = make(map[uuid.UUID]*room)
	rid_to_room.mutex = sync.RWMutex{}

	// making QUICK LOOKUP data structures
	roomname_to_rid.name_to_rid = make(map[string]uuid.UUID)
	roomname_to_rid.mutex = sync.RWMutex{}
	sid_to_sock.sid_to_sock = make(map[uuid.UUID]*socket)
	sid_to_sock.mutex = sync.RWMutex{}

	// IF HEROKU BUILD
	//goth.UseProviders(google.New(
	//	"676118167957-4bpa000p9bfaf5vu6halmen6nfjnuo1r.apps.googleusercontent.com",
	//	"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",
	//	"https://testfdfdfd-504b74ad04fd.herokuapp.com/callback",
	//))
	
	goth.UseProviders(google.New(
		"676118167957-4bpa000p9bfaf5vu6halmen6nfjnuo1r.apps.googleusercontent.com",
		"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",
		"http://ec2-3-21-28-93.us-east-2.compute.amazonaws.com:3000/callback",
	))


	//port = os.Getenv("PORT")
	port = "3000"
	// ALSO TURN :WS TO :WSS IN CHAT.HTML

	/*
		// IF LOCAL BUILD
		goth.UseProviders(google.New("676118167957-4bpa000p9bfaf5vu6halmen6nfjnuo1r.apps.googleusercontent.com", "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", "http://localhost:3000/callback"))
		port = "3000"
		// REMEMBER TO CHANGE TO WS IN CHAT.HTML
	*/

	// init
	secret_key := []byte("secret-password") // used to encrypt cookies, have to use a stronger password, but used a simple one for illustration purposes
	store = sessions.NewCookieStore(secret_key)
	store.Options.HttpOnly = true // just to make sure javascript doesn't access cookies
	gothic.Store = store          // not sure why we do this, but if we do, gothic will handle all cookie authentication

	newRoom("global")

	mux := http.NewServeMux() // create the default multiplexer
	fs := http.FileServer(http.Dir("./static"))
	mux.Handle("/static/", http.StripPrefix("/static/", fs))

	mux.HandleFunc("/game", gameHandler)

	mux.HandleFunc("/chat/", chatHandler)
	mux.HandleFunc("/lobby", lobbyHandler)
	mux.HandleFunc("/room", roomHandler)
	mux.HandleFunc("/createRoom", createRoomHandler)
	mux.HandleFunc("/deleteRoom", deleteRoomHandler)

	mux.HandleFunc("/callback", callbackHandler)
	mux.HandleFunc("/login", loginHandler)
	mux.HandleFunc("/logout/google", logoutHandler)

	mux.HandleFunc("/success", successHandler)
	mux.HandleFunc("/", indexHandler)

	// run the server
	log.Println("listening on localhost:" + port)
	log.Fatal(http.ListenAndServe(":"+port, mux))
}

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

			// addd ws to global socket data structure
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
					fmt.Println("set bool false")
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

func printUserBase() {
	println("printing userbase")
	uid_to_user.mutex.RLock()
	for i, j := range uid_to_user.users {
		j.mutex.RLock()
		println(i.String(), j.email)
		j.mutex.RUnlock()
	}
	uid_to_user.mutex.RUnlock()
}
func printRoomBase() {
	println("printing roombase")
	rid_to_room.mutex.RLock()
	for i, j := range rid_to_room.rooms {
		j.mutex.RLock()
		println(i.String(), j.room_name)
		j.mutex.RUnlock()
	}
	rid_to_room.mutex.RUnlock()
}

func (u *user) read(s *socket) {
	defer func() {
		fmt.Printf("read socket closing: %+v\n", s)
		if s.open {
			fmt.Println("should be false here when closing room w/ user in it")
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

// don't see why this needs to be running infinitely, can just call a goroutine when needed
// maybe having two goroutines using the WriteJSON function at the same time isn't the smartest
func (u *user) write(s *socket) {
	for msg := range s.incoming_message { // special trait of a channel, will block until something is in the channel or it is closed
		if err := s.socket.WriteJSON(msg); err != nil {
			fmt.Println("error writing to socket")
			fmt.Println(err)
			break
		} else {
			fmt.Println("successful write")
		}
	}
}

func indexHandler(res http.ResponseWriter, req *http.Request) {
	log.Println("/")
	t := template.Must(template.ParseFiles(filepath.Join("static", "index.html")))
	t.Execute(res, nil)
}

func loginHandler(res http.ResponseWriter, req *http.Request) {
	fmt.Println("/login (Redirected to Google)")

	session, err := store.Get(req, "session-name")
	fmt.Println("session values: ", session.Values)
	fmt.Println("session id: ", session.ID)
	if err != nil {
		fmt.Println("error here")
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}
	if auth, ok := session.Values["authenticated"].(bool); ok && auth {
		log.Println("1")
		t := template.Must(template.ParseFiles(filepath.Join("static", "user.html")))
		session, err := store.Get(req, "session-name")
		if err != nil {
			http.Error(res, err.Error(), http.StatusInternalServerError)
			return
		}

		uid, _ := uuid.Parse(session.Values["uid"].(string))

		user_object := &user{
			mutex: sync.RWMutex{},

			uid:          uid,
			email:        session.Values["Email"].(string),
			nickname:     "bob",
			currency:     0,
			admin_status: false,

			rid_to_sid_to_socket: make(map[uuid.UUID]map[uuid.UUID]*socket),
			rid_to_room:          make(map[uuid.UUID]*room),
		}

		uid_to_user.mutex.Lock()
		uid_to_user.users[uid] = user_object
		uid_to_user.mutex.Unlock()

		log.Println("2")

		var user2 = &struct {
			Email        string
			AvatarURL    string
			UserID       string
			AccessToken  string
			ExpiresAt    string
			RefreshToken string
		}{
			Email:        session.Values["Email"].(string),
			AvatarURL:    session.Values["AvatarURL"].(string),
			UserID:       session.Values["UserID"].(string),
			AccessToken:  session.Values["AccessToken"].(string),
			ExpiresAt:    session.Values["ExpiresAt"].(string),
			RefreshToken: session.Values["RefreshToken"].(string),
		}

		log.Println("3")
		t.Execute(res, user2)
	} else {
		gothic.BeginAuthHandler(res, req)
	}
}

func callbackHandler(res http.ResponseWriter, req *http.Request) {

	fmt.Println("/callback")

	user2, err := gothic.CompleteUserAuth(res, req)
	if err != nil {
		fmt.Fprintln(res, err)
		return
	}

	t := template.Must(template.ParseFiles(filepath.Join("static", "user.html")))

	session, err := store.Get(req, "session-name")
	if err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}

	not_assigned := true
	var random_number uuid.UUID
	for not_assigned {
		random_number = uuid.New()
		uid_to_user.mutex.Lock()
		if _, found := uid_to_user.users[random_number]; !found {
			uid_to_user.mutex.Unlock()
			not_assigned = false
			fmt.Println("assigning UID:", random_number)
		}
	}

	session.Values["Email"] = user2.Email
	session.Values["authenticated"] = true
	session.Values["uid"] = random_number.String()

	session.Values["UserID"] = user2.UserID
	session.Values["AvatarURL"] = user2.AvatarURL
	session.Values["AccessToken"] = user2.AccessToken
	session.Values["ExpiresAt"] = user2.ExpiresAt.String()
	session.Values["RefreshToken"] = user2.RefreshToken

	fmt.Println("session val:", session.Values)
	fmt.Println("session ID:", session.ID)
	err = session.Save(req, res)
	if err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}

	user_object := &user{
		mutex: sync.RWMutex{},

		uid:          random_number,
		email:        user2.Email,
		nickname:     "bob",
		currency:     0,
		admin_status: false,

		rid_to_sid_to_socket: make(map[uuid.UUID]map[uuid.UUID]*socket),
		rid_to_room:          make(map[uuid.UUID]*room),
	}

	uid_to_user.mutex.Lock()
	uid_to_user.users[random_number] = user_object
	uid_to_user.mutex.Unlock()

	fmt.Printf("user object: %+v\n", uid_to_user.users[random_number])
	fmt.Printf("OAUTH object: %+v\n", user2)

	t.Execute(res, user2)
}

func successHandler(res http.ResponseWriter, req *http.Request) {
	// Check if user is already authenticated
	session, err := store.Get(req, "session-name")
	fmt.Println(session)
	if err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}
	if auth, ok := session.Values["authenticated"].(bool); ok && auth {
		res.Write([]byte("hi " + session.Values["Email"].(string)))
	} else {
		//fmt.Println("User is not authenticated, redirecting to home page")
		http.Redirect(res, req, "/", http.StatusSeeOther)
	}
}

func chatHandler(res http.ResponseWriter, req *http.Request) {

	fmt.Println("/chat (joining the chat room)")
	// Check if user is already authenticated
	session, err := store.Get(req, "session-name")
	//fmt.Printf("session: %+v\n", session)
	if err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}
	if auth, ok := session.Values["authenticated"].(bool); ok && auth {

		// have the break down URL
		fmt.Println("URL: ", req.URL.Path)
		parts := strings.Split(req.URL.Path, "/")

		if len(parts) >= 3 {
			fmt.Println("parts: ", parts)
			if rid, err := uuid.Parse(parts[2]); err == nil {

				rid_to_room.mutex.RLock()
				_, ok := rid_to_room.rooms[rid]
				rid_to_room.mutex.RUnlock()

				if ok {
					t := template.Must(template.ParseFiles(filepath.Join("static", "chat.html")))

					data := map[string]string{"email": session.Values["Email"].(string), "host": req.Host, "room_id": parts[2]}
					t.Execute(res, data)
				} else {
					session.Values["Room_error"] = 6
					err = session.Save(req, res)
					if err != nil {
						http.Error(res, err.Error(), http.StatusInternalServerError)
						return
					}
					http.Redirect(res, req, "/lobby", http.StatusFound)
				}

			} else {
				http.Redirect(res, req, "/lobby", http.StatusSeeOther)
			}
		} else {
			http.Redirect(res, req, "/lobby", http.StatusSeeOther)
		}

	} else {
		//fmt.Println("User is not authenticated, redirecting to home page")
		http.Redirect(res, req, "/", http.StatusSeeOther)
	}
}

func logoutHandler(res http.ResponseWriter, req *http.Request) {
	log.Println(("/logout/google"))
	gothic.Logout(res, req)
	// TODO: have to get rid of the session here, maybe even screw with the client's cookie, but that's  not as important as removing the session
	res.Header().Set("Location", "/")
	res.WriteHeader(http.StatusTemporaryRedirect)
}

func roomHandler(res http.ResponseWriter, req *http.Request) {
	log.Println("/room (just a wrapper function to call the ws creating function)")

	u := req.URL
	parameters, err := url.ParseQuery(u.RawQuery)
	fmt.Println(u, parameters, err)

	if err != nil {
		// vulnerability here
	} else {
		fmt.Println(parameters["rid"][0])
		i, err := uuid.Parse(parameters["rid"][0])
		if err != nil {
			// ... handle error
			panic(err)
		} else {
			rid_to_room.rooms[i].ServeHTTP(res, req)
		}
	}
}

func gameHandler(res http.ResponseWriter, req *http.Request) {
	//c, _ := req.Cookie("session-name")

	// Check if user is already authenticated
	session, err := store.Get(req, "session-name")
	fmt.Println(session)
	if err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}
	if auth, ok := session.Values["authenticated"].(bool); ok && auth {
		t := template.Must(template.ParseFiles(filepath.Join("static", "game.html")))
		t.Execute(res, req)
	} else {
		fmt.Println("User is not authenticated, redirecting to home page")
		http.Redirect(res, req, "/", http.StatusSeeOther)
	}
}

func lobbyHandler(res http.ResponseWriter, req *http.Request) {
	fmt.Println("/lobby")

	session, err := store.Get(req, "session-name")
	//fmt.Println("session values: ", session.Values)
	if err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}
	if auth, ok := session.Values["authenticated"].(bool); ok && auth {

		data := struct {
			Room_error string
			Table      map[string]uuid.UUID
		}{}

		if session.Values["Room_error"] != 0 {
			switch session.Values["Room_error"] {
			case 1:
				data.Room_error = "this room exists already"
			case 2:
				data.Room_error = "error retrieving this room name"
			case 3:
				data.Room_error = "cannot have no name"
			case 4:
				data.Room_error = "cannot delete this room"
			case 5:
				data.Room_error = "please do not delete global"
			case 6:
				data.Room_error = "this room no longer exists"
			}
			session.Values["Room_error"] = 0
			err = session.Save(req, res)
			if err != nil {
				http.Error(res, err.Error(), http.StatusInternalServerError)
				return
			}
		}

		data.Table = roomname_to_rid.name_to_rid

		t := template.Must(template.ParseFiles(filepath.Join("static", "lobby.html")))
		t.Execute(res, data)
	} else {
		fmt.Println("User is not authenticated, redirecting to home page")
		http.Redirect(res, req, "/", http.StatusSeeOther)
	}
}

// TODO: note that allowing the user to create their own room name is a vulnerability
func createRoomHandler(res http.ResponseWriter, req *http.Request) {
	fmt.Println("/createRoom")

	fmt.Printf("%+v\n", store)

	session, err := store.Get(req, "session-name")
	if err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}

	u := req.URL
	parameters, err := url.ParseQuery(u.RawQuery)

	if err != nil {

		session.Values["Room_error"] = 2
		err = session.Save(req, res)
		if err != nil {
			http.Error(res, err.Error(), http.StatusInternalServerError)
			return
		}

		http.Redirect(res, req, "/lobby", http.StatusFound)
	} else {
		fmt.Println(parameters["room_name"][0]) // vulnerability here, need to check if this exists

		roomname_to_rid.mutex.Lock()
		_, ok := roomname_to_rid.name_to_rid[parameters["room_name"][0]]
		roomname_to_rid.mutex.Unlock()

		if ok {
			session.Values["Room_error"] = 1
			err = session.Save(req, res)
			if err != nil {
				http.Error(res, err.Error(), http.StatusInternalServerError)
				return
			}
			http.Redirect(res, req, "/lobby", http.StatusFound)
		} else {
			// TODO: probably have to do more string cleanup here
			if len(strings.TrimSpace(parameters["room_name"][0])) == 0 {
				session.Values["Room_error"] = 3
				err = session.Save(req, res)
				if err != nil {
					http.Error(res, err.Error(), http.StatusInternalServerError)
					return
				}
			} else {
				newRoom(parameters["room_name"][0])
			}

		}
		http.Redirect(res, req, "/lobby", http.StatusFound)
	}
}

func deleteRoomHandler(res http.ResponseWriter, req *http.Request) {

	session, err := store.Get(req, "session-name")
	if err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}

	u := req.URL
	parameters, err := url.ParseQuery(u.RawQuery)
	fmt.Println(u)

	if err != nil {
		session.Values["Room_error"] = 2
		err = session.Save(req, res)
		if err != nil {
			http.Error(res, err.Error(), http.StatusInternalServerError)
			return
		}
		http.Redirect(res, req, "/lobby", http.StatusFound)
	} else {

		fmt.Println(parameters["room_name"][0]) // vulnerability here, need to check if this exists

		roomname_to_rid.mutex.Lock()
		rid, ok := roomname_to_rid.name_to_rid[parameters["room_name"][0]]
		roomname_to_rid.mutex.Unlock()
		if !ok {
			session.Values["Room_error"] = 4
			err = session.Save(req, res)
			if err != nil {
				http.Error(res, err.Error(), http.StatusInternalServerError)
				return
			}
			http.Redirect(res, req, "/lobby", http.StatusFound)
		} else {
			if parameters["room_name"][0] == "global" {
				session.Values["Room_error"] = 5
				err = session.Save(req, res)
				if err != nil {
					http.Error(res, err.Error(), http.StatusInternalServerError)
					return
				}
				http.Redirect(res, req, "/lobby", http.StatusFound)
			} else {
				fmt.Println("1")
				rid_to_room.rooms[rid].open <- false
				http.Redirect(res, req, "/lobby", http.StatusFound)
			}

		}
	}
}
