package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"              // <--- ADD THIS (The Postgres Driver)
	"github.com/redis/go-redis/v9"
	"github.com/joho/godotenv"
)

//The Global State
//these variables will hold our open connection for the entire life of app

var (
	ctx = context.Background()
	rdb *redis.Client
	db *sql.DB
)

func initInfraStructure(){
	//loading the .evn file
	if err := godotenv.Load(); err != nil{
       log.Println("No env file found")
	}

	//Connecting to REDIS
	rdb = redis.NewClient(&redis.Options{
		Addr : "localhost:6379",
		Password: os.Getenv("REDIS_PASSWORD"),
		DB: 0,
	})

	//check if redis is actually alive 
	if _, err := rdb.Ping(ctx).Result(); err != nil{
		log.Fatalf("Redis is down: %v", err)
	}
	fmt.Println("connected to redis")

	//connecting to postgres
	dsn := fmt.Sprintf("host=localhost port=5433 user=%s password=%s dbname=%s sslmode=disable",
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
	)
	var err error
	db, err = sql.Open("postgres", dsn) // open the pool
	if err != nil {
		log.Fatalf("driver error : %v", err)
	}

	//checking if postgres is alive 
	if err = db.Ping(); err != nil {
		log.Fatalf("❌ Postgres is down: %v", err)
	}
	fmt.Println("✅ Connected to Postgres")

	// creating a table automatically
	// Fixed typos: "EXISTS" and "BIGINT"
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS urls(
		id SERIAL PRIMARY KEY,
		short_code VARCHAR(10) UNIQUE NOT NULL,
		long_url TEXT NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP, 
		clicks BIGINT DEFAULT 0
	);`)
	if err != nil {
		log.Fatalf("❌ Schema error: %v", err)
	}
}

func main(){
	initInfraStructure()//connecting to DBs
	r := gin.Default()
	//simple health check
	r.GET("/ping", func(c *gin.Context){
		c.JSON(200,gin.H{"message":"pong"})
	})
	r.Run(":8080")//listen on port 8080

}