package handlers

import (
	"elasticsearch_recommender/internal/db"
	types "elasticsearch_recommender/internal/models"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/dgrijalva/jwt-go"
)

type Handler struct {
	store db.Store
}

func NewHandler(store db.Store) *Handler {
	return &Handler{store: store}
}

func handleError(w http.ResponseWriter, message string, status int) {
	http.Error(w, message, status)
}

func getPageParam(r *http.Request) (int, error) {
	pageStr := r.URL.Query().Get("page")
	if pageStr == "" {
		pageStr = "1"
	}
	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		return 0, fmt.Errorf("invalid 'page' value': '%s'", pageStr)
	}
	return page, nil
}

func encodeJSON(w http.ResponseWriter, response interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *Handler) HtmlHandler(w http.ResponseWriter, r *http.Request) {
	// Get the "page" parameter from the request
	page, err := getPageParam(r)
	if err != nil {
		handleError(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Define limit and offset
	limit := 10
	offset := (page - 1) * limit

	// Get data from the store
	places, total, err := h.store.GetPlaces(limit, offset)
	if err != nil {
		http.Error(w, "error retrieving data", http.StatusInternalServerError)
		return
	}

	// Generate HTML response
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprintf(w, `<!doctype html>
<html>
<head>
    <meta charset="utf-8">
    <title>Places</title>
    <meta name="description" content="">
    <meta name="viewport" content="width=device-width, initial-scale=1">
	<style>
	.pagination a {
		margin-right: 15px; 
	}
	</style>
</head>
<body>
<h5>Total: %d</h5>
<ul>`, total)

	for _, place := range places {
		fmt.Fprintf(w, `<li>
            <div>%s</div>
            <div>%s</div>
            <div>%s</div>
        </li>`, place.Name, place.Address, place.Phone)
	}

	// Add page navigation
	totalPages := (total + limit - 1) / limit // Calculate the total number of pages
	if page > totalPages {
		http.Error(w, fmt.Sprintf("invalid 'page' value: '%d'", page), http.StatusBadRequest)
		return
	}
	fmt.Fprint(w, `<div class="pagination">`)
	if page > 1 {
		fmt.Fprintf(w, `<a href="?page=%d">« First </a>`, 1)
		fmt.Fprintf(w, `<a href="?page=%d">« Previous </a>`, page-1)
	}

	if page < totalPages {
		fmt.Fprintf(w, `<a href="?page=%d">  Next »</a>`, page+1)
		fmt.Fprintf(w, `<a href="?page=%d">Last »</a>`, totalPages)
	}
	fmt.Fprint(w, `</div>`)

	fmt.Fprint(w, `</ul>
</body>
</html>`)
}

func (h *Handler) JsonHandler(w http.ResponseWriter, r *http.Request) {
	page, err := getPageParam(r)
	if err != nil {
		handleError(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Define limit and offset
	limit := 10
	offset := (page - 1) * limit

	// Get data from the store
	places, totalPlaces, err := h.store.GetPlaces(limit, offset)
	if err != nil {
		http.Error(w, "Error fetching places", http.StatusInternalServerError)
		return
	}
	// Calculate the last page
	lastPage := (totalPlaces + limit - 1) / limit

	// Check if the requested page exceeds the last one
	if page > lastPage {
		http.Error(w, fmt.Sprintf(`{"error": "Invalid 'page' value: '%d'"}`, page), http.StatusBadRequest)
		return
	}
	var prevPage, nextPage int
	if page > 1 {
		prevPage = page - 1
	}
	if page < lastPage {
		nextPage = page + 1
	}

	response := types.PlacesResponse{
		Name:     "Places",
		Total:    totalPlaces,
		Places:   places,
		PrevPage: prevPage,
		NextPage: nextPage,
		LastPage: lastPage,
	}

	encodeJSON(w, response)
}

func (h *Handler) RecommendHandler(w http.ResponseWriter, r *http.Request) {
	// Get the "query" parameter from the request
	latStr := r.URL.Query().Get("lat")
	lonStr := r.URL.Query().Get("lon")
	lat, err := strconv.ParseFloat(latStr, 64)
	if err != nil {
		http.Error(w, fmt.Sprintf("Invalid 'lat' value: '%s'", latStr), http.StatusBadRequest)
		return
	}
	lon, err := strconv.ParseFloat(lonStr, 64)
	if err != nil {
		http.Error(w, fmt.Sprintf("Invalid 'lat' value: '%s'", lonStr), http.StatusBadRequest)
		return
	}
	// log.Printf("Received Latitude: %f, Longitude: %f", lat, lon)
	places, err := db.GetClosestRestaurants(lat, lon, 3)
	if err != nil {
		http.Error(w, "Error fetching places", http.StatusInternalServerError)
		return
	}

	response := struct {
		Name   string        `json:"name"`
		Places []types.Place `json:"places"`
	}{
		Name:   "Recommendation",
		Places: places,
	}

	encodeJSON(w, response)
}

func TokenHandler(w http.ResponseWriter, r *http.Request) {
	// Verify token
	// if r.Method != http.MethodPost {
	// 	http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	// 	return
	// }
	claims := jwt.MapClaims{
		"admin": true,
		"exp":   time.Now().Add(time.Hour).Unix(), // Token expires in 1 hour
		"name":  "Nikola",
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte("secret"))
	if err != nil {
		http.Error(w, "Error creating token", http.StatusInternalServerError)
		return
	}
	encodeJSON(w, map[string]string{"token": tokenString})
}

func JWTMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenString := r.Header.Get("Authorization")
		if tokenString == "" {
			http.Error(w, "Missing token", http.StatusUnauthorized)
			return
		}

		// Remove "Bearer " from the token
		tokenString = tokenString[len("Bearer "):]

		// Verify the token
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			// Check the signing method
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, http.ErrNoCookie
			}
			return []byte("secret"), nil
		})

		if err != nil || !token.Valid {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r) // If the token is valid, continue processing the request
	})
}

// // CORSHandler adds CORS headers
// func CORSHandler(next http.Handler) http.Handler {
// 	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		w.Header().Set("Access-Control-Allow-Origin", "*") // Allow all origins
// 		w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
// 		if r.Method == http.MethodOptions {
// 			w.WriteHeader(http.StatusNoContent)
// 			return
// 		}
// 		next.ServeHTTP(w, r)
// 	})
// }
