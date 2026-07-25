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

type Config struct {
	ApiPort   string
	DnsPort   string
	DBQueries *database.Queries
	DBURL     string
}

func main() {
	err := godotenv.Load()
	if err != nil {
		// log fatal closes the program - the more you know
		log.Fatalf("error loading env variables: %v", err)
	}

	config := Config{
		ApiPort: os.Getenv("API_PORT"),
		DnsPort: os.Getenv("DNS_PORT"),
		DBURL:   os.Getenv("DB_URL"),
	}

	// Connect to database
	db, err := sql.Open("postgres", config.DBURL)
	if err != nil {
		log.Printf("Failure to open connection to backend DB: %v", err)
		return
	}

	config.DBQueries = database.New(db)

	apiServer := api.NewServer(config.ApiPort)
	dnsServer := dns.NewServer(config.DnsPort)

	api.Run(apiServer)
	dns.Run(dnsServer)
}
