package client

const (
	WebPlatform     ClientPlatform = "WebClient"
	AndroidPlatform ClientPlatform = "AndroidClient"
	UnknownPlatform ClientPlatform = "Unknown"
)

type ClientPlatform string
