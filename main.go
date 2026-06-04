package main

import (
	"fmt"
	"net"
	"flag"

	"multiproxy/router"
	"multiproxy/protocols"
	"multiproxy/config"
)


func main() {
	configPath := flag.String("config", "proxy_config.json", "Path to config file")
	flag.Parse()

	cfg, err := config.LoadConfig(*configPath)
	if err != nil {
		fmt.Println("Config not found. Creating default config:", *configPath)

		defaultCfg := config.DefaultConfig()

		err := config.SaveConfig(*configPath, defaultCfg)
		if err != nil {
			fmt.Println("Failed to create default config:", err)
			return
		}

		fmt.Println("Default config created. Edit it and restart.")
		return
	}

	address := fmt.Sprintf("%s:%d", cfg.ListenIP, cfg.ListenPort)

	ln, err := net.Listen("tcp", address)

	if err != nil {
		fmt.Println(err)
	}
	defer ln.Close()

	fmt.Println("Multiproxy started on", address)

	socks5 := protocols.NewSocks5Proxy(cfg.Socks5Username, cfg.Socks5Password)
	socks4 := protocols.NewSocks4Proxy(cfg.Socks4Username)
	http   := protocols.NewHTTPProxy("", "", "")
	httpWeb := protocols.NewHTTPWebProxy(cfg.WebProxyHost, cfg.WebProxyPort)
	mc     := protocols.NewMCProxy(cfg.MCServerIP, cfg.MCServerPort)

	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println(err)
		}

		go router.Route(conn, socks5, socks4, http, httpWeb, mc)

	}

}
