// Config is put into a different package to prevent cyclic imports in case
// it is needed in several locations

package config

import "time"

type Config struct {
	Period time.Duration `config:"period"`
	ApiKey string `config:"cloudstackkey"`
	ApiSecret string `config:"cloudstacksecret"`
	ApiUrl string `config:"cloudstackurl"`
}
//
var DefaultConfig = Config{
	Period: 1 * time.Second,
	ApiKey: "test",
	ApiSecret: "test",
	ApiUrl: "http://localhost:8080/client/api",
}

