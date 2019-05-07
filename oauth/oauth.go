// Library that helps to create http client with OAuth authentication headers.
package oauth

import (
	"context"
	"fmt"
	"golang.org/x/oauth2"
	"net/http"
	"ojoj.ch/conf"
	"os"

	"github.com/golang/glog"
)

type OAuth struct {
	conf     *oauth2.Config
	filePath string
}

// Creates the client, clientName will be used to specify configuration files.
func Create(clientName string) (*OAuth, error) {
	var config oauth2.Config
	err := conf.LoadConfigFromJson("oauth-"+clientName, &config)
	if err != nil {
		return nil, err
	}

	return &OAuth{
		filePath: os.Getenv("HOME") + "/.credentials/oauth-" + clientName + ".gob",
		conf:     &config,
	}, nil
}

func (o *OAuth) VerifyToken(ctx context.Context) error {
	_, err := o.getToken(ctx, true)
	return err
}

func (o *OAuth) CreateAuthenticatedHttpClient(ctx context.Context) (*http.Client, error) {
	token, err := o.getToken(ctx, false)
	if err != nil {
		return nil, err
	}
	client, err := o.conf.Client(ctx, token), nil
	if err != nil {
		return nil, err
	}
	return client, nil
}

func (o *OAuth) loadToken() (*oauth2.Token, error) {
	var token oauth2.Token
	err := conf.LoadFromFile(o.filePath, &token)
	return &token, err
}

func (o *OAuth) getToken(ctx context.Context, interactive bool) (*oauth2.Token, error) {
	token, err := o.loadToken()
	if err == nil {
		return token, nil
	} else {
		glog.Warningf("Could not load token (%v), will try interactive.", err)
	}
	if !interactive {
		return nil, fmt.Errorf("file with the token was not found")
	}
	url := o.conf.AuthCodeURL("state", oauth2.AccessTypeOffline)
	fmt.Printf("Visit the URL for the auth dialog:\n\n%v\n\nand paste the code:\n", url)

	var code string
	_, err = fmt.Scan(&code)
	if err != nil {
		return nil, err
	}
	token, err = o.conf.Exchange(ctx, code)
	if err != nil {
		return nil, err
	}
	err = conf.SaveToFile(o.filePath, token)
	if err != nil {
		fmt.Printf("Could not save token: %v\n", err)
	}
	return token, nil
}
