package sessionsvc

import (
	"log"
	"unitly/internal/domain"

	"github.com/mileusna/useragent"
)

const pwa = "PWA"

// DeviceInfo holds the device fields persisted on a Session and shown on
// the sessions page.
type DeviceInfo struct {
	DeviceType  domain.DeviceType
	OSName      string
	OSVersion   string
	ClientLabel string
}

func resolveDevice(userAgent string) DeviceInfo {
	return parseUserAgent(userAgent)
}

func parseUserAgent(rawUA string) DeviceInfo {
	if rawUA == "" {
		d := DeviceInfo{DeviceType: domain.DeviceTypeUnknown}
		return d
	}

	ua := useragent.Parse(rawUA)
	log.Printf("%+v\n", ua)

	d := DeviceInfo{
		OSName:    ua.OS,
		OSVersion: ua.OSVersion,
	}

	switch {
	case ua.Bot:
		d.DeviceType = domain.DeviceTypeBot
	case ua.IsAndroid():
		d.DeviceType = domain.DeviceTypeAndroid
	case ua.IsIOS():
		d.DeviceType = domain.DeviceTypeIOS
	case ua.Desktop:
		d.DeviceType = domain.DeviceTypeDesktop
	default:
		d.DeviceType = domain.DeviceTypeUnknown
	}

	d.ClientLabel = ua.Name

	return d
}
