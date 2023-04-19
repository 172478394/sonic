package handler

import (
    "context"
    "github.com/go-sonic/sonic/log"
    "github.com/go-sonic/sonic/service"
    "go.uber.org/zap"
    "time"
)

func StartGenContent(gptService service.GPTService) {
    categoryMap := map[int32]string{
        2: "星座",
        3: "风水",
        4: "解梦",
        5: "取名",
    }

    ctx := context.TODO()
    go func() {
        for {
            // 获取当前时间
            now := time.Now()
            // 计算下一个0点的时间
            next := now.Add(time.Hour * 24)
            next = time.Date(next.Year(), next.Month(), next.Day(), 0, 0, 0, 0, next.Location())

            // 计算等待时间
            t := time.NewTimer(next.Sub(now))
            // 等待时间到达时执行任务
            <-t.C
            // 执行任务
            for i := 0; i < 3; i++ {
                for categoryId, categoryName := range categoryMap {
                    keywords, err := gptService.GenKeywords(ctx, categoryName, categoryId)
                    if err != nil {
                        log.Error("gen keywords error", zap.Error(err))
                        continue
                    }
                    _ = gptService.GenContent(ctx, keywords)
                }
            }
            log.Debug("文章生成完毕，等待下一次执行")
        }
    }()
}
