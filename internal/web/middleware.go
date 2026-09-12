package web

import (
	"net"
	"net/http"
)

func (s *Server) limit(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := getClientIP(r)
		if !s.limiter.Eligible(ip) {
			http.Error(w, "too many requests", http.StatusTooManyRequests)
			return
		}

		handler.ServeHTTP(w, r)
	})
}

func getClientIP(r *http.Request) string {
	ip := net.ParseIP(r.Header.Get("X-Real-IP"))

	if ip == nil {
		host, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			host = r.RemoteAddr
		}
		ip = net.ParseIP(host)
	}

	return ip.String()
}
