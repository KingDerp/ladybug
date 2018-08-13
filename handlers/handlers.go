package handlers

import (
	"net/http"

	"ladybug/database"
	"ladybug/server"
)

type Handler struct {
	http.Handler
}

func NewHandler(db *database.DB) *Handler {

	a := &authMiddleware{db: db}
	us := server.NewBuyerServer(db)
	u := newBuyerHandler(us)

	vs := server.NewVendorServer(db)
	v := newVendorHandler(vs)

	mux := http.NewServeMux()

	//TODO(mac): I'd like to be able to specify the route method and narrow it down to a function
	//call that handles that verb
	//buyer endpoints
	mux.Handle("/buyer/login", http.HandlerFunc(u.buyerLogin))
	mux.Handle("/buyer/sign-up", http.HandlerFunc(u.buyerSignUp))
	mux.Handle("/buyer", a.CheckBuyerSessionCookie(http.HandlerFunc(u.buyer)))
	mux.Handle("/products", http.HandlerFunc(u.buyerProducts))
	//TODO make a /products/category endpoint that lets you search products by category
	mux.Handle("/buyer/conversations", a.CheckBuyerSessionCookie(http.HandlerFunc(u.getPagedBuyerConversations)))
	//TODO make this function should mark a conversation as unread
	mux.Handle("/buyer/conversations/unread", a.CheckBuyerSessionCookie(http.HandlerFunc(u.getBuyerConversationsUnread)))
	mux.Handle("/buyer/conversation", a.CheckBuyerSessionCookie(http.HandlerFunc(u.pagedBuyerMessagesByConversationId)))
	mux.Handle("/buyer/conversation/message", a.CheckBuyerSessionCookie(http.HandlerFunc(u.postBuyerMessageToConversation)))

	//Product endpoints
	//2) get trial product
	//3) buy a product
	//1) review a product
	//4) change product review
	mux.Handle("/buyer/product/trial", a.CheckBuyerSessionCookie(http.HandlerFunc(u.buyerProductTrial)))
	mux.Handle("/buyer/product/review", a.CheckBuyerSessionCookie(http.HandlerFunc(u.buyerProductReview)))

	//vendor endpoints
	mux.Handle("/vendor/sign-up", http.HandlerFunc(v.vendorSignUp))
	mux.Handle("/vendor/product", a.CheckVendorSessionCookie(http.HandlerFunc(v.vendorProduct)))
	mux.Handle("/vendor/conversations", a.CheckVendorSessionCookie(http.HandlerFunc(v.getPagedVendorConversations)))
	mux.Handle("/vendor/conversations/unread", a.CheckVendorSessionCookie(http.HandlerFunc(v.getVendorConversationsUnread)))
	mux.Handle("/vendor/conversations/", a.CheckVendorSessionCookie(http.HandlerFunc(v.pagedVendorMessagesByConversationId)))
	mux.Handle("/vendor/conversations/message", a.CheckVendorSessionCookie(http.HandlerFunc(v.postVendorMessageToConversation)))
	//mux.Handle("/vendor/messages", a.CheckVendorSessionCookie(http.HandlerFunc(v.vendorMessage)))

	return &Handler{Handler: mux}
}
