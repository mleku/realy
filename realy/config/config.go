package config

type C struct {
	AppName        string   `json:"app_name" doc:"application name" default:"realy"`
	AllowList      []string `json:"allow_list" doc:"List of allowed IP addresses"`
	BlockList      []string `json:"block_list" doc:"list of IP addresses that will be ignored"`
	Admins         []string `json:"admins" doc:"list of npubs that have admin access"`
	Owners         []string `json:"owners" doc:"list of owner npubs whose follow lists set the whitelisted users and enables auth implicitly for all writes"`
	AuthRequired   bool     `json:"auth_required" doc:"authentication is required for read and write" default:"false"`
	PublicReadable bool     `json:"public_readable" doc:"authentication is relaxed for read except privileged events" default:"false"`
	LogLevel       string   `json:"log_level" doc:"Log level" doc:"info"`
	DBLogLevel     string   `json:"db_log_level" default:"info" doc:"database log level"`
	LogTimestamp   bool     `json:"log_timestamp" default:"false" doc:"print log timestamp"`
}
