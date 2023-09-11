package main

import (
	"embed"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func page_paste(c *gin.Context) {

	fmt.Println("in page_timeNow")

	path := c.Param("path")
	c.HTML(http.StatusOK, "paste.html", gin.H{
		"time": time.Now(),
		"path": path,
	})

}

func page_index(c *gin.Context) {

	fmt.Println("in page_index")

	message := c.PostForm("message")

	fmt.Printf("message: %v\n", message)

	c.HTML(http.StatusOK, "index.html", "")

}

type Paste struct {
	Content string
}

func bookNewPostHandler(c *gin.Context) {

	book := &Paste{}
	if err := c.Bind(book); err != nil {
		return
	}

	fmt.Println(book)

	c.Redirect(http.StatusFound, "/books/")
}

//go:embed assets/*
var fullAssets embed.FS

func handleLogin(c *gin.Context) {
	emailValue := c.PostForm("email")
	passwordValue := c.PostForm("password")

	fmt.Println(emailValue)
	fmt.Println(passwordValue)

	c.Redirect(302, "/paste/"+emailValue)
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
	g1.GET("/paste/:path", page_paste)

	g1.GET("/", page_index)

	router.POST("/login", handleLogin)

	router.Run(":8080")

}

func main() {
	startServer()
}

type FormStruct struct {
	Content string
}
