package server

import (
	"io"
	"log"
	"net/http"
)

const (
	connected = "200 Connected to Gee RPC"
	defaultRPCPath   = "/_geeprc_"
	defaultDebugPath = "/debug/geerpc"
)

func (s *Server) HandleHTTP() {
	// Register HTTP Handler to particular path
	http.Handle(defaultRPCPath, s)
}

func (s *Server) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	if req.Method != "CONNECT" {
		// support CONNECT only
		w.Header().Set("Content-Type", "test/plain; charset=utf-8")
		w.WriteHeader(http.StatusMethodNotAllowed)
		_, _ = io.WriteString(w, "405 must CONNECT\n")
		return
	}

	// After Hijack, the management of connection is the responsibility of caller
	conn, _, err := w.(http.Hijacker).Hijack()
	if err != nil {
		log.Print("prc hijacking", req.RemoteAddr, ":", err.Error())
		return
	}

	_, _ = io.WriteString(conn, "HTTP/1.0"+connected+"\n\n")
	// start exchanging Options and messages
	s.HandleConn(conn)
}

func HandleHTTP() {
	DefaultServer.HandleHTTP()
}

