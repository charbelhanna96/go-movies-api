// Models for the movie entity.
package model

type Director struct {
	ID        int    `json:"id"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
}

type Genre struct {
	ID    int    `json:"id"`
	Title string `json:"title"`
}

type Movie struct {
	ID           int      `json:"id"`
	Title        string   `json:"title"`
	YearReleased int      `json:"yearReleased"`
	Rating       float64  `json:"rating"`
	DurationMins int      `json:"durationMins"`
	Genre        Genre    `json:"genre"`
	Director     Director `json:"director"`
}
