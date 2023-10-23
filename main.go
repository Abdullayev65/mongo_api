package main

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"strconv"
)

func main() {
	ctx := context.Background()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(
		"mongodb://127.0.1:27017"))
	if err != nil {
		panic(err)
	}
	defer func(client *mongo.Client, ctx context.Context) {
		err := client.Disconnect(ctx)
		if err != nil {
			fmt.Println(err)
		}
	}(client, ctx)

	e := gin.Default()
	posts := client.Database("index").Collection("posts")

	e.POST("/create", func(c *gin.Context) {
		data := new(Post)
		err := c.Bind(data)
		if err != nil {
			c.String(400, err.Error())
			return
		}
		data.Id = primitive.NewObjectID().Hex()
		_, err = posts.InsertOne(c.Request.Context(), data)
		if err != nil {
			c.String(400, err.Error())
			return
		}

		c.JSON(200, data)
	})
	e.GET("/get", func(c *gin.Context) {
		data := make([]*Post, 0)
		limit, _ := strconv.Atoi(c.Query("limit"))
		lmt := int64(limit)
		cur, err := posts.Find(c.Request.Context(), bson.D{}, &options.FindOptions{Limit: &lmt})
		if err != nil {
			c.String(400, err.Error())
			return
		}

		err = cur.All(c.Request.Context(), &data)
		if err != nil {
			c.String(400, err.Error())
			return
		}

		c.JSON(200, data)
	})
	e.GET("/get/:id", func(c *gin.Context) {
		data := new(Post)
		data.Id = c.Param("id")
		id, err := primitive.ObjectIDFromHex(data.Id)
		if err != nil {
			c.String(400, err.Error())
			return
		}

		err = posts.FindOne(c.Request.Context(), bson.D{{"_id", id}}).Decode(data)
		if err != nil {
			c.String(400, err.Error())
			return
		}

		c.JSON(200, data)
	})
	e.PUT("/update/:id", func(c *gin.Context) {
		data := new(Post)
		err := c.Bind(data)
		if err != nil {
			c.String(400, err.Error())
			return
		}

		id, err := primitive.ObjectIDFromHex(c.Param("id"))
		if err != nil {
			panic(err)
		}

		_, err = posts.ReplaceOne(c.Request.Context(),
			bson.D{{"_id", id}}, data)
		if err != nil {
			c.String(400, err.Error())
			return
		}

		c.JSON(200, data)
	})
	e.DELETE("/delete/:id", func(c *gin.Context) {
		_, err := posts.DeleteOne(c.Request.Context(), bson.D{{"_id", c.Param("id")}})
		if err != nil {
			c.String(400, err.Error())
			return
		}

		c.JSON(200, "deleted")
	})

	err = e.Run(":90")
	if err != nil {
		fmt.Println(err)
	}
}

type Post struct {
	Id          string `json:"id" bson:"_id,omitempty"`
	Title       string `json:"title" bson:"title"`
	Description string `json:"description" bson:"description"`
}
