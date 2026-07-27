package api

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/cantr1/GoDNS/internal/auth"
	"github.com/cantr1/GoDNS/internal/database"
	"github.com/google/uuid"
)

// Dependencies is a struct containing dependencies for the API server meant to be passed to the API server
type Dependencies struct {
	DBQueries *database.Queries
	APIKey    string
	DEVMode   string
	Port      string
}

// Server is a struct used at runtime by the API to process dependencies
type Server struct {
	DBQueries *database.Queries
	APIKey    string
	DEVMode   string
	Port      string
}

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

// RecordCreate --- Struct definition for creating a new record
type RecordCreate struct {
	Name  string `json:"name"`
	TTL   int32  `json:"ttl"`
	Class string `json:"class"`
	Type  string `json:"type"`
	Value string `json:"value"`
}

func (r RecordCreate) Validate() bool {
	if r.Name == "" {
		return false
	}
	if r.TTL <= 0 {
		return false
	}
	if r.Class == "" {
		return false
	}
	if r.Type == "" {
		return false
	}
	if r.Value == "" {
		return false
	}

	return true
}

func NewServer(dependencies *Dependencies) *http.Server {
	var mux = http.NewServeMux()

	// Create struct for runtime dependencies
	apiServer := Server{
		DBQueries: dependencies.DBQueries,
		APIKey:    dependencies.APIKey,
		DEVMode:   dependencies.DEVMode,
		Port:      dependencies.Port,
	}

	apiServer.registerRoutes(mux)

	// Start Server
	server := &http.Server{
		Addr:    apiServer.Port,
		Handler: mux,
	}

	return server
}

func (apiServer *Server) registerRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/health", apiServer.healthHandler)
	mux.HandleFunc("GET /api/records", apiServer.getRecords)
	mux.HandleFunc("POST /api/records", apiServer.createRecord)
	mux.HandleFunc("DELETE /api/records", apiServer.deleteRecords)
}

func (apiServer *Server) healthHandler(w http.ResponseWriter, r *http.Request) {
	key, err := auth.GetBearerToken(*r)
	if err != nil {
		http.Error(w, "Failed to retrieve token", http.StatusUnauthorized)
		return
	}
	if key != apiServer.APIKey {
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (apiServer *Server) createRecord(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	key, err := auth.GetBearerToken(*r)
	if err != nil {
		http.Error(w, "Failed to retrieve token", http.StatusUnauthorized)
		return
	}
	if key != apiServer.APIKey {
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}

	decoder := json.NewDecoder(r.Body)
	parameters := RecordCreate{}
	err = decoder.Decode(&parameters)
	if err != nil {
		http.Error(w, "Error decoding parameters", http.StatusBadRequest)
		return
	}

	if !parameters.Validate() {
		http.Error(w, "Invalid parameters", http.StatusBadRequest)
		return
	}

	// Parse the RecordCreate struct into the required DB parameters
	dbParams := database.CreateDNSRecordParams{
		Name:  parameters.Name,
		Class: parameters.Class,
		Type:  parameters.Type,
		Value: parameters.Value,
		Ttl:   parameters.TTL,
	}

	newRecord, err := apiServer.DBQueries.CreateDNSRecord(r.Context(), dbParams)
	if err != nil {
		http.Error(w, "Failed to create record", http.StatusInternalServerError)
		return
	}

	// Marshal new record and return data
	response, err := json.Marshal(newRecord)
	if err != nil {
		http.Error(w, "Failed to marshal response", http.StatusInternalServerError)
		return
	}
	w.Write(response)
}

func (apiServer *Server) deleteRecords(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	key, err := auth.GetBearerToken(*r)
	if err != nil {
		http.Error(w, "Failed to retrieve token", http.StatusUnauthorized)
		return
	}
	if key != apiServer.APIKey {
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}

	// Check for query
	recordName := r.URL.Query().Get("name")
	recordValue := r.URL.Query().Get("value")

	if recordName != "" && recordValue != "" {
		http.Error(w, "Cannot query for both name and value", http.StatusBadRequest)
		return
	}

	if recordName == "" && recordValue == "" {
		// Allow full DB reset if in dev mode and not targeting specific record
		if apiServer.DEVMode != "true" {
			http.Error(w, "DEV mode is disabled", http.StatusUnauthorized)
			return
		}

		err = apiServer.DBQueries.RemoveRecords(r.Context())
		if err != nil {
			http.Error(w, "Failed to delete records", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	} else if recordName != "" {
		err = apiServer.DBQueries.RemoveRecordByName(r.Context(), recordName)
		if err != nil {
			http.Error(w, "Failed to delete record", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	} else {
		err = apiServer.DBQueries.RemoveRecordByValue(r.Context(), recordValue)
		if err != nil {
			http.Error(w, "Failed to delete record", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}

func (apiServer *Server) getRecords(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	recordName := r.URL.Query().Get("name")
	recordValue := r.URL.Query().Get("value")

	// If both query for both, fail out
	if recordName != "" && recordValue != "" {
		http.Error(w, "Cannot query for both name and value", http.StatusBadRequest)
		return
	}

	key, err := auth.GetBearerToken(*r)
	if err != nil {
		http.Error(w, "Failed to retrieve token", http.StatusUnauthorized)
		return
	}
	if key != apiServer.APIKey {
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}

	var dbRecords []database.DnsRecord
	if recordName == "" && recordValue == "" {
		// Get all records from the database
		dbRecords, err = apiServer.DBQueries.GetDNSRecords(r.Context())
		if err != nil {
			http.Error(w, "Failed to retrieve records", http.StatusInternalServerError)
			return
		}
	} else if recordName != "" {
		dbRecords, err = apiServer.DBQueries.GetDNSRecordByName(r.Context(), recordName)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				http.Error(w, "Record not found", http.StatusNotFound)
				return
			}
			http.Error(w, "Failed to retrieve record", http.StatusInternalServerError)
			return
		}
	} else {
		// Return records by value
		dbRecords, err = apiServer.DBQueries.GetDNSRecordByValue(r.Context(), recordValue)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				http.Error(w, "Record not found", http.StatusNotFound)
				return
			}
			http.Error(w, "Failed to retrieve record", http.StatusInternalServerError)
			return
		}
	}

	// Create a slice to store parsed records
	var records []Record
	for _, record := range dbRecords {
		tmp := Record{
			ID:        record.ID,
			Name:      record.Name,
			TTL:       record.Ttl,
			Class:     record.Class,
			Type:      record.Type,
			Value:     record.Value,
			CreatedAt: record.CreatedAt,
			UpdatedAt: record.UpdatedAt,
		}
		records = append(records, tmp)
	}

	if err := json.NewEncoder(w).Encode(records); err != nil {
		http.Error(w, "Failed to encode records", http.StatusInternalServerError)
		return
	}

}

func Run(server *http.Server) error {
	return server.ListenAndServe()
}
