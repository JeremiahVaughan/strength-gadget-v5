package main

import (
	"log"
	"net"
	"net/http"
)

func IpFilterMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		clientIP, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			log.Printf("error, when attempting to split the remote address into host and port: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		ip := net.ParseIP(clientIP)
		for _, block := range AllowedIpRanges {
			if block.Contains(ip) {
				next.ServeHTTP(w, r)
				return
			}
		}
		http.Error(w, "Forbidden", http.StatusForbidden)
		log.Printf("a request was filtered out because the source IP was not from Cloudflares whitelist. IP: %s", clientIP)
		return
	})
}
