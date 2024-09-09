package discovery

import (
	"net"
	"testing"
)

func TestGetPortsWith(t *testing.T) {
	ports, err := GetPortsWithErr(3)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if len(ports) != 3 {
		t.Fatal("expected to get 3 ports")
	}
	for _, port := range ports {
		ln, err := net.ListenTCP("tcp", &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: port})
		if err != nil {
			t.Fatal("expected port to be free")
		}
		ln.Close()
	}
}

func TestGetPorts(t *testing.T) {
	ports := GetPorts(3)
	if len(ports) != 3 {
		t.Fatal("expected to get 3 ports")
	}
	for _, port := range ports {
		ln, err := net.ListenTCP("tcp", &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: port})
		if err != nil {
			t.Fatal("expected port to be free")
		}
		ln.Close()
	}
}

func TestGetSPorts(t *testing.T) {
	ports := GetSPorts(3)
	if len(ports) != 3 {
		t.Fatal("expected to get 3 ports")
	}
}
