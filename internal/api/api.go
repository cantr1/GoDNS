package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"

	"github.com/cantr1/GoDNS/internal/database"
)

// Record --- Struct definition for returning a full record
type Record struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	TTL       int32     `json:"ttl"`
	Class     string    `json:"class"`
	Type      string    `json:"type"`
	Value     string    `json:"value"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Server --- Struct definition for database queries
type Server struct {
	DBQueries *database.Queries
}

func NewServer(port string, dbQueries *database.Queries) *http.Server {
	var mux = http.NewServeMux()

	// Create server struct for DB queries
	apiServer := &Server{
		DBQueries: dbQueries,
	}

	apiServer.registerRoutes(mux)

	// Start Server
	server := &http.Server{
		Addr:    port,
		Handler: mux,
	}

	return server
}

func (apiServer *Server) registerRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/health", apiServer.healthHandler)
	mux.HandleFunc("GET /api/records", apiServer.GetAllRecords)
}

func (apiServer *Server) healthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func (apiServer *Server) GetAllRecords(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Get the records from the database
	dbRecords, err := apiServer.DBQueries.GetDNSRecords(r.Context())
	if err != nil {
		http.Error(w, "Failed to retrieve records", http.StatusInternalServerError)
		return
	}

	// Create a slice to store parsed records
	var records []Record
	for _, record := range dbRecords {
		tmp := Record{
			ID:        record.ID,
			TTL:       record.Ttl,
			Class:     record.Class,
			Type:      record.Type,
			Value:     record.Value,
			CreatedAt: record.CreatedAt,
			UpdatedAt: record.UpdatedAt,
		}
		records = append(records, tmp)
	}

	// Parse to JSON
	jsonData, err := json.Marshal(records)
	if err != nil {
		http.Error(w, "Failed to marshal records", http.StatusInternalServerError)
		return
	}

	_, err = w.Write(jsonData)
	if err != nil {
		http.Error(w, "Failed to write response", http.StatusInternalServerError)
		return
	}
}

func Run(server *http.Server) error {
	return server.ListenAndServe()
}
