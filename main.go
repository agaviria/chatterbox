package main

import (
	"flag"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
	"sync"
)

func init() {
	// set root template directory.
	http.Handle("/", &templateHandler{filename: "chatbox.html"})

	// serve template to http.ResponseWriter object.
	// http.HandleFunc("/room", ServeHTTP)
}

var addr = flag.String("addr", ":8080", "http service address")

// templateHandler is responsible for loading, compiling and delivering templates
// through the method ServeHTTP which satisfies the http.Handler interface.
//
// filename is a new type that will compile the template once using (sync.once).
// this sync.once type will keep a reference and then respond to HTTP req's.
// tmpl, represents a single html/template.
type templateHandler struct {
	once     sync.Once
	filename string
	tmpl     *template.Template
}

// newRoom() is a helper function.  It initializes all struct fields for Room.
func newRoom() *Room {
	return &Room{
		broadcast: make(chan []byte),
		join:      make(chan *Client),
		leave:     make(chan *Client),
		clients:   make(map[*Client]bool),
	}
}

// ServeHTTP is the single method for templateHandler which handles the HTTP request.
// It's responsibility is to load templateHandler, compile tmpl, execute it and
// write the output to the specified http.ResponseWriter object.
func (t *templateHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	t.once.Do(func() {
		t.tmpl = template.Must(template.ParseFiles(filepath.Join("templates", t.filename)))
	})
	t.tmpl.Execute(w, nil)
}

func main() {
	flag.Parse() // parse the flags

	chatroom := newRoom()

	// get the run() method on room started.  This for loop will only execute a
	// single case per call.  It runs forever until the program is exited.
	go chatroom.run()

	// enable room handler requests.
	http.HandleFunc("/room", func(w http.ResponseWriter, r *http.Request) {
		serveWs(chatroom, w, r)
	})

	// http.Handle("/room", chatroom)

	// log print to stdout
	log.Println("Listening and serving on: ", *addr)
	// start web server
	if err := http.ListenAndServe(*addr, nil); err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
