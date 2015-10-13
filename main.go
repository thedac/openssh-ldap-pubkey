package main

import (
	"flag"
	"fmt"
	"log"

	"crypto/tls"
	"crypto/x509"

	"gopkg.in/ldap.v2"
)

const (
	sshPublicKeyName = "sshPublicKey"
)

type ldapEnv struct {
	host   string
	port   int
	base   string
	filter string
	tls    bool
	uid    string
}

func (l *ldapEnv) argparse() {
	h := flag.String("host", l.host, "LDAP server host")
	p := flag.Int("port", l.port, "LDAP server port")
	b := flag.String("base", l.base, "search base")
	f := flag.String("filter", l.filter, "search filter")
	t := flag.Bool("tls", l.tls, "LDAP connect over TLS")
	flag.Parse()

	if l.host != *h {
		l.host = *h
	}
	if l.port != *p {
		l.port = *p
	}
	if l.base != *b {
		l.base = *b
	}
	if l.filter != *f {
		l.filter = *f
	}
	if l.tls != *t {
		l.tls = *t
	}

	if len(flag.Args()) != 1 {
		log.Fatal("Specify username")
	}
	l.uid = flag.Args()[0]
}

func main() {
	l := &ldapEnv{}
	l.loadNslcdConf()
	l.argparse()

	c := &ldap.Conn{}
	var err error
	if l.tls {
		certs := *x509.NewCertPool()
		tlsConfig := &tls.Config{
			RootCAs: &certs,
		}

		// ToDo: check FQDN or IP address, set true when IPAddress.
		if true {
			tlsConfig.InsecureSkipVerify = true
		}

		c, err = ldap.DialTLS("tcp", fmt.Sprintf("%s:%d", l.host, l.port), tlsConfig)
		if err != nil {
			log.Fatal(err)
		}
		defer c.Close()
	} else {
		c, err = ldap.Dial("tcp", fmt.Sprintf("%s:%d", l.host, l.port))
		if err != nil {
			log.Fatal(err)
		}
		defer c.Close()
	}

	bindRequest := ldap.NewSimpleBindRequest("", "", nil)
	_, err = c.SimpleBind(bindRequest)
	if err != nil {
		log.Fatal(err)
	}

	searchRequest := ldap.NewSearchRequest(
		l.base, ldap.ScopeWholeSubtree, ldap.NeverDerefAliases,
		0, 0, false,
		fmt.Sprintf(l.filter, l.uid), []string{sshPublicKeyName}, nil)
	sr, err := c.Search(searchRequest)
	if err != nil {
		log.Fatal(err, sshPublicKeyName)
	}
	if len(sr.Entries) != 1 {
		log.Fatal("User does not exist or too many entries returned")
	}

	if len(sr.Entries[0].GetAttributeValues("sshPublicKey")) == 0 {
		log.Fatal("User does not use ldapPublicKey.")
	}
	for _, pubkey := range sr.Entries[0].GetAttributeValues("sshPublicKey") {
		fmt.Println(pubkey)
	}
}