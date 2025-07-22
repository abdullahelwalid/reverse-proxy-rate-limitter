package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/abdullahelwalid/go-rate-limiter/pkg/config"
	"github.com/abdullahelwalid/go-rate-limiter/pkg/limitter"
)

func ErrorHandler(w http.ResponseWriter, r *http.Request, e error) {
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadGateway)
	json.NewEncoder(w).Encode(map[string]string{
		"error": "Bad Gateway",
	})
}

func RunServer(proxyConfig config.Proxy) error {
	mux := http.NewServeMux()
	for _, resource := range proxyConfig.Resources {
		url, _ := url.Parse("http://" + resource.DomainName + ":" + strconv.Itoa(resource.Port))
		proxy := httputil.NewSingleHostReverseProxy(url)
		proxy.ErrorHandler = ErrorHandler

		handler := func(p *httputil.ReverseProxy) func(w http.ResponseWriter, r *http.Request) {
			return func(w http.ResponseWriter, r *http.Request) {
				r.URL.Host = url.Host
				r.URL.Scheme = url.Scheme
				r.Header.Set("X-Forwarded-Host", r.Header.Get("Host"))
				r.Host = url.Host
				// trim reverseProxyRoutePrefix
				path := r.URL.Path
				var endpoint string = "/"
				r.URL.Path = strings.TrimLeft(path, endpoint)
				ipAddr := strings.Split(r.RemoteAddr, ":")[0]
				clientToken := &limitter.ClientToken{IPAddr: ipAddr}
				err := clientToken.Consume()

				if err != nil && err.Error() == "Limit Exceeded" {
					w.Header().Add("Content-Type", "application/json")
					w.WriteHeader(http.StatusTooManyRequests)
					json.NewEncoder(w).Encode(map[string]string{"error": "Too many requests"})
					return
				} else if err != nil {
					fmt.Println(err)
					w.Header().Add("Content-Type", "application/json")
					w.WriteHeader(http.StatusInternalServerError)
					json.NewEncoder(w).Encode(map[string]string{"error": "An error has occurred"})
					return
				}
				// Note that ServeHttp is non blocking and uses a go routine under the hood
				fmt.Printf("[ TinyRP ] Redirecting request to %s from %s at %s\n", r.URL, ipAddr, time.Now().UTC())
				proxy.ServeHTTP(w, r)
			}
		}
		mux.HandleFunc(resource.Endpoint, handler(proxy))
	}

	if err := http.ListenAndServe(proxyConfig.DomainName+":"+strconv.Itoa(proxyConfig.Port), mux); err != nil {
		return fmt.Errorf("Failed to start the server: %v", err)
	}
	return nil
}
