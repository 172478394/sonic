package service

import (
    "context"
    "github.com/go-sonic/sonic/model/param"
)

type GPTService interface {
    GenContent(ctx context.Context) ([]param.Post, error)
}
