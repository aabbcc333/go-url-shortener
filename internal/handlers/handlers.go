package handlers

import (
	"math/rand"
	"net/http"

	"github.com/aabbcc333/go-url-shortener/internal/models"
	//"github.com/aabbcc333/go-url-shortener/internal/store"
	"github.com/gin-gonic/gin"
)


type URLStore interface {
    SaveURL(string, string) error
    GetURL(string) (string, error)
}

type UrlHandler struct {
    Store URLStore // <--- CHANGE THIS from *store.StorageService
}
func NewUrlHandler(s URLStore) *UrlHandler{
	return &UrlHandler{Store :s}
}

//helper function to generate ids 
func generateShortCode() string{
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	// Note: In newer Go versions, global rand seeding is automatic
	b := make([]byte, 6)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}

//main fucntoin to shorten the url 
func (h *UrlHandler) CreateShortUrl(c *gin.Context){
	var req models.ShortenRequest
	if err := c.ShouldBindJSON(&req); err != nil { 
		c.JSON(http.StatusBadRequest, gin.H{"error":"invalid json"})
		return 
	}
	code := generateShortCode()

	//calling the store layer 
	err := h.Store.SaveURL(code, req.URL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error":"failed to create url"})
		return
	}
	c.JSON(http.StatusOK, models.ShortenResponse{ShortCode: code})

}

func (h *UrlHandler) ResolveUrl(c *gin.Context){
	code := c.Param("shortCode")
	url, err := h.Store.GetURL(code)
	if err != nil { 
		c.JSON(http.StatusNotFound, gin.H{"error":"URL not found"})
	 return
	}
	// 301 = Permanent Redirect, 302 = Temporary
	// We use 302 here because 301 gets cached by the Browser so hard 
	// that analytics might stop working (the browser skips your server).
	c.Redirect(http.StatusFound, url)
}