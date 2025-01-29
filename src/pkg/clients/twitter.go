package clients

import (
	"context"
	"fmt"

	"github.com/michimani/gotwi"
	"github.com/michimani/gotwi/fields"
	"github.com/michimani/gotwi/user/userlookup"
	"github.com/michimani/gotwi/user/userlookup/types"
)

type TwitterClient struct {
	client gotwi.Client
	user   *types.User
}

func NewTwitterClient() (*TwitterClient, error) {
	c, err := gotwi.NewClient(&gotwi.NewClientInput{
		AuthenticationMethod: gotwi.AuthenMethodOAuth2BearerToken,
	})
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	p := &types.GetByUsernameInput{
		Username: "michimani210",
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
	u, err := userlookup.GetByUsername(context.Background(), c, p)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	return &TwitterClient{
		client: c,
		user:   u,
	}, nil
}

func (t *TwitterClient) PostTweet(tweet string) error {
	return nil
}
