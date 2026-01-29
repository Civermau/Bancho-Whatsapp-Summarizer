package main

import (
	"encoding/json"
	"fmt"
	"os"
)

// PromptsConfig represents the prompts configuration JSON structure
type PromptsConfig struct {
	InfoString        string `json:"InfoString"`
	VersionString     string `json:"VersionString"`
	PersonalityPrompt string `json:"PersonalityPrompt"`
	LengthShort       string `json:"LengthShort"`
	LengthMedium      string `json:"LengthMedium"`
	LengthLong        string `json:"LengthLong"`
}

// DebugPrint prints the PromptsConfig in a pretty JSON format for debugging.
//
// Usage:
//
//	pc, _ := ReadPromptsConfig("prompts.json")
//	pc.DebugPrint()
func (pc *PromptsConfig) DebugPrint() {
	j, err := json.MarshalIndent(pc, "", "  ")
	if err != nil {
		fmt.Println("PromptsConfig DebugPrint error:", err)
		return
	}
	fmt.Println("PromptsConfig DEBUG:")
	lines := []byte(string(j))
	start := 0
	for i := 0; i < len(lines); i++ {
		if lines[i] == '\n' || i == len(lines)-1 {
			end := i
			if i == len(lines)-1 {
				end = i + 1
			}
			row := string(lines[start:end])
			if len(row) > 40 {
				fmt.Println(row[:40] + "...")
			} else {
				fmt.Println(row)
			}
			start = i + 1
		}
	}
}

// Config represents the main configuration JSON structure
type Config struct {
	Token          string   `json:"Token"`
	OwnerLID       string   `json:"OwnerLID"`
	GroupWhitelist []string `json:"GroupWhitelist"`
	UserWhitelist  []string `json:"UserWhitelist"`
}

// DebugPrint prints the Config in a pretty JSON format for debugging.
//
// Usage:
//
//	cfg, _ := ReadConfig("config.json")
//	cfg.DebugPrint()
func (c *Config) DebugPrint() {
	j, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		fmt.Println("Config DebugPrint error:", err)
		return
	}
	fmt.Println("Config DEBUG:")
	lines := []byte(string(j))
	start := 0
	for i := 0; i < len(lines); i++ {
		if lines[i] == '\n' || i == len(lines)-1 {
			end := i
			if i == len(lines)-1 {
				end = i + 1
			}
			row := string(lines[start:end])
			if len(row) > 40 {
				fmt.Println(row[:40] + "...")
			} else {
				fmt.Println(row)
			}
			start = i + 1
		}
	}
}

// ReadPromptsConfig reads and parses a prompts configuration JSON file
func ReadPromptsConfig(filePath string) (*PromptsConfig, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var config PromptsConfig
	err = json.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}

// ReadConfig reads and parses a main configuration JSON file
func ReadConfig(filePath string) (*Config, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var config Config
	err = json.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}

func WriteConfig(filePath string) error {
	// TODO: implement file write
	return nil
}
