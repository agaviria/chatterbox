package main

import "net/http"

type authHandler struct {
	next http.Handler
}

// ServeHTTP satisfies the http.Handler interface.
func (h *authHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	_, err := r.Cookie("auth")
	if err == http.ErrNoCookie {
		// not authenticated
		w.Header().Set("Location", "/login")
		w.WriteHeader(http.StatusTemporaryRedirect)
		return
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// success - call next handler
	h.next.ServeHTTP(w, r)
}

// MustAuth wraps templateHandler, causing execution to run first through our
// authHandler, and only to templateHandler if the request is authenticated.
func MustAuth(handler http.Handler) http.Handler {
	return &authHandler{next: handler}
}
