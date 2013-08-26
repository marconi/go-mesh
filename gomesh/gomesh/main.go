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

	"github.com/marconi/go-mesh/gomesh"
	"github.com/marconi/go-mesh/gomesh/utils"
	"github.com/nu7hatch/gouuid"
)

const (
	TTL        = 5
	PORT       = 6346
	G2_MIME    = "application/x-gnutella2"
	USER_AGENT = "GoMesh 0.1"
)

var Payload = map[string]byte{
	"PING":     0x00,
	"PONG":     0x01,
	"BYE":      0x02,
	"PUSH":     0x40,
	"QUERY":    0x80,
	"QUERYHIT": 0x81,
}

type Peer struct {
	Guid *uuid.UUID
	Port int
	// Peers     map[net.Conn]Peer
	// Listener  net.Listener
	Hostcache *gomesh.HostCache
	Hub       net.Conn
}

type Message struct {
	Guid    *uuid.UUID
	p_type  byte
	ttl     byte
	hops    byte
	p_len   int
	payload []byte
}

func (p *Peer) CachePeers(t_peers string) {
	peers := strings.Split(t_peers, ",")
	for _, addr := range peers {
		p.Hostcache.Add(strings.Split(addr, " ")[0])
	}
	p.Hostcache.Save()
}

func (p *Peer) Connect(conn net.Conn) error {
	fmt.Println("Getting IP...")
	ip, err := utils.GetNetIP()
	if err != nil {
		return err
	}

	fmt.Println("Connecting...")
	l_ip := fmt.Sprintf("%s:%d", ip, PORT)
	r_ip := fmt.Sprintf("%s", conn.RemoteAddr())
	payload := fmt.Sprintf("GNUTELLA CONNECT/0.6\r\n"+
		"Listen-IP: %s\r\n"+
		"Remote-IP: %s\r\n"+
		"User-Agent: %s\r\n"+
		"Accept: application/x-gnutella2\r\n"+
		"X-Hub: False\r\n\r\n", l_ip, r_ip, USER_AGENT)

	fmt.Println(payload, "\n\n")
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
		fmt.Println("Parsing ultra-peers...")
		p.CachePeers(u_peers)
	}

	// save any try-hubs we get
	if t_hubs, ok := h["X-Try-Hubs"]; ok {
		fmt.Println("Parsing try-hubs...")
		p.CachePeers(t_hubs)
	}

	// handle any non-200 response
	if !strings.Contains(h["Title"], "200") {
		return errors.New(h["Title"])
	}

	// if we have a 200 response,
	// check that its reply has g2 header
	if h["Content-Type"] != G2_MIME {
		return errors.New(fmt.Sprintf("Invalid content-type header: ", h["Content-Type"]))
	}

	// check that it accepts g2 header
	if h["Accept"] != G2_MIME {
		return errors.New(fmt.Sprintf("Invalid accept header: ", h["Accept"]))
	}

	// check that its a hub or an ultra-peer, we don't connect to leaf
	hub_found := false
	if x_hub, ok := h["X-Hub"]; !hub_found && ok && x_hub == "True" {
		hub_found = true
	}
	if u_peer, ok := h["X-Ultrapeer"]; !hub_found && ok && u_peer == "True" {
		hub_found = true
	}

	if !hub_found {
		return errors.New("Peer not a hub or ultra-peer")
	}

	// OPTIONAL: check for deflate encoding

	// if everything is fine, accept the hub
	fmt.Println("Accepting hub...")
	payload = fmt.Sprintf("GNUTELLA/0.6 200 OK\r\n"+
		"Content-Type: %s\r\n"+
		"X-Hub: False\r\n\r\n", G2_MIME)

	fmt.Println(payload, "\n\n")
	if _, err := conn.Write([]byte(payload)); err != nil {
		return err
	}
	return nil
}

func (p *Peer) Ping() error {
	// TODO: update this to G2 spec.

	fmt.Println("Pinging...")
	buf := new(bytes.Buffer)
	buf.Write(p.Guid[:])
	buf.WriteByte(byte(Payload["PING"]))
	buf.WriteByte(byte(TTL))
	buf.WriteByte(byte(0)) // hops

	// payload length
	p_len := make([]byte, 3)
	binary.LittleEndian.PutUint16(p_len, uint16(0))
	buf.Write(p_len)

	if _, err := p.Hub.Write(buf.Bytes()); err != nil {
		return errors.New(fmt.Sprintf("Unable to write PING: ", err))
	}

	fmt.Println("Waiting for PONG...")
	buf = new(bytes.Buffer)
	b := make([]byte, 1024) // 1kb
	for {
		n, err := p.Hub.Read(b)
		if err != nil || n == 0 {
			return errors.New(fmt.Sprintf("Unable to read PONG: ", err))
		}
		buf.Write(b[:n])
	}

	fmt.Println("Got: ", string(buf.Bytes()))
	return nil
}

func (p *Peer) Pong(conn net.Conn)             {}
func (p *Peer) Query()                         {}
func (p *Peer) QueryHit()                      {}
func (p *Peer) Push()                          {}
func (p *Peer) Bye()                           {}
func (p *Peer) ValidatePing(msg *Message) bool { return false }

func (p *Peer) ParseMessage(raw_msg []byte) *Message {
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
		Guid:    guid,
		p_type:  p_type,
		ttl:     ttl,
		hops:    hops,
		p_len:   int(p_len),
		payload: payload.Bytes(),
	}
}

// func (p *Peer) HandleConn(conn net.Conn) {
// 	result := new(bytes.Buffer)
// 	buf := make([]byte, 4119) // read 4119kb = 4kb payload + 23b header
// 	for {
// 		fmt.Println("Waiting for response...")
// 		n, err := conn.Read(buf)
// 		if err != nil || n == 0 {
// 			fmt.Println("Error reading: ", err)
// 			conn.Close()
// 			break
// 		}
// 		result.Write(buf[:n])
// 		msg := p.ParseMessage(result.Bytes())
// 		switch msg.p_type {
// 		case Payload["PING"]:
// 			fmt.Println("Got PINGED!")
// 			if p.ValidatePing(msg) {
// 				p.Pong(conn)
// 			}
// 		case Payload["PONG"]:
// 			fmt.Println("Got PONGED!")
// 		default:
// 			fmt.Println("Unknown message: ", msg)
// 		}
// 	}
// }

// func (p *Peer) Start() {
//  if len(p.Peers) == 0 {
//      fmt.Println("No peers found!")
//  }

//  for _, addr := range p.Hostcache.Items() {
//      port := strings.Split(addr, ":")[1]
//      listener, err := net.Listen("tcp", fmt.Sprintf(":%s", port))
//      if err != nil {
//          continue
//      }

//      fmt.Fprintf(os.Stdout, "%s: listening at port %s\n", p.Guid, port)
//      p.Listener = listener
//      break
//  }

//  if p.Listener == nil {
//      log.Fatalln("Unable to start servant.")
//  }

//  for {
//      conn, err := p.Listener.Accept()
//      if err != nil {
//          log.Println("Unable to accept: ", err)
//          continue
//      }
//      go p.HandleConn(conn)
//  }
// }

func (p *Peer) FindHub() error {
	for _, addr := range p.Hostcache.Items() {
		fmt.Println("Dialing: ", addr, "...")
		conn, err := net.DialTimeout("tcp", addr, time.Minute)
		if err != nil {
			p.Hostcache.Delete(addr)
			fmt.Println("Dial error on ", addr, ": ", err)
			continue
		}

		if err := p.Connect(conn); err != nil {
			p.Hostcache.Delete(addr)
			fmt.Println("Connect error on ", addr, ": ", err)
			continue
		}

		p.Hub = conn
		break
	}

	if p.Hub == nil {
		return errors.New("No hub found.")
	}

	fmt.Println("Connected to: ", p.Hub.RemoteAddr())
	return nil
}

func NewPeer(hc *gomesh.HostCache) *Peer {
	return &Peer{
		Guid:      utils.GenPeerId(),
		Port:      PORT,
		Hostcache: hc,
	}
}

func main() {
	hc, err := gomesh.NewHostCache()
	if err != nil {
		log.Fatalln(err)
	}

	p := NewPeer(hc)
	if err := p.FindHub(); err != nil {
		log.Fatalln(err)
	}

	if err := p.Ping(); err != nil {
		log.Fatalln(err)
	}

}
