package dns

import (
	"fmt"

	"github.com/miekg/dns"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

type Server struct {
	port     int
	protocol string

	resolver *Resolver
	server   *dns.Server
}

func NewServer(protocol string, port int, resolver *Resolver) *Server {
	server := &dns.Server{Addr: fmt.Sprintf(":%d", port), Net: protocol}
	ds := &Server{
		port:     port,
		protocol: protocol,
		resolver: resolver,
		server:   server,
	}
	server.Handler = dns.HandlerFunc(ds.handleRequest)
	return ds
}

func (ds *Server) handleRequest(w dns.ResponseWriter, req *dns.Msg) {
	log.Log.Info("dns server receive", "request", req)
	ret := new(dns.Msg)
	ret.SetReply(req)

	if len(req.Question) == 0 {
		ret.Rcode = dns.RcodeFormatError
		if err := w.WriteMsg(ret); err != nil {
			log.Log.Error(err, "dns server failed to write msg")
		}
		return
	}
	// handle query, only support A, AAAA and CNAME
	q := req.Question[0]
	switch q.Qtype {
	case dns.TypeA, dns.TypeAAAA, dns.TypeCNAME:
		answers, err := ds.resolver.Resolve(q.Qtype, q.Name)
		if err != nil {
			ret.Rcode = dns.RcodeServerFailure
			if err := w.WriteMsg(ret); err != nil {
				log.Log.Error(err, "dns server failed to write msg")
			}
			return
		}
		ret.Answer = answers
		ret.Authoritative = true
		// if len(ret.Answer) == 0 {
		// 	ret.Rcode = dns.RcodeNameError
		// }
	default:
		ret.Rcode = dns.RcodeRefused
	}

	log.Log.Info("dns response", "request", req, "response", ret)
	if err := w.WriteMsg(ret); err != nil {
		log.Log.Error(err, "dns server failed to write msg")
	}
}

func (ds *Server) Serve() error {
	return ds.server.ListenAndServe()
}

func (ds *Server) Shutdown() {
	ds.server.Shutdown()
}
