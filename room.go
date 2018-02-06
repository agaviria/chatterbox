package main

// Room defines our chatroom.
type Room struct {
	// broadcast is a channel that holds incoming messages
	// that should be forwarded to the other clients.
	broadcast chan []byte
	// join is a channel for clients in-transit to join this room.
	join chan *Client
	// leave is a channel for joined clients wishing to leave the room.
	leave chan *Client
	// clients holds all current clients in this room. The boolean value is set
	// to true only when a user joins a room, see run() method for more details.
	clients map[*Client]bool
}

// run method will be watching three channels: join, leave and fwd.
// If a message is received ona any of the aforementioned channels, the select
// statement will run the block code for the particular case.
func (r *Room) run() {
	for {
		select {
		case client := <-r.join:
			// joining
			r.clients[client] = true
		case client := <-r.leave:
			// leaving
			if _, ok := r.clients[client]; ok {
				delete(r.clients, client)
				close(client.send)
			}
		case msg := <-r.broadcast:
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
