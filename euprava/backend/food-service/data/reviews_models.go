package data

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Rating struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	FoodID    primitive.ObjectID `bson:"foodId" json:"foodId"`
	UserID    primitive.ObjectID `bson:"userId" json:"userId"`
	Rating    int                `bson:"rating" json:"rating"` // 1..5
	CreatedAt time.Time          `bson:"createdAt" json:"createdAt"`
	UpdatedAt time.Time          `bson:"updatedAt" json:"updatedAt"`
}

type Comment struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	FoodID    primitive.ObjectID `bson:"foodId" json:"foodId"`
	UserID    primitive.ObjectID `bson:"userId" json:"userId"`
	Author    string             `bson:"author" json:"author"` // first_name + last_name
	Text      string             `bson:"text" json:"text"`
	CreatedAt time.Time          `bson:"createdAt" json:"createdAt"`
}

type ReviewSummary struct {
	FoodID       primitive.ObjectID `json:"foodId"`
	AvgRating    float64            `json:"avgRating"`
	RatingCount  int64              `json:"ratingCount"`
	CommentCount int64              `json:"commentCount"`
	CanReview    bool               `json:"canReview"`
	MyRating     int                `json:"myRating"`
}

type SetRatingRequest struct {
	Rating int `json:"rating"`
}

type AddCommentRequest struct {
	Text string `json:"text"`
}

type BatchSummaryRequest struct {
	FoodIds []string `json:"foodIds"`
}
