package wechat

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"oj/internal/user/domain"
)

var (
	redirectUri = url.PathEscape("https://qiyi.com/oauth/wechat/callback")
)

type Service interface {
	AuthURL(ctx context.Context, state string) (string, error)
	VerifyCode(ctx context.Context, code string) (domain.WeChatInfo, error)
}

type AuthSvc struct {
	appId     string
	appSecret string
	client    *http.Client
}

type Result struct {
	ErrCode int64  `json:"errCode"`
	ErrMsg  string `json:"errMsg"`

	AccessToken  string `json:"access_token"`
	ExpiresIn    int64  `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`

	OpenId  string `json:"openid"`
	Scope   string `json:"scope"`
	UnionId string `json:"unionid"`
}

func NewService(appId string, appSecret string) Service {
	return &AuthSvc{
		appId:     appId,
		appSecret: appSecret,
		client:    http.DefaultClient,
	}
}

func (s *AuthSvc) AuthURL(ctx context.Context, state string) (string, error) {
	const urlPattern = "https://open.weixin.qq.com/connect/qrconnect?appid=%s&redirect_uri=%s&response_type=code&scope=snsapi_login&state=%s#wechat_redirect"
	return fmt.Sprintf(urlPattern, s.appId, redirectUri, state), nil
}

func (s *AuthSvc) VerifyCode(ctx context.Context, code string) (domain.WeChatInfo, error) {
	const targetPattern = "https://api.weixin.qq.com/sns/oauth2/access_token?appid=%s&secret=%s&code=%s&grant_type=authorization_code"
	target := fmt.Sprintf(targetPattern, s.appId, s.appSecret, code)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, target, nil)
	if err != nil {
		return domain.WeChatInfo{}, err
	}

	var resp *http.Response
	resp, err = s.client.Do(req)
	if err != nil {
		return domain.WeChatInfo{}, err
	}

	decoder := json.NewDecoder(resp.Body)
	var res Result
	err = decoder.Decode(&res)
	if err != nil {
		return domain.WeChatInfo{}, err
	}

	if res.ErrCode != 0 {
		return domain.WeChatInfo{}, fmt.Errorf("微信返回错误码 %d，错误信息 %s", res.ErrCode, res.ErrMsg)
	}

	return domain.WeChatInfo{
		OpenID:  res.OpenId,
		UnionID: res.UnionId,
	}, nil
}
