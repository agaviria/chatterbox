package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
	"sync"
)

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
	r := newRoom()
	// set root template directory
	http.Handle("/", &templateHandler{filename: "chat.html"})
	http.Handle("/room", r)

	// get the run() method on room started.  This for loop will only execute a
	// single case per call.  It runs forever until the program is exited.
	go r.run()

	// set the port address and print to stdout
	var addr = ":8080"
	fmt.Println("Listening and serving on: ", addr)

	// start the web server
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatal("ListenAndServe:", err)
	}
}
