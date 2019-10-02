package main

import (
	"crypto/sha256"
	"encoding/json"
	"flag"
	"golang.org/x/crypto/chacha20poly1305"
	"golang.org/x/crypto/pbkdf2"
	"io/ioutil"
	"log"
	"net"
	"os"
	"time"
	"udpmux"
)

func prepare() *udpmux.Instance {
	var flagHelp bool
	var flagConfigFile string
	flag.BoolVar(&flagHelp, "h", false, "show usage")
	flag.StringVar(&flagConfigFile, "c", "", "use config file")
	config := Config{
		Address:    ":2333",
		ServerMode: false,
		Net:        "udp",
	}
	flag.StringVar(&config.Output, "o", config.Output, "output file in tsv format")
	flag.Parse()
	if flagHelp {
		flag.Usage()
		os.Exit(0)
	}
	if flagConfigFile != "" {
		b, err := ioutil.ReadFile(flagConfigFile)
		if err != nil {
			log.Fatalln("read config:", err)
		}
		err = json.Unmarshal(b, &config)
		if err != nil {
			log.Fatalln("parse config:", err)
		}
	}
	if config.Key == "" {
		log.Fatalln("Key is required.")
	}

	addr, err := net.ResolveUDPAddr(config.Net, config.Address)
	if err != nil {
		log.Fatalln(err)
	}
	remote := (*net.UDPAddr)(nil)
	if config.Remote != "" {
		remote, err = net.ResolveUDPAddr(config.Net, config.Remote)
		if err != nil {
			log.Fatalln(err)
		}
	}

	k := pbkdf2.Key([]byte(config.Key), []byte(salt), 4096, chacha20poly1305.KeySize, sha256.New)
	aead, err := chacha20poly1305.NewX(k)
	if err != nil {
		log.Fatalln("crypto:", err)
	}

	var outFile *udpmux.RotateWriter
	if config.Output != "" {
		outFile = udpmux.CreateRotateWriteCloser(config.Output, nil)
	}

	instance := &udpmux.Instance{
		Network: config.Net,
		Addr:    addr,
		Remote:  remote,
		AEAD:    aead,
		Tag:     []byte(tag),
	}

	instance.Channels = make([]*net.UDPAddr, 0)
	for _, s := range config.Channels {
		addr, err := net.ResolveUDPAddr(config.Net, s)
		if err != nil {
			log.Fatalln(err)
		}
		instance.Channels = append(instance.Channels, addr)
	}

	if !config.ServerMode {
		instance.ClientInstance = &udpmux.ClientInstance{
			Interval: time.Duration(config.Keepalive) * time.Second,
			Output:   outFile,
		}
	}

	// discard sensitive info explicitly
	config.Key = ""

	return instance
}

func main() {
	i := prepare()
	log.Println("starting")
	i.Run()
}
