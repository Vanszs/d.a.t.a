package characters

import (
	"encoding/json"
	"fmt"
	"os"
)

type CharacterConfig struct {
	Name   string   `json:"name"`
	System string   `json:"system"`
	Bio    []string `json:"bio"`
	Lore   []string `json:"lore"`
	Style  struct {
		Tone        []string `json:"tone"`
		Actions     []string `json:"actions"`
		Constraints []string `json:"constraints"`
	} `json:"style"`
	Topics []string `json:"topics"`
	Goals  []Goal   `json:"goals"`
}

type Goal struct {
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Priority    float64 `json:"priority"`
}

func LoadFromFile(path string) (*Character, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading file: %w", err)
	}

	var config CharacterConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("parsing json: %w", err)
	}

	return NewCharacter(config), nil
}

func NewCharacter(config CharacterConfig) *Character {
	return &Character{
		Name:        config.Name,
		System:      config.System,
		Bio:         config.Bio,
		Lore:        config.Lore,
		Style:       StyleGuide(config.Style),
		Topics:      config.Topics,
		Goals:       config.Goals,
		Preferences: make(map[string]float64),
	}
}
