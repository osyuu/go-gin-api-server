package model

import (
	"encoding/base64"
	"encoding/json"
	"time"
)

type Cursor struct {
	ID        string    `json:"id"`
	CreatedAt time.Time `json:"created_at"`
}

type CursorRequest struct {
	Cursor   string  `json:"cursor" form:"cursor"`
	Limit    int     `json:"limit" form:"limit" binding:"min=1,max=100"`
	AuthorID *string `json:"author_id,omitempty" form:"author_id"`
}

type CursorResponse[T any] struct {
	Data    []T    `json:"data"`
	Next    string `json:"next_cursor,omitempty"`
	HasMore bool   `json:"has_more"`
}

// set defaults
func (c *CursorRequest) SetDefaults() {
	if c.Limit <= 0 {
		c.Limit = 10
	}
	if c.Limit > 100 {
		c.Limit = 100
	}
}

func EncodeCursor(cursor Cursor) string {
	data, err := json.Marshal(cursor)
	if err != nil {
		return ""
	}
	return base64.StdEncoding.EncodeToString(data)
}

func DecodeCursor(cursorStr string) (Cursor, error) {
	data, err := base64.StdEncoding.DecodeString(cursorStr)
	if err != nil {
		return Cursor{}, err
	}

	var cursor Cursor
	err = json.Unmarshal(data, &cursor)
	return cursor, err
}
