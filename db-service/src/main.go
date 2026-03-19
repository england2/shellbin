package main

import (
	"crypto/md5"
	"database/sql"
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-sql-driver/mysql"
)

var db *sql.DB

type Paste struct {
	Hash         string `json:"Hash"`
	Content      string `json:"Content"`
	Created      string `json:"Created"`
	LastAccessed string `json:"LastAccessed"`
}

func tryGetEnv(key string) string {
	val, err := os.LookupEnv(key)
	if !err {
		log.Panicf("env var %v is not set", key)
	}
	return val
}

func getEnvOrDefault(key string, fallback string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}
	return fallback
}

func setupDb() {

	cfg := mysql.Config{
		DBName:               getEnvOrDefault("DATABASE_NAME", "shellbin"),
		Net:                  "tcp",
		Addr:                 tryGetEnv("DBADDR"),
		User:                 tryGetEnv("DBUSER"),
		Passwd:               tryGetEnv("DBPASS"),
		AllowNativePasswords: true,
	}

	var err error
	db, err = sql.Open("mysql", cfg.FormatDSN())

	if err != nil {
		setupDb()
	}

}

func servePaste(c *gin.Context) {

	var incomingJson Paste
	if err := c.BindJSON(&incomingJson); err != nil {
		log.Printf("%v\n", err)
	}

	hash := strings.Replace(incomingJson.Hash, "/", "", -1)
	fmt.Printf("(in servePaste) hash: %v\n", hash)

	paste, err := getPasteByHash(hash)
	fmt.Printf("paste.Hash: %v\n", paste.Hash)

	if err == nil {
		fmt.Printf("(in servePaste) 200 paste found ")
		fmt.Printf("paste: %v\n", paste)
		c.IndentedJSON(200, paste)
	} else {
		fmt.Printf("(in servePaste) 404 paste not found ")
		fmt.Println(err)
		c.IndentedJSON(404, Paste{})
	}

}

// POST endpoint receiving paste content. writes content of paste to the database
func processInput(c *gin.Context) {

	var incomingJson Paste
	if err := c.BindJSON(&incomingJson); err != nil {
		log.Printf("%v\n", err)
	}
	content := incomingJson.Content
	fmt.Println(content) //t

	hashToSend, err := insertPaste(content)
	if err == nil {
		c.IndentedJSON(200, Paste{
			Hash: hashToSend,
		})
	} else {
		c.IndentedJSON(503, Paste{})
	}
}

func getPasteByHash(hash string) (Paste, error) {

	var paste Paste

	row := db.QueryRow("SELECT * FROM pastes WHERE hash = ?", hash)
	if err := row.Scan(&paste.Hash, &paste.Content, &paste.Created, &paste.LastAccessed); err != nil {
		if err == sql.ErrNoRows {
			// expected error
			return paste, err
		} else {
			// unexpected error
			fmt.Println(err)
			return paste, fmt.Errorf("getPasteByHash: %v", err)
		}
	}

	fmt.Println("returning paste", paste) //t
	return paste, nil
}

func hashExists(contentHash string) bool {

	var paste Paste

	row := db.QueryRow("SELECT hash FROM pastes WHERE hash = ?", contentHash)
	err := row.Scan(&paste.Hash)
	if err == sql.ErrNoRows {
		fmt.Println("hashExists: returning false") //t
		return false
	}

	fmt.Println("hashExists: returning true") //t
	return true
}

// timeout after 5 seconds, return generic error to user
func insertPaste(content string) (string, error) {

	contentHash := getMD5Hash(content)

	if hashExists(contentHash) {
		fmt.Println("callDb: hash exists. returning hash.") //t
		return contentHash, nil                             // maybe this should return a full paste???
	}

	fmt.Println("callDb: hash does not exist. creating paste.") //t
	err := addPaste(Paste{
		Hash:         contentHash,
		Content:      content,
		Created:      "curdate()",
		LastAccessed: "curdate()",
	})

	if err != nil {
		return "", err
	}

	return contentHash, nil

}

func addPaste(paste Paste) error {

	phrase := fmt.Sprintf("INSERT INTO pastes (hash, content, submission_date, last_used) VALUES (\"%v\", \"%v\", %v, %v);\n",
		paste.Hash, paste.Content, "curdate()", "curdate()")
	result, err := db.Exec(phrase)

	if err != nil {
		return fmt.Errorf("phrase: %v \n sql error: %v", phrase, err)
	}

	fmt.Println(result.RowsAffected())

	return nil
}

func getMD5Hash(text string) string {
	hash := md5.Sum([]byte(text))
	return hex.EncodeToString(hash[:4])
}

func main() {

	setupDb()

	router := gin.Default()
	router.POST("/processInput", processInput)
	router.POST("/servePaste", servePaste)
	// TODO helm env vars for port
	router.Run("0.0.0.0:7272")

}
