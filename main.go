package main

import (
    "html/template"
    "os"
)

type Artist struct {
    Name        string
    ImageURL    string
    Birthdate   string
    LastConcert string
    LastAlbum   string
}

func main() {
    tmpl, err := template.ParseFiles("template.html")
    if err != nil {
        panic(err)
    }

    artists := []Artist{
        {
            Name:        "Imagine Dragons",
            ImageURL:    "images/imagine-dragons.jpg",
            Birthdate:   "2008",
            LastConcert: "15 août 2023",
            LastAlbum:   "3 septembre 2021",
        },
        {
            Name:        "Coldplay",
            ImageURL:    "images/coldplay.jpg",
            Birthdate:   "1996",
            LastConcert: "10 juillet 2023",
            LastAlbum:   "15 octobre 2021",
        },
        {
            Name:        "Muse",
            ImageURL:    "images/muse.jpg",
            Birthdate:   "1994",
            LastConcert: "20 juin 2023",
            LastAlbum:   "26 août 2022",
        },
    }

    for _, artist := range artists {
        file, err := os.Create(artist.Name + ".html")
        if err != nil {
            panic(err)
        }
        defer file.Close()

        err = tmpl.Execute(file, artist)
        if err != nil {
            panic(err)
        }
    }
}