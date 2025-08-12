package model

type WeightedTag struct {
	Name   string `json:"name" db:"name"`
	Weight int    `json:"weight" db:"weight"`
}
