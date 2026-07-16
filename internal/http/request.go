package http

import (
	"net/http"
	"unitly/internal/client"
	"unitly/internal/http/header"
)

func IsWebPlatform(r *http.Request) bool {
	return header.GetClientPlatform(r) == client.WebPlatform
}
