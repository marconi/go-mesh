package main

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"log"
	"net"
	"strings"
	"time"

	"github.com/marconi/go-mesh/gomesh/libs"
	"github.com/marconi/go-mesh/gomesh/utils"
	"github.com/nu7hatch/gouuid"
)

const (
	TTL        = 5
	PORT       = 6346
	USER_AGENT = "GoMesh/1.0"
)

var PAYLOAD = map[string]byte{
	"PING":     0x00,
	"PONG":     0x01,
	"BYE":      0x02,
	"PUSH":     0x40,
	"QUERY":    0x80,
	"QUERYHIT": 0x81,
}

type Message struct {
	GUID    *uuid.UUID
	p_type  byte
	ttl     byte
	hops    byte
	p_len   int
	payload []byte
}

type Servent struct {
	GUID      *uuid.UUID
	Port      int
	Hostcache *libs.HostCache
	Peer      net.Conn
}

func (s *Servent) CachePeers(t_peers string) {
	peers := strings.Split(t_peers, ",")
	for _, addr := range peers {
		s.Hostcache.Add(strings.Split(addr, " ")[0])
	}
	s.Hostcache.Save()
}

func (s *Servent) Handshake(conn net.Conn) error {
	fmt.Println("Handshaking...")
	payload := fmt.Sprintf("GNUTELLA CONNECT/0.6\r\n"+
		"User-Agent: %s\r\n"+
		"X-Ultrapeer: False\r\n"+
		"X-Query-Routing: 0.1\r\n"+
		"Pong-Caching: 0.1\r\n"+
		"GGEP: 0.5\r\n\r\n", USER_AGENT)
	fmt.Println(payload)
	if _, err := conn.Write([]byte(payload)); err != nil {
		return err
	}

	fmt.Println("Reading response...")
	buf := new(bytes.Buffer)
	b := make([]byte, 1024) //1kb per read
	for {
		n, err := conn.Read(b)
		if err != nil || n == 0 {
			break
		}
		buf.Write(b[:n])
	}

	if buf.Len() == 0 {
		return errors.New("Unable to read GNUTELLA CONNECT response")
	}

	raw_h := string(buf.Bytes())
	h := utils.ParseHeaders(raw_h)
	fmt.Printf("%#v\n", h)

	// save any ultra-peers we get
	if u_peers, ok := h["X-Try-Ultrapeers"]; ok {
		fmt.Println("Parsing x-ultra-peers...")
		s.CachePeers(u_peers)
	}

	// save any try-hubs we get
	if t_hubs, ok := h["X-Try"]; ok {
		fmt.Println("Parsing x-tries...")
		s.CachePeers(t_hubs)
	}

	// handle any non-200 and incompatible protocol response
	if !strings.Contains(h["Title"], "200") || !strings.Contains(h["Title"], "GNUTELLA/0.6") {
		return errors.New(h["Title"])
	}

	// if everything is fine, accept the hub
	fmt.Println("Accepting hub...")
	payload = "GNUTELLA/0.6 200 OK\r\n\r\n"
	fmt.Println(payload)
	if _, err := conn.Write([]byte(payload)); err != nil {
		return err
	}
	return nil
}

func (s *Servent) Bootstrap() error {
	for _, addr := range s.Hostcache.Items() {
		fmt.Println("Dialing: ", addr, "...")
		conn, err := net.DialTimeout("tcp", addr, time.Minute)
		if err != nil {
			s.Hostcache.Delete(addr)
			fmt.Println("Dial error on ", addr, ": ", err)
			continue
		}

		if err := s.Handshake(conn); err != nil {
			conn.Close()
			s.Hostcache.Delete(addr)
			fmt.Println("Handshake error on ", addr, ": ", err)
			continue
		}

		s.Peer = conn
		break
	}

	if s.Peer == nil {
		return errors.New("No peer found.")
	}

	fmt.Println("Peer found: ", s.Peer.RemoteAddr())
	return nil
}

func NewServent() *Servent {
	hc, err := libs.NewHostCache()
	if err != nil {
		log.Fatalln(err)
	}

	return &Servent{
		GUID:      utils.GenPeerId(),
		Port:      PORT,
		Hostcache: hc,
	}
}

/**
 * Payload type messages
 */

func (s *Servent) Ping() error {
	fmt.Println("Pinging...")
	buf := new(bytes.Buffer)
	buf.Write(s.GUID[:])
	buf.WriteByte(byte(PAYLOAD["PING"]))
	buf.WriteByte(byte(TTL))
	buf.WriteByte(byte(0)) // hops

	// payload length
	p_len := make([]byte, 3)
	binary.LittleEndian.PutUint16(p_len, uint16(0))
	buf.Write(p_len)

	if _, err := s.Peer.Write(buf.Bytes()); err != nil {
		return errors.New(fmt.Sprintf("Unable to write PING: ", err))
	}

	fmt.Println("Waiting for PONG...")
	buf = new(bytes.Buffer)
	b := make([]byte, 1024) // 1kb
	for {
		n, err := s.Peer.Read(b)
		if err != nil || n == 0 {
			return errors.New(fmt.Sprintf("Unable to read PONG: ", err))
		}
		buf.Write(b[:n])
	}

	fmt.Println("Got: ", string(buf.Bytes()))
	return nil
}
func (s *Servent) Pong(conn net.Conn) {}
func (s *Servent) Query()             {}
func (s *Servent) QueryHit()          {}
func (s *Servent) Push()              {}
func (s *Servent) Bye()               {}
func (s *Servent) ParseMessage(raw_msg []byte) *Message {
	guid, err := uuid.Parse(raw_msg[:16])
	if err != nil {
		fmt.Println("Unable to parse GUID: ", err)
	}
	p_type := raw_msg[16]
	ttl := raw_msg[17]
	hops := raw_msg[18]
	p_len := binary.LittleEndian.Uint16(raw_msg[19:22])
	payload := new(bytes.Buffer)

	if p_len > 0 {
		payload.Write(raw_msg[22:])
	}

	return &Message{
		GUID:    guid,
		p_type:  p_type,
		ttl:     ttl,
		hops:    hops,
		p_len:   int(p_len),
		payload: payload.Bytes(),
	}
}

func main() {
	servent := NewServent()
	if err := servent.Bootstrap(); err != nil {
		log.Fatalln(err)
	}
}
