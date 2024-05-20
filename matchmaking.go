package main

import (
	"fmt"
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
			fmt.Println("message~", msg.Name, ms.m.mid, ": ", msg.Message)

			if msg.Event == "ping" {
				fmt.Println("ping!")
			} else if msg.Event == "hello" {
				ms.m.broadcast <- msg
			}
		} else {
			fmt.Println("error reading from socket")
			break
		}
	}
}
