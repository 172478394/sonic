package dto

type Keyword struct {
    Category        []int32 `json:"category"`
    TagIds          []int32 `json:"tagIds"`
    Keyword         string  `json:"keyword"`
    Title           string  `json:"title"`
    MetaDescription string  `json:"metaDescription"`
}
