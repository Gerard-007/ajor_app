package repository

import (
	"context"
	"errors"
	"time"

	"github.com/Gerard-007/ajor_app/internal/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func GetAllUsers(db *mongo.Collection) ([]*models.User, error) {
	var users []*models.User
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	opts := options.Find().SetProjection(bson.M{"password": 0})
	cursor, err := db.Find(ctx, bson.M{}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		var user models.User
		if err := cursor.Decode(&user); err != nil {
			return nil, err
		}
		users = append(users, &user)
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}

	return users, nil
}

func CreateUser(db *mongo.Collection, user *models.User) error {
	_, err := db.InsertOne(context.TODO(), user)
	return err
}

type UserUpdate struct {
	Email    string             `bson:"email,omitempty"`
	Username string             `bson:"username,omitempty"`
	Phone    string             `bson:"phone,omitempty"`
	Verified bool               `bson:"verified,omitempty"`
	IsAdmin  bool               `bson:"is_admin,omitempty"`
	WalletID primitive.ObjectID `bson:"wallet_id,omitempty"`
}

func UpdateUser(db *mongo.Database, id primitive.ObjectID, userUpdate *UserUpdate) (*models.User, error) {
	usersCollection := db.Collection("users")

	// Check for duplicate email or username
	if userUpdate.Email != "" {
		var existing models.User
		err := usersCollection.FindOne(context.TODO(), bson.M{"email": userUpdate.Email, "_id": bson.M{"$ne": id}}).Decode(&existing)
		if err == nil {
			return nil, errors.New("email already exists")
		}
		if err != mongo.ErrNoDocuments {
			return nil, err
		}
	}
	if userUpdate.Username != "" {
		var existing models.User
		err := usersCollection.FindOne(context.TODO(), bson.M{"username": userUpdate.Username, "_id": bson.M{"$ne": id}}).Decode(&existing)
		if err == nil {
			return nil, errors.New("username already exists")
		}
		if err != mongo.ErrNoDocuments {
			return nil, err
		}
	}

	update := bson.M{
		"$set": bson.M{
			"email":      userUpdate.Email,
			"username":   userUpdate.Username,
			"phone":      userUpdate.Phone,
			"verified":   userUpdate.Verified,
			"is_admin":   userUpdate.IsAdmin,
			"wallet_id":  userUpdate.WalletID,
			"updated_at": time.Now(),
		},
	}

	var updatedUser models.User
	err := usersCollection.FindOneAndUpdate(
		context.TODO(),
		bson.M{"_id": id},
		update,
		options.FindOneAndUpdate().SetReturnDocument(options.After),
	).Decode(&updatedUser)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, errors.New("user not found")
		}
		return nil, err
	}

	return &updatedUser, nil
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