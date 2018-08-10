package handlers

import (
	"context"
	"encoding/json"
	"ladybug/database"
	"ladybug/server"
	"net/http"
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

func GetBuyer(ctx context.Context) *database.Buyer {
	buyer, _ := ctx.Value(buyerContextKey).(*database.Buyer)
	return buyer
}

//TODO(mac): write put for this function
func (u *buyerHandler) buyer(w http.ResponseWriter, req *http.Request) {

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

		return

	}

	if req.Method == "PUT" {
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

		return

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

	sign_up_resp, err := u.buyerServer.BuyerSignUp(ctx, &sign_up_req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		logrus.Errorf("%+v", err)
		return
	}

	http.SetCookie(w, &http.Cookie{Name: "session", Value: sign_up_resp.Session.Id,
		Expires: sign_up_resp.Session.CreatedAt.Add(730 * time.Hour)})
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

	session, err := u.buyerServer.BuyerLogIn(ctx, &log_in_req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{Name: "session", Value: session.Id,
		Expires: session.CreatedAt.Add(730 * time.Hour)})
}

func (u *buyerHandler) buyerProducts(w http.ResponseWriter, req *http.Request) {
	if req.Method != "GET" {
		http.Error(w, "method not allowed", http.StatusBadRequest)
		return
	}

	ctx := req.Context()

	decoder := json.NewDecoder(req.Body)
	var products_req server.ProductRequest
	err := decoder.Decode(&products_req)
	if err != nil {
		http.Error(w, "unable to parse json", http.StatusInternalServerError)
		return
	}

	products, err := u.buyerServer.BuyerProducts(ctx, &products_req)
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

func (u *buyerHandler) getPagedBuyerConversation(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()

	if req.Method == "GET" {
		buyer_pk := GetBuyerPk(req.Context())

		decoder := json.NewDecoder(req.Body)
		var conversation_req server.PagedBuyerConversationsReq
		err := decoder.Decode(&conversation_req)
		if err != nil {
			http.Error(w, "unable to parse json", http.StatusInternalServerError)
			return
		}

		conversation_req.BuyerPk = buyer_pk

		conversations, err := u.buyerServer.GetPagedBuyerConversations(ctx, &conversation_req)
		if err != nil {
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}

		b, err := json.Marshal(conversations)
		if err != nil {
			http.Error(w, "server error", http.StatusInternalServerError)
			return
		}

		h := w.Header()
		h.Set("Content-Type", "application/json")
		w.Write(b)

		return

	}

	http.Error(w, "method not allowed", http.StatusBadRequest)
	return

}
