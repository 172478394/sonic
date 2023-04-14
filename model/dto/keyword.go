package dto

type Keyword struct {
    Data []struct {
        Category        []int32 `json:"category"`
        Keyword         string  `json:"keyword"`
        Title           string  `json:"title"`
        MetaDescription string  `json:"metaDescription"`
    } `json:"data"`
}
