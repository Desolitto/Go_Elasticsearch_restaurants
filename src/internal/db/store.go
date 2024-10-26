package db

import (
	"context"
	types "elasticsearch_recommender/internal/models"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"
)

type Store interface {
	// Returns a list of elements, the total number of results, and/or an error if one occurs
	GetPlaces(limit int, offset int) ([]types.Place, int, error)
}

type ElasticsearchStore struct {
	client *elasticsearch.Client
}

func NewElasticsearchStore(client *elasticsearch.Client) *ElasticsearchStore {
	return &ElasticsearchStore{client: client}
}

func (es *ElasticsearchStore) GetPlaces(limit int, offset int) ([]types.Place, int, error) {
	query := fmt.Sprintf(`{
		"query": {
			"match_all": {}
		},
		"from": %d,
		"size": %d,
		"sort": [{ "id": "asc" }]
	}`, offset, limit)
	// Execute the request to Elasticsearch
	res, err := es.client.Search(
		es.client.Search.WithContext(context.Background()),
		es.client.Search.WithIndex("places"),
		es.client.Search.WithBody(strings.NewReader(query)),
		es.client.Search.WithTrackTotalHits(true),
	)
	if err != nil {
		return nil, 0, fmt.Errorf("error executing request to Elasticsearch: %s", err)
	}
	defer res.Body.Close()
	if res.IsError() {
		return nil, 0, fmt.Errorf("error in response from Elasticsearch: %s", res.String())
	}

	// Decode the response from Elasticsearch
	var response map[string]interface{}
	decoder := json.NewDecoder(res.Body)
	err = decoder.Decode(&response)
	if err != nil {
		return nil, 0, fmt.Errorf("error decoding response from Elasticsearch: %s", err)
	}

	// Extract "hits" and check if it's a map
	hits, ok := response["hits"].(map[string]interface{})
	if !ok {
		return nil, 0, fmt.Errorf("error: 'hits' is not of type map[string]interface{}")
	}

	// Extract "total" from hits and check if it's a map
	total, ok := hits["total"].(map[string]interface{})
	if !ok {
		return nil, 0, fmt.Errorf("error: 'total' is not map[string]interface{}")
	}

	value, ok := total["value"].(float64)
	if !ok {
		return nil, 0, fmt.Errorf("error: value of 'value' is not float64")
	}

	totalHits := int(value)

	// Get the array of documents from "hits"
	// places := make([]types.Place, len(response.Hits.Hits))
	// for i, hit := range response.Hits.Hits {
	// 	places[i] = hit.Source
	// }
	hitsArray, ok := hits["hits"].([]interface{})
	if !ok {
		return nil, 0, fmt.Errorf("error: 'hits' is not []interface{}")
	}

	var places []types.Place
	for _, hit := range hitsArray {
		var place types.Place
		// Extract "_source" from the current document
		source := hit.(map[string]interface{})["_source"]
		// Serialize the "_source" map into JSON bytes
		sourceBytes, err := json.Marshal(source)
		if err != nil {
			log.Println("Error serializing '_source':", err)
			continue
		}
		if err := json.Unmarshal(sourceBytes, &place); err != nil {
			continue
		}
		places = append(places, place)
	}
	return places, totalHits, nil
}

func GetClosestRestaurants(lat, lon float64, limit int) ([]types.Place, error) {
	es, err := elasticsearch.NewDefaultClient()
	if err != nil {
		return nil, fmt.Errorf("error creating Elasticsearch client: %s", err)
	}
	query := fmt.Sprintf(`{
		"query": {
			"match_all": {}
		},
		"sort": [
			{
				"_geo_distance": {
					"location": {
						"lat": %f,
						"lon": %f
					},
					"order": "asc",
					"unit": "km"
				}
			}
		],
		"size": %d
	}`, lat, lon, limit)
	req := esapi.SearchRequest{
		Index: []string{"places"},
		Body:  strings.NewReader(query),
	}
	res, err := req.Do(context.Background(), es)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, fmt.Errorf("error: %s", res.String())
	}

	var response struct {
		Hits struct {
			Hits []struct {
				Source types.Place `json:"_source"`
			} `json:"hits"`
		} `json:"hits"`
	}

	if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
		return nil, err
	}

	places := make([]types.Place, len(response.Hits.Hits))
	for i, hit := range response.Hits.Hits {
		places[i] = hit.Source
	}

	return places, nil
}

/*
http://127.0.0.1:8888/api/recommend?lat=37.666&lon=55.674
*/
