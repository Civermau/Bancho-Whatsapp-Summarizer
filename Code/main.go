package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"go.mau.fi/whatsmeow"
)

var (
	GlobalConfig        *Config
	GlobalPromptsConfig *PromptsConfig
	GlobalClient        *whatsmeow.Client
	GlobalAppDB         *AppDB
	GlobalWhitelistCache *WhitelistCache
	GlobalAliasCache     *AliasCache
	GlobalImageDescriptionCache *ImageDescriptionCache
)

func main() {
	ctx := context.Background()

	var err error
	GlobalConfig, err = ReadConfig("config.json")
	if err != nil {
		panic("Failed to read config.json: " + err.Error())
	}
	GlobalConfig.DebugPrint()

	GlobalPromptsConfig, err = ReadPromptsConfig("prompts.json")
	if err != nil {
		panic("Failed to read prompts.json: " + err.Error())
	}
	GlobalPromptsConfig.DebugPrint()

	GlobalAppDB, err = OpenAppDB(ctx, "")
	if err != nil {
		panic("Failed to open database: " + err.Error())
	}

	// Initialize caches
	GlobalWhitelistCache = &WhitelistCache{
		groups: make(map[string]bool),
		users:  make(map[string]bool),
	}
	GlobalAliasCache = &AliasCache{
		aliases: make(map[string]string),
	}
	GlobalImageDescriptionCache = &ImageDescriptionCache{
		descriptions: make(map[string]string),
	}

	GlobalClient, err = initializeClient(ctx)
	if err != nil {
		panic(err)
	}

	err = connectClient(ctx, GlobalClient)
	if err != nil {
		panic(err)
	}

	// Listen to Ctrl+C (you can also do something else that prevents the program from exiting)
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c

	if GlobalAppDB != nil {
		_ = GlobalAppDB.Close()
	}
	GlobalClient.Disconnect()
}
