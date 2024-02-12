package main

type Product struct {
	Name string `json:"name"`
	Desc string `json:"desc"`
	Img  string `json:"img" bson:"img"`
}
