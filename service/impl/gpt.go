package impl

import (
    "context"
    "errors"
    "fmt"
    "github.com/cloudwego/kitex/client"
    "github.com/go-sonic/sonic/config"
    "github.com/go-sonic/sonic/consts"
    "github.com/go-sonic/sonic/kitex_gen/api"
    "github.com/go-sonic/sonic/kitex_gen/api/chat"
    "github.com/go-sonic/sonic/log"
    "github.com/go-sonic/sonic/model/dto"
    "github.com/go-sonic/sonic/model/param"
    "github.com/go-sonic/sonic/service"
    "github.com/go-sonic/sonic/util"
    "github.com/go-sonic/sonic/util/xerr"
    "github.com/mozillazg/go-pinyin"
    "github.com/russross/blackfriday/v2"
    "go.uber.org/zap"
    "gorm.io/gorm"
    "strings"
    "time"
)

type gptServiceImpl struct {
    service.BasePostService
    client      chat.Client
    PostService service.PostService
    TagService  service.TagService
}

func NewGPTService(conf *config.Config, postService service.PostService, tagService service.TagService) service.GPTService {
    chatClient, err := chat.NewClient("chat", client.WithHostPorts(conf.Gpt), client.WithConnectTimeout(5*time.Second))
    if err != nil {
        panic(err)
    }
    return &gptServiceImpl{
        client:      chatClient,
        PostService: postService,
        TagService:  tagService,
    }
}

func (g *gptServiceImpl) GenKeywords(ctx context.Context, categoryName string, categoryId int32) ([]dto.Keyword, error) {
    keywordPrompt := "I want you to act as a market research expert that speaks and writes fluent Chinese. Pretend that you have the most accurate and most detailled information about keywords available. Pretend that you are able to develop a full SEO content plan in fluent Chinese. I will give you the target keyword %s. From this keyword create a markdown table with a keyword list for an SEO content strategy plan on the topic %s. This form is required to have 20 records. Cluster the keywords according to the top 100 super categories, Randomly select categories from these 100 categories without sorting them in order each time, and name the super category in the first column called Tag. Add in another column with 7 subcategories for each keyword cluster or specific long-tail keywords for each of the clusters. Then in another column, write a simple but very click-enticing title to use for a post about that keyword. Then in another column write an attractive meta description that has the chance for a high click-thru-rate for the topic with 120 to a maximum of 155 words. The meta description shall be value based, so mention value of the article and have a simple call to action to cause the searcher to click. Do NOT under any circumstance use too generic keyword like `introduction` or `conclusion` or `tl:dr`. Focus on the most specific keywords only. Do not use single quotes, double quotes or any other enclosing characters in any of the columns you fill in. Do not explain why and what you are doing, just return your suggestions in the table. The markdown table shall be written entirely in Chinese language and have the following columns: tag, keyword, title, meta description."
    //keywordPrompt := "我希望您扮演一位能够书写流利中文的市场调研专家。假设您拥有最准确、最详细的关键词信息。假设您能够使用流利的中文制定完整的SEO内容计划。现在给您一个目标关键词“%s”。请根据这个关键词，创建一个带有关键词列表的Markdown表格，用于“%s”主题的SEO内容战略计划。这个表格需要包含20条记录。根据前100个超级分类对关键词进行分组，每次随机选择这100个分类中的分类，并在第一列中命名为“标签”。为每个关键词簇或特定的长尾关键词添加另一列中的7个子分类。接下来，在另一列中，编写一个简单但非常引人点击的标题，用于生成关于该关键词的文章。然后，在另一列中，编写一个有吸引力的meta描述，描述文章的价值，并包含120到155个字的简单呼叫性行动，以引导搜索者点击。meta描述应基于价值，不要使用太过通用的关键词，如“简介”、“结论”或“tl:dr”。只专注于最具体的关键词。在填写的任何列中，请不要使用单引号、双引号或任何其他封闭字符。不需要解释您的操作和目的，只需在表格中返回您的建议。Markdown表格应使用中文，并包含以下列：标签、关键词、标题、meta描述。"

    keywords := make([]dto.Keyword, 0)
    var request api.Request
    //request.Temperature = 0.75
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
    content := resp.Choices[0].Message.Content
    contents := strings.Split(content, "\n")
    if len(contents) < 3 {
        return nil, xerr.WithMsg(nil, "content len < 3")
    }
    contents = contents[2:]
    for _, row := range contents {
        rows := strings.Split(row, "|")
        if len(rows) < 6 {
            continue
        }

        rowKeyword := strings.TrimSpace(rows[2])
        title := strings.TrimSpace(rows[3])
        metaDescription := strings.TrimSpace(rows[4])
        k := dto.Keyword{
            Category:        []int32{categoryId},
            Keyword:         rowKeyword,
            Title:           title,
            MetaDescription: metaDescription,
        }
        tagName := strings.TrimSpace(rows[1])
        if tagName == "" {
            log.Debug("tag name is empty", zap.String("row", row))
            if rowKeyword != "" {
                tagName = strings.ReplaceAll(rowKeyword, " ", "")
            } else {
                continue
            }
        }
        existTag, err := g.TagService.GetByNameLike(ctx, fmt.Sprintf("%%%s%%", tagName))
        if err != nil {
            if existTag == nil || errors.Is(err, gorm.ErrRecordNotFound) {
                paramTag := param.Tag{
                    Name:  tagName,
                    Color: "#cfd3d7",
                }
                py := pinyin.LazyConvert(tagName, nil)
                //slug := strings.ReplaceAll(py[0], " ", "-")
                slug := strings.Join(py, "-")
                paramTag.Slug = util.Slug(slug)
                existTag, err = g.TagService.Create(ctx, &paramTag)
                if err == nil {
                    k.TagIds = []int32{existTag.ID}
                }
            }
        } else {
            k.TagIds = []int32{existTag.ID}
        }

        keywords = append(keywords, k)
    }

    return keywords, nil
}

func (g *gptServiceImpl) GenContent(ctx context.Context, keywords []dto.Keyword) error {
    //systemPrompt := "I Want You To Act As A Content Writer Very Proficient SEO Writer Writes Fluently Chinese."
    userPrompt := "I Want You To Act As A Content Writer Very Proficient SEO Writer Writes Fluently Chinese. Start with an introduction paragraph. Write a 3000-word 100% Unique, SEO-optimized, Human-Written article in Chinese with at least 15 headings and subheadings that covers the topic provided in the Prompt. Write The article In Your Own Words Rather Than Copying And Pasting From Other Sources. Consider perplexity and burstiness when creating content, ensuring high levels of both without losing specificity or context. Use fully detailed paragraphs that engage the reader. Write In A Conversational Style As Written By A Human (Use An Informal Tone, Utilize Personal Pronouns, Keep It Simple, Engage The Reader, Use The Active Voice, Keep It Brief, Use Rhetorical Questions, and Incorporate Analogies And Metaphors). End with a conclusion paragraph. Using Markdown syntax to write the article. Now Write An Article On This Topic '%s' in Chinese."
    //userPrompt := "我希望您以一位内容作家和精通SEO写手的身份撰写一篇文章，流利地用中文写作。文章应以介绍段落开始，并在中文中写一篇3000字、100%独特、经过SEO优化、人工撰写的文章，其中包含至少15个标题和副标题，涵盖了提示中提供的主题。请用自己的话语写文章，而不是从其他来源复制和粘贴。在创作内容时考虑到复杂性和突发性，确保二者在高水平的同时不失去具体性和语境。使用充分详细的段落吸引读者。以一种对话的方式写作，就像人类写的一样（使用非正式的语气，使用个人代词，保持简洁，引起读者兴趣，使用主动语态，简洁明了，使用修辞性问题，引入类比和隐喻）。文章以结论段落结束。请使用Markdown语法撰写文章。现在请以“%s”为主题写一篇文章"
    for _, data := range keywords {
        var request api.Request
        //request.Temperature = 0.7
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
        resp, err := g.client.Completion(ctx, &request)
        if err != nil {
            log.Error("gpt gen content err", zap.Error(err))
            continue
        }
        //log.Debug("resp:", zap.String("content", resp.Choices[0].Message.Content))
        postParam := param.Post{
            Title:           data.Title,
            Status:          consts.PostStatusPublished,
            Slug:            fmt.Sprintf("article-%s", time.Now().Local().Format("20060102030405")),
            OriginalContent: resp.Choices[0].Message.Content,
            Content:         util.Bytes2str(blackfriday.Run(util.Str2bytes(resp.Choices[0].Message.Content))),
            CategoryIDs:     data.Category,
            TagIDs:          data.TagIds,
            MetaKeywords:    data.Keyword,
            MetaDescription: data.MetaDescription,
        }
        _, err = g.PostService.Create(ctx, &postParam)
        if err != nil {
            log.Error("create content err", zap.Error(err))
        }
    }

    return nil
}
