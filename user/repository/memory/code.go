package memory

import (
	"context"
	"fmt"
)

type CodeMem struct {
}

func NewCodeMem() *CodeMem {
	return &CodeMem{}
}

func (cm *CodeMem) Set(ctx context.Context, biz, phone, code string) error {

}

func (cm *CodeMem) Verify(ctx context.Context, biz, phone, code string) error {

}

func (cm *CodeMem) key(biz, phone string) string {
	return fmt.Sprintf("phone_code:%s:%s", biz, phone)
}
