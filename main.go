package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
)

type Artist struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Image string `json:"image"`
}

type Relation struct {
	ID              int                 `json:"id"`
	DatesLocations  map[string][]string `json:"datesLocations"`
	ArtistName      string
	ArtistNameLower string
	ArtistImage     string `json:"image"`
}

type APIResponse struct {
	Index []Relation `json:"index"`
}

var tmpl = template.Must(template.ParseFiles("templates/index.html"))

func getArtists() ([]Artist, error) {
	url := "https://groupietrackers.herokuapp.com/api/artists"
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var artists []Artist
	err = json.Unmarshal(body, &artists)
	if err != nil {
		return nil, err
	}

	return artists, nil
}

func getRelations() ([]Relation, error) {
	url := "https://groupietrackers.herokuapp.com/api/relation"
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var apiResponse APIResponse
	err = json.Unmarshal(body, &apiResponse)
	if err != nil {
		return nil, err
	}

	return apiResponse.Index, nil
}

func handler(w http.ResponseWriter, r *http.Request) {
	relations, err := getRelations()
	if err != nil {
		http.Error(w, "Erreur de récupération des données", http.StatusInternalServerError)
		return
	}

	artists, err := getArtists()
	if err != nil {
		http.Error(w, "Erreur de récupération des artistes", http.StatusInternalServerError)
		return
	}

	filter := strings.ToLower(r.URL.Query().Get("filter"))
	filteredMap := make(map[int]Relation)
	var filtered []Relation

	for _, relation := range relations {
		for location, dates := range relation.DatesLocations {
			for _, date := range dates {
				ArtistNameLower := strings.ReplaceAll(strings.ToLower(artists[relation.ID-1].Name), " ", "")
				if strings.Contains(strings.ToLower(location), filter) ||
					strings.Contains(strings.ToLower(date), filter) ||
					strings.Contains(strings.ToLower(artists[relation.ID-1].Name), filter) {
					relation.ArtistImage = artists[relation.ID-1].Image
					relation.ArtistName = artists[relation.ID-1].Name
					relation.ArtistNameLower = ArtistNameLower
					if _, exists := filteredMap[relation.ID]; !exists {
						filteredMap[relation.ID] = relation
						filtered = append(filtered, relation)
					}
					break
				}
			}
		}
	}

	err = tmpl.Execute(w, filtered)
	if err != nil {
		log.Printf("Erreur lors de l'exécution du template : %v", err)
		http.Error(w, "Erreur interne du serveur", http.StatusInternalServerError)
	}
}

func main() {
	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))
	http.HandleFunc("/", handler)
	fmt.Println("Serveur démarré sur : http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
