package main

type Config struct {
	HistoryLimit           int
	AutoSaveInteralSeconds uint
	EnableVowpalWabbit     bool
}

func NewConfig() *Config {
	return &Config{
		HistoryLimit:           0,
		AutoSaveInteralSeconds: 300,
		EnableVowpalWabbit:     true,
	}
}
