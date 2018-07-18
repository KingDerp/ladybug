package handlers

import (
	"context"
	"encoding/json"
	"ladybug/database"
	"ladybug/server"
	"net/http"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

type contextKey int

const (
	userContextKey contextKey = iota
	userPkContextKey
)

type userHandler struct {
	userServer *server.UserServer
}

func newUserHandler(server *server.UserServer) *userHandler {
	return &userHandler{userServer: server}
}

func WithUser(ctx context.Context, user *database.User) context.Context {
	return context.WithValue(ctx, userContextKey, user)
}

func WithUserPk(ctx context.Context, pk int64) context.Context {
	return context.WithValue(ctx, userPkContextKey, pk)
}

func GetUserPk(ctx context.Context) int64 {
	pk, _ := ctx.Value(userPkContextKey).(int64)
	return pk
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

func (u *userHandler) userHandler(w http.ResponseWriter, req *http.Request) {

	//TODO(mac): write GET and PUT for user
	if req.Method != "GET" || req.Method != "PUT" {
		http.Error(w, "method not allowed", http.StatusBadRequest)
		return
	}

	ctx := req.Context()
	if req.Method == "GET" {
		user_pk := GetUserPk(req.Context())

		user_response, err := u.userServer.GetUser(ctx,
			&server.GetUserRequest{UserPk: user_pk})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		b, err := json.Marshal(user_response)
		if err != nil {
			http.Error(w, "server error", http.StatusInternalServerError)
			return
		}

		h := w.Header()
		h.Set("Content-Type", "application/json")
		w.Write(b)

	}

	if req.Method == "PUT" {
		//user := GetUser(req.Context())

		//u.UserServer.Update
	}
}

func (u *userHandler) userSignUpHandler(w http.ResponseWriter, req *http.Request) {

	ctx := req.Context()

	if req.Method != "POST" {
		http.Error(w, "method not allowed", http.StatusBadRequest)
		return
	}

	decoder := json.NewDecoder(req.Body)
	var sign_up_req server.SignUpRequest
	err := decoder.Decode(&sign_up_req)
	if err != nil {
		http.Error(w, "unable to parse json", http.StatusInternalServerError)
		return
	}

	sign_up_resp, err := u.userServer.SignUp(ctx, &sign_up_req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		logrus.Errorf("%+v", err)
		return
	}

	//TODO(mac) write a policy to determine how long until cookie expires
	http.SetCookie(w, &http.Cookie{Name: "session", Value: sign_up_resp.Session.Id,
		Expires: sign_up_resp.Session.CreatedAt.Add(24 * time.Hour)})
}

func (u *userHandler) userLogInHandler(w http.ResponseWriter, req *http.Request) {
	if req.Method != "POST" {
		http.Error(w, "method not allowed", http.StatusBadRequest)
		return
	}

	ctx := req.Context()

	decoder := json.NewDecoder(req.Body)
	var log_in_req server.LogInRequest
	err := decoder.Decode(&log_in_req)
	if err != nil {
		http.Error(w, "unable to parse json", http.StatusInternalServerError)
		return
	}

	session, err := u.userServer.LogIn(ctx, &log_in_req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{Name: "session", Value: session.Id,
		Expires: session.CreatedAt.Add(24 * time.Hour)})
}

//TODO(mac): remove handler as a portion of the naming conventions it doesn't add anything of value
func (u *userHandler) userProducts(w http.ResponseWriter, req *http.Request) {
	if req.Method != "GET" {
		http.Error(w, "method not allowed", http.StatusBadRequest)
		return
	}

	ctx := req.Context()

	category_param := req.URL.Query().Get("categories")
	if category_param == "" {
		category_param = "random"
	}

	categories := strings.Split(category_param, ",")
	products_req := &server.ProductRequest{
		ProductCategories: categories,
	}

	products, err := u.userServer.UserProducts(ctx, products_req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	b, err := json.Marshal(products)
	if err != nil {
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}

	h := w.Header()
	h.Set("Content-Type", "application/json")
	w.Write(b)
}
