package core

import "github.com/carv-protocol/d.a.t.a/src/pkg/clients"

type SocialClient interface {
	SendMessage(msg SocialMessage) error
}

type SocialMessage struct {
	Type        string
	Content     string
	Platform    string
	TargetUsers []string
	Context     map[string]interface{}
}

// SocialClientImpl handles social media interactions
type SocialClientImpl struct {
	twitterClient *clients.TwitterClient
	// discordBot    *client.DiscordBot
	outputQueue chan SocialMessage
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
