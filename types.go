package main

type Product struct {
	Name       string `bson:"name" json:"name"`
	Desc       string `bson:"desc" json:"desc"`
	Img        string `bson:"img" json:"img"`
	ImgContent []byte `bson:"imgContent" json:"imgContent"`
}
