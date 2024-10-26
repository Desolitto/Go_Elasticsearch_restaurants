package main

import (
	"elasticsearch_recommender/internal/db"
	client "elasticsearch_recommender/internal/elasticsearch"
	"elasticsearch_recommender/internal/handlers"
	"flag"
	"log"
	"net/http"
)

func main() {
	// Define the -t flag
	tokenFlag := flag.Bool("t", false, "Use token handlers")
	flag.Parse()

	// Create Elasticsearch client
	esClient, err := client.NewElasticSearchClient()
	if err != nil {
		log.Fatalf("error creating Elasticsearch client: %s", err)
	}

	// Create index and mapping
	err = client.CreateIndex(esClient)
	if err != nil {
		log.Fatalf("error creating index: %s", err)
	}

	// Load data from CSV file
	places, err := db.LoadPlaces("internal/data/data.csv")
	if err != nil {
		log.Fatalf("error loading data from file: %s\n", err)
	}
	log.Printf("Loaded %d places from CSV", len(places))

	err = client.BulkInsert(esClient, places)
	if err != nil {
		log.Fatalf("error inserting places: %s\n", err)
	}

	// Create store
	store := db.NewElasticsearchStore(esClient)

	// Create handlers
	handler := handlers.NewHandler(store)

	// Register routes
	http.HandleFunc("/", handler.HtmlHandler)
	http.HandleFunc("/api/places", handler.JsonHandler)
	http.HandleFunc("/api/get_token", handlers.TokenHandler)

	if *tokenFlag {
		http.Handle("/api/recommend", handlers.JWTMiddleware(http.HandlerFunc(handler.RecommendHandler)))
	} else {
		http.HandleFunc("/api/recommend", handler.RecommendHandler)
	}

	log.Println("Server started at http://localhost:8888")

	if err = http.ListenAndServe(":8888", nil); err != nil {
		log.Fatalf("Error starting server: %s", err)
	}
}

/*
// http.Handle("/api/recommend", handlers.CORSHandler(handlers.JWTMiddleware(http.HandlerFunc(handler.RecommendHandler))))
http://127.0.0.1:8888/api/get_token
eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJhZG1pbiI6dHJ1ZSwiZXhwIjoxNzI4OTIxMDk1LCJuYW1lIjoiTmlrb2xhIn0.nWNX4Vv0bmhmp6h5cIeyT2WxZZvjwVa1aXCJFPPYJYA
➜  ~ curl -X GET "http://localhost:8888/api/recommend?lat=55.674&lon=37.666" -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJhZG1pbiI6dHJ1ZSwiZXhwIjoxNzI4OTIxMDk1LCJuYW1lIjoiTmlrb2xhIn0.nWNX4Vv0bmhmp6h5cIeyT2WxZZvjwVa1aXCJFPPYJYA"
curl -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJhZG1pbiI6dHJ1ZSwiZXhwIjoxNzI4OTMxMTc1LCJuYW1lIjoiTmlrb2xhIn0.0SwRtDfvNohWVpY1R5sqOALiS2xWzFgMw3PdIQZtDaM" "http://localhost:8888/api/recommend?lat=55.674&lon=37.666" | jq .
-H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJhZG1pbiI6dHJ1ZSwiZXhwIjoxNzI4OTMxMTc1LCJuYW1lIjoiTmlrb2xhIn0.0SwRtDfvNohWVpY1R5sqOALiS2xWzFgMw3PdIQZtDaM" http://localhost:8888/api/recommend?lat=55.674&lon=37.666
*/
