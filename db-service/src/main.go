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
	router.POST("/processInput", processInput)
	router.POST("/servePaste", servePaste)
	// TODO helm env vars for port
	router.Run("0.0.0.0:7272")

}

func servePaste(c *gin.Context) {

	var incomingJson PasteJson
	if err := c.BindJSON(&incomingJson); err != nil {
		log.Printf("%v\n", err)
	}
	hash := incomingJson.Hash

	// TODO make go awk helper library and use it here:
	//  hash := incomingJson.Hash.goawkHelper("'/hash/' { gsub("/"",""); print $2 }")
	hashArr := strings.Split(hash, "")
	if hashArr[0] == "/" {
		hash = strings.Join(hashArr[1:], "")
	}

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
		c.IndentedJSON(404, PasteJson{})
	}

}

// POST endpoint receiving paste content. writes content of paste to the database
func processInput(c *gin.Context) {

	var incomingJson PasteJson
	if err := c.BindJSON(&incomingJson); err != nil {
		log.Printf("%v\n", err)
	}
	content := incomingJson.Content
	fmt.Println(content) //t

	hashToSend, err := insertPaste(content)
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
