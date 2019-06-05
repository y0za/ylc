package main

import (
	"context"
	"fmt"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

type OAuthManager struct {
	config    *oauth2.Config
	tokenRepo TokenRepository
}

func NewOAuthManager(clientID, clientSecret string, tokenRepo TokenRepository) *OAuthManager {
	config := &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Endpoint:     google.Endpoint,
		RedirectURL:  "urn:ietf:wg:oauth:2.0:oob",
		Scopes: []string{
			"https://www.googleapis.com/auth/youtube",
			"https://www.googleapis.com/auth/youtube.readonly",
			"https://www.googleapis.com/auth/youtube.force-ssl",
		},
	}
	return &OAuthManager{
		config:    config,
		tokenRepo: tokenRepo,
	}
}

func (oa *OAuthManager) requestOAuthToken(ctx context.Context) (*oauth2.Token, error) {
	fmt.Println("Access to this URL and get auth code.")
	fmt.Println(oa.config.AuthCodeURL(""))
	fmt.Print("Input auth code: ")
	var code string
	_, err := fmt.Scanf("%s\n", &code)
	if err != nil {
		return nil, fmt.Errorf("failed to scan auth code: %v", err)
	}

	return oa.config.Exchange(ctx, code)
}

type TokenRepository interface {
	Save(*oauth2.Token) error
	Load() (*oauth2.Token, error)
}

type cachedTokenSource struct {
	base oauth2.TokenSource
	repo TokenRepository
}

func (cts *cachedTokenSource) Token() (*oauth2.Token, error) {
	t, err := cts.base.Token()
	if err != nil {
		return nil, err
	}
	if err := cts.repo.Save(t); err != nil {
		return nil, err
	}
	return t, nil
}

func (oa *OAuthManager) TokenSource(ctx context.Context) (oauth2.TokenSource, error) {
	t, err := oa.tokenRepo.Load()
	if err != nil {
		// if cached token doesn't exist, request new token
		t, err = oa.requestOAuthToken(ctx)
		if err != nil {
			return nil, err
		}
	}

	return &cachedTokenSource{
		base: oa.config.TokenSource(ctx, t),
		repo: oa.tokenRepo,
	}, nil
}
