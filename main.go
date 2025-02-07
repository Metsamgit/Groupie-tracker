package main

import (
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/zmb3/spotify"
	"golang.org/x/oauth2/clientcredentials"
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

func IndexHandler(w http.ResponseWriter, r *http.Request) {
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

func SpotifyHandler(w http.ResponseWriter, r *http.Request) {
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 3 {
		http.Error(w, "ID de l'artiste manquant", http.StatusBadRequest)
		return
	}
	artistID := pathParts[2] // Exemple : "/spotify/1" → "1"

	// Récupérer la liste des artistes depuis ton API locale
	artists, err := getArtists()
	if err != nil {
		http.Error(w, "Erreur lors de la récupération des artistes", http.StatusInternalServerError)
		return
	}

	var artistName string
	for _, artist := range artists {
		if fmt.Sprint(artist.ID) == artistID {
			artistName = artist.Name
			break
		}
	}

	// Si l'artiste n'est pas trouvé
	if artistName == "" {
		http.Error(w, "Artiste non trouvé", http.StatusNotFound)
		return
	}

	authConfig := &clientcredentials.Config{
		ClientID:     "f7e0d63c72e14c4690c5f4e3c956fc9f",
		ClientSecret: "110f9b30484946108d3c1feea68e473c",
		TokenURL:     spotify.TokenURL,
	}

	accessToken, err := authConfig.Token(context.Background())
	if err != nil {
		log.Fatalf("error retrieve access token: %v", err)
	}

	client := spotify.Authenticator{}.NewClient(accessToken)

	results, err := client.Search(artistName, spotify.SearchTypeArtist)

	if err != nil {
		log.Fatalf("error searching for artist: %v", err)
	}

	if len(results.Artists.Artists) == 0 {
		log.Fatalf("no artist found for %s", artistName)
		return
	}

	topArtist := results.Artists.Artists[0]

	topTracks, err := client.GetArtistsTopTracks(topArtist.ID, "FR")
	if err != nil {
		log.Fatalf("Erreur lors de la récupération des top tracks: %v", err)
	}

	// Vérifier si des tracks existent
	if len(topTracks) == 0 {
		log.Println("Aucun top track trouvé")
		return
	}
	
	response := map[string]string{
		"iframe": fmt.Sprintf(`<iframe src="https://open.spotify.com/embed/track/%s" width="300" height="380" frameborder="0" allowtransparency="true" allow="encrypted-media"></iframe>`, topTracks[0].ID),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func suggestionsHandler(w http.ResponseWriter, r *http.Request) {
	artists, err := getArtists()
	if err != nil {
		http.Error(w, "Erreur de récupération des artistes", http.StatusInternalServerError)
		return
	}

	query := strings.ToLower(r.URL.Query().Get("q"))
	var matches []Artist

	for _, artist := range artists {
		if strings.HasPrefix(strings.ToLower(artist.Name), query) { // On vérifie si ça commence par la lettre tapée
			matches = append(matches, artist)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(matches)
}

func ArtistHandler(w http.ResponseWriter, r *http.Request) {
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

	path := strings.Split(r.URL.Path, "/")

	artistID := path[len(path)-1]

	var selectedArtist *Artist
	var selectedRelations *Relation

	for _, artist := range artists {
		if fmt.Sprint(artist.ID) == artistID {
			selectedArtist = &artist
			break
		}
	}

	for _, relation := range relations {
		if fmt.Sprint(relation.ID) == artistID {
			selectedRelations = &relation
			break
		}
	}

	if selectedArtist == nil || selectedRelations == nil {
		http.Error(w, "Artiste non trouvé", http.StatusNotFound)
		return
	}

	selectedRelations.ArtistName = selectedArtist.Name
	selectedRelations.ArtistImage = selectedArtist.Image

	tmpl, err := template.ParseFiles("templates/artist.html")
	if err != nil {
		http.Error(w, "Erreur de chargement du template", http.StatusInternalServerError)
		return
	}

	tmpl.Execute(w, selectedRelations)

}

func main() {
	http.HandleFunc("/", IndexHandler)
	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))
	http.HandleFunc("/suggestions", suggestionsHandler)
	http.HandleFunc("/artist/", ArtistHandler)
	http.HandleFunc("/spotify/", SpotifyHandler)
	fmt.Println("Serveur démarré sur : http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
