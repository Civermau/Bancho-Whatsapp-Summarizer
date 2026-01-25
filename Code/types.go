package main

import (
	"sync"
	"time"

	waProto "go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/types"
)

type MediaMeta struct {
	MimeType  string
	SizeBytes int64
	Width     int
	Height    int
	Duration  float64
	Hash      string
}

type MessageContext struct {
	MessageID  string
	ChatID     types.JID
	SenderID   types.JID
	SenderName string
	IsGroup    bool

	Text      string
	MediaType string
	MediaMeta *MediaMeta

	Timestamp  time.Time
	Mentions   []string
	IsFromMe   bool
	RawMessage *waProto.Message
}

type SummaryInfo struct {
	MessageCount int
	Style        string
	Media        bool
	Reason       bool
}

type WhitelistCache struct {
	mu     sync.RWMutex
	groups map[string]bool
	users  map[string]bool
}

type AliasCache struct {
	mu      sync.RWMutex
	aliases map[string]string
}
