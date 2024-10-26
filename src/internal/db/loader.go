package db

import (
	types "elasticsearch_recommender/internal/models"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
)

func LoadPlaces(filePath string) ([]types.Place, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, errors.New("error opening file: " + err.Error())
	}
	defer file.Close()
	reader := csv.NewReader(file)
	reader.Comma = '\t'
	if _, err = reader.Read(); err != nil {
		return nil, err
	}
	var places []types.Place

	for {
		record, err := reader.Read()
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}
		if len(record) < 6 {
			return nil, fmt.Errorf("record has insufficient fields: %v", record)
		}
		id, err := strconv.Atoi(record[0])
		if err != nil {
			return nil, fmt.Errorf("error converting ID to int: %s", err)
		}
		// широта и долгота перупутаны в csv, переворачиваем
		lat, _ := strconv.ParseFloat(record[4], 64)
		lon, _ := strconv.ParseFloat(record[5], 64)
		place := types.Place{
			ID:       id,
			Name:     record[1],
			Address:  record[2],
			Phone:    record[3],
			Location: types.GeoPoint{Lat: lon, Lon: lat},
		}
		places = append(places, place)
	}
	return places, nil
}
