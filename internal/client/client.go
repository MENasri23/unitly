package client

const (
	WebPlatform     ClientPlatform = "WebClient"
	AndroidPlatform ClientPlatform = "AndroidClient"
	UnknownPlatform ClientPlatform = "Unknown"
)

type ClientPlatform string

func GetClientPlatform(platformStr string) ClientPlatform {
	p := ClientPlatform(platformStr)
	switch p {
	case WebPlatform, AndroidPlatform:
		return p
	default:
		return UnknownPlatform
	}
}
