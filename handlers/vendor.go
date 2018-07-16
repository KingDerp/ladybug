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

func (v *vendorHandler) vendorSignUpHandler(w http.ResponseWriter, req *http.Request) {
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
		Expires: sign_up_resp.Session.CreatedAt.Add(24 * time.Hour)})
}
