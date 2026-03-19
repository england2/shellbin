package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"time"
)

var dbserviceAddr string

func main() {

	dbserviceAddr = os.Getenv("DBSERVICEADDR")
	url := "0.0.0.0:6262"

	var listener net.Listener
	for {
		var err error
		listener, err = net.Listen("tcp", url)
		if err == nil {
			break
		}
		time.Sleep(time.Second * 2)
	}

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

	hash, err := addContentDb(content)
	if err == nil {
		conn.Write([]byte(fmt.Sprintf("paste.cat-z.xyz/%v\n", hash)))
	} else {
		conn.Write([]byte("internal error\n"))
	}
}

func addContentDb(content string) (string, error) {

	values := map[string]string{"Content": content}
	json_data, err := json.Marshal(values)
	panicErr(err)

	resp, err := http.Post("http://"+dbserviceAddr+"/processInput",
		"application/json", bytes.NewBuffer(json_data))
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("db-service returned status %d", resp.StatusCode)
	}

	var jsonResp PasteJson
	body, err := io.ReadAll(resp.Body)
	panicErr(err)
	json.Unmarshal(body, &jsonResp)

	fmt.Println(jsonResp)

	return jsonResp.Hash, nil
}

func panicErr(e error) {
	if e != nil {
		panic(e)
	}
}
