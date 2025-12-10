package models

//shortenRequest is what we expect from the user 
type ShortenRequest struct { 
	URL string `json:"url" binding:"required"`
}

type ShortenResponse struct {
	ShortCode string `json:"short_code"`
}