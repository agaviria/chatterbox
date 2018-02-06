package main

import (
	"bytes"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second
	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second
	// Send pings to peer with this period. Must be less than pongWait!
	pingPeriod = (pongWait * 9) / 10
	// maxMessageSize allowed from peer.
	maxMessageSize = 512
)

var (
	newLine = []byte{'\n'}
	space   = []byte{' '}
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

// Client represents a single chatterbox user and is the middleman b/w the
// socket connection and the room hub.
type Client struct {
	// websocket for our client.
	socket *websocket.Conn
	// send is a channel on which outbound messages are sent.
	send chan []byte
	// room is the chatroom this client is chatting.
	room *Room
}

// read allows our client to read from the websocket via the ReadMessage method.
// It's continually sending any received messages to broadcast channel on room type.
func (c *Client) read() {
	defer func() {
		c.room.leave <- c
		c.socket.Close()
	}()
	c.socket.SetReadLimit(maxMessageSize)
	c.socket.SetReadDeadline(time.Now().Add(pongWait))
	c.socket.SetPongHandler(func(string) error {
		c.socket.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, msg, err := c.socket.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway) {
				log.Printf("error: %v", err)
			}
			break
		}
		msg = bytes.TrimSpace(bytes.Replace(msg, newLine, space, -1))
		c.room.broadcast <- msg
	}
}

// write method continually accepts messages from the broadcast channel.
// It executes via the WriteMessage method within a goroutine.
func (c *Client) write() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.socket.Close()
	}()
	for {
		select {
		case msg, ok := <-c.send:
			c.socket.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// the room hub closed the channel.
				c.socket.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.socket.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(msg)

			// Add queued chat mesages to the current socket message.
			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write(newLine)
				w.Write(<-c.send)
			}

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			c.socket.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.socket.WriteMessage(websocket.PingMessage, []byte{}); err != nil {
				return
			}
		}
	}
}

// serveWs method means a room can now act as a handler.  When a request comes
// in via the ServeHTTP method, we get the socket by calling upgrader.Upgrade method.
// If all is well, we initialize our client and pass it into the join channel
// for the current room object.
func serveWs(room *Room, w http.ResponseWriter, req *http.Request) {
	socket, err := upgrader.Upgrade(w, req, nil)
	if err != nil {
		log.Println(err)
		return
	}
	client := &Client{
		room:   room,
		socket: socket,
		send:   make(chan []byte, 256),
	}
	client.room.join <- client
	// write() method will run in a different thread or goroutine.
	go client.write()
	// read() method will be called in the main thread.
	go client.read()
}
