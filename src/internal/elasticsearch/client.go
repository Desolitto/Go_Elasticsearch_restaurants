package client

import (
	types "elasticsearch_recommender/internal/models"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/elastic/go-elasticsearch/v8"
)

func NewElasticSearchClient() (*elasticsearch.Client, error) {
	ctg := elasticsearch.Config{Addresses: []string{"http://localhost:9200"}}
	es, err := elasticsearch.NewClient(ctg)
	if err != nil {
		return nil, err
	}
	res, err := es.Info()
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	log.Printf("Elasticsearch info: %s", res.String())

	return es, nil
}

// Set max_result_window to allow retrieval of up to 30,000 documents per query,
// exceeding Elasticsearch's default limit of 10,000. This enables efficient
// pagination and retrieval of large datasets without hitting the default cap.
func CreateIndex(es *elasticsearch.Client) error {
	indexBody := `{
		"settings": {
			"index": {
				"max_result_window": 20000
			}
		},
		"mappings": {
			"properties": {
				"id": {
					"type": "long"
				},
				"name": {
					"type":  "text"
				},
				"address": {
					"type":  "text"
				},
				"phone": {
					"type":  "text"
				},
				"location": {
					"type": "geo_point"
				}
			}
		}
	}`
	res, err := es.Indices.Delete([]string{"places"}, es.Indices.Delete.WithIgnoreUnavailable(true))
	if err != nil {
		return fmt.Errorf("error deleting indices: %s", err)
	}
	defer res.Body.Close()
	res, err = es.Indices.Create("places", es.Indices.Create.WithBody(strings.NewReader(indexBody)))
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.IsError() {
		return fmt.Errorf("error creating index: %s", res.String())
	}
	log.Println("Index 'places' created successfully.")
	return nil
}

func BulkInsert(es *elasticsearch.Client, places []types.Place) error {
	var request strings.Builder
	for i, place := range places {
		// Формируем метаданные
		meta := fmt.Sprintf(`{"index": {"_index": "places", "_id": "%d"}}`, i+1)
		request.WriteString(meta + "\n") // добавляем новую строку после метаданных

		// Сериализуем данные
		data, err := json.Marshal(place)
		if err != nil {
			return fmt.Errorf("error marshaling data: %s", err)
		}
		request.WriteString(string(data) + "\n") // добавляем новую строку после данных
	}

	res, err := es.Bulk(strings.NewReader(request.String()))
	if err != nil {
		return fmt.Errorf("error executing bulk request: %s", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		var errorResponse map[string]interface{}
		if err := json.NewDecoder(res.Body).Decode(&errorResponse); err != nil {
			return fmt.Errorf("error parsing the response body: %s", err)
		}
		log.Printf("Error response: %v", errorResponse)
		return fmt.Errorf("error executing bulk request: %v", errorResponse)
	}

	log.Println("Data inserted successfully.")
	defer res.Body.Close()
	return nil
}

/*
	// mapping := `{
	// 	"mappings": {
	// 		"properties": {
	// 			"id": {
	// 				"type": "long"
	// 			},
	// 			"name": {
	// 				"type":  "text"
	// 			},
	// 			"address": {
	// 				"type":  "text"
	// 			},
	// 			"phone": {
	// 				"type":  "text"
	// 			},
	// 			"location": {
	// 				"type": "geo_point"
	// 			}
	// 		}
	// 	}
	// }`

*/
