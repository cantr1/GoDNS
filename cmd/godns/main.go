package main

import (
	"database/sql"
	"log"
	"os"
	"strconv"

	"github.com/cantr1/GoDNS/internal/api"
	"github.com/cantr1/GoDNS/internal/database"
	"github.com/cantr1/GoDNS/internal/dns"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

// Config is a struct containing configuration variables for the application
type Config struct {
	APIPort string
	DNSPort int
	APIKey  string
	DBURL   string
	DEVMode string
}

func main() {
	err := godotenv.Load()
	if err != nil {
		// log fatal closes the program - the more you know
		log.Fatalf("error loading env variables: %v", err)
	}

	// Convert DNS port to string
	dnsPortStr, err := strconv.Atoi(os.Getenv("DNS_PORT"))
	if err != nil {
		log.Fatalf("error converting DNS port to string: %v", err)
	}

	config := Config{
		APIPort: os.Getenv("API_PORT"),
		DNSPort: dnsPortStr,
		APIKey:  os.Getenv("API_KEY"),
		DBURL:   os.Getenv("DB_URL"),
		DEVMode: os.Getenv("DEV_MODE"),
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
		Port:      config.APIPort,
	}

	// Set up dependencies for DNS server
	dnsDependencies := dns.Dependencies{
		DBQueries: database.New(db),
		Port:      config.DNSPort,
	}

	apiServer := api.NewServer(&apiDependencies)
	dnsServer := dns.NewServer(&dnsDependencies)

	go api.Run(apiServer)
	dns.Run(dnsServer)
}
