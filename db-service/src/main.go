package main

import (
	"crypto/md5"
	"database/sql"
	"encoding/hex"
	"fmt"
	"log"
	"os"

	"github.com/go-sql-driver/mysql"
)

var db *sql.DB

type Paste struct {
	Hash         string
	Content      string
	Created      string
	LastAccessed string
}

func tryGetEnv(key string) string {
	val, err := os.LookupEnv(key)
	if !err {
		log.Panicf("env var %v is not set", key)
	}
	return val
}

func setup() {

	cfg := mysql.Config{
		DBName:               "shellbin", // TODO hardcoded. must be synced with mysql-chart
		Net:                  "tcp",
		User:                 tryGetEnv("DBUSER"),
		Passwd:               tryGetEnv("DBPASS"),
		Addr:                 tryGetEnv("DBSERVICEADDR"),
		AllowNativePasswords: true,
	}

	var err error
	db, err = sql.Open("mysql", cfg.FormatDSN())
	if err != nil {
		log.Fatal(err)
	}

	// testing ???
	pingErr := db.Ping()
	if pingErr != nil {
		log.Fatal(pingErr)
	}
	fmt.Println("Connected!")

}

func main() {

	setup()

	alb, err := pasteByHash("abc")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Paste found: %v\n", alb)

	if err != nil {
		log.Fatal(err)
	}

	newPasteHash, err := addPaste("aaa")
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println(newPasteHash)

}

func pasteByHash(hash string) (Paste, error) {
	var paste Paste

	row := db.QueryRow("SELECT content FROM pastes WHERE hash = ?", hash)
	if err := row.Scan(&paste.Hash, &paste.Content, &paste.Content, &paste.LastAccessed); err != nil {
		if err == sql.ErrNoRows {
			return paste, fmt.Errorf("albumsById %v: no such paste", hash)
		}
		return paste, fmt.Errorf("contentByHash %v: %v", hash, err)
	}
	return paste, nil
}

func addPaste(content string) (string, error) {

	contentHash := getMD5Hash(content)

	paste := Paste{
		Hash:         contentHash,
		Content:      content,
		Created:      "curdate()",
		LastAccessed: "curdate()",
	}

	// TODO ?:
	// if tableHasHash(contentHash) ...

	result, err := db.Exec("INSERT INTO album (hash, content, submission_date, last_used,) VALUES (?, ?, ?, ?)",
		paste.Hash, paste.Content, "curdate()", "")
	if err != nil {
		return "", fmt.Errorf("addPaste: %v", err)
	}

	fmt.Println(result.RowsAffected())

	return contentHash, nil
}

func getMD5Hash(text string) string {
	hash := md5.Sum([]byte(text))
	return hex.EncodeToString(hash[:10])
}

// package main

// import (
// 	"crypto/md5"
// 	"encoding/hex"
// 	"log"

// 	"github.com/gin-gonic/gin"
// )

// func main() {

// 	router := gin.Default()

// 	router.POST("/upload", upload)

// 	router.Run("localhost:7272")
// }

// type jsonStruct struct {
// 	FIELD string `json:"field"`
// }

// func upload(c *gin.Context) {

// 	var incomingJson jsonStruct
// 	if err := c.BindJSON(&incomingJson); err != nil {
// 		log.Fatalf("%v\n", err)
// 	}
// 	content := incomingJson.FIELD

// 	var hashToSend string

// 	if dbHasContent(content) {
// 		hashToSend = getMD5Hash(content)
// 	} else {
// 		dbNewEntry(content)
// 	}

// 	c.IndentedJSON(200, jsonStruct{
// 		FIELD: hashToSend,
// 	})

// }
