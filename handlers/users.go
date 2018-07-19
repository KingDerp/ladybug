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
	buyerContextKey contextKey = iota
	buyerPkContextKey
)

type buyerHandler struct {
	buyerServer *server.BuyerServer
}

func newBuyerHandler(server *server.BuyerServer) *buyerHandler {
	return &buyerHandler{buyerServer: server}
}

func WithBuyer(ctx context.Context, buyer *database.Buyer) context.Context {
	return context.WithValue(ctx, buyerContextKey, buyer)
}

func WithBuyerPk(ctx context.Context, pk int64) context.Context {
	return context.WithValue(ctx, buyerPkContextKey, pk)
}

func GetBuyerPk(ctx context.Context) int64 {
	pk, _ := ctx.Value(buyerPkContextKey).(int64)
	return pk
}

//TODO(Mac): why don't we handle the error here? is it just a best effort? same in the func above
func GetBuyer(ctx context.Context) *database.Buyer {
	buyer, _ := ctx.Value(buyerContextKey).(*database.Buyer)
	return buyer
}

//TODO(mac): write put for this function
func (u *buyerHandler) buyer(w http.ResponseWriter, req *http.Request) {

	if req.Method != "GET" {
		http.Error(w, "method not allowed", http.StatusBadRequest)
		return
	}

	ctx := req.Context()
	if req.Method == "GET" {
		buyer_pk := GetBuyerPk(req.Context())

		buyer_response, err := u.buyerServer.GetBuyer(ctx,
			&server.GetBuyerRequest{BuyerPk: buyer_pk})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		b, err := json.Marshal(buyer_response)
		if err != nil {
			http.Error(w, "server error", http.StatusInternalServerError)
			return
		}

		h := w.Header()
		h.Set("Content-Type", "application/json")
		w.Write(b)

	}
}

func (u *buyerHandler) buyerSignUp(w http.ResponseWriter, req *http.Request) {

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

	sign_up_resp, err := u.buyerServer.SignUp(ctx, &sign_up_req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		logrus.Errorf("%+v", err)
		return
	}

	//TODO(mac) write a policy to determine how long until cookie expires. I'd prefer it if the
	//cookie was valid as long as the browser was open
	http.SetCookie(w, &http.Cookie{Name: "session", Value: sign_up_resp.Session.Id,
		Expires: sign_up_resp.Session.CreatedAt.Add(24 * time.Hour)})
}

func (u *buyerHandler) buyerLogin(w http.ResponseWriter, req *http.Request) {
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

	session, err := u.buyerServer.LogIn(ctx, &log_in_req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{Name: "session", Value: session.Id,
		Expires: session.CreatedAt.Add(24 * time.Hour)})
}

//TODO(mac): remove handler as a portion of the naming conventions it doesn't add anything of value
func (u *buyerHandler) buyerProducts(w http.ResponseWriter, req *http.Request) {
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

	products, err := u.buyerServer.BuyerProducts(ctx, products_req)
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

func (u *buyerHandler) sendBuyerMessage(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()

	if req.Method == "GET" {

		buyer_pk := GetBuyerPk(req.Context())

		messages, err := u.buyerServer.GetBuyerMessages(ctx, &server.GetBuyerMessageRequest{
			BuyerPk: buyer_pk})
		if err != nil {
			http.Error(w, "server error", http.StatusInternalServerError)
			return
		}

		b, err := json.Marshal(messages)
		if err != nil {
			http.Error(w, "server error", http.StatusInternalServerError)
			return
		}

		h := w.Header()
		h.Set("Content-Type", "application/json")
		w.Write(b)

	}

	if req.Method == "POST" {

		decoder := json.NewDecoder(req.Body)
		var message_req server.PostBuyerMessageRequest
		err := decoder.Decode(&message_req)
		if err != nil {
			http.Error(w, "unable to parse json", http.StatusInternalServerError)
			return
		}

		buyer_pk := GetBuyerPk(ctx)
		message_req.BuyerPk = buyer_pk

		message_resp, err := u.buyerServer.PostBuyerMessage(ctx, &message_req)
		if err != nil {
			http.Error(w, "server error", http.StatusInternalServerError)
			return
		}

		b, err := json.Marshal(message_resp)
		if err != nil {
			http.Error(w, "server error", http.StatusInternalServerError)
			return
		}

		h := w.Header()
		h.Set("Content-Type", "application/json")
		w.Write(b)
	}

	http.Error(w, "method not allowed", http.StatusBadRequest)
	return
}
