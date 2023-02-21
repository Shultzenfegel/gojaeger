package main

import (
	"context"
	"log"
	"math/rand"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/shultzenfegel/gojaeger/tracer"

	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
)

var ginTracer = otel.Tracer("gin-example")

func init() {
	rand.Seed(time.Now().UnixNano())
}

func main() {
	// tp, err := tracer.InitStdoutTracer()
	tp, err := tracer.InitJaegerTracer("http://localhost:14268/api/traces")
	if err != nil {
		log.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Cleanly shutdown and flush telemetry when the application exits.
	defer func(ctx context.Context) {
		// Do not make the application hang when it is shutdown.
		ctx, cancel = context.WithTimeout(ctx, time.Second*5)
		defer cancel()
		if err := tp.Shutdown(ctx); err != nil {
			log.Fatal(err)
		}
	}(ctx)

	router := gin.Default()
	router.Use(otelgin.Middleware("x-demo-traces-server"))

	router.GET("/albums", getAlbums)
	router.GET("/albums/:id", getAlbumByID)
	router.POST("/albums", postAlbums)

	router.Run("localhost:8080")
}

// album represents data about a record album.
type album struct {
	ID     string  `json:"id"`
	Title  string  `json:"title"`
	Artist string  `json:"artist"`
	Price  float64 `json:"price"`
}

// albums slice to seed record album data.
var albums = []album{
	{ID: "1", Title: "Blue Train", Artist: "John Coltrane", Price: 56.99},
	{ID: "2", Title: "Jeru", Artist: "Gerry Mulligan", Price: 17.99},
	{ID: "3", Title: "Sarah Vaughan and Clifford Brown", Artist: "Sarah Vaughan", Price: 39.99},
}

// getAlbums responds with the list of all albums as JSON.
func getAlbums(c *gin.Context) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	ctx, span := ginTracer.Start(ctx, "getAlbums_span")
	defer span.End()

	randomSleep(ctx)
	getAlbumsInternal(ctx)
	c.IndentedJSON(http.StatusOK, albums)
}

func randomSleep(ctx context.Context) {
	n := rand.Intn(1000)

	tr := otel.Tracer("randomSleep")
	_, span := tr.Start(ctx, "randomSleep_span")
	span.SetAttributes(attribute.Key("sleep_time").Int(n))
	defer span.End()

	time.Sleep(time.Duration(n) * time.Millisecond)
}

func getAlbumsInternal(ctx context.Context) []album {
	_, span := ginTracer.Start(ctx, "getAlbumsInternal_span")
	defer span.End()

	n := rand.Intn(1000)
	time.Sleep(time.Duration(n) * time.Millisecond)

	return albums
}

// getAlbumByID locates the album whose ID value matches the id
// parameter sent by the client, then returns that album as a response.
func getAlbumByID(c *gin.Context) {
	id := c.Param("id")

	// Loop over the list of albums, looking for
	// an album whose ID value matches the parameter.
	for _, a := range albums {
		if a.ID == id {
			c.IndentedJSON(http.StatusOK, a)
			return
		}
	}
	c.IndentedJSON(http.StatusNotFound, gin.H{"message": "album not found"})
}

// postAlbums adds an album from JSON received in the request body.
func postAlbums(c *gin.Context) {
	var newAlbum album

	// Call BindJSON to bind the received JSON to
	// newAlbum.
	if err := c.BindJSON(&newAlbum); err != nil {
		return
	}

	// Add the new album to the slice.
	albums = append(albums, newAlbum)
	c.IndentedJSON(http.StatusCreated, newAlbum)
}
