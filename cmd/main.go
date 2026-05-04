package main

import (
	"net/http"
	"unitly/bootstrap"
)

func main() {
	app := bootstrap.CreateAppliaction()
	app.Run(http.NewServeMux())
}