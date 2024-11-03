package web

import (
	"bytes"
	"encoding/json"
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
	"net/http"
	"net/http/httptest"
	"oj/internal/article/domain"
	"oj/internal/article/service"
	"oj/internal/article/service/svcmocks"
	ijwt "oj/internal/user/middleware/jwt"
	"testing"
)

func TestArticleHandler_Publish(t *testing.T) {
	testCases := []struct {
		name string

		mock func(ctrl *gomock.Controller) service.ArticleService

		reqBody string

		wantCode int

		wantRes Response
	}{
		{
			name: "新建并发表",
			reqBody: `
					{
						"title":"我的标题",
						"content":"我的内容"
					}`,
			mock: func(ctrl *gomock.Controller) service.ArticleService {
				svc := svcmocks.NewMockArticleService(ctrl)
				svc.EXPECT().Publish(gomock.Any(), domain.Article{
					Title:   "我的标题",
					Content: "我的内容",
					Author: domain.Author{
						Id: 1,
					},
				}).Return(uint64(1), nil)
				return svc
			},
			wantCode: http.StatusOK,
			wantRes: Response{
				Data: float64(1),
				Msg:  "OK",
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			server := server.Default()

			server.Use(func(c *app.RequestContext) {
				c.Set("claims", ijwt.Claims{
					Id: 1,
				})
			})

			h := NewArticleHandler(tc.mock(ctrl), zap.NewNop())
			h.RegisterRoute(server)

			req, err := http.NewRequest(http.MethodPost, "/articles/publish", bytes.NewBuffer([]byte(tc.reqBody)))
			require.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")

			resp := httptest.NewRecorder()
			// 这就是 HTTP 请求进入 Gin 的地方
			// 当这样调用时，Gin 就会处理这个请求
			// 响应写回到 resp 里
			server.ServeHTTP(resp, req)

			assert.Equal(t, tc.wantCode, resp.Code)
			var webRes Response
			err = json.NewDecoder(resp.Body).Decode(&webRes)
			require.NoError(t, err)
			assert.Equal(t, tc.wantRes, webRes)
		})
	}
}
