package main

import "time"

func main() {
	s := &Service{
		ID:            "foo",
		ContainerPort: "80",
		Protocol:      "tcp",
		Port:          "5000",
		Selector:      map[string]string{"foo": "bar"},
	}

	sm := newServiceManager("0.0.0.0")
	sm.add(s)
	time.Sleep(time.Duration(1000 * time.Minute))
}
