package main

import (
	"crypto/md5"
	"database/sql"
	"encoding/hex"
	"fmt"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/go-sql-driver/mysql"
)

var db *sql.DB

// can we bind json to this?
type Paste struct {
	Hash         string
	Content      string
	Created      string
	LastAccessed string
}

// or use this instead of the above?
type PasteJson struct {
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

func logFatalErr(e error) {
	if e != nil {
		fmt.Println("FATAL:")
		log.Fatal(e)
	}
}

func setupDb() {

	cfg := mysql.Config{
		DBName:               "shellbin",
		Net:                  "tcp",
		Addr:                 tryGetEnv("DBADDR"),
		User:                 tryGetEnv("DBUSER"),
		Passwd:               tryGetEnv("DBPASS"),
		AllowNativePasswords: true,
	}

	var err error
	db, err = sql.Open("mysql", cfg.FormatDSN())
	logFatalErr(err)

	pingErr := db.Ping()
	logFatalErr(pingErr)

}

func main() {

	setupDb()

	router := gin.Default()
	router.POST("/process", process)
	router.Run("0.0.0.0:7272")

}

func process(c *gin.Context) {

	var incomingJson PasteJson
	if err := c.BindJSON(&incomingJson); err != nil {
		log.Fatalf("%v\n", err)
	}
	content := incomingJson.Content
	fmt.Println(content) //t

	hashToSend, err := callDb(content)
	if err == nil {
		c.IndentedJSON(200, PasteJson{
			Content: hashToSend,
		})
	} else {
		c.IndentedJSON(503, PasteJson{
			Content: "server error",
		})
	}
}

func getPasteByHash(hash string) (Paste, error) {
	var paste Paste

	row := db.QueryRow("SELECT hash FROM pastes WHERE hash = ?", hash)
	if err := row.Scan(&paste.Hash, &paste.Content, &paste.Created, &paste.LastAccessed); err != nil {
		if err == sql.ErrNoRows {
			return paste, err
		}
		return paste, fmt.Errorf("getPasteByHash %v: %v", hash, err)
	}
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
func callDb(content string) (string, error) {

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
