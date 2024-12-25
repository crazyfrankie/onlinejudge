package integration

import (
	"bytes"
	"encoding/json"
	"github.com/crazyfrankie/onlinejudge/internal/article/repository/dao"
	"github.com/crazyfrankie/onlinejudge/ioc"
	"gorm.io/gorm"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// ArticleTestSuite 测试套件
type ArticleTestSuite struct {
	suite.Suite
	server *gin.Engine
	db     *gorm.DB
}

func TestArticle(t *testing.T) {
	suite.Run(t, &ArticleTestSuite{})
}

func (s *ArticleTestSuite) SetupSuite() {
	s.server = ioc.InitGin()

	db := InitDB()
	s.db = db
}

func (s *ArticleTestSuite) TearDownTest() {
	s.db.Exec("TRUNCATE TABLE article")
}

func (s *ArticleTestSuite) TestArticleHandler_Edit() {
	t := s.T()
	testCases := []struct {
		name string

		// 集成测试准备数据
		before func(t *testing.T)
		// 集成测试验证数据
		after func(t *testing.T)

		// 预期输入
		art Article
		// HTTP 响应码
		wantCode int
		// 希望带上文章的 ID
		wantResult Result[uint64]
	}{
		{
			name: "新建帖子",
			before: func(t *testing.T) {

			},
			after: func(t *testing.T) {
				// 验证数据库
				var art dao.Article

				err := s.db.Where("id = ?", 1).First(&art).Error
				assert.NoError(t, err)
				assert.True(t, art.Ctime > 0)
				assert.True(t, art.Utime > 0)
				art.Ctime = 0
				art.Utime = 0
				assert.Equal(t, dao.Article{
					ID:       1,
					Title:    "我的标题",
					Content:  "我的内容",
					AuthorID: 1,
				}, art)
			},

			art: Article{
				Title:   "我的标题",
				Content: "我的内容",
			},
			wantCode: http.StatusOK,
			wantResult: Result[uint64]{
				Data: 1,
				Msg:  "OK",
			},
		},
		{
			name: "更新帖子",
			before: func(t *testing.T) {
				err := s.db.Create(&dao.Article{
					ID:       2,
					Title:    "我的标题",
					Content:  "我的内容",
					AuthorID: 1,
					// 跟时间有关的测试，不必要不要用 time.Now()
					// 很难断言
					Ctime: 123,
					Utime: 234,
				}).Error
				assert.NoError(t, err)
			},
			after: func(t *testing.T) {
				// 验证数据库
				var art dao.Article

				err := s.db.Where("id = ?", 2).First(&art).Error
				assert.NoError(t, err)
				assert.True(t, art.Utime > 234)
				// 为了确保我更新了
				art.Utime = 0
				assert.Equal(t, dao.Article{
					ID:       2,
					Title:    "新的标题",
					Content:  "新的内容",
					Ctime:    123,
					AuthorID: 1,
				}, art)
			},

			art: Article{
				ID:      2,
				Title:   "新的标题",
				Content: "新的内容",
			},
			wantCode: http.StatusOK,
			wantResult: Result[uint64]{
				Data: 2,
				Msg:  "OK",
			},
		},
		{
			name: "修改别人的帖子",
			before: func(t *testing.T) {
				err := s.db.Create(&dao.Article{
					ID:       3,
					Title:    "我的标题",
					Content:  "我的内容",
					AuthorID: 2,
					// 跟时间有关的测试，不必要不要用 time.Now()
					// 很难断言
					Ctime: 123,
					Utime: 234,
				}).Error
				assert.NoError(t, err)
			},
			after: func(t *testing.T) {
				// 验证数据库
				var art dao.Article

				err := s.db.Where("id = ?", 3).First(&art).Error
				assert.NoError(t, err)
				assert.Equal(t, dao.Article{
					ID:       3,
					Title:    "我的标题",
					Content:  "我的内容",
					Ctime:    123,
					Utime:    234,
					AuthorID: 2,
				}, art)
			},

			art: Article{
				ID:      3,
				Title:   "新的标题",
				Content: "新的内容",
			},
			wantCode: http.StatusInternalServerError,
			wantResult: Result[uint64]{
				Data: 0,
				Msg:  "system errors",
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// 构造请求
			// 执行
			// 验证结果
			tc.before(t)
			reqBody, err := json.Marshal(tc.art)
			assert.NoError(t, err)
			req, err := http.NewRequest(http.MethodPost, "/articles/edit", bytes.NewBuffer(reqBody))
			require.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")

			resp := httptest.NewRecorder()
			// 这就是 HTTP 请求进入 Gin 的地方
			// 当这样调用时，Gin 就会处理这个请求
			// 响应写回到 resp 里
			s.server.ServeHTTP(resp, req)

			assert.Equal(t, tc.wantCode, resp.Code)
			var webRes Result[uint64]
			err = json.NewDecoder(resp.Body).Decode(&webRes)
			require.NoError(t, err)
			assert.Equal(t, tc.wantResult, webRes)
			tc.after(t)
		})
	}
}

type Article struct {
	ID      uint64 `json:"id"`
	Title   string `json:"title"`
	Content string `json:"content"`
}

type Result[T any] struct {
	Data T      `json:"data"`
	Msg  string `json:"msg"`
	Code int    `json:"code"`
}
