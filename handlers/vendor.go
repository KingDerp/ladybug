package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"ladybug/database"
	"ladybug/server"

	"github.com/sirupsen/logrus"
)

const (
	vendorContextKey contextKey = iota
	vendorPkContextKey
)

type vendorHandler struct {
	vendorServer *server.VendorServer
}

func newVendorHandler(server *server.VendorServer) *vendorHandler {
	return &vendorHandler{vendorServer: server}
}

func WithVendor(ctx context.Context, vendor *database.Vendor) context.Context {
	return context.WithValue(ctx, vendorContextKey, vendor)
}

func WithVendorPk(ctx context.Context, pk int64) context.Context {
	return context.WithValue(ctx, vendorPkContextKey, pk)
}

func GetVendorPk(ctx context.Context) int64 {
	pk, _ := ctx.Value(vendorPkContextKey).(int64)
	return pk
}

func GetVendor(ctx context.Context) *database.Vendor {
	vendor, _ := ctx.Value(vendorContextKey).(*database.Vendor)
	return vendor
}

func (v *vendorHandler) vendorSignUp(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()

	if req.Method != "POST" {
		http.Error(w, "method not allowed", http.StatusBadRequest)
		return
	}

	decoder := json.NewDecoder(req.Body)
	var sign_up_req server.VendorSignUpRequest
	err := decoder.Decode(&sign_up_req)
	if err != nil {
		http.Error(w, "unable to parse json", http.StatusInternalServerError)
		return
	}

	sign_up_resp, err := v.vendorServer.VendorSignUp(ctx, &sign_up_req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		logrus.Errorf("%+v", err)
		return
	}

	http.SetCookie(w, &http.Cookie{Name: "session", Value: sign_up_resp.Session.Id,
		Expires: sign_up_resp.Session.CreatedAt.Add(730 * time.Hour)})

	b, err := json.Marshal(sign_up_resp)
	if err != nil {
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}

	h := w.Header()
	h.Set("Content-Type", "application/json")
	w.Write(b)
}

func (v *vendorHandler) vendorProduct(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()

	if req.Method != "POST" {
		http.Error(w, "method not allowed", http.StatusBadRequest)
		return
	}

	decoder := json.NewDecoder(req.Body)
	var product_request server.RegisterProductRequest
	err := decoder.Decode(&product_request)
	if err != nil {
		http.Error(w, "unable to parse json", http.StatusInternalServerError)
		return
	}

	vendor_pk := GetVendorPk(req.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	product_request.VendorPk = vendor_pk

	register_prod_response, err := v.vendorServer.RegisterProduct(ctx, &product_request)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		logrus.Errorf("%+v", err)
		return
	}

	b, err := json.Marshal(register_prod_response)
	if err != nil {
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}

	h := w.Header()
	h.Set("Content-Type", "application/json")
	w.Write(b)
}

func (v *vendorHandler) getPagedVendorConversations(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()

	if req.Method == "GET" {

		decoder := json.NewDecoder(req.Body)
		var conversation_req server.PagedVendorConversationReq
		err := decoder.Decode(&conversation_req)
		if err != nil {
			http.Error(w, "unable to parse json", http.StatusInternalServerError)
			return
		}

		conversation_req.VendorPk = GetVendorPk(req.Context())

		conversations, err := v.vendorServer.GetPagedVendorConversations(ctx, &conversation_req)
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

func (v *vendorHandler) getVendorConversationsUnread(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()

	if req.Method == "GET" {
		req := &server.VendorConversationsUnreadReq{VendorPk: GetVendorPk(req.Context())}

		conversations, err := v.vendorServer.GetVendorCoversationsUnread(ctx, req)
		if err != nil {
			http.Error(w, "server error", http.StatusInternalServerError)
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

func (v *vendorHandler) pagedVendorMessagesByConversationId(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()

	if req.Method == "GET" {

		decoder := json.NewDecoder(req.Body)
		var conversation_req server.PagedVendorMessagesByConversationIdReq
		err := decoder.Decode(&conversation_req)
		if err != nil {
			http.Error(w, "unable to parse json", http.StatusInternalServerError)
			return
		}

		messages, err := v.vendorServer.PagedVendorMessagesByConversationId(ctx, &conversation_req)
		if err != nil {
			http.Error(w, "internal server error", http.StatusInternalServerError)
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

		return

	}

	http.Error(w, "method not allowed", http.StatusBadRequest)
	return
}

func (v *vendorHandler) postVendorMessageToConversation(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()

	if req.Method == "POST" {
		decoder := json.NewDecoder(req.Body)
		var conversation_req server.PostVendorMessageToConversationReq
		err := decoder.Decode(&conversation_req)
		if err != nil {
			http.Error(w, "unable to parse json", http.StatusInternalServerError)
			return
		}

		conversation_req.VendorPk = GetVendorPk(req.Context())

		messages, err := v.vendorServer.PostVendorMessageToConversation(ctx, &conversation_req)
		if err != nil {
			http.Error(w, "internal server error", http.StatusInternalServerError)
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

		return
	}

	http.Error(w, "method not allowed", http.StatusBadRequest)
	return
}
