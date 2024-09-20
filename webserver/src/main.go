package main

import (
	"bytes"
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

//go:embed assets/*
var fullAssets embed.FS

// helper
func unmarshalPaste(resp http.Response) Paste {

	var jsonResp Paste
	body, err := ioutil.ReadAll(resp.Body)
	panicErr(err)
	json.Unmarshal(body, &jsonResp)

	fmt.Println(jsonResp)
	return jsonResp
}

func routePaste(c *gin.Context) {

	fmt.Println("in routePaste")

	path := c.Param("path")

	resp := getContentDb(path)

	if resp.Status == "200" {

		paste := unmarshalPaste(*resp)

		c.HTML(http.StatusOK, "paste.html", gin.H{
			"content": paste.Content,
		})
	} else if resp.Status == "404" {
		c.HTML(http.StatusOK, "paste.html", "")
	}

}

func routeIndex(c *gin.Context) {
	c.HTML(http.StatusOK, "index.html", "")
}

func handleSubmit(c *gin.Context) {
	content := c.PostForm("content")

	fmt.Println(content)

	resp := addContentDb(content)

	if resp.Status == "200" {
		c.Redirect(302, "/paste/"+content)
	} else if resp.Status == "504" {
		fmt.Println("db service down")
		c.HTML(http.StatusBadRequest, "erorr.html", "")
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
			Status: "504",
		}
	}

	defer resp.Body.Close()

	return resp

}

func getContentDb(hash string) *http.Response {

	json_data, err := json.Marshal(map[string]string{
		"Hash": hash})
	panicErr(err)

	resp, err := http.Post("http://"+dbserviceAddr+"/servePaste",
		"application/json", bytes.NewBuffer(json_data))
	// TODO recover from this error, reoute to page 404
	panicErr(err)

	defer resp.Body.Close()

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
	g1.GET("/paste/:path", routePaste)

	g1.GET("/", routeIndex)

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
