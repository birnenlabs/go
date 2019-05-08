package spotify

import (
	"context"
	"ojoj.ch/oauth"
)

type Spotify struct {
	oauthClient *oauth.OAuth
}

func New(ctx context.Context) *Spotify {
	// First create OAuth.
	oauthClient, err := oauth.Create("spotify")
	if err != nil {
		panic(err)
	}

	// Verify the token
	err = oauthClient.VerifyToken(ctx)
	if err != nil {
		panic(err)
	}

	return &Spotify{
		oauthClient: oauthClient,
	}
}
