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

func DeleteUserAndProfile(db *mongo.Database, userID primitive.ObjectID) error {
	session, err := db.Client().StartSession()
	if err != nil {
		return err
	}
	defer session.EndSession(context.TODO())

	err = mongo.WithSession(context.TODO(), session, func(sc mongo.SessionContext) error {
		// Delete user
		usersCollection := db.Collection("users")
		_, err := usersCollection.DeleteOne(sc, bson.M{"_id": userID})
		if err != nil {
			return err
		}

		// Delete profile
		profilesCollection := db.Collection("profiles")
		_, err = profilesCollection.DeleteOne(sc, bson.M{"user_id": userID})
		if err != nil {
			return err
		}

		return session.CommitTransaction(sc)
	})

	if err != nil {
		session.AbortTransaction(context.TODO())
		return err
	}
	return nil
}