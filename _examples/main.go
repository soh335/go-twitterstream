package main

import (
	"encoding/json"
	"flag"
	"github.com/soh335/go-twitterstream"
	"log"
)

var (
	consumerKey    = flag.String("consumerKey", "", "consumerKey")
	consumerSecret = flag.String("consumerSecret", "", "consumerSecret")
	token          = flag.String("token", "", "token")
	tokenSecret    = flag.String("tokenSecret", "", "tokenSecret")
)

func main() {
	flag.Parse()

	client := &twitterstream.Client{
		ConsumerKey:     *consumerKey,
		ConsumerSecret:  *consumerSecret,
		Token:           *token,
		TokenSecret:     *tokenSecret,
		GzipCompression: true,
	}

	conn, err := client.Userstream("POST", map[string]string{"stringify_friend_ids": "true"})

	if err != nil {
		log.Fatal(err)
	}

	defer conn.Close()

	for {
		line, err := conn.Next()
		if err != nil {
			log.Fatal(err)
		}
		var item map[string]interface{}
		if err := json.Unmarshal(line, &item); err != nil {
			log.Fatal("json decode failed:" + err.Error())
		}
		log.Println(item)
	}
}
