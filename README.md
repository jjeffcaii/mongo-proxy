# mongo-proxy ( STILL WORKING ON IT!!! DO NOT USE IT IN PRODUCTION!!! )
an easy toolkit used to create your own proxy server for MongoDB.

## Code Example

> The following codes are very ugly. I'll try to optimize them next time.

```go
package main

import (
	"fmt"
	"log"

	"github.com/jjeffcaii/mongo-proxy"
	"github.com/jjeffcaii/mongo-proxy/middware"
	"github.com/jjeffcaii/mongo-proxy/protocol"
)

func main() {
	var port = 27018
	server := pxmgo.NewServer(fmt.Sprintf(":%d", port))
	log.Printf("start proxy server: port=%d\n", port)
	server.Serve(func(c1 pxmgo.Context) {
		// create security manager
		securityManager := middware.NewSecurityManager(func(database *string) (*string, *string, error) {
			if "test" == *database {
				user, passwd := "foo", "bar"
				return &user, &passwd, nil
			}
			return nil, nil, fmt.Errorf("access deny for db: %s", *database)
		})
		// register frontend context middlewares.
		c1.Use(middware.SkipIsMaster, protocol.OpCodeQuery)
		c1.Use(securityManager.AsMiddware())
		var db *string
		// pick first message after security passed.
		var first protocol.Message
		for {
			msg, _ := c1.Next()
			if v, ok := securityManager.Ok(); ok {
				first = msg
				db = v
				break
			}
		}
		log.Printf("security verify success: database=%s.\n", *db)
		// choose mongo host and port
		var mgoHostAndPort = "127.0.0.1:27017"
		// connect backend begin!
		backend := pxmgo.NewStaticBackend(mgoHostAndPort)
		err := backend.Serve(func(c2 pxmgo.Context) {
			if bs, err := first.Encode(); err == nil {
				c2.Send(bs)
			}
			pxmgo.Pump(c1, c2)
		})
		if err != nil {
			log.Println("fire backend endpoint failed:", err)
			c1.Close()
		}
	})

}

```

## Todo List
 - [ ] Remove ugly handlers(onRes,onNext).
 - [ ] Support OP_CMD,OP_CMD_REPLY for mongo cli.
 - [ ] More graceful APIs.