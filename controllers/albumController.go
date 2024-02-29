package controllers

import (
	"database/sql"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// album represents data about a record album.
type Album struct {
    ID     int64  `json:"id"`
    Title  string  `json:"title"`
    Artist string  `json:"artist"`
    Price  float64 `json:"price"`
    Created *time.Time `json:"created,omitempty"`
}

func PostAlbums(c *gin.Context,  db *sql.DB){
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

func UpdateAlbum(c *gin.Context, db *sql.DB){
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

func GetAlbumByID(c *gin.Context, db *sql.DB) {
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

func GetAlbums(c *gin.Context, db *sql.DB){
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

func DeleteAlbum(c *gin.Context, db *sql.DB){
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