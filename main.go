package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/munene-m/go-rest-api/controllers"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

// album represents data about a record album.
type Album struct {
    ID     int64  `json:"id"`
    Title  string  `json:"title"`
    Artist string  `json:"artist"`
    Price  float64 `json:"price"`
    Created *time.Time `json:"created,omitempty"`
}

// var dataMap = make(map[string]string)

func main(){
    err := godotenv.Load(".env")
    if err != nil {
        log.Fatalf("Error loading .env file: %v", err)
    }
    dbPassword := os.Getenv("DB_PASSWORD")
    dbName := os.Getenv("DB_NAME")
    dbPort := os.Getenv("DB_PORT")

    connectionStr := fmt.Sprintf("postgres://postgres:%s@localhost:%s/%s?sslmode=disable", dbPassword, dbPort, dbName)
    db, err := sql.Open("postgres", connectionStr)
    if err != nil{
        log.Fatal(err)
    }
    defer db.Close()

    if err != nil {
        log.Fatal(err)
    }
    if err = db.Ping(); err != nil {
        log.Fatal(err)
    }
    createAlbumStore(db)
	r := gin.Default()

    postAlbumsHandler := func(db *sql.DB) gin.HandlerFunc {
        return func(c *gin.Context) {
            controllers.PostAlbums(c, db)
        }
    }
    getAlbumHandler := func(db *sql.DB) gin.HandlerFunc {
        return func(c *gin.Context){
            controllers.GetAlbumByID(c, db)
        }
    }
    getAlbumsHandler := func(db *sql.DB) gin.HandlerFunc {
        return func(c *gin.Context){
            controllers.GetAlbums(c, db)
        }
    }
    updateAlbumHandler := func(db *sql.DB) gin.HandlerFunc {
        return func(c *gin.Context){
            controllers.UpdateAlbum(c, db)
        }
    }
    deleteAlbumHandler := func(db *sql.DB) gin.HandlerFunc{
        return func(c *gin.Context){
            controllers.DeleteAlbum(c, db)
        }
    }

	r.GET("/albums/:id", getAlbumHandler(db))
    r.GET("/albums", getAlbumsHandler(db))
	r.POST("/albums", postAlbumsHandler(db) )
    r.PUT("/albums/update/:id", updateAlbumHandler(db))
    r.DELETE("/albums/delete/:id", deleteAlbumHandler((db)))

	r.Run(":8080") //listen and serve on 0.0.0.0:8000 
}

func createAlbumStore(db *sql.DB){
    query := `CREATE TABLE IF NOT EXISTS albumStore (
        id SERIAL PRIMARY KEY,
        Title VARCHAR(100) NOT NULL,
        Artist VARCHAR(100) NOT NULL,
        Price NUMERIC(10,2) NOT NULL, 
        Created timestamp DEFAULT NOW()
    )`
    _, err := db.Exec(query)
    if err != nil {
        log.Fatal(err)
    }
}