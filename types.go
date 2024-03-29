package main

import "go.mongodb.org/mongo-driver/bson/primitive"

type Product struct {
	ID         primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	Name       string             `bson:"name" json:"name"`
	Desc       string             `bson:"desc" json:"desc"`
	Img        string             `bson:"img" json:"img"`
	ImgContent []byte             `bson:"imgContent" json:"imgContent"`
}
