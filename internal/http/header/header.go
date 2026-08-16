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
)

func GetClientPlatform(r *http.Request) client.ClientPlatform {
	return client.GetClientPlatform(r.Header.Get(ClientPlatform))
}
