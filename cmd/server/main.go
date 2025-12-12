package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/aabbcc333/go-url-shortener/internal/handlers"
	"github.com/aabbcc333/go-url-shortener/internal/middleware"
	"github.com/aabbcc333/go-url-shortener/internal/store"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/redis/go-redis/v9"
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
	
	r.GET("/:shortCode",urlHandler.ResolveUrl)
	
	api := r.Group("/api")
	api.Use(middleware.RateLimiter(rdb))
	{
		api.POST("/shorten", urlHandler.CreateShortUrl)
	}
	srv := http.Server{
		Addr : ":8081",
		Handler: r,
		ReadHeaderTimeout: 2 * time.Second, 
		WriteTimeout: 30 * time.Second,
		ReadTimeout: 10 * time.Second,
		IdleTimeout: 60 * time.Second,
	}
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	go func(){
	log.Printf("listen and server on 8081")
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed{
		log.Fatalf("could not start the server")
	}
    }()
	sig := <- stop 
	log.Printf("recived %s signal",sig)

	ctx, cancel := context.WithTimeout(context.Background(), 10 * time.Second)
	defer cancel() 

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("❌ Server forced to shutdown: %v", err)
	}

	log.Println("closing datbase connections")
	if err:= db.Close(); err != nil{
		log.Printf("error clsoing postgres")
	}
	if err:= rdb.Close(); err != nil{
		log.Printf("Error closing redis")
	}
	log.Println("server existed gracefully")
}