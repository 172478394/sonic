package impl

import (
    "errors"
    "github.com/cloudwego/kitex/client"
    "github.com/go-sonic/sonic/config"
    "github.com/go-sonic/sonic/kitex_gen/api/chat"
    "github.com/go-sonic/sonic/model/param"
    "github.com/go-sonic/sonic/service"
    "github.com/go-sonic/sonic/service/assembler"
    "time"
)

type gptServiceImpl struct {
    service.BasePostService
    client *chat.Client
}

func NewGPTService(conf *config.Config, postTagService service.PostTagService, postCategoryService service.PostCategoryService, categoryService service.CategoryService, postAssembler assembler.PostAssembler, tagService service.TagService) service.GPTService {
    chatClient, err := chat.NewClient("chat", client.WithHostPorts(conf.Gpt), client.WithConnectTimeout(5*time.Second))
    if err != nil {
        panic(err)
    }
    return &gptServiceImpl{
        client: &chatClient,
    }
}

func (g *gptServiceImpl) GenContent() ([]param.Post, error) {
    return nil, errors.New("not implemented")
}
