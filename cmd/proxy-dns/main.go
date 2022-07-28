package main

import (
	"context"
	"fmt"
	"net"

	"github.com/mkeeler/dns-proxy/internal/proto-gen/dnsproxy"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	addr := net.UDPAddr{
		Port: 8600,
		IP:   net.ParseIP("127.0.0.1"),
	}

	conn, err := net.ListenUDP("udp", &addr)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	grpcConn, err := grpc.Dial("localhost:8502", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		panic(err)
	}

	proxyDNS(conn, dnsproxy.NewDnsClient(grpcConn))
}

func proxyDNS(conn *net.UDPConn, client dnsproxy.DnsClient) {
	buf := make([]byte, 1520)

	for {
		_, addr, err := conn.ReadFrom(buf)
		if err != nil {
			fmt.Printf("error reading from conn: %v\n", err)
			continue
		}

		req := &dnsproxy.ResolveRequest{
			Message: buf,
		}

		resp, err := client.Resolve(context.Background(), req)
		if err != nil {
			fmt.Printf("error resolving request: %v\n", err)
			continue
		}

		_, err = conn.WriteTo(resp.Message, addr)
		if err != nil {
			fmt.Printf("error sending response: %v\n", err)
		}
	}
}
