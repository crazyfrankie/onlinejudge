package github

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"oj/internal/user/domain"
)

var (
	redirectUri = url.PathEscape("http://localhost:9000/oauth/github/callback")
)

type Service interface {
	AuthUrl(ctx context.Context, state string) (string, error)
	VerifyCode(ctx context.Context, code string) (Result, error)
	AcquireUserInfo(ctx context.Context, token string) (domain.GithubInfo, error)
}

type AuthService struct {
	clientId     string
	clientSecret string
	client       *http.Client
}

type Result struct {
	AccessToken string `json:"access_token"`
	Scope       string `json:"scope"`
	TokenType   string `json:"token_type"`
}

func NewService() Service {
	return &AuthService{
		clientId:     "Ov23li68AdQDNiAvw3O8",
		clientSecret: "a44e01be45bb0d80e8d065aa62f385b83baed815",
		client:       http.DefaultClient,
	}
}

func (svc *AuthService) AuthUrl(ctx context.Context, state string) (string, error) {
	const urlPattern = "https://github.com/login/oauth/authorize?client_id=%s&redirect_uri=%s&state=%s"
	return fmt.Sprintf(urlPattern, svc.clientId, redirectUri, state), nil
}

func (svc *AuthService) VerifyCode(ctx context.Context, code string) (Result, error) {
	const targetPattern = "https://github.com/login/oauth/access_token?client_id=%s&client_secret=%s&code=%s&redirect_uri=%s"
	target := fmt.Sprintf(targetPattern, svc.clientId, svc.clientSecret, code, redirectUri)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, target, nil)
	if err != nil {
		return Result{}, err
	}

	req.Header.Set("Accept", "application/json")

	var resp *http.Response

	resp, err = svc.client.Do(req)
	if err != nil {
		return Result{}, err
	}

	var res Result
	decoder := json.NewDecoder(resp.Body)
	if err = decoder.Decode(&res); err != nil {
		return Result{}, err
	}

	return res, nil
}

func (svc *AuthService) AcquireUserInfo(ctx context.Context, token string) (domain.GithubInfo, error) {
	target := "https://api.github.com/user"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, target, nil)
	if err != nil {
		return domain.GithubInfo{}, err
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	var resp *http.Response
	resp, err = svc.client.Do(req)
	if err != nil {
		return domain.GithubInfo{}, err
	}

	var userInfo domain.GithubInfo
	decoder := json.NewDecoder(resp.Body)
	if err := decoder.Decode(&userInfo); err != nil {
		return domain.GithubInfo{}, err
	}

	return userInfo, nil
}
