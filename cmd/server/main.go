package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/aabbcc333/go-url-shortener/internal/handlers"
	"github.com/aabbcc333/go-url-shortener/internal/store"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
    _ "github.com/lib/pq"
)





func main(){
	//loading config 
	if err := godotenv.Load(); err != nil{
		log.Println("⚠️  .env file not found, checking system vars")
	}
	//setting up redis 
	rdb := redis.NewClient(&redis.Options{
		Addr : "localhost:6379",
		Password : os.Getenv("REDIS_PASSWORD"),

	})
	// 3Stup Postgres Connection
	dsn := fmt.Sprintf("host=localhost port=5433 user=%s password=%s dbname=%s sslmode=disable",
		os.Getenv("DB_USER"), os.Getenv("DB_PASSWORD"), os.Getenv("DB_NAME"))
	
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatal("Failed to connect to DB:", err)
	}

	// 4. Run Migrations (Create Table)
	// Ideally this is a separate script, but we keep it here for simplicity
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS urls(
		id SERIAL PRIMARY KEY,
		short_code VARCHAR(10) UNIQUE NOT NULL,
		long_url TEXT NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP, 
		clicks BIGINT DEFAULT 0
	);`)
	if err != nil {
		log.Fatal("Schema error:", err)
	}
	//initalize layers 
	storageService := store.NewStorage(db,rdb)
	urlHandler := handlers.NewUrlHandler(storageService)

	//setupRouter 
	r := gin.Default()
	r.GET("/ping",func(c *gin.Context){
		c.JSON(200, gin.H{"message":"pong"})

	})
	r.POST("/api/shorten", urlHandler.CreateShortUrl)
	r.GET("/:shortCode",urlHandler.ResolveUrl)
	fmt.Println("🚀 Server running on :8080")
	if err := r.Run(":8081"); err != nil {
		log.Fatal("Server failed:", err)
	}
}