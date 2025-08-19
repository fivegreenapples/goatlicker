package model

type Payment struct {
	PersonId int    `json:"person_id"`
	Name     string `json:"name"`
	Amount   int    `json:"amount"`
}
