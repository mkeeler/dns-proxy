package main

import (
	"context"
	"fmt"
	"net"

	"github.com/miekg/dns"

	"github.com/mkeeler/dns-proxy/internal/proto-gen/dnsproxy"
	"google.golang.org/grpc"
)

type DNSServer struct {
}

func newDNSServer() *DNSServer {
	return &DNSServer{}
}

func (d *DNSServer) Resolve(ctx context.Context, req *dnsproxy.ResolveRequest) (*dnsproxy.ResolveResponse, error) {
	var dnsRequest dns.Msg

	if err := dnsRequest.Unpack(req.Message); err != nil {
		return nil, fmt.Errorf("error decoding DNS request: %w", err)
	}

	resp := &dns.Msg{}
	resp.SetReply(&dnsRequest)
	resp.SetRcode(&dnsRequest, dns.RcodeSuccess)
	aRecord := &dns.A{
		Hdr: dns.RR_Header{
			Name:   dnsRequest.Question[0].Name,
			Rrtype: dns.TypeA,
			Class:  dns.ClassINET,
			Ttl:    uint32(5),
		},
		A: net.IPv4(1, 2, 3, 4),
	}

	resp.Answer = append(resp.Answer, aRecord)

	out, err := resp.Pack()
	if err != nil {
		return nil, fmt.Errorf("failure to pack DNS response: %w", err)
	}

	return &dnsproxy.ResolveResponse{
		Message: out,
	}, nil
}

func main() {
	lis, err := net.Listen("tcp", "127.0.0.1:8502")
	if err != nil {
		panic(err)
	}

	grpcServer := grpc.NewServer()

	srv := newDNSServer()

	dnsproxy.RegisterDnsServer(grpcServer, srv)
	grpcServer.Serve(lis)
}
