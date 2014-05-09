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
	client := twitterstream.NewClient(
		*consumerKey,
		*consumerSecret,
		*token,
		*tokenSecret,
	)
	client.GzipCompression = true

	client.OnLine(func(line []byte) {
		var item map[string]interface{}
		if err := json.Unmarshal(line, &item); err != nil {
			log.Fatal("json decode failed:" + err.Error())
		}
		log.Println(item)
	})
	log.Print(client.Userstream("POST", map[string]string{"stringify_friend_ids": "true"}))
}
