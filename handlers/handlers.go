package handlers

import (
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/cors"

	"ladybug/database"
	"ladybug/server"
)

type Handler struct {
	http.Handler
}

func NewHandler(db *database.DB) *Handler {

	r := chi.NewRouter()

	// Basic CORS
	// for more ideas, see: https://developer.github.com/v3/#cross-origin-resource-sharing
	cors := cors.New(cors.Options{
		// AllowedOrigins: []string{"https://foo.com"}, // Use this to allow specific origin
		// hosts
		AllowedOrigins: []string{"*"}, //TODO(mac): before this roles out to prod this needs to include a config option which includes where this is publicly hosted
		// AllowOriginFunc:  func(r *http.Request, origin string) bool { return true
		// },
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300, // Maximum value not ignored by any of major browsers
	})
	r.Use(cors.Handler)

	//a := &authMiddleware{db: db}
	bs := server.NewBuyerServer(db)
	u := newBuyerHandler(bs)

	//vs := server.NewVendorServer(db)
	//v := newVendorHandler(vs)

	r.Post("/api/buyer/sign-up", http.HandlerFunc(u.buyerSignUp))

	/*
		mux := http.NewServeMux()


		r.Mount("/", mux)

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
	*/

	return &Handler{Handler: r}
}
