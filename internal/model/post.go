package model

type Post struct {
	ID      string `json:"id"`                       // 記事ID
	Title   string `json:"title" binding:"required"` // タイトル
	Content string `json:"content"`                  // 本文
	Author  string `json:"author"`                   // 執筆者名
}
