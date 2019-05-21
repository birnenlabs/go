package spotify

import (
	"birnenlabs.com/oauth"
	"birnenlabs.com/ratelimit"
	"context"
	"time"
)

type Spotify struct {
	// Client with 1 qps limit
	httpClient ratelimit.AnyClient
	// market to search songs (e.g. "pl")
	market string
}

func New(ctx context.Context, market string) (*Spotify, error) {
	// First create OAuth.
	oauthClient, err := oauth.Create("spotify")
	if err != nil {
		return nil, err
	}

	// Verify the token
	err = oauthClient.VerifyToken(ctx)
	if err != nil {
		return nil, err
	}

	// Get http client with Bearer
	httpClient, err := oauthClient.CreateAuthenticatedHttpClient(ctx)
	if err != nil {
		return nil, err
	}

	return &Spotify{
		httpClient: ratelimit.New(httpClient, time.Second),
		market:     market,
	}, nil
}
