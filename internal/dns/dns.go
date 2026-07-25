package dns

import (
	"context"
	"log"
	"net"
	"strings"
	"time"

	"github.com/cantr1/GoDNS/internal/database"
	"golang.org/x/net/dns/dnsmessage"
)

// Dependencies is a struct containing dependencies for the DNS server
type Dependencies struct {
	DBQueries *database.Queries
	Port      int
}

// Server is a struct used at runtime by the DNS server to process dependencies
type Server struct {
	DBQueries *database.Queries
	Port      int
}

func parseIPv4(value string) ([4]byte, bool) {
	ip := net.ParseIP(value).To4()
	if ip == nil {
		return [4]byte{}, false
	}

	return [4]byte{ip[0], ip[1], ip[2], ip[3]}, true
}

func (s *Server) getRecordsByName(name string) ([]database.DnsRecord, error) {
	// db queries require background context to function - use with timeout in case of long reads
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	return s.DBQueries.GetDNSRecordByName(ctx, name)
}

func NewServer(dependencies *Dependencies) *Server {
	dnsServer := &Server{
		DBQueries: dependencies.DBQueries,
		Port:      dependencies.Port,
	}

	return dnsServer
}

func Run(dnsServer *Server) {
	addr := &net.UDPAddr{Port: dnsServer.Port}
	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		panic(err)
	}
	// Buffer to store incoming UDP packets (Standard DNS UDP limit is 512 bytes)
	buf := make([]byte, 512)

	for {
		// Read incoming packets
		n, remoteAddr, err := conn.ReadFromUDP(buf)
		if err != nil {
			log.Printf("Error reading packet: %v\n", err)
			continue
		}

		// Parse raw data into a DNS Message struct
		var msg dnsmessage.Message
		if err := msg.Unpack(buf[:n]); err != nil {
			log.Printf("Error unpacking message: %v\n", err)
			continue
		}

		// Skip if there are no questions asked
		if len(msg.Questions) == 0 {
			continue
		}

		// Process the first question (standard behavior for simple servers)
		question := msg.Questions[0]
		log.Printf("Received query for: %s (Type: %v)\n", question.Name.String(), question.Type)

		// Formulate the response skeleton
		reply := dnsmessage.Message{
			Header: dnsmessage.Header{
				ID:            msg.ID, // Match request ID
				Response:      true,   // Mark as response packet
				OpCode:        msg.OpCode,
				Authoritative: true,
			},
			Questions: msg.Questions, // Return original question block
		}

		// Handle 'A' Record Lookup (IPv4 address queries)
		if question.Type == dnsmessage.TypeA {
			var dbRecords []database.DnsRecord
			// Query the DB based on the question name / value
			if question.Name.String() != "" {
				name := strings.ToLower(strings.TrimSuffix(question.Name.String(), "."))
				log.Printf("Querying database for: %s\n", name)
				dbRecords, err = dnsServer.getRecordsByName(name)
				if err != nil {
					log.Printf("Error querying database: %v\n", err)
					continue
				}
			} else {
				log.Printf("Unhandled query received")
			}

			for _, record := range dbRecords {
				ip, ok := parseIPv4(record.Value)
				if !ok {
					continue
				}
				reply.Answers = append(reply.Answers, dnsmessage.Resource{
					Header: dnsmessage.ResourceHeader{
						Name:  question.Name,
						Type:  dnsmessage.TypeA,
						Class: dnsmessage.ClassINET,
						TTL:   uint32(record.Ttl), // Time-to-live: 5 minutes
					},
					Body: &dnsmessage.AResource{
						A: ip, // Resolves to localhost
					},
				})
			}
		}

		// 6. Pack the struct back into binary wire format
		packed, err := reply.Pack()
		if err != nil {
			log.Printf("Error packing reply: %v\n", err)
			continue
		}

		// 7. Write the packed byte stream back to the client
		_, err = conn.WriteToUDP(packed, remoteAddr)
		if err != nil {
			log.Printf("Error sending response: %v\n", err)
		}
	}
}
