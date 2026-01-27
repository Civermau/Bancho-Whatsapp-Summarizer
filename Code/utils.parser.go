package main

import (
	"encoding/hex"
	"fmt"
	"time"

	"go.mau.fi/whatsmeow/types/events"
)

// Helper function to safely dereference a string pointer
func stringValue(ptr *string) string {
	if ptr == nil {
		return ""
	}
	return *ptr
}

// Helper function to safely dereference a uint64 pointer
func uint64Value(ptr *uint64) int64 {
	if ptr == nil {
		return 0
	}
	return int64(*ptr)
}

// Helper function to safely dereference a uint32 pointer
func uint32Value(ptr *uint32) int {
	if ptr == nil {
		return 0
	}
	return int(*ptr)
}

// Helper function to safely dereference a uint32 pointer for duration
func durationValue(ptr *uint32) float64 {
	if ptr == nil {
		return 0
	}
	return float64(*ptr)
}

// Helper function to encode media key to hex string
func encodeMediaKey(mediaKey []byte) string {
	if len(mediaKey) == 0 {
		return ""
	}
	return hex.EncodeToString(mediaKey)
}

func buildMediaMeta(mimeType *string, fileLength *uint64, mediaKey []byte, width, height *uint32, duration *uint32) *MediaMeta {
	return &MediaMeta{
		MimeType:  stringValue(mimeType),
		SizeBytes: uint64Value(fileLength),
		Width:     uint32Value(width),
		Height:    uint32Value(height),
		Duration:  durationValue(duration),
		Hash:      encodeMediaKey(mediaKey),
	}
}

func ParseMessageEvent(evt *events.Message) (*MessageContext, error) {
	msg := &MessageContext{
		IsGroup:    evt.Info.IsGroup,
		MessageID:  evt.Info.ID,
		ChatID:     evt.Info.Chat,
		SenderID:   evt.Info.Sender,
		Timestamp:  evt.Info.Timestamp,
		IsFromMe:   evt.Info.IsFromMe,
		RawMessage: evt.Message,
	}

	if evt.Info.PushName != "" {
		msg.SenderName = evt.Info.PushName
	}

	// Extract mentions from ExtendedTextMessage if available
	if ext := evt.Message.GetExtendedTextMessage(); ext != nil && ext.ContextInfo != nil {
		for _, m := range ext.ContextInfo.MentionedJID {
			msg.Mentions = append(msg.Mentions, m)
		}
	}

	if conv := evt.Message.GetConversation(); conv != "" {
		msg.Text = conv
	} else if ext := evt.Message.GetExtendedTextMessage(); ext != nil && ext.Text != nil {
		msg.Text = *ext.Text
	}

	// Handle different media types
	switch {
	case evt.Message.StickerMessage != nil:
		// Stickers are detected as images since they are webp
		stk := evt.Message.StickerMessage
		msg.MediaType = "image"
		msg.MediaMeta = buildMediaMeta(stk.Mimetype, stk.FileLength, stk.MediaKey, stk.Width, stk.Height, nil)

	case evt.Message.ImageMessage != nil:
		img := evt.Message.ImageMessage
		msg.MediaType = "image"
		msg.MediaMeta = buildMediaMeta(img.Mimetype, img.FileLength, img.MediaKey, img.Width, img.Height, nil)

	case evt.Message.VideoMessage != nil:
		vid := evt.Message.VideoMessage
		msg.MediaType = "video"
		msg.MediaMeta = buildMediaMeta(vid.Mimetype, vid.FileLength, vid.MediaKey, vid.Width, vid.Height, vid.Seconds)

	case evt.Message.AudioMessage != nil:
		aud := evt.Message.AudioMessage
		msg.MediaType = "audio"
		msg.MediaMeta = buildMediaMeta(aud.Mimetype, aud.FileLength, aud.MediaKey, nil, nil, aud.Seconds)

	case evt.Message.DocumentMessage != nil:
		doc := evt.Message.DocumentMessage
		msg.MediaType = "document"
		msg.MediaMeta = buildMediaMeta(doc.Mimetype, doc.FileLength, doc.MediaKey, nil, nil, nil)

	default:
		msg.MediaType = "text"
		msg.MediaMeta = nil
	}

	return msg, nil
}

func (msg *MessageContext) Print() {
	fmt.Printf("----------------------------\n")
	fmt.Printf("MessageID: %s\n", msg.MessageID)
	fmt.Printf("ChatID: %s\n", msg.ChatID.String())
	fmt.Printf("SenderID: %s\n", msg.SenderID.String())
	fmt.Printf("SenderName: %s\n", msg.SenderName)
	fmt.Printf("IsGroup: %v\n", msg.IsGroup)
	fmt.Printf("Text: %s\n", msg.Text)
	fmt.Printf("MediaType: %s\n", msg.MediaType)
	if msg.MediaMeta != nil {
		fmt.Printf("MediaMeta: MimeType=%s, SizeBytes=%d, Width=%d, Height=%d, Duration=%.2f, Hash=%s\n",
			msg.MediaMeta.MimeType, msg.MediaMeta.SizeBytes, msg.MediaMeta.Width,
			msg.MediaMeta.Height, msg.MediaMeta.Duration, msg.MediaMeta.Hash)
	} else {
		fmt.Printf("MediaMeta: nil\n")
	}
	fmt.Printf("Timestamp: %s\n", msg.Timestamp.Format(time.RFC3339))
	fmt.Printf("Mentions: %v\n", msg.Mentions)
	fmt.Printf("IsFromMe: %v\n", msg.IsFromMe)
	// fmt.Printf("Raw message data: %v\n", msg.RawMessage)
	fmt.Printf("----------------------------\n")
}
