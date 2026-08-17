package header

import (
	"net/http"
	"unitly/internal/client"
)

const (
	ClientPlatform = "X-Client-Platform"
	Authorization  = "Authorization"
	ContentType    = "Content-Type"
	Accept         = "Accept"
	UserAgent      = "User-Agent"
)

func GetClientPlatform(r *http.Request) client.ClientPlatform {
	return client.GetClientPlatform(r.Header.Get(ClientPlatform))
}

func GetUserAgent(r *http.Request) string {
	return r.Header.Get(UserAgent)
}
