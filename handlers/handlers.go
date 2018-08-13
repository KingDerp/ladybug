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
	mux.Handle("/buyer/conversations",
		a.CheckBuyerSessionCookie(http.HandlerFunc(u.getPagedBuyerConversations)))
	mux.Handle("/buyer/conversations/unread",
		a.CheckBuyerSessionCookie(http.HandlerFunc(u.getBuyerConversationsUnread)))
	mux.Handle("/buyer/conversation",
		a.CheckBuyerSessionCookie(http.HandlerFunc(u.pagedBuyerMessagesByConversationId)))
	mux.Handle("/buyer/conversation/message",
		a.CheckBuyerSessionCookie(http.HandlerFunc(u.pagedBuyerMessagesByConversationId)))
	//mux.Handle("/buyer/messages", a.CheckBuyerSessionCookie(http.HandlerFunc(u.sendBuyerMessage)))
	//mux.Handle("/buyer/messages/unread", a.CheckBuyerSessionCookie(http.HandlerFunc(u.sendBuyerMessage)))

	//TODO make this function should mark a conversation as unread
	//mux.Handle("/buyer/message/unread", a.CheckBuyerSessionCookie(http.HandlerFunc(u.unreadConversation)))

	//vendor endpoints
	mux.Handle("/vendor/sign-up", http.HandlerFunc(v.vendorSignUp))
	mux.Handle("/vendor/product", a.CheckVendorSessionCookie(http.HandlerFunc(v.vendorProduct)))
	//mux.Handle("/vendor/messages", a.CheckVendorSessionCookie(http.HandlerFunc(v.vendorMessage)))

	return &Handler{Handler: mux}
}
