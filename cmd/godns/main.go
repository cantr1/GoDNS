package main

import (
	"database/sql"
	"log"
	"os"

	"github.com/cantr1/GoDNS/internal/api"
	"github.com/cantr1/GoDNS/internal/database"
	"github.com/cantr1/GoDNS/internal/dns"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

// Config is a struct containing configuration variables for the application
type Config struct {
	APIPort string
	DNSPort string
	APIKey  string
	DBURL   string
	DEVMode bool
}

func main() {
	err := godotenv.Load()
	if err != nil {
		// log fatal closes the program - the more you know
		log.Fatalf("error loading env variables: %v", err)
	}

	config := Config{
		APIPort: os.Getenv("API_PORT"),
		DNSPort: os.Getenv("DNS_PORT"),
		APIKey:  os.Getenv("API_KEY"),
		DBURL:   os.Getenv("DB_URL"),
		DEVMode: os.Getenv("DEV_MODE") == "true",
	}

	// Connect to database
	db, err := sql.Open("postgres", config.DBURL)
	if err != nil {
		log.Printf("Failure to open connection to backend DB: %v", err)
		return
	}

	// Set up dependencies for the API server
	apiDependencies := api.Dependencies{
		DBQueries: database.New(db),
		APIKey:    config.APIKey,
		DEVMode:   config.DEVMode,
	}

	apiServer := api.NewServer(config.APIPort, &apiDependencies)
	dnsServer := dns.NewServer(config.DNSPort)

	api.Run(apiServer)
	dns.Run(dnsServer)
}
