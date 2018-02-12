package main

import (
	"fmt"
	"log"

	"github.com/jjeffcaii/mongo-proxy"
	"github.com/jjeffcaii/mongo-proxy/middware"
)

func main() {
	server := pxmgo.NewServer(":27018")
	log.Println("proxy server start")
	validator := func(db string) (*pxmgo.Identifier, error) {
		if "test" == db {
			user, passwd := "foo", "bar"
			return &pxmgo.Identifier{
				Username: user,
				Password: passwd,
			}, nil
		}
		return nil, fmt.Errorf("access deny for db: %s", db)
	}
	server.Serve(func(c1 pxmgo.Context) {
		// create authenticator.
		authenticator := middware.NewAuthenticator(validator)
		// register frontend context middlewares.
		c1.Use(&middware.SkipIsMaster{}, authenticator)
		// wait for auth finish.
		db, ok := authenticator.Wait()
		if !ok {
			return
		}
		log.Printf("connect database %s success\n", *db)
		// choose mongo host and port
		var mgoHostAndPort = "127.0.0.1:27017"
		// connect backend begin!
		backend := pxmgo.NewBackend(mgoHostAndPort)
		err := backend.Serve(func(c2 pxmgo.Context) {
			go pxmgo.Pump(c1, c2)
		})
		if err != nil {
			log.Println("fire backend endpoint failed:", err)
			c1.Close()
		}
	})

}
