package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)



type MockStore struct {
	data map[string]string
}

func (m *MockStore) SaveURL(shortcode string, originalUrl string) error {
	m.data[shortcode] = originalUrl
	return nil
}

func (m *MockStore) GetURL(shortcode string) (string, error){
	val, exist := m.data[shortcode]
	if !exist {
		return "", http.ErrNoCookie // Just a dummy error
	}
	return val, nil
}
// --- STEP 2: The Test ---
func TestCreateShortUrl(t *testing.T) {
	// A. Create the Fake Store
	mockDb := &MockStore{data: make(map[string]string)}

	// B. Create the Handler with the FAKE store
	// THIS IS THE MAGIC: The handler accepts it because it satisfies the interface!
	h := NewUrlHandler(mockDb)

	// C. Setup Gin (Fake Browser)
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	// D. Create a Fake Request
	fakeBody := strings.NewReader(`{"url": "https://google.com"}`)
	c.Request, _ = http.NewRequest("POST", "/api/shorten", fakeBody)

	// E. Run the Function
	h.CreateShortUrl(c)

	// F. Assertions
	// Did we get a 200 OK?
	assert.Equal(t, 200, w.Code)
	
	// Did the data actually save to our fake map?
	// We can check mockDb.data directly!
	assert.NotEmpty(t, mockDb.data, "Map should not be empty")
}