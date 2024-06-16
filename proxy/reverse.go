package proxy

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
)

func NewReverseProxy(targetHost string) (*httputil.ReverseProxy, error) {
	targetUrl, err := url.Parse(targetHost)
	if err != nil {
		return nil, err
	}

	return httputil.NewSingleHostReverseProxy(targetUrl), nil
}

func ReverseHandler(proxy *httputil.ReverseProxy) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		proxy.ServeHTTP(w, r)
	}
}

func StartReverseProxy(listenAddr string, targetAddr string) {
	proxy, err := NewReverseProxy("http://" + targetAddr + "/")
	if err != nil {
		panic(err)
	}

	http.HandleFunc("/", ReverseHandler(proxy))
	log.Fatal(http.ListenAndServe(listenAddr, nil))
}
