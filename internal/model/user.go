package model

type User struct {
	ID   string `json:"id"`
	Name string `json:"name" binding:"required,custom_validator"`
	Age  int    `json:"age" binding:"gte=0,lte=120"`
}
