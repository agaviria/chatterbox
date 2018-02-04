package main

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

// room defines our chatroom.
type room struct {
	// fwd is a channel that holds incoming messages
	// that should be forwarded to the other clients.
	fwd chan []byte
	// join is a channel for clients in-transit to join this room.
	join chan *client
	// leave is a channel for joined clients wishing to leave the room.
	leave chan *client
	// clients holds all current clients in this room. The boolean value is set
	// to true only when a user joins a room, see run() method for more details.
	clients map[*client]bool
}

// newRoom() makes a room that is ready to go.
func newRoom() *room {
	return &room{
		fwd:     make(chan []byte),
		join:    make(chan *client),
		leave:   make(chan *client),
		clients: make(map[*client]bool),
	}
}

// run method will be watching three channels: join, leave and fwd.
// If a message is received ona any of the aforementioned channels, the select
// statement will run the block code for the particular case.
func (r *room) run() {
	for {
		select {
		case client := <-r.join:
			// joining
			r.clients[client] = true
		case client := <-r.leave:
			// leaving
			delete(r.clients, client)
			close(client.send)
		case msg := <-r.fwd:
			// forward message to all clients
			for client := range r.clients {
				select {
				case client.send <- msg:
					// send the message
				default:
					// failed to send
					delete(r.clients, client)
					close(client.send)
				}
			}
		}
	}
}

const (
	socketBufferSize  = 1024
	messageBufferSize = 256
)

var upgrader = &websocket.Upgrader{ReadBufferSize: socketBufferSize, WriteBufferSize: socketBufferSize}

// ServeHTTP method means a room can now act as a handler.  When a request comes
// in via the ServeHTTP method, we get the socket by calling upgrader.Upgrade method.
// If all is well, we initialize our client and pass it into the join channel
// for the current room object.
func (r *room) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	socket, err := upgrader.Upgrade(w, req, nil)
	if err != nil {
		log.Fatal("ServeHTTP:", err)
		return
	}
	client := &client{
		socket: socket,
		send:   make(chan []byte, messageBufferSize),
		room:   r,
	}
	r.join <- client
	// Once a user leaves a room, the defer function will be responsible for tyding up.
	defer func() { r.leave <- client }()
	// write() method will run in a different thread or goroutine.
	go client.write()
	// read() method will be called in the main thread.  It will block operations
	// keeping connections alive until it's time to close it.
	client.read()
}
