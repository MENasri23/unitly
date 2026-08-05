package db

import "strings"

type Args map[string]string

type DBConfig struct {
	Name     string
	Host     string
	Port     string
	Username string
	Password string
	Driver   string
	Args     Args
}

func (c *DBConfig) GetURL() string {
	var sb strings.Builder

	sb.WriteString(c.Driver)
	sb.WriteString("://")

	if len(c.Username) > 0 || len(c.Password) > 0 {
		sb.WriteString(c.Username)
		if len(c.Password) > 0 {
			sb.WriteRune(':')
			sb.WriteString(c.Password)
		}
		sb.WriteRune('@')
	}

	sb.WriteString(c.Host)

	if len(c.Port) > 0 {
		sb.WriteRune(':')
		sb.WriteString(c.Port)
	}

	sb.WriteRune('/')
	sb.WriteString(c.Name)

	if len(c.Args) > 0 {
		sb.WriteRune('?')
		first := true
		for key, val := range c.Args {
			if !first {
				sb.WriteRune('&')
			}

			sb.WriteString(key)
			sb.WriteRune('=')
			sb.WriteString(val)

			if first {
				first = false
			}
		}
	}

	return sb.String()
}
