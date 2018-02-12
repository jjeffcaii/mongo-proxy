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
	"github.com/jjeffcaii/mongo-proxy/middleware"
)

func main() {
	proxy := pxmgo.NewServer(":27018")
	log.Println("proxy server start")
	// custom your auth validator.
	validator := func(db string) (*middleware.Identifier, error) {
		// use foo/bar to login test db.
		if "test" == db {
			user, passwd := "foo", "bar"
			return &middleware.Identifier{
				Username: user,
				Password: passwd,
			}, nil
		}
		return nil, fmt.Errorf("access deny for db: %s", db)
	}
	// begin serve.
	proxy.Serve(func(c1 pxmgo.Context) {
		// skip is master
		skipIsMaster := middleware.NewSkipIsMaster()
		// create authenticator.
		authenticator := middleware.NewAuthenticator(validator)
		// register frontend context middlewares.
		c1.Use(skipIsMaster, authenticator)
		// wait for auth finish.
		db, ok := authenticator.Wait()
		if !ok {
			log.Println("authenticate failed")
			return
		}
		log.Printf("connect database %s success\n", *db)
		// choose mongo host and port.
		var mongoURI = "127.0.0.1:27017"
		// connect backend begin.
		pxmgo.NewBackend(mongoURI).Serve(func(c2 pxmgo.Context) {
			// just pump frontend and backend.
			pxmgo.Pump(c1, c2)
		})
	})

}

```

## Todo List
 - [x] Remove ugly handlers(onRes,onNext).
 - [ ] Support OP_CMD,OP_CMD_REPLY for mongo cli.
 - [ ] More graceful APIs.