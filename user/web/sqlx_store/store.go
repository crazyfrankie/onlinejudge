package sqlx_store

import (
	"net/http"

	ginSession "github.com/gin-contrib/sessions"
	"github.com/gorilla/sessions"
)

type Store struct {
}

func (st *Store) Get(r *http.Request, name string) (*sessions.Session, error) {
	//TODO implement me
	panic("implement me")
}

func (st *Store) New(r *http.Request, name string) (*sessions.Session, error) {
	//TODO implement me
	panic("implement me")
}

func (st *Store) Save(r *http.Request, w http.ResponseWriter, s *sessions.Session) error {
	//TODO implement me
	panic("implement me")
}

func (st *Store) Options(options ginSession.Options) {
	//TODO implement me
	panic("implement me")
}
