package spotify

import (
	"birnenlabs.com/oauth"
	"birnenlabs.com/ratelimit"
	"context"
	"time"
)

type connector struct {
	// Client with 1 qps limit
	httpClient ratelimit.AnyClient
	// market to search songs (e.g. "pl")
	market string
}

type Spotify struct {
	connector *connector
	cache     Cache
}

func newConnector(ctx context.Context, market string) (*connector, error) {
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

	return &connector{
		httpClient: ratelimit.New(httpClient, time.Second),
		market:     market,
	}, nil
}

func New(ctx context.Context, market string) (*Spotify, error) {
	c, err := newConnector(ctx, market)
	if err != nil {
		return nil, err
	}
	return &Spotify{
		connector: c,
		cache:     newCache(),
	}, nil
}
