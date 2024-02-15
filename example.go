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

	r.POST("/name/:name", postName)
	r.GET("/names/:name", getNames)
	// r.GET("/albums", getAlbums)
	r.GET("/albums/:id", getAlbumHandler(db))
	r.POST("/albums", postAlbumsHandler(db) )
	// r.DELETE("/albums/delete/:id", deleteAlbum)

	r.Run(":8080") //listen and serve on 0.0.0.0:8000 
}

// func getAlbums(c *gin.Context) {
//     c.IndentedJSON(http.StatusOK, albums)
// }

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
        // Album with the same title already exists in the database
        c.JSON(http.StatusBadRequest, gin.H{"message": "Album already exists"})
        return
    }

    // Add the new album to the slice.
    // albums = append(albums, newAlbum)
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
            // If the album is not found, return a "not found" message in the response.
            c.JSON(http.StatusNotFound, gin.H{"message": "album not found"})
            return
        }
        // Handle other errors
        log.Fatal(err)
        return
    }

    // If the album is found, return its details in the response.
    c.JSON(http.StatusOK, album)
}


// func deleteAlbum(c *gin.Context) {
//     // Get the ID of the album to be deleted from the URL parameters.
//     id := c.Param("id")

//     // Loop through the albums slice to find the album with the specified ID.
//     for i, album := range albums {
//         if album.ID == id {
//             // Remove the album from the slice using slicing technique.
//             albums = append(albums[:i], albums[i+1:]...)
//             // Respond with a success message.
//             c.JSON(http.StatusOK, gin.H{
//                 "message": "Album deleted successfully",
//                 "album":   album,
//             })
//             return
//         }
//     }

//     // If the album with the specified ID is not found, respond with a 404 Not Found error.
//     c.JSON(http.StatusNotFound, gin.H{
//         "error": "Album not found",
//     })
// }


func postName(c *gin.Context){
	name := c.Param("name")
		dataMap[name] = name

		c.JSON(http.StatusOK, gin.H{
			"message":"Name posted successfully.",
			
		})
}

func getNames(c *gin.Context){
	paramName := c.Param("name")

	// Check if the provided name exists in the map
	name, ok := dataMap[paramName]
	if !ok {
		// If the name does not exist in the map, return a 404 Not Found response
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Name not found",
		})
		return
	}

	// Respond with the name from the map in the JSON response
	c.JSON(http.StatusOK, gin.H{
		"name": name,
	})
}