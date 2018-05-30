package handlers

import (
	"context"
	"encoding/json"
	"ladybug/database"
	"net/http"

	"golang.org/x/crypto/bcrypt"
)

type contextKey int

const (
	userContextKey contextKey = iota
)

func WithUser(ctx context.Context, user *database.User) context.Context {
	return context.WithValue(ctx, userContextKey, user)
}

func GetUser(ctx context.Context) *database.User {
	user, _ := ctx.Value(userContextKey).(*database.User)
	return user
}

func rootHandler(w http.ResponseWriter, req *http.Request) {
	// The "/" pattern matches everything, so we need to check
	// that we're at the root here.
	if req.URL.Path != "/" {
		http.NotFound(w, req)
		return
	}
	user := GetUser(req.Context())

	b, err := json.Marshal(user)
	if err != nil {
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}

	h := w.Header()
	h.Set("Content-Type", "application/json")
	w.Write(b)
}

type incomingUser struct {
	fullName  string `json:"fullName"`
	password  string `json:"password"`
	fullEmail string `json:"fullEmail"`
}

//TODO(mac): this feels wrong to get at the database this way. find a better way
func (a *authMiddleware) userHandler(w http.ResponseWriter, req *http.Request) {

	if req.Method == "DELETE" {
		http.Error(w, "method not allowed", http.StatusBadRequest)
		return
	}

	decoder := json.NewDecoder(req.Body)
	var incoming_user incomingUser
	err := decoder.Decode(&incoming_user)
	if err != nil {
		http.Error(w, "unable to parse json", http.StatusInternalServerError)
		return
	}

	//TODO(mac): need to validate email as well
	err = validateUser(&incoming_user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(incoming_user.password), bcrypt.MaxCost)
	if err != nil {
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}

	err = a.db.CreateUserWithEmailNoReturn(incomingUser.fullName, incomingUser.fullEmail, hash)
	if err != nil {
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}

	/*
			//TODO(mac): what do I need to do here? generate a session and send it back in a cookie for
			//further use?
		h := w.Header()
		h.Set("Content-Type", "application/json")
		w.Write(b)
	*/
}
