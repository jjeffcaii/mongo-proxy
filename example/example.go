package main

import (
	"fmt"
	"log"

	"github.com/jjeffcaii/mongo-proxy"
)

func main() {
	var port = 27018
	server := pxmgo.NewServer(fmt.Sprintf(":%d", port))
	log.Printf("start proxy server: port=%d\n", port)
	validator := func(db string) (*string, *string, error) {
		if "test" == db {
			user, passwd := "foo", "bar"
			return &user, &passwd, nil
		}
		return nil, nil, fmt.Errorf("access deny for db: %s", db)
	}
	server.Serve(func(c1 pxmgo.Context) {
		// create security manager
		// register frontend context middlewares.
		c1.Use(&pxmgo.SkipIsMaster{}, pxmgo.NewSecurityManager(validator))
		// choose mongo host and port
		var mgoHostAndPort = "127.0.0.1:27017"
		// connect backend begin!
		backend := pxmgo.NewStaticBackend(mgoHostAndPort)
		err := backend.Serve(func(c2 pxmgo.Context) {
			go pxmgo.Pump(c1, c2)
		})
		if err != nil {
			log.Println("fire backend endpoint failed:", err)
			c1.Close()
		}
	})

}
