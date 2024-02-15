package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

// album represents data about a record album.
type album struct {
    ID     int64  `json:"id"`
    Title  string  `json:"title"`
    Artist string  `json:"artist"`
    Price  float64 `json:"price"`
    Created *time.Time `json:"created,omitempty"`
}

var dataMap = make(map[string]string)

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
            postAlbums(c, db)
        }
    }
    getAlbumHandler := func(db *sql.DB) gin.HandlerFunc {
        return func(c *gin.Context){
            getAlbumByID(c, db)
        }
    }

	r.GET("/albums/:id", getAlbumHandler(db))
	r.POST("/albums", postAlbumsHandler(db) )

	r.Run(":8080") //listen and serve on 0.0.0.0:8000 
}

func createAlbumStore(db *sql.DB){
    query := `CREATE TABLE IF NOT EXISTS albumStore (
        id SERIAL PRIMARY KEY,
        Title VARCHAR(100) NOT NULL,
        Artist VARCHAR(100) NOT NULL,
        Price NUMERIC(6,2) NOT NULL,
        Created timestamp DEFAULT NOW()
    )`
    _, err := db.Exec(query)
    if err != nil {
        log.Fatal(err)
    }
}

func postAlbums(c *gin.Context,  db *sql.DB){
    var newAlbum album
    if err := c.BindJSON(&newAlbum); err != nil {
        return
    }

	existsQuery := "SELECT EXISTS (SELECT 1 FROM albumStore WHERE title = $1)"
    var exists bool
    existsErr := db.QueryRow(existsQuery, newAlbum.Title).Scan(&exists)
    if existsErr != nil {
        // Handle query error
        log.Fatal(existsErr)
        return
    }

    if exists {
        c.JSON(http.StatusBadRequest, gin.H{"message": "Album already exists"})
        return
    }

    query := `INSERT INTO albumStore (title, artist, price)
    VALUES ($1, $2, $3) RETURNING id`
    
    var pk int
    err := db.QueryRow(query, newAlbum.Title, newAlbum.Artist, newAlbum.Price).Scan(&pk)

    if err != nil {
        log.Fatal(err)
    }
    c.IndentedJSON(http.StatusCreated, gin.H{"message": "Album created successfully","id":pk})
}
func getAlbumByID(c *gin.Context, db *sql.DB) {
    id := c.Param("id")

    // Query the database to find the album with the specified ID.
    var album album
    query := "SELECT * FROM albumStore WHERE id = $1"
    err := db.QueryRow(query, id).Scan(&album.ID, &album.Title, &album.Artist, &album.Price, &album.Created)
    if err != nil {
        if err == sql.ErrNoRows {
            c.JSON(http.StatusNotFound, gin.H{"message": "album not found"})
            return
        }
        log.Fatal(err)
        return
    }

    c.JSON(http.StatusOK, album)
}