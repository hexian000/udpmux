package main

const (
	salt = "udpmux"
	tag  = "udpmux_server"
)

type Config struct {
	Net        string
	Address    string
	Remote     string
	Channels   []string
	Key        string
	ServerMode bool

	Keepalive int
	Output    string
}
