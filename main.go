package main

import (
	"log"
	"net"
	"net/http"
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
	service := &Service{
		ID:            "hello",
		ContainerPort: "80",
		Protocol:      "tcp",
		Port:          "5000",
		Selector:      map[string]string{"environment": "prod"},
	}
	sm.add(service)

	hostPort := net.JoinHostPort(bindIP, "8000")
	log.Fatal(http.ListenAndServe(hostPort, nil))
}
