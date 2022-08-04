package cacheTypes

import "go.mongodb.org/mongo-driver/bson/primitive"

type UserSession struct {
	AccessLevel int                `bson:"accessLevel"`
	UserID      primitive.ObjectID `bson:"userID"`
}
