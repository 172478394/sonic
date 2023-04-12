package service

import (
    "github.com/go-sonic/sonic/model/param"
)

type GPTService interface {
    GenContent() ([]param.Post, error)
}
