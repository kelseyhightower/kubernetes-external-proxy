package main

import (
	"io"
	"log"
	"net"
	"net/http"
	"net/rpc"
	"os"
)

var (
	apiserver string
	bindIP    string
)

func main() {
	apiserver = os.Getenv("KUBERNETES_API_SERVER")
	if apiserver == "" {
		log.Fatal("KUBERNETES_API_SERVER cannot be empty")
	}

	bindIP = os.Getenv("KEP_BIND_IP")
	if bindIP == "" {
		bindIP = "0.0.0.0"
	}
	sm := newServiceManager()
	rpc.Register(sm)
	http.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		defer req.Body.Close()
		w.Header().Set("Content-Type", "application/json")
		res := NewRPCRequest(req.Body).Call()
		_, err := io.Copy(w, res)
		if err != nil {
			log.Println(err)
		}
	})
	hostPort := net.JoinHostPort(bindIP, "8000")
	log.Println("starting kubernetes-external-proxy service...")
	log.Printf("accepting RPC request on http://%s", hostPort)
	log.Fatal(http.ListenAndServe(hostPort, nil))
}
