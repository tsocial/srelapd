package ldap

import (
	"io"
	"log"
	"net"
	"strings"

	"github.com/nmcclain/asn1-ber"
	"github.com/tsocial/catoolkit/tlsproxy"
)

type Binder interface {
	Bind(bindDN, bindSimplePw string, conn net.Conn) (LDAPResultCode, error)
}
type Searcher interface {
	Search(boundDN string, req SearchRequest, conn net.Conn) (ServerSearchResult, error)
}
type Unbinder interface {
	Unbind(boundDN string, conn net.Conn) (LDAPResultCode, error)
}
type Closer interface {
	Close(boundDN string, conn net.Conn) error
}

//
type Server struct {
	BindFns   map[string]Binder
	SearchFns map[string]Searcher
	UnbindFns map[string]Unbinder
	CloseFns  map[string]Closer
}

type ServerSearchResult struct {
	Entries    []*Entry
	Referrals  []string
	Controls   []Control
	ResultCode LDAPResultCode
}

//
func NewServer() *Server {
	s := new(Server)

	d := defaultHandler{}
	s.BindFns = make(map[string]Binder)
	s.SearchFns = make(map[string]Searcher)
	s.UnbindFns = make(map[string]Unbinder)
	s.CloseFns = make(map[string]Closer)
	s.BindFunc("", d)
	s.SearchFunc("", d)
	s.UnbindFunc("", d)
	s.CloseFunc("", d)
	return s
}
func (server *Server) BindFunc(baseDN string, f Binder) {
	server.BindFns[baseDN] = f
}
func (server *Server) SearchFunc(baseDN string, f Searcher) {
	server.SearchFns[baseDN] = f
}
func (server *Server) UnbindFunc(baseDN string, f Unbinder) {
	server.UnbindFns[baseDN] = f
}
func (server *Server) CloseFunc(baseDN string, f Closer) {
	server.CloseFns[baseDN] = f
}

func (server *Server) ListenAndServe(listen string, p *tlsproxy.TlsParams) {
	tlsproxy.RunWithParams(listen, func(conn net.Conn) {
		defer Handler(func(err error) {
			log.Println(err)
		})
		server.handleConnection(conn)
	}, p)
}

//
func (server *Server) handleConnection(conn net.Conn) {
	boundDN := "" // "" == anonymous

handler:
	for {
		// read incoming LDAP packet
		packet, err := ber.ReadPacket(conn)
		if err == io.EOF { // Client closed connection
			break
		} else if err != nil {
			log.Printf("handleConnection ber.ReadPacket ERROR: %s", err.Error())
			break
		}

		// sanity check this packet
		if len(packet.Children) < 2 {
			log.Print("len(packet.Children) < 2")
			break
		}
		// check the message ID and ClassType
		messageID, ok := packet.Children[0].Value.(uint64)
		if !ok {
			log.Print("malformed messageID")
			break
		}
		req := packet.Children[1]
		if req.ClassType != ber.ClassApplication {
			log.Print("req.ClassType != ber.ClassApplication")
			break
		}
		// handle controls if present
		controls := []Control{}
		if len(packet.Children) > 2 {
			for _, child := range packet.Children[2].Children {
				controls = append(controls, DecodeControl(child))
			}
		}

		//log.Printf("DEBUG: handling operation: %s [%d]", ApplicationMap[req.Tag], req.Tag)
		//ber.PrintPacket(packet) // DEBUG

		// dispatch the LDAP operation
		switch req.Tag { // ldap op code
		case ApplicationBindRequest:
			ldapResultCode := HandleBindRequest(req, server.BindFns, conn)
			if ldapResultCode == LDAPResultSuccess {
				boundDN, ok = req.Children[1].Value.(string)
				if !ok {
					log.Printf("Malformed Bind DN")
					break handler
				}
			}
			responsePacket := encodeBindResponse(messageID, ldapResultCode)
			if err = sendPacket(conn, responsePacket); err != nil {
				log.Printf("sendPacket error %s", err.Error())
				break handler
			}
		case ApplicationSearchRequest:
			if err := HandleSearchRequest(req, &controls, messageID, boundDN, server, conn); err != nil {
				log.Printf("handleSearchRequest error %s", err.Error()) // TODO: make this more testable/better err handling - stop using log, stop using breaks?
				e := err.(*Error)
				if err = sendPacket(conn, encodeSearchDone(messageID, e.ResultCode)); err != nil {
					log.Printf("sendPacket error %s", err.Error())
					break handler
				}
				break handler
			} else {
				if err = sendPacket(conn, encodeSearchDone(messageID, LDAPResultSuccess)); err != nil {
					log.Printf("sendPacket error %s", err.Error())
					break handler
				}
			}
		case ApplicationUnbindRequest:
			break handler // simply disconnect
		}
	}

	for _, c := range server.CloseFns {
		c.Close(boundDN, conn)
	}

	conn.Close()
}

//
func sendPacket(conn net.Conn, packet *ber.Packet) error {
	_, err := conn.Write(packet.Bytes())
	if err != nil {
		log.Printf("Error Sending Message: %s", err.Error())
		return err
	}
	return nil
}

//
func routeFunc(dn string, funcNames []string) string {
	bestPick := ""
	for _, fn := range funcNames {
		if strings.HasSuffix(dn, fn) {
			l := len(strings.Split(bestPick, ","))
			if bestPick == "" {
				l = 0
			}
			if len(strings.Split(fn, ",")) > l {
				bestPick = fn
			}
		}
	}
	return bestPick
}

//
type defaultHandler struct {
}

func (h defaultHandler) Bind(bindDN, bindSimplePw string, conn net.Conn) (LDAPResultCode, error) {
	return LDAPResultInvalidCredentials, nil
}
func (h defaultHandler) Search(boundDN string, req SearchRequest, conn net.Conn) (ServerSearchResult, error) {
	return ServerSearchResult{make([]*Entry, 0), []string{}, []Control{}, LDAPResultSuccess}, nil
}
func (h defaultHandler) Unbind(boundDN string, conn net.Conn) (LDAPResultCode, error) {
	return LDAPResultSuccess, nil
}
func (h defaultHandler) Close(boundDN string, conn net.Conn) error {
	conn.Close()
	return nil
}
