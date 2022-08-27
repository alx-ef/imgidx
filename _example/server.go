package main

import (
	"fmt"
	"github.com/alef-ru/imgidx/embedders"
	"github.com/alef-ru/imgidx/index"
	"gorm.io/driver/sqlite"
	"image"
	"log"
	"mime/multipart"
	"net/http"
	"net/url"
	"strings"
)
import "github.com/gin-gonic/gin"

var idx index.Index

type AddImageRequest struct {
	Url        string      `json:"url"  binding:"required"`
	Attributes interface{} `json:"attrs"`
}

func newIndex() index.Index {
	kd3Idx, err := index.NewKDTreeImageIndex(
		embedders.Composition([]embedders.ImageEmbedder{
			embedders.NewAspectRatioEmbedder(),
			embedders.NewColorDispersionEmbedder(),
			embedders.NewLowResolutionEmbedder(8, 8),
		}),
	)
	if err != nil {
		log.Fatalf("Failed to create KDTreeIndex index : %v", err)
	}
	persistentIdx, err := index.NewPersistentIndex(sqlite.Open("imgidx.db"), kd3Idx)
	if err != nil {
		log.Fatalf("Failed to create persistent index : %v", err)
	}
	return persistentIdx
}
func initAndRunWebServer() {
	r := gin.New()
	gin.EnableJsonDecoderDisallowUnknownFields()
	r.POST("/images/", addImage)     // Add new Images to the index
	r.GET("/images/*url", findByURL) // Find the most similar image by URL
	r.POST("/find-similar-to-file/", findByFile)
	r.StaticFile("/", "_example/spa.html")
	r.MaxMultipartMemory = 8 << 20 // 8 MiB
	err := r.Run(":8080")
	if err != nil {
		log.Fatal(err)
	}
}

func validationError(c *gin.Context, message string) {
	c.JSON(http.StatusBadRequest, gin.H{
		"message": message,
	})
}

func addImage(c *gin.Context) {
	req := AddImageRequest{}
	if err := c.BindJSON(&req); err != nil {
		validationError(c, err.Error())
		return
	}
	if _, err := url.ParseRequestURI(req.Url); err != nil {
		validationError(c, err.Error())
		return
	}
	_, err := index.AddImageUrl(idx, req.Url, req.Attributes)
	if err != nil {
		validationError(c, err.Error())
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "done",
	})
}

func findByURL(c *gin.Context) {
	imgUrl := strings.TrimPrefix(c.Param("url"), "/")
	if _, err := url.ParseRequestURI(imgUrl); err != nil {
		validationError(c, err.Error())
		return
	}
	nearestImgUrl, attrs, dist, err := index.NearestByURL(idx, imgUrl)
	if err != nil {
		validationError(c, err.Error())
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"url":                nearestImgUrl,
		"additional_details": attrs,
		"distance":           dist,
	})
}

func findByFile(c *gin.Context) {
	file, err := c.FormFile("image-file")
	if err != nil {
		validationError(c, err.Error())
		return
	}
	f, err := file.Open()
	if err != nil {
		validationError(c, err.Error())
		return
	}
	defer func(f multipart.File) {
		err := f.Close()
		if err != nil {
			log.Printf("Failed to close file from HTTP request: %v", err)
		}
	}(f)
	queryImg, _, err := image.Decode(f)
	if err != nil {
		validationError(c, fmt.Sprintf("Failed to decode image %v", err))
		return
	}

	nearestImgUrl, attrs, dist, err := idx.Nearest(queryImg)

	if err != nil {
		validationError(c, fmt.Sprintf("Failed to find similar image : %v", err))
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"url":                nearestImgUrl,
		"additional_details": attrs,
		"distance":           dist,
	})
}

func main() {
	idx = newIndex()
	initAndRunWebServer()
}
