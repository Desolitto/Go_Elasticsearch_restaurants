# Elasticsearch-restaurants

## Overview

Elasticsearch-restaurants is a Go project implementing a backend service for managing and searching a large dataset of restaurants in Moscow using Elasticsearch. The service supports:

- Loading and indexing restaurant data into Elasticsearch with proper geo-point mappings.
- Providing a simple HTML interface to browse places with pagination.
- A JSON API to fetch paginated places.
- Geo-based search to find the closest restaurants.
- JWT-based authentication for protected endpoints.

---

## Table of Contents

- [Introduction](#introduction)  
- [Features](#features)  
- [Getting Started](#getting-started)  
- [Usage](#usage)  
  - [Exercise 00: Loading Data](#exercise-00-loading-data)  
  - [Exercise 01: Simplest Interface](#exercise-01-simplest-interface)  
  - [Exercise 02: Proper API](#exercise-02-proper-api)  
  - [Exercise 03: Closest Restaurants](#exercise-03-closest-restaurants)  
  - [Exercise 04: JWT Authentication](#exercise-04-jwt-authentication)  
- [Project Structure](#project-structure)  

---

## Introduction

This project provides a simple recommendation system for restaurants based on geolocation.

- The dataset contains over 13,000 restaurants in Moscow with fields: ID, Name, Address, Phone, Longitude, Latitude.
- Elasticsearch is used as the backend search engine with geo-point support.
- The service exposes HTTP endpoints for browsing, JSON API access, and geo-distance based recommendations.
- JWT authentication secures sensitive endpoints.

---

## Features

- Create Elasticsearch index with explicit mappings for text and geo_point fields.
- Bulk upload of restaurant data into Elasticsearch.
- HTTP server running on port 8888 serving:
  - HTML pages with paginated lists of restaurants.
  - JSON API endpoints with pagination and error handling.
  - Geo-distance based search for closest restaurants.
- Input validation with proper HTTP 400/401 error responses.
- JWT token generation and middleware for securing API endpoints.

---

## Getting Started

### Prerequisites

- Go 1.16+
- Elasticsearch 8.x running locally or remotely
- Git

### Installation

```bash
git clone https://github.com/Desolitto/Go_Elasticsearch_restaurants
cd Go_Elasticsearch_restaurants
go mod tidy
```

### Running Elasticsearch

Start Elasticsearch as per your installation, e.g.:

```bash
/path/to/elasticsearch/bin/elasticsearch
```

---

## Usage

### Exercise 00: Loading Data

- Create an Elasticsearch index named `places` with mapping specifying types for `name`, `address`, `phone`, and `location` (geo_point).
- Use Go Elasticsearch client to create the index and apply mapping programmatically.
- Bulk upload restaurant data into the `places` index.
- Validate the index and documents via Elasticsearch API or Kibana.

### Exercise 01: Simplest Interface

- HTTP server listens on port 8888.
- Serve an HTML page listing restaurant names, addresses, and phones.
- Support pagination via `?page=N` query parameter.
- Return HTTP 400 for invalid page parameters.
- Hide "Previous" link on first page and "Next" link on last page.
- Adjust Elasticsearch index settings to allow pagination beyond 10,000 results if needed.

Example page URL:

```
http://127.0.0.1:8888/?page=2
```

### Exercise 02: Proper API

- Add JSON API endpoint `/api/places` supporting pagination with `?page=N`.
- Return JSON response with fields `name`, `total`, `places` (list), `prev_page`, `next_page`, `last_page`.
- Validate `page` parameter and return HTTP 400 with JSON error on invalid input.

Example API call:

```bash
curl "http://127.0.0.1:8888/api/places?page=3"
```

### Exercise 03: Closest Restaurants

- Implement `/api/recommend?lat=<latitude>&lon=<longitude>` endpoint.
- Return JSON list of 3 closest restaurants sorted by geo-distance.
- Validate `lat` and `lon` parameters, return HTTP 400 on invalid input.

Example API call:

```bash
curl "http://127.0.0.1:8888/api/recommend?lat=55.674&lon=37.666"
```

### Exercise 04: JWT Authentication

- Implement `/api/get_token` endpoint to generate and return JWT tokens.
- Protect `/api/recommend` endpoint with JWT middleware.
- Require `Authorization: Bearer <token>` header to access protected endpoint.
- Return HTTP 401 Unauthorized if token is missing or invalid.

Example token response:

```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

---

## Project Structure

```
src/
├── cmd/
│   └── main.go          # Entry point for the HTTP server
├── internal/
│   ├── data/            # Data loading and processing utilities
│   ├── db/              # Elasticsearch client wrappers and DB interface implementations
│   ├── elasticsearch/   # Elasticsearch specific logic: index creation, bulk upload, queries
│   ├── handlers/        # HTTP handlers for API endpoints and web pages
│   └── models/          # Data models and types (e.g., Place struct)
├── go.mod
├── go.sum
└── README.md
```

---
