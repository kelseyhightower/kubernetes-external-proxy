package main

import (
	"container/ring"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

type Service struct {
	ID            string            `json:"id"`
	ContainerPort string            `json:"containerPort"`
	Protocol      string            `json:"protocol"`
	Port          string            `json:"port"`
	Selector      map[string]string `json:"selector"`
}

type ServiceProxy struct {
	service  *Service
	shutdown chan bool
	done     chan bool
	sync.Mutex
	pods []string
	r    *ring.Ring
}

func newServiceProxy(service *Service) *ServiceProxy {
	shutdown := make(chan bool)
	done := make(chan bool)
	return &ServiceProxy{done: done, shutdown: shutdown, service: service}
}

func (sp *ServiceProxy) start() error {
	if err := sp.updatePods(); err != nil {
		return err
	}

	hostPort := net.JoinHostPort(bindIP, sp.service.Port)
	ln, err := net.Listen(sp.service.Protocol, hostPort)
	if err != nil {
		return err
	}

	var shutdown bool
	go func() {
		log.Printf("accepting new connections for the %s service", sp.service.ID)
		log.Printf("proxy established %s <-> %s", sp.service.Port, sp.service.ContainerPort)
		for {
			if shutdown {
				goto Shutdown
			}
			conn, err := ln.Accept()
			if err != nil {
				if shutdown {
					goto Shutdown
				}
				log.Printf("error accepting connections for service: %s", sp.service.ID)
				time.Sleep(time.Duration(5 * time.Second))
				continue
			}
			go sp.handleConnection(conn)
		}

	Shutdown:
		log.Printf("stopping the %s service...", sp.service.ID)
		sp.done <- true
		return
	}()
	go func() {
		select {
		case <-sp.shutdown:
			shutdown = true
			ln.Close()
		}
	}()
	return nil
}

func (sp *ServiceProxy) stop() error {
	sp.shutdown <- true
	<-sp.done
	return nil
}

func (sp *ServiceProxy) updatePods() error {
	pods, err := podsFromLabelQuery(sp.service.Selector)
	if err != nil {
		return err
	}
	sp.Lock()
	sp.pods = pods
	r := ring.New(len(pods))
	for i := 0; i < r.Len(); i++ {
		r.Value = pods[i]
		r = r.Next()
	}
	sp.r = r
	sp.Unlock()
	return nil
}

func (sp *ServiceProxy) nextPod() string {
	sp.Lock()
	sp.Unlock()
	if sp.r == nil {
		return ""
	}
	sp.r = sp.r.Next()
	return net.JoinHostPort(sp.r.Value.(string), sp.service.ContainerPort)
}

func (sp *ServiceProxy) handleConnection(conn net.Conn) {
	hostPort := sp.nextPod()
	if hostPort == "" {
		log.Printf("error cannot service request for %s: no pods available", sp.service.ID)
		conn.Close()
		return
	}
	proxyConn, err := net.Dial(sp.service.Protocol, hostPort)
	if err != nil {
		log.Println(err)
		return
	}

	var wg sync.WaitGroup
	wg.Add(2)

	go copyData(proxyConn, conn, &wg)
	go copyData(conn, proxyConn, &wg)
	wg.Wait()
	conn.Close()
	proxyConn.Close()
}

func copyData(in, out net.Conn, wg *sync.WaitGroup) {
	defer wg.Done()
	_, err := io.Copy(out, in)
	if err != nil {
		log.Println(err)
	}
	out.(*net.TCPConn).CloseWrite()
	in.(*net.TCPConn).CloseRead()
}

func podsFromLabelQuery(selector map[string]string) ([]string, error) {
	var pods []string

	labels := []string{}
	for k, v := range selector {
		labels = append(labels, fmt.Sprintf("%s=%s", k, v))
	}

	u := &url.URL{Scheme: "http", Path: "/api/v1beta1/pods", Host: apiserver}
	q := u.Query()
	q.Set("labels", strings.Join(labels, ","))
	u.RawQuery = q.Encode()

	req := &http.Request{Method: "GET", URL: u}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Println("error retrieving pods:", err)
		return nil, errors.New("error retrieving pods")
	}

	if resp.StatusCode != http.StatusOK {
		log.Printf("error retrieving pods from %s", u)
		return nil, errors.New("non 200 status code")
	}
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var ps PodList
	err = json.Unmarshal(data, &ps)
	if err != nil {
		return nil, err
	}

	for _, p := range ps.Items {
		if p.CurrentState.Status == "Running" {
			pods = append(pods, p.CurrentState.PodIP)
		}
	}

	return pods, nil
}
