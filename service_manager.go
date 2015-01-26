package main

import (
	"fmt"
	"log"
	"net"
	"sync"
)

type ServiceManager struct {
	mu sync.Mutex
	m  map[string]*ServiceProxy
}

func newServiceManager() *ServiceManager {
	m := make(map[string]*ServiceProxy)
	return &ServiceManager{m: m}
}

func (sm *ServiceManager) Add(args *Service, reply *string) error {
	if err := sm.add(args); err != nil {
		log.Println(err)
		return err
	}
	*reply = net.JoinHostPort(bindIP, args.Port)
	log.Println("new service added", args)
	return nil
}

func (sm *ServiceManager) Del(args *string, reply *bool) error {
	if err := sm.del(*args); err != nil {
		log.Println(err)
		return err
	}
	*reply = true
	log.Println("service service deleted", args)
	return nil
}

func (sm *ServiceManager) add(service *Service) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	if _, ok := sm.m[service.ID]; ok {
		err := fmt.Errorf("service %s already exist", service.ID)
		log.Println(err)
		return err
	}
	sp := newServiceProxy(service)
	err := sp.start()
	if err != nil {
		log.Printf("error adding service %s: %v\n", service.ID, err)
		return err
	}
	sm.m[service.ID] = sp
	return nil
}

func (sm *ServiceManager) del(id string) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	if v, ok := sm.m[id]; ok {
		delete(sm.m, id)
		if err := v.stop(); err != nil {
			return err
		}
	}
	return nil
}
