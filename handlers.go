package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"go.mongodb.org/mongo-driver/v2/mongo"
)

func handleHealth(writer http.ResponseWriter, req *http.Request) {
	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(http.StatusOK)
	writer.Write([]byte(`{"status":"healthy","message":"Service is running properly"}`))
}

func handleGetOriginalLink(writer http.ResponseWriter, req *http.Request) {
	ip := req.RemoteAddr
	rateLimited := limiting(ip, func() {

		coll := client.Database(dbName).Collection(collName)
		id := req.PathValue("id")
		redirectUrl, err := getShortLink(coll, id)

		if err != nil && err != mongo.ErrNoDocuments {
			fmt.Println("Redirect link not found")
			http.NotFound(writer, req)
			return
		}

		if redirectUrl != nil {
			fmt.Println("Redirecting to link:", redirectUrl)
			http.Redirect(writer, req, redirectUrl.OriginalLink, http.StatusFound)
			return
		}
	})

	if rateLimited {
		fmt.Println("Rate limite exceeded")
		http.Error(writer, "Rate limit exceeded", http.StatusTooManyRequests)
		return
	}
}

func handleShorten(writer http.ResponseWriter, req *http.Request) {
	ip := req.RemoteAddr
	rateLimited := limiting(ip, func() {
		writer.Header().Set("Content-Type", "application/json")
		var body ShortenBody
		err := json.NewDecoder(req.Body).Decode(&body)

		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}

		parsedUrl, err := url.Parse(body.OriginalUrl)

		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}

		if parsedUrl.Scheme == "" {
			parsedUrl.Scheme = "http"
		}

		url := parsedUrl.String()

		coll := client.Database(dbName).Collection(collName)
		duplicateLink, err := checkDuplicateLink(coll, url)

		if err != nil && err != mongo.ErrNoDocuments {
			response := ShortenResponse{
				Success: false,
				Error:   "Database error: " + err.Error(),
			}
			writer.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(writer).Encode(response)
			return
		}

		if duplicateLink != nil {
			fmt.Println("Duplicated link", body.OriginalUrl)
			response := ShortenResponse{
				Success:      true,
				OriginalURL:  duplicateLink.OriginalLink,
				ShortenedURL: duplicateLink.ShortenLink,
			}
			writer.WriteHeader(http.StatusOK)
			json.NewEncoder(writer).Encode(response)
			return
		}

		shortenLink := generateShortLink()
		doc := Link{OriginalLink: url, ShortenLink: shortenLink}
		result, err := coll.InsertOne(context.TODO(), doc)

		if err != nil {
			response := ShortenResponse{
				Success: false,
				Error:   "Failed to create shortened URL: " + err.Error(),
			}
			writer.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(writer).Encode(response)
			return
		}

		fmt.Printf("Success created shorten link %v\n", result.InsertedID)

		response := ShortenResponse{
			Success:      true,
			OriginalURL:  url,
			ShortenedURL: shortenLink,
		}

		writer.WriteHeader(http.StatusCreated)
		json.NewEncoder(writer).Encode(response)
	})

	if rateLimited {
		http.Error(writer, "Rate limit exceeded", http.StatusTooManyRequests)
		return
	}
}
