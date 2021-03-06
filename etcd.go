package main

import (
	"net"
	"strconv"
	"strings"

	"github.com/coreos/go-etcd/etcd"
	"github.com/miekg/dns"
)

func toPath(s string) string {
	l := dns.SplitDomainName(s)
	for i, j := 0, len(l)-1; i < j; i, j = i+1, j-1 {
		l[i], l[j] = l[j], l[i]
	}
	// TODO(miek): escape slashes in s.
	return strings.Join(l, "/")
}

// questionToPath converts a DNS question to a etcd key. If the questions looks
// like service.staging.skydns.local SRV, the resulting key
// will by /local/skydns/staging/service/SRV .
func questionToPath(q string, t uint16) string {
	return "/" + toPath(q) + "/" + dns.TypeToString[t]
}

func parseA(v string) (net.IP, error)    { return net.ParseIP(v).To4(), nil }
func parseAAAA(v string) (net.IP, error) { return net.ParseIP(v).To16(), nil }
func parseSRV(v string) (uint16, uint16, uint16, string, error) {
	p := strings.Split(v, " ") // Stored as space separated values.
	prio, _ := strconv.Atoi(p[0])
	weight, _ := strconv.Atoi(p[1])
	port, _ := strconv.Atoi(p[2])
	return uint16(prio), uint16(weight), uint16(port), p[3], nil
}

func parseValue(t uint16, value string, h dns.RR_Header) dns.RR {
	switch t {
	case dns.TypeA:
		a := new(dns.A)
		a.Hdr = h
		a.A, _ = parseA(value)
		return a
	case dns.TypeAAAA:
		aaaa := new(dns.AAAA)
		aaaa.Hdr = h
		aaaa.AAAA, _ = parseAAAA(value)
		return aaaa
	case dns.TypeSRV:
		srv := new(dns.SRV)
		srv.Hdr = h
		srv.Priority, srv.Weight, srv.Port, srv.Target, _ = parseSRV(value)
		return srv
	}
	return nil
}

func get(e *etcd.Client, q string, t uint16) ([]dns.RR, error) {
	path := questionToPath(q, t)
	r, err := e.Get(path, false, false)
	if err != nil {
		return nil, err
	}
	h := dns.RR_Header{Name: q, Rrtype: t, Class: dns.ClassINET, Ttl: 60} // Ttl is overridden
	rr := parseValue(t, r.Node.Value, h)
	return []dns.RR{rr}, nil
}
