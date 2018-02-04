package main

import "github.com/gorilla/websocket"

// client represents a single chatterbox user
type client struct {
	// websocket for our client.
	socket *websocket.Conn
	// send is a channel on which messages are sent.
	send chan []byte
	// room is the chatroom this client is chatting.
	room *room
}

// read allows our client to read from the websocket via the ReadMessage method.
// It is continually sending any received messages to the fwd channel on the room type.
func (c *client) read() {
	for {
		if _, msg, err := c.socket.ReadMessage(); err == nil {
			c.room.fwd <- msg
		} else {
			break
		}
	}
	c.socket.Close()
}

// write method continually accepts messages from the send channel writing everything
// out of the websocket via the WriteMessage method.
func (c *client) write() {
	for msg := range c.send {
		if err := c.socket.WriteMessage(websocket.TextMessage, msg); err != nil {
			break
		}
	}
	c.socket.Close()
}
