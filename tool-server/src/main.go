package main

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

var dbServiceAddr string

func init() {
	var hasDbServiceAddrEnv bool
	dbServiceAddr, hasDbServiceAddrEnv = os.LookupEnv("DBSERVICEADDR")
	if !hasDbServiceAddrEnv {
		panic("Env var DBSERVICEADDR is not set")
	}
}

func main() {

	router := gin.Default()

	router.POST("/handleUpload", handleUpload)

	router.Run("localhost:6262")
}

type jsonStruct struct {
	FIELD string `json:"field"`
}

func handleUpload(c *gin.Context) {

	var incomingJson jsonStruct
	if err := c.BindJSON(&incomingJson); err != nil {
		log.Fatalf("%v\n", err)
	}

	c.IndentedJSON(200, jsonStruct{
		FIELD: getMD5Hash(incomingJson.FIELD),
	})

}

func callDatabase(input string) (string, error) {

	postBody, _ := json.Marshal(map[string]string{
		"field": string(input),
	})

	postBuffer := bytes.NewBuffer(postBody)
	url := dbServiceAddr + "/dbCallTool"
	resp, err := http.Post(url, "application/json", postBuffer)
	panicErr(err)

	defer resp.Body.Close()

	var binLink jsonStruct
	body, err := ioutil.ReadAll(resp.Body)
	panicErr(err)
	json.Unmarshal(body, &binLink)

	return binLink.FIELD, nil
}

// simulates the db return
func getMD5Hash(text string) string {
	hash := md5.Sum([]byte(text))
	return hex.EncodeToString(hash[:])
}

func panicErr(e error) {
	if e != nil {
		panic(e)
	}
}
