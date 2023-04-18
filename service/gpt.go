package service

import (
    "context"
    "github.com/go-sonic/sonic/model/dto"
)

type GPTService interface {
    GenContent(ctx context.Context, keywords []dto.Keyword) error
    GenKeywords(ctx context.Context, categoryName string, categoryId int32) ([]dto.Keyword, error)
}
