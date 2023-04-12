package handler

import (
    "context"
    "github.com/go-sonic/sonic/log"
    "github.com/go-sonic/sonic/service"
    "go.uber.org/zap"
    "time"
)

func StartGenContent(postService service.PostService, gptService service.GPTService) {
    // 获取当前时间
    now := time.Now()
    // 计算下一个0点的时间
    next := now.Add(time.Hour * 24)
    next = time.Date(next.Year(), next.Month(), next.Day(), 0, 0, 0, 0, next.Location())

    // 计算等待时间
    t := time.NewTimer(next.Sub(now))
    // 等待时间到达时执行任务
    go func() {
        <-t.C
        // 执行任务
        posts, err := gptService.GenContent()
        if err != nil {
            log.Error("gen content error", zap.Error(err))
        } else {
            for _, post := range posts {
                _, _ = postService.Create(context.TODO(), &post)
            }
        }

        // 定期执行任务
        ticker := time.NewTicker(time.Hour * 24)
        for {
            select {
            case <-ticker.C:
                // 执行任务
                posts, err = gptService.GenContent()
                if err != nil {
                    log.Error("gen content error", zap.Error(err))
                    continue
                }
                for _, post := range posts {
                    _, _ = postService.Create(context.TODO(), &post)
                }
            }
        }
    }()
}
