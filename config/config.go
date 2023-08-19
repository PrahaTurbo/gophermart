package config

import (
	"flag"
	"os"
)

type Config struct {
	RunAddr        string
	DatabaseURI    string
	AccrualSysAddr string
	JWTSecret      string
}

func Load() Config {
	var c Config
	flag.StringVar(&c.RunAddr, "a", "localhost:8080", "server address in a form host:port")
	flag.StringVar(&c.DatabaseURI, "d", "", "database address")
	flag.StringVar(&c.AccrualSysAddr, "r", "http://localhost:8081", "accrual system address")

	flag.Parse()

	c.loadEnvVars()

	return c
}

func (c *Config) loadEnvVars() {
	if envRunAddr := os.Getenv("RUN_ADDRESS"); envRunAddr != "" {
		c.RunAddr = envRunAddr
	}

	if envDatabaseURI := os.Getenv("DATABASE_URI"); envDatabaseURI != "" {
		c.DatabaseURI = envDatabaseURI
	}

	if envAccrualSysAddr := os.Getenv("ACCRUAL_SYSTEM_ADDRESS"); envAccrualSysAddr != "" {
		c.AccrualSysAddr = envAccrualSysAddr
	}

	if envJWTSecret := os.Getenv("JWT_SECRET"); envJWTSecret != "" {
		c.JWTSecret = envJWTSecret
	} else {
		c.JWTSecret = "secret_for_tests"
	}
}
