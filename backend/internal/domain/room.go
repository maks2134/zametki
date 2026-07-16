package domain

import "time"

type Member struct {
	ID       string    `bson:"id"       json:"id"`
	Name     string    `bson:"name"     json:"name"`
	Color    string    `bson:"color"    json:"color"`
	JoinedAt time.Time `bson:"joinedAt" json:"joinedAt"`
}

type Room struct {
	ID        string    `bson:"_id"       json:"id"`
	Code      string    `bson:"code"      json:"code"`
	Title     string    `bson:"title"     json:"title"`
	CreatedAt time.Time `bson:"createdAt" json:"createdAt"`
	Members   []Member  `bson:"members"   json:"members"`
}
