package main

import (
	"net/http"
	"realtimechat/router"
)

func main() {

	routes := router.NewRouter()

	server := &http.Server{
		Addr:    ":8888",
		Handler: routes,
	}

	err := server.ListenAndServe()
	if err != nil {
		panic(err)
	}

}
