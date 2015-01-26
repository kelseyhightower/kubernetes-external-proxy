package main

import (
	"errors"
	"log"
	"net"
	"sync"
)

// A ServiceManager manages service proxies.
type ServiceManager struct {
	mu sync.Mutex
	m  map[string]*ServiceProxy
}

func newServiceManager() *ServiceManager {
	m := make(map[string]*ServiceProxy)
	return &ServiceManager{m: m}
}

// Add creates a new ServiceProxy based on args.
func (sm *ServiceManager) Add(args *Service, reply *string) error {
	if err := sm.add(args); err != nil {
		log.Println(err)
		return err
	}
	*reply = net.JoinHostPort(bindIP, args.Port)
	log.Printf("%s service added", args.ID)
	return nil
}

// Del deletes the ServiceProxy based on the given service ID (args).
// The ServiceProxy is stopped after active connections have terminated.
func (sm *ServiceManager) Del(args *string, reply *bool) error {
	if err := sm.del(*args); err != nil {
		log.Println(err)
		return err
	}
	*reply = true
	log.Printf("%s service deleted", *args)
	return nil
}

func (sm *ServiceManager) add(service *Service) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	if _, ok := sm.m[service.ID]; ok {
		err := errors.New("service already exist")
		log.Printf("error adding the %s service: %v", err)
		return err
	}
	sp := newServiceProxy(service)
	if err := sp.start(); err != nil {
		log.Printf("error adding the %s service: %v", service.ID, err)
		return err
	}
	sm.m[service.ID] = sp
	return nil
}

func (sm *ServiceManager) del(id string) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	if v, ok := sm.m[id]; ok {
		if err := v.stop(); err != nil {
			log.Printf("error stopping the %s service: %v", id, err)
			return err
		}
		delete(sm.m, id)
	}
	return nil
}
