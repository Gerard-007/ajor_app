package repository

import (
    "context"
    "github.com/Gerard-007/ajor_app/internal/models"
    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/bson/primitive"
)

func CreateUser(db *mongo.Collection, user *models.User) error {
    _, err := db.InsertOne(context.TODO(), user)
    return err
}

func GetUserByID(db *mongo.Collection, id primitive.ObjectID) (*models.User, error) {
    var user models.User
    err := db.FindOne(context.TODO(), bson.M{"_id": id}).Decode(&user)
    if err != nil {
        return nil, err
    }
    return &user, nil
}

func GetUserByEmail(db *mongo.Collection, email string) (*models.User, error) {
    var user models.User
    err := db.FindOne(context.TODO(), bson.M{"email": email}).Decode(&user)
    if err != nil {
        return nil, err
    }
    return &user, nil
}