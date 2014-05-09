# go-twitterstream

## usage

```go
package main

import (
	"encoding/json"
	"github.com/soh335/go-twitterstream"
	"log"
)

func main() {
	client := twitterstream.NewClient(
		consumerKey,
		consumerSecret,
		token,
		tokenSecret,
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
```
