package core

import "github.com/carv-protocol/d.a.t.a/src/pkg/clients"

type SocialClient interface {
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
