package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"os"
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

	listener, err := net.Listen("tcp", "localhost:8080")
	if err != nil {
		// handle error
	}
	for {
		conn, err := listener.Accept()
		fmt.Println("got connection")
		if err != nil {
		}
		go handleConnection(conn)
	}

}

func handleConnection(conn net.Conn) {
	readBuf := make([]byte, 65536)
	conn.Read(readBuf)
	fmt.Println(string(readBuf))

}

type jsonStruct struct {
	FIELD string `json:"field"`
}

func processClient(conn net.Conn) {
	_, err := io.Copy(os.Stdout, conn)
	if err != nil {
		fmt.Println(err)
	}
	conn.Close()
}

func callDatabase(input string) (string, error) {

	postBody, _ := json.Marshal(map[string]string{
		"field": string(input),
	})

	postBuffer := bytes.NewBuffer(postBody)
	url := dbServiceAddr + "/handleUpload"
	resp, err := http.Post(url, "application/json", postBuffer)
	panicErr(err)

	defer resp.Body.Close()

	var binLink jsonStruct
	body, err := ioutil.ReadAll(resp.Body)
	panicErr(err)
	json.Unmarshal(body, &binLink)

	return binLink.FIELD, nil
}

func panicErr(e error) {
	if e != nil {
		panic(e)
	}
}
