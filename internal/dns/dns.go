package dns

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"os"
	"strings"

	"github.com/miekg/dns"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// Resolver for dns
type Resolver struct {
	clientConfig *dns.ClientConfig
}

// NewResolver create a new resolver
// nameServers default use /etc/resolv.conf
func NewResolver(nameServers []string) (*Resolver, error) {
	r := &Resolver{}
	var reader io.Reader
	if len(nameServers) == 0 {
		data, err := os.ReadFile("/etc/resolv.conf")
		if err != nil {
			return nil, err
		}
		reader = bytes.NewReader(data)
	} else {
		buffer := bytes.NewBuffer(nil)
		for _, nameServer := range nameServers {
			if nameServer == "" {
				return nil, fmt.Errorf("empty name server")
			}
			buffer.WriteString(fmt.Sprintf("nameserver %s\n", nameServer))
		}
		reader = buffer
	}
	clientConfig, err := dns.ClientConfigFromReader(reader)
	if err != nil {
		return nil, err
	}
	r.clientConfig = clientConfig

	return r, nil
}

// Resolve resolve domain to ip, questionType support dns.TypeA, dns.TypeAAAA and dns.TypeCNAME.
func (r *Resolver) Resolve(questionType uint16, name string) ([]dns.RR, error) {
	if !strings.HasSuffix(name, ".") {
		name += "."
	}

	switch questionType {
	case dns.TypeA, dns.TypeAAAA:
		answers := r.exchange(questionType, name)
		return answers, nil
	case dns.TypeCNAME:
		answers := r.exchange(questionType, name)
		return answers, nil
	default:
		return nil, fmt.Errorf("unsupported question type: %d", questionType)
	}
}

func (r *Resolver) exchange(questionType uint16, name string) []dns.RR {
	client := &dns.Client{}
	m := new(dns.Msg)
	m.SetQuestion(name, questionType)

	answer := make([]dns.RR, 0)
	for _, server := range r.clientConfig.Servers {
		in, _, err := client.Exchange(m, net.JoinHostPort(server, r.clientConfig.Port))
		if err != nil {
			log.Log.Error(err, "dns client exchange error", "server", server)
			continue
		}
		for _, a := range in.Answer {
			switch a.Header().Rrtype {
			case dns.TypeA, dns.TypeAAAA:
				answer = append(answer, a)
			case dns.TypeCNAME:
				if questionType == dns.TypeCNAME || questionType == dns.TypeA || questionType == dns.TypeAAAA {
					answer = append(answer, a)
				}
			}
		}
		if len(answer) > 0 {
			break
		}
	}
	return answer
}
