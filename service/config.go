package main

type Config struct {
	AutoSaveInteralSeconds uint
	EnableVowpalWabbit     bool
	LogLevel               string
}

func NewConfig() *Config {
	return &Config{
		AutoSaveInteralSeconds: 300,
		EnableVowpalWabbit:     true,
		LogLevel:               "info",
	}
}
