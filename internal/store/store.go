package store

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)


type StorageService struct {
	Postgres *sql.DB
	Redis *redis.Client
}

//create a new storageService service 

func NewStorage(pg *sql.DB, rdb *redis.Client) *StorageService{
	return &StorageService{
		Postgres: pg,
		Redis: rdb,
	}
}

//SaveURL handles he write-through logic (postgres + redis)
func (s *StorageService) SaveURL(shortCode string, originalURL string) error {
	ctx := context.Background()
	//write to postgres, local varialbe (s)
	_, err := s.Postgres.Exec("INSERT INTO urls (short_code, long_url ) VALUES ( $1 , $2)", shortCode , originalURL)
	if err != nil {
		return fmt.Errorf("postgres insert error: %w", err)
	}

	//write to redis 
	err = s.Redis.Set(ctx, "short:"+shortCode, originalURL, 24*time.Hour).Err()
	if err != nil {
		// Log but don't fail, because Postgres succeeded
		log.Printf("⚠️ Redis write failed: %v", err)
	}
	return nil
	
}

//get-url handles the chahce aside logic 
func (s *StorageService) GetURL(shortCode string)(string, error){
	ctx := context.Background()
	//check redis 
	val, err := s.Redis.Get(ctx, "short:"+shortCode).Result()
	if err == nil { 
		return val, nil //cahche hit
	}
	//check postgres (cache miss)
	var longURL string
	err = s.Postgres.QueryRow("SELECT long_url FROM urls WHERE short_code = $1", shortCode).Scan(&longURL)
	if err != nil { 
		return "", err //not found in db either 
	}

	//popilate redis ( cahche warmign )
	s.Redis.Set(ctx, "short:"+shortCode, longURL, 24*time.Hour)
	return longURL, nil
}