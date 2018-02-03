package main

import "golang.org/x/net/websocket"

// client represents a single chatterbox user
type client struct {
	// websocket for our client.
	socket *websocket.Conn
	// send is a channel on which messages are sent.
	send chan []byte
	// room is the chatroom this client is chatting.
	room *room
}

// room defines our chatroom.
type room struct {
	// fwd is a channel that holds incoming messages
	// that should be forwarded to the other clients.
	fwd chan []byte
}
