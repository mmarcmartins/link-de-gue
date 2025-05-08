package main

import (
	"context"
	"feaven/url-shortener/utils"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"time"

	"golang.org/x/time/rate"

	_ "github.com/joho/godotenv/autoload"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.mongodb.org/mongo-driver/v2/mongo/readpref"
)

var rateLimiter = make(map[string]*rate.Sometimes)
var dbName = "link-de-gue"
var collName = "links"
var client *mongo.Client

type ShortenBody struct {
	OriginalUrl string
}

type Link struct {
	_id          primitive.ObjectID `bson:"_id,omitempty" json:"_id,omitempty"`
	OriginalLink string             `bson:"OriginalLink" json:"originalLink"`
	ShortenLink  string             `bson:"ShortenLink" json:"shortenLink"`
}

type ShortenResponse struct {
	Success      bool   `json:"success"`
	OriginalURL  string `json:"originalUrl"`
	ShortenedURL string `json:"shortenedUrl"`
	Error        string `json:"error,omitempty"`
}

func getShortLink(collection *mongo.Collection, shortLink string) (*Link, error) {
	var result *Link = &Link{}
	err := collection.FindOne(context.TODO(), bson.D{{Key: "ShortenLink", Value: shortLink}}).Decode(result)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, mongo.ErrNoDocuments
		}
		return nil, err
	}

	return result, nil
}

func checkDuplicateLink(collection *mongo.Collection, originalUrl string) (*Link, error) {
	var result *Link = &Link{}
	err := collection.FindOne(context.TODO(), bson.D{{Key: "OriginalLink", Value: originalUrl}}).Decode(result)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, mongo.ErrNoDocuments
		}
		return nil, err
	}

	return result, nil
}

func generateShortLink() string {
	newId := utils.ShortURLToID(primitive.NewObjectID().Hex())
	return utils.IdToShortURL(newId)
}

func connectDB() *mongo.Client {
	fmt.Println("Getting envs to connect to MongoDB")

	username := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	cluster := os.Getenv("DB_CLUSTER")

	fmt.Println("Connecting to MongoDB")

	serverAPI := options.ServerAPI(options.ServerAPIVersion1)
	opts := options.Client().ApplyURI("mongodb+srv://" +
		url.QueryEscape(username) + ":" +
		url.QueryEscape(password) + "@" +
		url.QueryEscape(cluster) + "/?retryWrites=true&w=majority&appName=shorten-url-go").SetServerAPIOptions(serverAPI)

	client, err := mongo.Connect(opts)

	if err != nil {
		panic(err)
	}

	if err := client.Ping(context.TODO(), readpref.Primary()); err != nil {
		panic(err)
	}

	fmt.Println("You successfully connected to MongoDB!")
	return client
}

func limiting(ip string, handler func()) bool {
	if rateLimiter[ip] == nil {
		rateLimiter[ip] = &rate.Sometimes{
			First:    2,
			Interval: 1 * time.Second,
		}
	}

	var rateLimited = true

	rateLimiter[ip].Do(func() {
		rateLimited = false
		handler()
	})

	return rateLimited
}
func init() {
	rateLimiter = make(map[string]*rate.Sometimes)
	client = connectDB()
}

func main() {
	port := os.Getenv("PORT")
	fmt.Println("Starting url shortener service")

	http.HandleFunc("GET /health", handleHealth)

	http.HandleFunc("GET /{id}", handleGetOriginalLink)

	http.HandleFunc("POST /shorten", handleShorten)

	fmt.Println("Starting server on port", port)
	http.ListenAndServe(":"+port, nil)
}
