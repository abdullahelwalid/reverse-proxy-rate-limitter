package server

import (
	"encoding/json"
	"errors"
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

func RunServer(proxyConfig config.Proxy) error {
	mux := http.NewServeMux()
	for _, resource := range proxyConfig.Resources {
		url, _ := url.Parse("http://" + resource.DomainName + ":" + strconv.Itoa(resource.Port))
		proxy := httputil.NewSingleHostReverseProxy(url)

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
				clientToken := &limitter.ClientToken{IPAddr: r.RemoteAddr}
				err := clientToken.Consume()
				var errLimitExceed limitter.ErrorClientTokenLimitExceed
				if errors.Is(err, &errLimitExceed) {
					w.Header().Add("Content-Type", "application/json")
					w.WriteHeader(http.StatusTooManyRequests)
					json.NewEncoder(w).Encode(map[string]string{"error": "Too many requests"})
					return
				} else if err != nil {
					w.Header().Add("Content-Type", "application/json")
					w.WriteHeader(http.StatusInternalServerError)
					json.NewEncoder(w).Encode(map[string]string{"error": "An error has occurred"})
					return
				}
				// Note that ServeHttp is non blocking and uses a go routine under the hood
				fmt.Printf("[ TinyRP ] Redirecting request to %s from %s at %s\n", r.URL, r.RemoteAddr, time.Now().UTC())
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
