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
    getAlbumsHandler := func(db *sql.DB) gin.HandlerFunc {
        return func(c *gin.Context){
            getAlbums(c, db)
        }
    }
    updateAlbumHandler := func(db *sql.DB) gin.HandlerFunc {
        return func(c *gin.Context){
            updateAlbum(c, db)
        }
    }
    deleteAlbumHandler := func(db *sql.DB) gin.HandlerFunc{
        return func(c *gin.Context){
            deleteAlbum(c, db)
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

func postAlbums(c *gin.Context,  db *sql.DB){
    var newAlbum Album
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

func updateAlbum(c *gin.Context, db *sql.DB){
    id := c.Param("id")
    row := db.QueryRow("SELECT * FROM albumStore WHERE id = $1", id)

    var album Album
    if err := row.Scan(&album.ID,&album.Title,&album.Artist, &album.Price,  &album.Created); err != nil{
        if err == sql.ErrNoRows{
            c.JSON(http.StatusNotFound, gin.H{"message":"Album not found."})
        }
        //any other error
        log.Fatal(err)
        c.JSON(http.StatusInternalServerError, gin.H{"message":"Internal server error"})
        return
    }

     // Bind JSON payload to updatedAlbum
     var updatedAlbum Album
     if err := c.BindJSON(&updatedAlbum); err != nil {
         c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
         return
     }
 
     // Update corresponding fields of existing album with values from updatedAlbum
     if updatedAlbum.Title != "" {
         album.Title = updatedAlbum.Title
     }
     if updatedAlbum.Artist != "" {
         album.Artist = updatedAlbum.Artist
     }
     if updatedAlbum.Price != 0 {
         album.Price = updatedAlbum.Price
     }

      // Execute SQL UPDATE statement to update the album record in the database
    _, err := db.Exec("UPDATE albumStore SET title = $1, artist = $2, price = $3 WHERE id = $4",
    album.Title, album.Artist, album.Price, album.ID)
if err != nil {
    log.Fatal(err)
    c.JSON(http.StatusInternalServerError, gin.H{"message": "Internal server error"})
    return
}

// Return the updated album in the response
c.JSON(http.StatusOK, album)

}

func getAlbumByID(c *gin.Context, db *sql.DB) {
    id := c.Param("id")

    // Query the database to find the album with the specified ID.
    var album Album
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
func getAlbums(c *gin.Context, db *sql.DB){
    data := []Album{}
    rows, err := db.Query("SELECT title, artist, price, id FROM albumStore")
    if err != nil {
        log.Fatal(err)
    }
    defer rows.Close()
    var title string
    var artist string
    var price float64
    var id int64

    for rows.Next(){
        err := rows.Scan(&title, &artist, &price, &id)
        if err != nil {
            log.Fatal(err)
        }
        data = append(data, Album{Title: title, Artist: artist, Price: price, ID: int64(id)})
    }
    c.JSON(http.StatusOK, data)
}

func deleteAlbum(c *gin.Context, db *sql.DB){
    id := c.Param("id")

    var exists bool
    err := db.QueryRow("SELECT EXISTS (SELECT 1 FROM albumStore WHERE id = $1)", id).Scan(&exists)
    if err != nil {
        // Log the error message
        log.Println("Error checking album existence:", err)
        // Return an error response to the client
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check album existence"})
        return
    }
    
    if !exists {
        // Album does not exist, return a 404 Not Found response
        c.JSON(http.StatusNotFound, gin.H{"error": "Album not found"})
        return
    }
    
    query := `DELETE FROM albumStore WHERE id = $1`
    _ , deletionErr := db.Exec(query, id)
    if deletionErr != nil {
        log.Println("Error deleting album:", deletionErr)
        // Return an error response to the client
        c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to delete album"})
        return
    }
    c.JSON(http.StatusOK, gin.H{"message": "Album deleted successfully"})
}