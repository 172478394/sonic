package impl

import (
    "context"
    "encoding/json"
    "errors"
    "fmt"
    "github.com/cloudwego/kitex/client"
    "github.com/go-sonic/sonic/config"
    "github.com/go-sonic/sonic/consts"
    "github.com/go-sonic/sonic/kitex_gen/api"
    "github.com/go-sonic/sonic/kitex_gen/api/chat"
    "github.com/go-sonic/sonic/model/dto"
    "github.com/go-sonic/sonic/model/param"
    "github.com/go-sonic/sonic/service"
    "github.com/go-sonic/sonic/util"
    "github.com/russross/blackfriday/v2"
    "time"
)

type gptServiceImpl struct {
    service.BasePostService
    client chat.Client
}

func NewGPTService(conf *config.Config) service.GPTService {
    chatClient, err := chat.NewClient("chat", client.WithHostPorts(conf.Gpt), client.WithConnectTimeout(5*time.Second))
    if err != nil {
        panic(err)
    }
    return &gptServiceImpl{
        client: chatClient,
    }
}

func (g *gptServiceImpl) GenKeywords(ctx context.Context) ([]dto.Keyword, error) {
    keywordPrompt := "I want you to act as a market research expert that speaks and writes fluent Chinese. Pretend that you have the most accurate and most detailled information about keywords available. Pretend that you are able to develop a full SEO content plan in fluent Chinese. I will give you the target keyword %s. From this keyword create a markdown table with a keyword list for an SEO content strategy plan on the topic %s. This form is required to have 20 records. Cluster the keywords according to the top 100 super categories, Randomly select categories from these 100 categories without sorting them in order each time, and name the super category in the first column called Tag. Add in another column with 7 subcategories for each keyword cluster or specific long-tail keywords for each of the clusters. Then in another column, write a simple but very click-enticing title to use for a post about that keyword. Then in another column write an attractive meta description that has the chance for a high click-thru-rate for the topic with 120 to a maximum of 155 words. The meta description shall be value based, so mention value of the article and have a simple call to action to cause the searcher to click. Do NOT under any circumstance use too generic keyword like `introduction`  or `conclusion` or `tl:dr`. Focus on the most specific keywords only. Do not use single quotes, double quotes or any other enclosing characters in any of the columns you fill in. Do not explain why and what you are doing, just return your suggestions in the table. The markdown table shall be in Chinese language and have the following columns:  TAG, keyword,  title, meta description."
    categoryMap := map[int]string{
        2: "星座",
        3: "风水",
        4: "解梦",
        5: "取名",
    }
    for categoryId, categoryName := range categoryMap {
        var request api.Request
        request.Temperature = 0.75
        request.Messages = []*api.ChatCompletionMessage{
            //{
            //    Role:    "system",
            //    Content: systemPrompt,
            //},
            {
                Role:    "user",
                Content: fmt.Sprintf(keywordPrompt, categoryName, categoryName),
            },
        }
        resp, _err := g.client.Completion(ctx, &request)
        if _err != nil {
            return nil, _err
        }
    }
}

func (g *gptServiceImpl) GenContent(ctx context.Context) ([]param.Post, error) {
    fileName := "./keywords/" + time.Now().Local().Format("2006-01-02") + ".json"
    if !util.FileIsExisted(fileName) {
        return nil, errors.New("file not exists")
    }
    dataBytes, err := util.ReadFile(fileName)
    if err != nil {
        return nil, err
    }

    var keywords dto.Keyword
    err = json.Unmarshal(dataBytes, &keywords)
    if err != nil {
        return nil, err
    }

    posts := make([]param.Post, 0, len(keywords.Data))
    //systemPrompt := "I Want You To Act As A Content Writer Very Proficient SEO Writer Writes Fluently Chinese."
    userPrompt := "I Want You To Act As A Content Writer Very Proficient SEO Writer Writes Fluently Chinese. Start with an introduction paragraph. Write a 3000-word 100% Unique, SEO-optimized, Human-Written article in Chinese with at least 15 headings and subheadings that covers the topic provided in the Prompt. Write The article In Your Own Words Rather Than Copying And Pasting From Other Sources. Consider perplexity and burstiness when creating content, ensuring high levels of both without losing specificity or context. Use fully detailed paragraphs that engage the reader. Write In A Conversational Style As Written By A Human (Use An Informal Tone, Utilize Personal Pronouns, Keep It Simple, Engage The Reader, Use The Active Voice, Keep It Brief, Use Rhetorical Questions, and Incorporate Analogies And Metaphors). End with a conclusion paragraph. Using Markdown language to write the article. Now Write An Article On This Topic '%s'"
    for i, data := range keywords.Data {
        var request api.Request
        request.Temperature = 0.75
        request.Messages = []*api.ChatCompletionMessage{
            //{
            //    Role:    "system",
            //    Content: systemPrompt,
            //},
            {
                Role:    "user",
                Content: fmt.Sprintf(userPrompt, data.Title),
            },
        }
        resp, _err := g.client.Completion(ctx, &request)
        if _err != nil {
            return nil, _err
        }
        //log.Debug("resp:", zap.String("content", resp.Choices[0].Message.Content))
        postParam := param.Post{
            Title:           data.Title,
            Status:          consts.PostStatusPublished,
            Slug:            fmt.Sprintf("article-%s-%d", time.Now().Local().Format("20060102"), i+1),
            OriginalContent: resp.Choices[0].Message.Content,
            Content:         util.Bytes2str(blackfriday.Run(util.Str2bytes(resp.Choices[0].Message.Content))),
            CategoryIDs:     data.Category,
            TagIDs:          data.TagIds,
            MetaKeywords:    data.Keyword,
            MetaDescription: data.MetaDescription,
        }
        posts = append(posts, postParam)
    }

    return posts, nil
}
