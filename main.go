package main

import (
	"github.com/miekg/dns"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"net"
)

var (
	bindAddr string

	mainCmd = &cobra.Command{}

	serveCmd = &cobra.Command{
		Use:   "serve",
		Short: "Serve as a simple DNS A resolver/forwarder",
		Run:   runDns,
	}
)

func runDns(cmd *cobra.Command, args []string) {
	dns.HandleFunc(".", handler)
	log.Fatalln(dns.ListenAndServe(bindAddr, "udp", nil))
}

func handler(w dns.ResponseWriter, r *dns.Msg) {
	m := new(dns.Msg)
	m.SetReply(r)
	m.RecursionAvailable = true
	m.RecursionDesired = true
	for _, q := range r.Question {
		switch q.Qtype {
		case dns.TypeA:
			ips, err := net.LookupIP(q.Name)
			if err != nil {
				log.Warnf("Lookup '%s' failed: %s", q.Name, err.Error())
			} else {
				for _, ip := range ips {
					ip4 := ip.To4()
					if ip4 == nil {
						continue
					}
					rr := new(dns.A)
					rr.Hdr = dns.RR_Header{Name: q.Name, Rrtype: dns.TypeA, Class: dns.ClassINET, Ttl: 0}
					rr.A = ip.To4()
					m.Answer = append(m.Answer, rr)
				}
			}
		}
	}
	w.WriteMsg(m)
}

func main() {
	mainCmd.AddCommand(serveCmd)
	serveCmd.PersistentFlags().StringVarP(&bindAddr, "bind", "b", ":8053", "Bind address. Address and port to bind and serve requests on.")
	mainCmd.Execute()
}
