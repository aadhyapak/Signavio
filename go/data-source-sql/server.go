package main

// Signavio Workflow connector example that provides a data source for a list of countries.

import (
	"database/sql"
	"encoding/json"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"log"
	"net/http"
	"strings"
)

var address = ":9000"

type Option struct {
	Code string `json:"code"`
	Name string `json:"name"`
}

// Connects to the database and tests the connection
func database() (*sql.DB) {
	db, err := sql.Open("mysql", "cldr:cldr@tcp(127.0.0.1:3306)/cldr")
	if err != nil {
		panic(err.Error())
	}
	return db
}

// Finds a country by code.
func findOne(code string) (Option, bool) {
	db := database()
	defer db.Close()

	var country Option
	sql := "select code, name from countries where code = ?"
	err := db.QueryRow(sql, code).Scan(&country.Code, &country.Name)
	if err != nil {
		log.Fatal(err)
	  return Option{}, false
	}
	return country, true
}

// Serves a single country.
func country(response http.ResponseWriter, request *http.Request) {
  log.Printf("%v %v\n", request.Method, request.URL.Path)
	code := strings.TrimPrefix(request.URL.Path, "/country/")
	country, found := findOne(code)
	if found {
		json, _ := json.MarshalIndent(country, "", "  ")
		response.Header().Set("Content-Type", "application/json")
		response.Write(json)
	} else {
		response.WriteHeader(404) // Not Found
	}
}

// Finds countries whose names contain the given query.
func find(query string) []Option {
	sql := "select code, name from countries where '' = ? || lower(name) like ? order by name"
	db := database()
	results, err := db.Query(sql, query, fmt.Sprintf("%%%v%%", strings.ToLower(query)))
	if err != nil {
		panic(err.Error())
	}
	defer results.Close()

	var countries []Option
	for results.Next() {
		var country Option
		err = results.Scan(&country.Code, &country.Name)
		if err != nil {
			panic(err.Error())
		}

		fmt.Printf("%v %v\n", country.Code, country.Name)
		countries = append(countries, country)
	}

	return countries
}

// Serves the list of countries.
func options(response http.ResponseWriter, request *http.Request) {
	query := request.URL.Query().Get("filter")
	var options = find(query)
	json, _ := json.MarshalIndent(&options, "", "  ")
	response.Header().Set("Content-Type", "application/json")
	response.Write(json)
}

// Serves a single country option.
func option(response http.ResponseWriter, request *http.Request) {
	code := strings.TrimPrefix(request.URL.Path, "/country/options/")
	country, found := findOne(code)
	if found {
		var option = Option{country.Code, country.Name}
		json, _ := json.MarshalIndent(option, "", "  ")
		response.Header().Set("Content-Type", "application/json")
		response.Write(json)
	} else {
		response.WriteHeader(404) // Not Found
	}
}

func descriptor(response http.ResponseWriter, request *http.Request) {
	http.ServeFile(response, request, "descriptor.json")
}

// Serves a connector over HTTP.
func main() {
	http.HandleFunc("/", descriptor)
	http.HandleFunc("/country/options/", option)
	http.HandleFunc("/country/options", options)
	http.HandleFunc("/country/", country)
	println("Listening on " + address)
	panic(http.ListenAndServe(address, nil))
}
