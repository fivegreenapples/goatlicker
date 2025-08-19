package model

type Transaction struct {
	Id          int    `json:"id"`
	Description string `json:"description"`
	Date        int    `json:"date"`
	TotalAmount int    `json:"totalamount"`
}
