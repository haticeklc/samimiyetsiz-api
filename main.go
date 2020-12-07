package main

import (
	"context"
	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"os"
	"time"
)

func main() {
	db, err := Connect()

	if err != nil {
		log.Print(err.Error())
		os.Exit(0)
		return
	}

	userCollection := db.Collection("users")

	app := fiber.New()

	app.Post("/users", createUser(userCollection))
	app.Get("/users", readUsers(userCollection))
	app.Get("/users/:id", readUser(userCollection))
	app.Get("/users/un/:username", readUserByUsername(userCollection))
	app.Put("/users", updateUser(userCollection))
	app.Delete("/users/:id", deleteUser(userCollection))
	app.Listen(":8000")
}

func readUserByUsername(uc *mongo.Collection) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		username := ctx.Params("username")
		var user User

		err := uc.FindOne(context.Background(), bson.M{"username": username}).Decode(&user)
		if err != nil {
			return err
		}
		return ctx.JSON(user)

	}

}

func deleteUser(uc *mongo.Collection) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		id := ctx.Params("id")
		ID, err := primitive.ObjectIDFromHex(id)
		if err != nil {
			return err
		}
		one, err := uc.DeleteOne(context.Background(), bson.M{"_id": ID})
		if err != nil {
			return err
		}
		return ctx.JSON(fiber.Map{"status": one.DeletedCount})

	}
}

func updateUser(uc *mongo.Collection) fiber.Handler {

	return func(ctx *fiber.Ctx) error {
		var user User
		err := ctx.BodyParser(&user)
		if err != nil {
			return err
		}

		_, err = uc.UpdateOne(context.Background(), bson.M{"_id": user.ID}, bson.M{"$set": user})
		if err != nil {
			return err
		}
		return ctx.JSON(user)
	}
}

func readUser(uc *mongo.Collection) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		id := ctx.Params("id")
		ID, err := primitive.ObjectIDFromHex(id)
		if err != nil {
			return err
		}
		var user User
		err = uc.FindOne(context.Background(), bson.M{"_id": ID}).Decode(&user)
		if err != nil {
			return err
		}

		return ctx.JSON(user)

	}

}

func readUsers(uc *mongo.Collection) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		var users []User

		cursor, err := uc.Find(context.Background(), bson.D{})

		if err != nil {
			return err
		}

		for cursor.Next(context.TODO()) {
			var user User
			_ = cursor.Decode(&user)
			users = append(users, user)
		}
		return ctx.JSON(users)
	}
}

func createUser(uc *mongo.Collection) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		user := new(User)
		err := ctx.BodyParser(user)
		if err != nil {
			log.Print(err.Error())
			return err
		}
		user.ID = primitive.NewObjectID()
		_, err = uc.InsertOne(context.Background(), user)
		if err != nil {
			log.Print(err.Error())
			return err
		}
		return ctx.JSON(user)
	}

}

type User struct {
	ID          primitive.ObjectID `bson:"_id" json:"id"`
	Username    string             `bson:"username" json:"username"`
	Password    string             `bson:"password" json:"password"`
	CreatedDate time.Time          `json:"createdDate"`
}

func Connect() (*mongo.Database, error) {
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)

	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		return nil, err
	}
	err = client.Ping(context.Background(), nil)
	if err != nil {
		return nil, err
	}
	db := client.Database("example")

	return db, nil
}
