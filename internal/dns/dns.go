package dns

import "net/http"

func NewServer(port string) *http.Server {
	return &http.Server{
		Addr: port,
	}
}

func Run(server *http.Server) {
	//placeholder
}
