package domain

import "time"

type Category string

const (
	CategoryIdea    Category = "idea"
	CategoryDate    Category = "date"
	CategoryGift    Category = "gift"
	CategoryMovie   Category = "movie"
	CategoryTravel  Category = "travel"
	CategoryThought Category = "thought"
	CategoryOther   Category = "other"
)

func (c Category) Valid() bool {
	switch c {
	case CategoryIdea, CategoryDate, CategoryGift, CategoryMovie,
		CategoryTravel, CategoryThought, CategoryOther:
		return true
	default:
		return false
	}
}

type Reaction struct {
	MemberID string `bson:"memberId" json:"memberId"`
	Emoji    string `bson:"emoji"    json:"emoji"`
}

type Note struct {
	ID        string     `bson:"_id"                 json:"id"`
	RoomID    string     `bson:"roomId"              json:"roomId"`
	AuthorID  string     `bson:"authorId"            json:"authorId"`
	Title     string     `bson:"title,omitempty"     json:"title,omitempty"`
	Content   string     `bson:"content"             json:"content"`
	Category  Category   `bson:"category"            json:"category"`
	Color     string     `bson:"color,omitempty"     json:"color,omitempty"`
	Pinned    bool       `bson:"pinned"              json:"pinned"`
	Reactions []Reaction `bson:"reactions"           json:"reactions"`
	CreatedAt time.Time  `bson:"createdAt"           json:"createdAt"`
	UpdatedAt time.Time  `bson:"updatedAt"           json:"updatedAt"`
}

type NoteCreate struct {
	Title    string
	Content  string
	Category Category
	Color    string
	Pinned   bool
}

type NoteUpdate struct {
	Title    *string
	Content  *string
	Category *Category
	Color    *string
	Pinned   *bool
}

type NoteFilter struct {
	RoomID   string
	Category *Category
	Limit    int64
	Before   *time.Time
}
