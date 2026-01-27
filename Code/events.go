package main

import (
	"context"
	"fmt"
	"strings"

	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
)

func eventHandler(evt interface{}) {
	switch v := evt.(type) {
	case *events.Message:
		ctx, err := ParseMessageEvent(v)
		if err != nil {
			return
		}
		// TODO: implement whitelist check
		splitMessages(ctx)
		ctx.Print()
	}
}

// ? --------------------------------------------------------------------------------------------------------
// ? ----------------------------------------------Message Splitter------------------------------------------
// ? --------------------------------------------------------------------------------------------------------

func splitMessages(ctx *MessageContext) {
	switch ctx.MediaType {
	case "image":
		handleImageMessage(ctx)
		return
	case "video":
		handleVideoMessage(ctx)
		return
	case "audio":
		handleAudioMessage(ctx)
		return
	}

	if len(ctx.Text) > 0 && ctx.Text[0] == '-' && ctx.IsGroup == true {
		fmt.Print("Command triggered with -!\n")
		handleCommands(ctx)
		return
	}

	handleTextMessage(ctx)
}

// ? -----------------------------------------------------------------------------------------------------
// ? ----------------------------------------------Type Handlers------------------------------------------
// ? -----------------------------------------------------------------------------------------------------

// handleImageMessage handles incoming image messages.
func handleImageMessage(ctx *MessageContext) {
	fmt.Print("IMAGE DETECTED\n")
	imageCache, err := isImageCached(GlobalImageDescriptionCache, ctx.MediaMeta.Hash, GlobalAppDB)
	if err != nil {
		// TODO: Send to api, then store
		imageCache = "Placeholder" // ! placeholder, duh
		err = setNewImageCache(GlobalImageDescriptionCache, ctx.MediaMeta.Hash, imageCache, GlobalAppDB)
	}
	// TODO: Store imageCache in db
}

// handleVideoMessage handles incoming video messages.
func handleVideoMessage(ctx *MessageContext) {
	fmt.Print("VIDEO DETECTED\n")
	// TODO: Process video message (e.g., save video info, respond, etc.)
}

// handleAudioMessage handles incoming audio messages.
func handleAudioMessage(ctx *MessageContext) {
	fmt.Print("AUDIO DETECTED\n")
	// TODO: Process audio message (e.g., transcribe, respond, etc.)
}

// ? -----------------------------------------------------------------------------------------------------
// ? ----------------------------------------------Text Handlers------------------------------------------
// ? -----------------------------------------------------------------------------------------------------

func handleTextMessage(ctx *MessageContext) {
	selfID := GlobalClient.Store.LID.User + "@lid"

	for _, mention := range ctx.Mentions {
		if mention == selfID {
			SendTextMessage(GlobalClient, ctx.ChatID, "Soy ese") // TODO: Send random sticker
			break
		}
	}
}

func handleCommands(ctx *MessageContext) {
	words := strings.Split(ctx.Text, " ")

	switch words[0] {
	case "-s", "--summarize":
		fmt.Println("Summarize command detected!")
		// TODO: IMPLEMENT SUMMARIZE LOGIC

	// ? ===================================
	case "-v", "--version":
		SendTextMessage(GlobalClient, ctx.ChatID, GlobalPromptsConfig.VersionString)

	// ? ===================================
	case "-i", "--info":
		SendTextMessage(GlobalClient, ctx.ChatID, GlobalPromptsConfig.InfoString)

	// ? ===================================
	case "--whitelist":
		ownerJID, err := types.ParseJID(GlobalConfig.OwnerLID)
		if err != nil {
			SendTextMessage(GlobalClient, ctx.ChatID, "Owner not configured correctly.")
			break
		}
		if ctx.SenderID != ownerJID {
			SendTextMessage(GlobalClient, ctx.ChatID, "Only the owner can whitelist.")
			fmt.Printf("%s tried to whitelist!\n", ctx.SenderName)
			break
		}

		// TODO: implement whitelist in db
		fmt.Println("Whitelist command issued by owner.")

	// ? ===================================
	case "--alias":
		handleAliasCommand(ctx, words)

	// ? ===================================
	case "--disable":
		// TODO: Implement disable command handling

	case "--enable":
		// TODO: Implement disable command handling

	// ? ===================================
	case "--reload-json":
		ownerJID, err := types.ParseJID(GlobalConfig.OwnerLID)
		if err != nil {
			SendTextMessage(GlobalClient, ctx.ChatID, "Owner not configured correctly.")
			break
		}
		if ctx.SenderID != ownerJID {
			SendTextMessage(GlobalClient, ctx.ChatID, "Only the owner can reload configs.")
			break
		}
		err = reloadConfigs()
		if err != nil {
			SendTextMessage(GlobalClient, ctx.ChatID, "Failed to reload configs: "+err.Error())
		} else {
			SendTextMessage(GlobalClient, ctx.ChatID, "Configs reloaded successfully.")
		}
	}
}

// ? ----------------------------------------------Alias Handler----------------------------------------------
func handleAliasCommand(ctx *MessageContext, words []string) {
	if GlobalAppDB == nil {
		SendTextMessage(GlobalClient, ctx.ChatID, "Database not initialized.")
		return
	}
	if len(words) < 2 {
		SendTextMessage(GlobalClient, ctx.ChatID, "Usage: --alias <name>")
		return
	}

	dbCtx := context.Background()
	chatJID := ctx.ChatID.String()
	senderJID := ctx.SenderID.String()
	alias := words[1]

	if err := GlobalAppDB.SetAlias(dbCtx, chatJID, senderJID, alias); err != nil {
		SendReplyMessage(GlobalClient, ctx, "Failed to save alias")
		return
	}

	SendReplyMessage(GlobalClient, ctx, "Alias has been saved.")
}

// ? ----------------------------------------------Config Handler----------------------------------------------
func reloadConfigs() error {
	var err error

	GlobalConfig, err = ReadConfig("config.json")
	if err != nil {
		return err
	}

	GlobalPromptsConfig, err = ReadPromptsConfig("prompts.json")
	if err != nil {
		return err
	}

	return nil
}
