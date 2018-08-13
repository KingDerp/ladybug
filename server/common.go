package server

import "ladybug/database"

type Message struct {
	Id            string `json:"id"`
	BuyerSent     bool   `json:"buyerSent"`
	Description   string `json:"description"`
	CreatedAt     int64  `json:"createdAt"`
	MessageNumber int64  `json:"messageNumber"`
}

func MessageFromDB(message *database.Message) *Message {
	return &Message{
		Id:            message.Id,
		BuyerSent:     message.BuyerSent,
		Description:   message.Description,
		CreatedAt:     message.CreatedAt.Unix(),
		MessageNumber: message.ConversationNumber,
	}
}

func MessagesFromDB(messages []*database.Message) []*Message {
	out := []*Message{}
	for _, m := range messages {
		out = append(out, MessageFromDB(m))
	}

	return out
}

type Conversation struct {
	Id string `json:"id"`
}

func ConversationFromDB(conv *database.Conversation) *Conversation {
	return &Conversation{
		Id: conv.Id,
	}
}

func ConversationsFromDB(conversations []*database.Conversation) []*Conversation {
	out := []*Conversation{}
	for _, c := range conversations {
		out = append(out, ConversationFromDB(c))
	}

	return out
}
