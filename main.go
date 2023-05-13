package main

import (
    "flag"
    "log"
    "net"
    "time"

    "github.com/armon/go-socks5"
    "github.com/elazarl/goproxy"
    "golang.org/x/net/proxy"
)

var (
    listenAddr  = flag.String("listen-addr", "127.0.0.1:8080", "proxy server listen address")
    socks5Addr  = flag.String("socks5-addr", "127.0.0.1:1080", "socks5 server address")
    username    = flag.String("username", "", "proxy authentication username")
    password    = flag.String("password", "", "proxy authentication password")
    idleTimeout = flag.Duration("idle-timeout", 60*time.Second, "proxy idle timeout")
)

func main() {
    flag.Parse()

    // Create a SOCKS5 server.
    socks5Config := &socks5.Config{}
    socks5Server, err := socks5.New(socks5Config)
    if err != nil {
        log.Fatalf("failed to create SOCKS5 server: %v", err)
    }
    go func() {
        log.Printf("starting SOCKS5 server at %v", *socks5Addr)
        if err := socks5Server.ListenAndServe("tcp", *socks5Addr); err != nil {
            log.Fatalf("failed to start SOCKS5 server: %v", err)
        }
    }()

    // Create a HTTP/HTTPS proxy server.
    proxyServer := goproxy.NewProxyHttpServer()
    proxyServer.Tr.Proxy = func(req *http.Request) (*url.URL, error) {
        dialer, err := proxy.SOCKS5("tcp", *socks5Addr, nil, proxy.Direct)
        if err != nil {
            return nil, err
        }
        return dialer.Dial("tcp", req.URL.Host)
    }
    proxyServer.ConnectDial = proxyServer.Tr.Dial
    proxyServer.Logger = log.New(ioutil.Discard, "", 0) // disable logging
    proxyServer.OnRequest().DoFunc(func(req *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
        // Check if the request is from an authenticated user.
        if *username != "" && *password != "" {
            if user, pass, ok := req.BasicAuth(); !ok || user != *username || pass != *password {
                response := goproxy.NewResponse(req,
                    goproxy.ContentTypeText, http.StatusUnauthorized, "Unauthorized")
                response.Header.Set("Proxy-Authenticate", `Basic realm="Restricted"`)
                return nil, response
            }
        }
        return req, nil
    })
    httpServer := &http.Server{
        Addr:         *listenAddr,
        Handler:      proxyServer,
        ReadTimeout:  *idleTimeout,
        WriteTimeout: *idleTimeout,
        IdleTimeout:  *idleTimeout,
    }
    log.Printf("starting HTTP/HTTPS proxy server at %v", *listenAddr)
    if err := httpServer.ListenAndServe(); err != nil {
        log.Fatalf("failed to start HTTP/HTTPS proxy server: %v", err)
    }
}
