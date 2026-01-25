package main

import (
	"context"

	"go.mau.fi/whatsmeow"
	waProto "go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/types"
	"google.golang.org/protobuf/proto"
)

func SendTextMessage(client *whatsmeow.Client, chatJID types.JID, message string) error {
	_, err := client.SendMessage(context.Background(), chatJID, &waProto.Message{
		Conversation: &message,
	})
	return err
}

func SendReplyMessage(client *whatsmeow.Client, messageContext *MessageContext, message string) error {
	contextInfo := &waProto.ContextInfo{
		StanzaID:      &messageContext.MessageID,
		Participant:   proto.String(messageContext.SenderID.String()),
		QuotedMessage: messageContext.RawMessage,
	}

	msg := &waProto.Message{
		ExtendedTextMessage: &waProto.ExtendedTextMessage{
			Text:        &message,
			ContextInfo: contextInfo,
		},
	}

	_, err := client.SendMessage(context.Background(), messageContext.ChatID, msg)
	return err
}
