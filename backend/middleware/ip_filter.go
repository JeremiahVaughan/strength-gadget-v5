package middleware

import (
	"log"
	"net"
	"net/http"
	"strengthgadget.com/m/v2/config"
)

func IpFilterMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		//clientIP := strings.Split(r.RemoteAddr, ":")[0]
		clientIP := "10.0.0.2"
		ip := net.ParseIP(clientIP)
		for _, block := range config.AllowedIpRanges {
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
