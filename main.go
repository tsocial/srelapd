package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"

	"github.com/nmcclain/ldap"
	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	cfgFile = kingpin.Flag("config", "Config File").Required().File()
	port    = kingpin.Flag("port", "Port").Required().Int()
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
	s.EnforceLDAP = true
	handler := newConfigHandler(cfg)

	s.BindFunc("", handler)
	s.SearchFunc("", handler)
	s.CloseFunc("", handler)

	log.Printf("Frontend LDAP server listening on %v\n", *port)

	if err := s.ListenAndServe(fmt.Sprintf("0.0.0.0:%v", *port)); err != nil {
		panic(err)
	}
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
