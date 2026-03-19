package main

import (
	"bytes"
	"embed"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

//go:embed assets/*
var fullAssets embed.FS

// helper
func unmarshalPaste(resp *http.Response) (Paste, error) {
	var jsonResp Paste
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return jsonResp, err
	}
	if err := json.Unmarshal(body, &jsonResp); err != nil {
		return jsonResp, err
	}

	fmt.Println(jsonResp)
	return jsonResp, nil
}

func routePaste(c *gin.Context) {

	fmt.Println("simple trivial change")
	fmt.Println("in routePaste")

	path := c.Param("path")

	resp := getContentDb(path)

	if resp.StatusCode == http.StatusOK {
		paste, err := unmarshalPaste(resp)
		if err != nil {
			c.HTML(http.StatusBadGateway, "error.html", "")
			return
		}

		c.HTML(http.StatusOK, "paste.html", gin.H{
			"content": paste.Content,
		})
	} else if resp.StatusCode == http.StatusNotFound {
		c.HTML(http.StatusNotFound, "error.html", "")
	} else {
		c.HTML(http.StatusBadGateway, "error.html", "")
	}

}

func routeIndex(c *gin.Context) {
	c.HTML(http.StatusOK, "index.html", "")
}

func handleSubmit(c *gin.Context) {
	content := c.PostForm("content")

	fmt.Println(content)

	resp := addContentDb(content)

	if resp.StatusCode == http.StatusOK {
		paste, err := unmarshalPaste(resp)
		if err != nil {
			c.HTML(http.StatusBadRequest, "error.html", "")
			return
		}
		c.Redirect(http.StatusFound, "/"+paste.Hash)
	} else if resp.StatusCode == http.StatusGatewayTimeout {
		fmt.Println("db service down")
		c.HTML(http.StatusBadRequest, "error.html", "")
	} else {
		c.HTML(http.StatusBadRequest, "error.html", "")
	}
}

func addContentDb(content string) *http.Response {

	json_data, err := json.Marshal(map[string]string{
		"Content": content})
	panicErr(err)

	resp, err := http.Post("http://"+dbserviceAddr+"/processInput",
		"application/json", bytes.NewBuffer(json_data))

	if err != nil {
		return &http.Response{
			Status:     "504 Gateway Timeout",
			StatusCode: http.StatusGatewayTimeout,
		}
	}

	return resp

}

func getContentDb(hash string) *http.Response {

	json_data, err := json.Marshal(map[string]string{
		"Hash": hash})
	panicErr(err)

	resp, err := http.Post("http://"+dbserviceAddr+"/servePaste",
		"application/json", bytes.NewBuffer(json_data))
	if err != nil {
		return &http.Response{
			Status:     "504 Gateway Timeout",
			StatusCode: http.StatusGatewayTimeout,
		}
	}

	return resp
}

type Paste struct {
	Hash         string `json:"Hash"`
	Content      string `json:"Content"`
	Created      string `json:"Created"`
	LastAccessed string `json:"LastAccessed"`
}

func startServer() {

	router := gin.Default()

	assets, err := fs.Sub(fullAssets, "assets")
	router.LoadHTMLGlob("assets/templates/*")

	if err != nil {
		log.Fatal("Failed to load assets", err)
	}

	router.StaticFS("/assets", http.FS(assets))

	g1 := router.Group("/")
	g1.GET("/", routeIndex)
	g1.GET("/:path", routePaste)
	g1.GET("/paste/:path", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/"+c.Param("path"))
	})

	router.POST("/submit", handleSubmit)

	router.Run(hostAddr)

}

var dbserviceAddr string
var hostAddr string

func main() {
	dbserviceAddr = os.Getenv("DBSERVICEADDR")
	hostAddr = os.Getenv("HOSTADDR")

	startServer()
}

func panicErr(e error) {
	if e != nil {
		panic(e)
	}
}
