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
	"time"
)

var dbserviceAddr string

// crash if db-service is not responding or env var DBSERVICEADDR is empty
func init() {
	time.Sleep(time.Second * 4)

	dbserviceAddr = os.Getenv("DBSERVICEADDR")
	_, err := callDbService("")
	panicErr(err)
}

func main() {
	url := "127.0.0.1:6262"
	listener, err := net.Listen("tcp", url)
	panicErr(err)
	fmt.Printf("listening on %v\n", url)

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println(err)
		}
		go processClient(conn)
	}
}

type PasteJson struct {
	Content string `json:"Content"`
	Hash    string `json:"Hash"`
}

func processClient(conn net.Conn) {
	fmt.Printf("got connecton from %v\n", conn.RemoteAddr().String())
	defer conn.Close()
	buf := new(bytes.Buffer)
	n, err := io.Copy(buf, conn)
	if err != nil {
		fmt.Println(err)
	}

	if n > 1048576*2 { // 2MB
		conn.Write([]byte("Error: content exceeds 2MB\n"))
		conn.Close()
		return
	}

	content := buf.String()

	fmt.Println(content) //t

	hash, err := callDbService(content)
	if err == nil {
		conn.Write([]byte(fmt.Sprintf("paste.cat-z.xyz/%v\n", hash)))
	} else {
		conn.Write([]byte("internal error\n"))
	}
}

func callDbService(content string) (string, error) {

	values := map[string]string{"Content": content}
	json_data, err := json.Marshal(values)
	panicErr(err)

	resp, err := http.Post("http://"+dbserviceAddr+"/process", "application/json", bytes.NewBuffer(json_data))
	panicErr(err)

	defer resp.Body.Close()

	var jsonResp PasteJson
	body, err := ioutil.ReadAll(resp.Body)
	panicErr(err)
	json.Unmarshal(body, &jsonResp)

	fmt.Println(jsonResp)

	return jsonResp.Content, nil
}

func panicErr(e error) {
	if e != nil {
		panic(e)
	}
}
