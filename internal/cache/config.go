package cache

import "fmt"

type CacheConfig struct {
	Host     string
	Username string
	Password string
	Port     string
}

func (c *CacheConfig) GetAddress() string {
	return fmt.Sprintf("%s:%s", c.Host, c.Port)
}
