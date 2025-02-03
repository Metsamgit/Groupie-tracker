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
	ID    int          `json:"id"`
	Name  string       `json:"name"`
	Image template.URL `json:"image"`
}

type Relation struct {
	ID             int                 `json:"id"`
	DatesLocations map[string][]string `json:"datesLocations"`
	ArtistName     string
	ArtistImage    template.URL
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
	var filtered []Relation

	for _, artist := range artists {
		if strings.Contains(strings.ToLower(artist.Name), filter) {
			for _, relation := range relations {
				if relation.ID == artist.ID {
					relation.ArtistName = artist.Name
					relation.ArtistImage = artist.Image
					filtered = append(filtered, relation)
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
	http.HandleFunc("/", handler)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	fmt.Println("Serveur démarré sur : http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
