package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
)

type jsonStruct struct {
	FIELD string `json:"field"`
}

func main() {

	input, err := ioutil.ReadAll(bufio.NewReader(os.Stdin))
	panicErr(err)

	if len(input) > 2097152 {
		fmt.Println("Error: Maximum paste is 2MB")
		os.Exit(1)
	}

	postBody, _ := json.Marshal(map[string]string{
		"field": string(input),
	})

	postBuffer := bytes.NewBuffer(postBody)
	url := "http://127.0.0.1:8080/handleUpload"
	resp, err := http.Post(url, "application/json", postBuffer)
	panicErr(err)

	defer resp.Body.Close()

	var binLink jsonStruct
	body, err := ioutil.ReadAll(resp.Body)
	panicErr(err)
	json.Unmarshal(body, &binLink)

	fmt.Println(binLink.FIELD) //t

}

func panicErr(e error) {
	if e != nil {
		panic(e)
	}
}
