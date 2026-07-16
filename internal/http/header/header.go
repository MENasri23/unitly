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
	p := client.ClientPlatform(r.Header.Get(ClientPlatform))
	switch p {
	case client.WebPlatform:
	case client.AndroidPlatform:
		return p
	}
	return client.UnknownPlatform
}
