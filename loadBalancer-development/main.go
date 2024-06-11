package main

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
)

type Server interface{
	Address() string
	IsAlive() bool
	Serve(rw http.ResponseWriter, r *http.Request)
}
type simpleServer struct{
	addr string
	proxy httputil.ReverseProxy
}

func newSimpleServer(addr string) *simpleServer{
	serverUrl, err := url.Parse(addr)
	handleErr(err)

	return &simpleServer{
		addr: addr,
		proxy: *httputil.NewSingleHostReverseProxy(serverUrl),
	}
}

type loadBalancer struct {
	port string
	roundRobinCount int
	servers []Server
}

func NewLoadBalancer(port string, servers []Server) *loadBalancer{
	return &loadBalancer{
		port: port,
		roundRobinCount: 0,
		servers: servers,
	}

}

func handleErr(err error){

	if err != nil{
		fmt.Printf("error: %v\n",err)
		os.Exit(1);
	}
}

func (s *simpleServer) Address() string{
	return s.addr
}

func (s *simpleServer) IsAlive() bool{
	return true
}

func (s *simpleServer) Serve(rw http.ResponseWriter, req *http.Request){
	s.proxy.ServeHTTP(rw, req)
}

func (lb *loadBalancer) getNextAvailableServer() Server{
	server := lb.servers(lb.roundRobinCount%len(lb.servers))
	for !server.IsAlive(){
		lb.roundRobinCount++
		server = lb.servers[lb.roundRobinCount%len(lb.servers)]
	}
	lb.roundRobinCount++
	return server
}

func (lb *loadBalancer) serveProxy(rw http.ResponseWriter, req *http.Request) {

	targetServer := lb.getNextAvailableServer()
	fmt.Printf("forwarding request to Address %q\n", &targetServer.Address())
	targetServer.serve(rw,req)
}

func main(){
	servers := []Server{
		newSimpleServer("https://www.google.com"),
		newSimpleServer("https://www.facebook.com"),
		newSimpleServer("https://www.instagram.com"),
	}

	lb := NewLoadBalancer("8000",servers)
	handleRedirect := func(rw http.ResponseWriter, req *http.Request){
		lb.serveProxy(rw, req)
	}
	http.HandleFunc("/",handleRedirect)

	fmt.Printf("Server is working on PORT 8000: %s'\n",lb.port)
	http.ListenAndServe(":"+lb.port, nil)
}