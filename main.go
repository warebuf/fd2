package main

import (
	"log"
	"net/http"
	"net/url"
	"strconv"

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

	mid_to_msid_to_match_socket map[uuid.UUID]map[uuid.UUID]*match_socket
	mid_to_match                map[uuid.UUID]*match

	list_of_mmsockets map[uuid.UUID]*mmsocket
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

type match struct {
	mutex sync.RWMutex

	mid       uuid.UUID
	game_mode string
	capacity  uint

	broadcast        chan *message // a channel is a thread-safe queue, incoming messages
	participant_join chan *match_socket
	spectator_join   chan *match_socket // a channel for clients wishing to join
	leave            chan *match_socket // a channel for clients wishing to leave

	participant_signup  chan *mmsocket
	participant_signout chan *mmsocket

	participant_uid_to_msid_to_match_socket map[uuid.UUID]map[uuid.UUID]*match_socket
	participant_uid_to_user                 map[uuid.UUID]*user

	spectator_uid_to_msid_to_match_socket map[uuid.UUID]map[uuid.UUID]*match_socket
	spectator_uid_to_user                 map[uuid.UUID]*user

	message_logs []*message

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

type match_socket struct {
	mutex sync.RWMutex

	socket *websocket.Conn

	msid uuid.UUID

	u *user
	m *match

	incoming_message chan *message // send is a channel on which messages are sent.

	open bool
}

type mmsocket struct {
	socket *websocket.Conn

	mmid uuid.UUID

	u *user

	incoming_message chan *message // send is a channel on which messages are sent.

	open bool
}

type message struct {
	Name    string
	Message string
	When    time.Time
	Event   string
	MatchID uuid.UUID
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

type matchbase struct {
	match map[uuid.UUID]*match
	mutex sync.RWMutex
}

type mmbase struct {
	matchmaking map[uuid.UUID]*mmsocket
	mutex       sync.RWMutex
}

var uid_to_user userbase
var rid_to_room roombase
var mid_to_match matchbase
var mmid_to_matchmaking mmbase

// QUICK LOOK UP
type rname_to_rid struct {
	name_to_rid map[string]uuid.UUID
	mutex       sync.RWMutex
}
type sid_to_socket struct {
	sid_to_sock map[uuid.UUID]*socket
	mutex       sync.RWMutex
}
type msid_to_socket struct {
	msid_to_sock map[uuid.UUID]*match_socket
	mutex        sync.RWMutex
}

var roomname_to_rid rname_to_rid
var sid_to_sock sid_to_socket
var msid_to_sock msid_to_socket

func main() {

	// making GLOBAL data structures
	uid_to_user.users = make(map[uuid.UUID]*user)
	uid_to_user.mutex = sync.RWMutex{}
	rid_to_room.rooms = make(map[uuid.UUID]*room)
	rid_to_room.mutex = sync.RWMutex{}
	mid_to_match.match = make(map[uuid.UUID]*match)
	mid_to_match.mutex = sync.RWMutex{}
	mmid_to_matchmaking.matchmaking = make(map[uuid.UUID]*mmsocket)
	mmid_to_matchmaking.mutex = sync.RWMutex{}

	// making QUICK LOOKUP data structures
	roomname_to_rid.name_to_rid = make(map[string]uuid.UUID)
	roomname_to_rid.mutex = sync.RWMutex{}
	sid_to_sock.sid_to_sock = make(map[uuid.UUID]*socket)
	sid_to_sock.mutex = sync.RWMutex{}
	msid_to_sock.msid_to_sock = make(map[uuid.UUID]*match_socket)

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

	fs2 := http.FileServer(http.Dir("./music"))
	mux.Handle("/music/", http.StripPrefix("/music/", fs2))

	mux.HandleFunc("/game", gameHandler)
	mux.HandleFunc("/matchmaking", matchmakingHandler)
	mux.HandleFunc("/start", startHandler)
	mux.HandleFunc("/ingame/", ingameHandler)

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
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}
	if auth, ok := session.Values["authenticated"].(bool); ok && auth {
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

			mid_to_msid_to_match_socket: make(map[uuid.UUID]map[uuid.UUID]*match_socket),
			mid_to_match:                make(map[uuid.UUID]*match),
		}
		fmt.Println("created User Object")

		uid_to_user.mutex.Lock()
		uid_to_user.users[uid] = user_object
		uid_to_user.mutex.Unlock()

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

		mid_to_msid_to_match_socket: make(map[uuid.UUID]map[uuid.UUID]*match_socket),
		mid_to_match:                make(map[uuid.UUID]*match),
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
	//fmt.Println(session)
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
				rid_to_room.rooms[rid].open <- false
				http.Redirect(res, req, "/lobby", http.StatusFound)
			}

		}
	}
}

func gameHandler(res http.ResponseWriter, req *http.Request) {
	//c, _ := req.Cookie("session-name")

	/*
		files, err := os.ReadDir("./music")
		/if err != nil {
			log.Fatal(err)
		}
		randNum := rand.Intn(len(files))
	*/

	// Check if user is already authenticated
	session, err := store.Get(req, "session-name")
	fmt.Println(session)
	if err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}

	if auth, ok := session.Values["authenticated"].(bool); ok && auth {
		t := template.Must(template.ParseFiles(filepath.Join("static", "game.html")))
		//data := map[string]string{"email": session.Values["Email"].(string), "host": req.Host, "song": files[randNum].Name()}
		data := map[string]string{"email": session.Values["Email"].(string), "host": req.Host}
		t.Execute(res, data)
	} else {
		fmt.Println("User is not authenticated, redirecting to home page")
		http.Redirect(res, req, "/", http.StatusSeeOther)
	}

}

func matchmakingHandler(res http.ResponseWriter, req *http.Request) {
	log.Println("/matchmaking")

	// Check if user is already authenticated
	session, err := store.Get(req, "session-name")
	fmt.Println(session)
	if err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}

	// get the user's object from their UUID
	fmt.Println("UUID:", session.Values["uid"])
	uid, _ := uuid.Parse(session.Values["uid"].(string)) // convert string type to UUID type
	uid_to_user.mutex.RLock()
	requesting_user := uid_to_user.users[uid]
	uid_to_user.mutex.RUnlock()

	// create a web socket connection
	var upgrader = &websocket.Upgrader{ReadBufferSize: 1024, WriteBufferSize: 1024}
	sock, err := upgrader.Upgrade(res, req, nil)
	//sock.SetReadLimit(128000) //have to handle this read limit correctly
	if err != nil {
		log.Fatal("ServeHTTP:", err)
		return
	}

	// create a socket object for the user
	not_assigned := true
	var random_mmid uuid.UUID
	var temp *mmsocket
	for not_assigned {
		random_mmid = uuid.New()
		mmid_to_matchmaking.mutex.RLock()
		_, found := mmid_to_matchmaking.matchmaking[random_mmid]
		mmid_to_matchmaking.mutex.RUnlock()
		if !found {
			mmid_to_matchmaking.mutex.Lock()
			temp = &mmsocket{
				socket:           sock,
				mmid:             random_mmid,
				u:                requesting_user,
				incoming_message: make(chan *message),
				open:             true,
			}
			mmid_to_matchmaking.matchmaking[random_mmid] = temp
			mmid_to_matchmaking.mutex.Unlock()

			not_assigned = false
		}
	}

	go mm_write(temp)
	go mm_read(temp)

	// when a match socket is created, send them a list of all possible matches
	for _, i := range mid_to_match.match {
		temp.incoming_message <- &message{Event: "newMatch", Message: i.game_mode + strconv.Itoa(int(i.capacity)), When: time.Now(), MatchID: i.mid}
		for _, j := range i.participant_uid_to_user {
			temp.incoming_message <- &message{Name: j.email, Message: "participantJoinSuccess", Event: "participantJoinSuccess", When: time.Now(), MatchID: i.mid}
		}
	}
}

func startHandler(res http.ResponseWriter, req *http.Request) {
	fmt.Println("/start")

	session, err := store.Get(req, "session-name")
	if err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}

	u := req.URL
	parameters, err := url.ParseQuery(u.RawQuery)
	fmt.Println(parameters)

	if err != nil {

		session.Values["Room_error"] = 2
		err = session.Save(req, res)
		if err != nil {
			http.Error(res, err.Error(), http.StatusInternalServerError)
			return
		}

		http.Redirect(res, req, "/game", http.StatusFound)
	} else {
		fmt.Println(parameters["room_name"][0]) // vulnerability here, need to check if this exists

		http.Redirect(res, req, "/game", http.StatusFound)
	}

}

func ingameHandler(res http.ResponseWriter, req *http.Request) {
	log.Println("/ingame")

	// Check if user is already authenticated
	session, err := store.Get(req, "session-name")
	if err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}

	u := req.URL

	parsed := strings.Split(u.Path, `/`)

	if len(parsed) < 3 {
		fmt.Println("URL is weird/wrong")
		http.Redirect(res, req, "/game", http.StatusSeeOther)
	}

	mid, err2 := uuid.Parse(parsed[2])
	mtch, found := mid_to_match.match[mid]

	fmt.Println(mtch)

	if !found || (err2 != nil) {
		fmt.Println("could not find this MID")
		http.Redirect(res, req, "/game", http.StatusSeeOther)
	}

	if auth, ok := session.Values["authenticated"].(bool); ok && auth {
		fmt.Println("asdsa")
		data := map[string]string{"email": session.Values["Email"].(string), "mid": parsed[2]}
		t := template.Must(template.ParseFiles(filepath.Join("static", "user.html")))
		t.Execute(res, data)
		http.Redirect(res, req, u.Path, http.StatusSeeOther)

	} else {
		//fmt.Println("User is not authenticated, redirecting to home page")
		http.Redirect(res, req, "/", http.StatusSeeOther)
	}
}
