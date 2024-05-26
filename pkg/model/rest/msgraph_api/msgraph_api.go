package msgraph_api

import (
	"time"
)

type ChatsResponse struct {
	Value []Chat `json:"value"`
}

type Chat struct {
	ChatType  string         `json:"chatType"`
	Id        string         `json:"id"`
	LastMsg   LastMsgPreview `json:"lastMessagePreview,omitempty"`
	Topic     string         `json:"topic"`
	Viewpoint ChatViewpoint  `json:"viewpoint"`
}

type ChatViewpoint struct {
	LastMessageReadDateTime time.Time `json:"lastMessageReadDateTime,omitempty"`
}

type LastMsgPreview struct {
	Id              string      `json:"id"`
	Body            MsgBody     `json:"body"`
	CreatedDateTime time.Time   `json:"createdDateTime"`
	LastMsgFrom     LastMsgFrom `json:"from,omitempty"`
}

type MsgBody struct {
	Content     string `json:"content,omitempty"`
	ContentType string `json:"contentType"`
}

type LastMsgFrom struct {
	User User `json:"user"`
}

type User struct {
	Id          string `json:"id"`
	DisplayName string `json:"displayName"`
}

type MsgRequest struct {
	MsgBody MsgBody `json:"body"`
}

type MsgResponse struct {
	MsgBody MsgBody `json:"body"`
}
