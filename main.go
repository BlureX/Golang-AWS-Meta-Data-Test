package main

import (
	"context"
	"errors"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"strings"
	"time"
)

var block bool
var c *http.Client

func bd(ctx context.Context, network, addr string) (net.Conn, error) {
	conn, err := net.DialTimeout(network, addr, time.Second*5)
	if err != nil {
		return conn, err
	}
	log.Println("MADE REQUEST TO " + addr + " WHICH RESOLVED TO IP ADDR: " + conn.RemoteAddr().String())
	if strings.Contains(conn.RemoteAddr().String(), "169.254.169.254") && block {
		log.Println("OK The IP ADDR CONTAINED META DATA URL SO WE BLOCKING")
		return conn, errors.New("BLACKLIST WORKING")
	}
	return conn, err
}

func testMetaDataAccess(w http.ResponseWriter, r *http.Request) {
	tr := &http.Transport{DialContext: bd}
	c = &http.Client{Timeout: 5 * time.Second, Transport: tr}
	urlList := map[string]string{
		"http://361.596.361.596/":                                        "Hexadecimal Val",
		"http://169.254.169.254/latest/meta-data/instance-id":            "Ipv4 addr value",
		"http://instance-data/latest/meta-data/instance-id":              "DNS val",
		"http://[0:0:0:0:0:ffff:a9fe:a9fe]/latest/meta-data/instance-id": "IPV6 val",
		"http://0251.0376.0251.0376":                                     "Octal IP addr",
	}
	for url, val := range urlList {
		log.Println("-----------------------------------------")
		log.Printf("OK MAKING A REQUEST TO URL: %s which represents aws meta data in %s \n", url, val)
		log.Println("Going to make a request with blacklist off which gets me the following result:")
		makeReq(false, url)
		log.Println("   ")
		log.Println("Going to make a request with blacklist on which gets me the following result:")
		makeReq(true, url)
		log.Println("-----------------------------------------\n")
	}
}

func makeReq(shouldBlock bool, url string) {
	block = shouldBlock
	blockString := "BLOCKING"
	if !block {
		blockString = "NOT BLOCKING"
	}
	resp, err := c.Get(url)
	if err != nil {
		log.Printf("GOT ERROR WITH MAKING A REQUEST TO URL WHILE %s, GOT ERROR: %s", blockString, err.Error())
		return
	}
	if shouldBlock {
		log.Println("Ok you weren't meant to make it this far. Ip blocking didn't work for this URL: ", url)
		return
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println("OK THIS WASN'T SUPPOSE TO HAPPEN BUT CANT READ BODY. Got error: ", err)
		return
	}
	log.Println("RESP BODY IS:" + string(body))
}

func main() {
	http.HandleFunc("/", testMetaDataAccess) // set router
	log.Println("STARTING")
	err := http.ListenAndServe(":9090", nil) // set listen port
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
