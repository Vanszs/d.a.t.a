package core

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/carv-protocol/d.a.t.a/src/pkg/clients"
)

type SocialClient interface {
	SendMessage(msg SocialMessage) error
	GetMessageChannel() chan SocialMessage
}

// IntentType defines different types of intents
type IntentType string

const (
	IntentQuestion    IntentType = "question"
	IntentFeedback    IntentType = "feedback"
	IntentComplaint   IntentType = "complaint"
	IntentSuggestion  IntentType = "suggestion"
	IntentGreeting    IntentType = "greeting"
	IntentInquiry     IntentType = "inquiry"
	IntentRequest     IntentType = "request"
	IntentAcknowledge IntentType = "acknowledge"
)

// EntityType defines different types of entities
type EntityType string

const (
	EntityPerson   EntityType = "person"
	EntityProduct  EntityType = "product"
	EntityCompany  EntityType = "company"
	EntityLocation EntityType = "location"
	EntityDateTime EntityType = "datetime"
	EntityCrypto   EntityType = "crypto"
	EntityWallet   EntityType = "wallet"
	EntityContract EntityType = "contract"
)

// EmotionType defines different types of emotions
type EmotionType string

const (
	EmotionPositive EmotionType = "positive"
	EmotionNegative EmotionType = "negative"
	EmotionNeutral  EmotionType = "neutral"
)

type ProcessedMessage struct {
	Intent               IntentType  `json:"intent"`
	Entity               EntityType  `json:"entity"`
	Emotion              EmotionType `json:"emotion"`
	Confidence           float64     `json:"confidence"`
	ShouldReply          bool        `json:"should_reply"`
	ResponseMsg          string      `json:"response_msg"`
	ShouldGenerateTask   bool        `json:"should_generate_task"`
	ShouldGenerateAction bool        `json:"should_generate_action"`
}

type SocialMessage struct {
	Type        string
	Content     string
	Platform    string
	FromUser    string
	TargetUsers []string
	Context     map[string]interface{}
}

// SocialClientImpl handles social media interactions
type SocialClientImpl struct {
	twitterClient *clients.TwitterClient
	// discordBot    *client.DiscordBot
	socialMsgChannel chan SocialMessage
}

func NewSocialClient() *SocialClientImpl {
	return &SocialClientImpl{}
}

func (sc *SocialClientImpl) SendMessage(msg SocialMessage) error {
	switch msg.Platform {
	case "twitter":
		return sc.twitterClient.Tweet(msg.Content)
	// case "discord":
	// 	return sc.discordBot.SendMessage(msg.Content)
	case "all":
		// Send to all platforms
		if err := sc.twitterClient.Tweet(msg.Content); err != nil {
			return err
		}
		// return sc.discordBot.SendMessage(msg.Content)
	}
	return nil
}

func (sc *SocialClientImpl) GetMessageChannel() chan SocialMessage {
	return sc.socialMsgChannel
}

func parseAnalysis(response string) (*ProcessedMessage, error) {
	startTag := "```json\n"
	endTag := "\n```"

	startIndex := strings.Index(response, startTag)
	if startIndex == -1 {
		return nil, fmt.Errorf("start tag not found")
	}
	startIndex += len(startTag)

	endIndex := strings.Index(response[startIndex:], endTag)
	if endIndex == -1 {
		return nil, fmt.Errorf("end tag not found")
	}
	endIndex += startIndex

	jsonContent := response[startIndex:endIndex]

	var processedMsg ProcessedMessage
	if err := json.Unmarshal([]byte(jsonContent), &processedMsg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %w", err)
	}
	return &processedMsg, nil
}
