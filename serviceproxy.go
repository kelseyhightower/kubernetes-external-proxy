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

type ServiceManager struct {
	mu sync.Mutex
	m  map[string]*ServiceProxy
}

func newServiceManager() *ServiceManager {
	m := make(map[string]*ServiceProxy)
	return &ServiceManager{m: m}
}

func (sm *ServiceManager) AddService(args *Service, reply *int) error {
	log.Println("AddService called")
	/*
		if err := sm.add(args); err != nil {
			*reply = 1
			return err
		}
	*/
	if err := sm.hello("foo"); err != nil {
		*reply = 1
		return err
	}
	log.Println("got this far")
	// Never makes it to here, unless I remove the call the sm.add.
	// but the call to sm.add seems to work without error.
	*reply = 0
	return nil
}

func (sm *ServiceManager) hello(s string) error {
	log.Println(s)
	return nil
}

func (sm *ServiceManager) add(service *Service) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sp := newServiceProxy(service)
	err := sp.start()
	if err != nil {
		log.Printf("error adding service %s: %v\n", service.ID, err)
		return err
	}
	sm.m[service.ID] = sp
	return nil
}

type ServiceProxy struct {
	service *Service
	sync.Mutex
	pods []string
	r    *ring.Ring
}

func newServiceProxy(service *Service) *ServiceProxy {
	return &ServiceProxy{service: service}
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
	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Printf("error accepting connections for serviceID %s", sp.service.ID)
			time.Sleep(time.Duration(10 * time.Second))
			continue
		}
		go sp.handleConnection(conn)
	}
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
	sp.r = sp.r.Next()
	return net.JoinHostPort(sp.r.Value.(string), sp.service.ContainerPort)
}

func (sp *ServiceProxy) handleConnection(conn net.Conn) {
	hostPort := sp.nextPod()
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
	resp, err := http.Get(fmt.Sprintf("%s/api/v1beta1/pods?labels=environment=prod", apiserver))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
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
