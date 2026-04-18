package config

import (
	"encoding/json"
	"os"
)


type Config struct {
	Socks5Username string `json:"socks5_username"`
	Socks5Password string `json:"socks5_password"`
	Socks4Username string `json:"socks4_username"`

	MCServerIP   string `json:"mc_server_ip"`
	MCServerPort int    `json:"mc_server_port"`

	ListenIP   string `json:"listen_ip"`
	ListenPort int    `json:"listen_port"`
}

func DefaultConfig() *Config {
	return &Config{
		Socks5Username: "",
		Socks5Password: "",
		Socks4Username: "",
		MCServerIP:     "localhost",
		MCServerPort:   25565,
		ListenIP:       "localhost",
		ListenPort:     1080,
	}
}

func SaveConfig(path string, cfg *Config) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ") // pretty JSON

	return encoder.Encode(cfg)
}

func LoadConfig(path string) (*Config, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var cfg Config
	decoder := json.NewDecoder(file)

	if err := decoder.Decode(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

