package main

import (
	"fmt"
	"github.com/alef-ru/imgidx"
	"gorm.io/driver/sqlite"
	"image"
	"log"
	"mime/multipart"
	"net/http"
	"net/url"
	"strings"
)
import "github.com/gin-gonic/gin"

var idx imgidx.Index

type AddImageRequest struct {
	Url        string      `json:"url"  binding:"required"`
	Attributes interface{} `json:"attrs"`
}

func initAndRunWebServer() {
	r := gin.New()
	gin.EnableJsonDecoderDisallowUnknownFields()
	r.POST("/images/", addImage)     // Add new Images to the index
	r.GET("/images/*url", findByURL) // Find the most similar image by URL
	r.POST("/find-similar-to-file/", findByFile)
	r.StaticFile("/", "_examples/server/spa.html")
	r.StaticFile("/bootstrap.min.css", "_examples/server/bootstrap.min.css")
	r.MaxMultipartMemory = 8 << 20 // 8 MiB
	err := r.Run(":8080")
	if err != nil {
		log.Fatal(err)
	}
}

func validationError(c *gin.Context, err error) {
	code := http.StatusBadRequest
	if errors.Is(err, imgidx.URIAlreadyExists{}) {
		code = http.StatusConflict
	}
	c.JSON(code, gin.H{"message": err.Error()})
}

func addImage(c *gin.Context) {
	req := AddImageRequest{}
	if err := c.BindJSON(&req); err != nil {
		validationError(c, err)
		return
	}
	if _, err := url.ParseRequestURI(req.Url); err != nil {
		validationError(c, err)
		return
	}
	_, err := imgidx.AddImageUrl(idx, req.Url, req.Attributes)
	if err != nil {
		validationError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "done",
	})
}

func findByURL(c *gin.Context) {
	imgUrl := strings.TrimPrefix(c.Param("url"), "/")
	if _, err := url.ParseRequestURI(imgUrl); err != nil {
		validationError(c, err)
		return
	}
	nearestImgUrl, attrs, dist, err := imgidx.NearestByURL(idx, imgUrl)
	if err != nil {
		validationError(c, err)
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
		validationError(c, err)
		return
	}
	f, err := file.Open()
	if err != nil {
		validationError(c, err)
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
		validationError(c, fmt.Errorf("failed to decode image %w", err))
		return
	}

	nearestImgUrl, attrs, dist, err := idx.Nearest(queryImg)

	if err != nil {
		validationError(c, fmt.Errorf("failed to find similar image : %w", err))
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"url":                nearestImgUrl,
		"additional_details": attrs,
		"distance":           dist,
	})
}

func main() {
	var err error
	idx, err = imgidx.NewPersistentCompositeIndex(8, 8, sqlite.Open("imgidx.db"))
	if err != nil {
		log.Fatal(err)
	}
	initAndRunWebServer()
}
