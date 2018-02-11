package main

import (
	"fmt"
	"log"

	"github.com/jjeffcaii/mongo-proxy"
	"github.com/jjeffcaii/mongo-proxy/middware"
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
		c1.Use(&middware.SkipIsMaster{}, securityManager)
		// choose mongo host and port
		var mgoHostAndPort = "127.0.0.1:27017"
		// connect backend begin!
		backend := pxmgo.NewStaticBackend(mgoHostAndPort)
		err := backend.Serve(func(c2 pxmgo.Context) {
			pxmgo.Pump(c1, c2)
		})
		if err != nil {
			log.Println("fire backend endpoint failed:", err)
			c1.Close()
		}
	})

}
