package clients

import (
	"context"
	"fmt"
	"time"

	"github.com/michimani/gotwi"
	"github.com/michimani/gotwi/fields"
	"github.com/michimani/gotwi/resources"
	"github.com/michimani/gotwi/tweet/managetweet"
	"github.com/michimani/gotwi/user/userlookup"
	"github.com/michimani/gotwi/user/userlookup/types"

	manageTypes "github.com/michimani/gotwi/tweet/managetweet/types"
	"github.com/michimani/gotwi/tweet/searchtweet"
	searchTypes "github.com/michimani/gotwi/tweet/searchtweet/types"
)

type TwitterConfig struct {
	APIKey       string `mapstructure:"api_key"`
	APIKeySecret string `mapstructure:"api_key_secret"`
	AccessToken  string `mapstructure:"access_token"`
	TokenSecret  string `mapstructure:"token_secret"`
}

type TwitterClient struct {
	client *gotwi.Client
	user   *resources.User
	tweets []resources.Tweet
}

type Tweet struct {
	ID     string
	Text   string
	UserID string
}

func NewTwitterClient(twitterConfig *TwitterConfig) (*TwitterClient, error) {
	in := &gotwi.NewClientInput{
		AuthenticationMethod: gotwi.AuthenMethodOAuth1UserContext,
		OAuthToken:           twitterConfig.AccessToken,
		OAuthTokenSecret:     twitterConfig.TokenSecret,
		APIKey:               twitterConfig.APIKey,
		APIKeySecret:         twitterConfig.APIKeySecret,
	}

	c, err := gotwi.NewClient(in)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	p := &types.GetMeInput{
		Expansions: fields.ExpansionList{
			fields.ExpansionPinnedTweetID,
		},
		UserFields: fields.UserFieldList{
			fields.UserFieldCreatedAt,
		},
		TweetFields: fields.TweetFieldList{
			fields.TweetFieldCreatedAt,
		},
	}

	u, err := userlookup.GetMe(context.Background(), c, p)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	fmt.Println("me is ", u.Data.Username)

	return &TwitterClient{
		client: c,
		user:   &u.Data,
		tweets: u.Includes.Tweets,
	}, nil
}

func (t *TwitterClient) MonitorMentioned(ctx context.Context) ([]*Tweet, error) {
	// TODO: config time duration
	startTime := time.Now().Add(-15 * time.Minute)
	l := &searchTypes.ListRecentInput{

		StartTime: &startTime,
		SortOrder: searchTypes.ListSortOrderRecency,
		Query:     fmt.Sprintf("@%s", *t.user.Username),
	}

	output, err := searchtweet.ListRecent(ctx, t.client, l)
	if err != nil {
		fmt.Println(err.Error())
		return nil, err
	}

	result := make([]*Tweet, 0)
	for _, tweet := range output.Data {
		result = append(result, &Tweet{
			ID:     *tweet.ID,
			Text:   *tweet.Text,
			UserID: *tweet.AuthorID,
		})
	}

	return result, nil
}

func (t *TwitterClient) GetMe() string {
	return *t.user.ID
}

func (t *TwitterClient) Tweet(ctx context.Context, tweet string) error {
	p := &manageTypes.CreateInput{
		Text: gotwi.String(tweet),
	}

	_, err := managetweet.Create(ctx, t.client, p)
	if err != nil {
		fmt.Println(err.Error())
		return err
	}
	return nil
}
