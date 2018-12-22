package main

import (
	"encoding/json"
	"io/ioutil"

	"github.com/tsocial/catoolkit/tlsproxy"
	"github.com/tsocial/srelapd/ldap"
	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	cfgFile    = kingpin.Flag("config", "Config File").Required().File()
	listen     = kingpin.Flag("listen", "Listening on port").Default("389").String()
	caCert     = kingpin.Flag("ca-cert-file", "CA Cert File").String()
	certFile   = kingpin.Flag("cert-file", "Cert File").String()
	keyFile    = kingpin.Flag("key-file", "Key File").String()
	serverName = kingpin.Flag("server-name", "ServerName in tls Config").
			Default("localhost").String()
)

// interface for backend handler
type Backend interface {
	ldap.Binder
	ldap.Searcher
	ldap.Closer
}

type configUser struct {
	Name         string
	OtherGroups  []int
	PassSHA256   string
	PrimaryGroup int
	SSHKeys      []string
	OTPSecret    string
	Disabled     bool
	UnixID       int
	Mail         string
	LoginShell   string
	GivenName    string
	SN           string
	Homedir      string
}

type configGroup struct {
	Name          string
	UnixID        int
	IncludeGroups []int
}

type config struct {
	BaseDN string
	Groups []configGroup
	Users  []configUser
}

func main() {
	kingpin.Parse()

	cfg, err := doConfig()
	if err != nil {
		panic(err)
	}

	// configure the backend
	s := ldap.NewServer()
	handler := newConfigHandler(cfg)

	s.BindFunc("", handler)
	s.SearchFunc("", handler)
	s.CloseFunc("", handler)

	params := &tlsproxy.TlsParams{
		HardFail:   false,
		CACertFile: *caCert,
		CertFile:   *certFile,
		KeyFile:    *keyFile,
		ServerName: *serverName,
	}

	if *caCert == "" || *certFile == "" || *keyFile == "" {
		params.SkipTls = true
	}

	s.ListenAndServe(*listen, params)
}

// doConfig reads the cli flags and config file
func doConfig() (*config, error) {
	cfg := config{}

	// defer the closing of our jsonFile so that we can parse it later on
	defer (*cfgFile).Close()

	b, _ := ioutil.ReadAll(*cfgFile)

	if err := json.Unmarshal([]byte(b), &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
