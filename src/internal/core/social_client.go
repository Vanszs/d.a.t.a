package core

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/carv-protocol/d.a.t.a/src/pkg/clients"
)

type SocialClient interface {
	SendMessage(msg SocialMessage) error
	GetMessageChannel() chan SocialMessage
	MonitorMessages(ctx context.Context) error
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
	Metadata    map[string]interface{}
}

// SocialClientImpl handles social media interactions
type SocialClientImpl struct {
	twitterClient    *clients.TwitterClient
	discordBot       *clients.DiscordBot
	socialMsgChannel chan SocialMessage
}

func NewSocialClient(
	twitterConfig *clients.TwitterConfig,
	discordConfig *clients.DiscordConfig,
) *SocialClientImpl {
	cli := &SocialClientImpl{
		socialMsgChannel: make(chan SocialMessage),
	}
	if twitterConfig != nil {
		client, err := clients.NewTwitterClient(twitterConfig)
		if err != nil {
			fmt.Println("generate client err:", err)
			panic(err)
		}
		cli.twitterClient = client
	}
	if discordConfig != nil {
		cli.discordBot = clients.NewDiscordBot(discordConfig.APIToken)
	}

	return cli
}

func (sc *SocialClientImpl) SendMessage(msg SocialMessage) error {
	switch msg.Platform {
	case "twitter":
		return sc.twitterClient.Tweet(context.Background(), msg.Content)
	case "discord":
		return sc.discordBot.SendMessage(context.Background(), &clients.DiscordMsg{
			AuthorID:  msg.FromUser,
			Content:   msg.Content,
			ChannelID: msg.Metadata["channel_id"].(string),
		})
	case "all":
		// Send to all platforms
		if err := sc.twitterClient.Tweet(context.Background(), msg.Content); err != nil {
			return err
		}
		// return sc.discordBot.SendMessage(msg.Content)
	}

	return nil
}

func (sc *SocialClientImpl) GetMessageChannel() chan SocialMessage {
	return sc.socialMsgChannel
}

func (sc *SocialClientImpl) MonitorMessages(ctx context.Context) error {
	var wg sync.WaitGroup
	if sc.twitterClient != nil {
		wg.Add(1)
		go func() {
			defer wg.Done()
			sc.monitorTwitter(ctx)
		}()
	}

	if sc.discordBot != nil {
		// TODO fix context
		wg.Add(1)
		go func() {
			defer wg.Done()
			sc.monitorDiscord(ctx)
		}()
	}

	wg.Wait()
	return nil
}

func (sc *SocialClientImpl) monitorTwitter(ctx context.Context) {
	ticker := time.NewTicker(15 * time.Minute)
	defer ticker.Stop()

	// fmt.Println("try to monitor twitter")

	for {
		select {
		case <-ticker.C:
			tweets, err := sc.twitterClient.MonitorMentioned(context.Background())
			if err != nil {
				// TODO: handle error
				fmt.Println("monitor twitter err:", err)
				return
			}

			fmt.Println("got tweet!!!", tweets)
			for _, tweet := range tweets {
				sc.socialMsgChannel <- SocialMessage{
					Type:        "mention",
					Content:     tweet.Text,
					Platform:    "twitter",
					FromUser:    tweet.UserID,
					TargetUsers: []string{sc.twitterClient.GetMe()},
				}
			}
		case <-ctx.Done():
			return
		}
	}
}

func (sc *SocialClientImpl) monitorDiscord(ctx context.Context) {
	channel := sc.discordBot.GetMessageChannel()

	for {
		select {
		case msg := <-channel:
			fmt.Println("got discord msg!!!", msg)
			sc.socialMsgChannel <- SocialMessage{
				Type:     "message",
				Content:  msg.Content,
				Platform: "discord",
				FromUser: msg.AuthorID,
				Metadata: map[string]interface{}{"channel_id": msg.ChannelID},
			}
		case <-ctx.Done():
			return
		}
	}
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
