package server

import "ladybug/database"

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
