package tafgo

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

const (
	JCENORMAL = uint8(0)
	JCEONEWAY = uint8(1)
)

var ErrTafRPCTimeout = errors.New("Taf RPC timeout")

type rpcSession struct {
	ID int32
	ch chan *ResponsePacket
}

type rpcChannel struct {
	Conn    net.Conn
	ch      chan TafStruct
	idx     int
	counter int64
	running bool
}

func parseEndpoint(s string) (EndpointF, error) {
	e := EndpointF{}
	fields := strings.Fields(s)
	if strings.EqualFold(fields[0], "tcp") {
		e.Istcp = 1
	}
	restField := fields[1:]
	for i := 0; i < len(restField); i += 2 {
		switch restField[i] {
		case "-h":
			e.Host = restField[i+1]
		case "-p":
			i, err := strconv.ParseInt(restField[i+1], 10, 32)
			if err != nil {
				return e, err
			}
			e.Port = int32(i)
		case "-t":
			i, err := strconv.ParseInt(restField[i+1], 10, 32)
			if err != nil {
				return e, err
			}
			e.Timeout = int32(i)
		case "-g":
			i, err := strconv.ParseInt(restField[i+1], 10, 32)
			if err != nil {
				return e, err
			}
			e.Grid = int32(i)
		case "-f":
			i, err := strconv.ParseInt(restField[i+1], 10, 32)
			if err != nil {
				return e, err
			}
			e.GridFlag = int32(i)
		case "-w":
			i, err := strconv.ParseInt(restField[i+1], 10, 32)
			if err != nil {
				return e, err
			}
			e.Weight = int32(i)
		case "-v":
			i, err := strconv.ParseInt(restField[i+1], 10, 32)
			if err != nil {
				return e, err
			}
			e.WeightType = int32(i)
		case "-l":
			i, err := strconv.ParseInt(restField[i+1], 10, 32)
			if err != nil {
				return e, err
			}
			e.Cpuload = int32(i)
		case "-m":
			i, err := strconv.ParseInt(restField[i+1], 10, 64)
			if err != nil {
				return e, err
			}
			e.Sampletime = i
		case "-d":
			e.ContainerName = restField[i+1]
		default:
			log.Printf("Unknown arg:%s", restField[i])
		}

	}
	return e, nil
}

type Client struct {
	//Addr    string
	servant string
	Timeout time.Duration
	MaxConn int

	endpoints []EndpointF
	clients   []*rpcChannel
	sessions  map[int32]*rpcSession
	sid       int32

	sessionMutex   sync.Mutex
	clientsMutex   sync.Mutex
	endpointCursor int32
}

func (c *Client) selectEndpoint() EndpointF {
	cursor := c.endpointCursor
	if int(cursor) < len(c.endpoints) {
		atomic.AddInt32(&c.endpointCursor, 1)
		return c.endpoints[cursor]
	}
	atomic.StoreInt32(&c.endpointCursor, 1)
	return c.endpoints[0]
}

func (c *Client) newRPCSession(sid int32) *rpcSession {
	c.sessionMutex.Lock()
	s := new(rpcSession)
	s.ID = sid
	s.ch = make(chan *ResponsePacket)
	c.sessions[sid] = s
	c.sessionMutex.Unlock()
	return s
}
func (c *Client) closeRPCSession(sid int32) {
	c.sessionMutex.Lock()
	delete(c.sessions, sid)
	c.sessionMutex.Unlock()
}
func (c *Client) getRPCSession(sid int32) *rpcSession {
	c.sessionMutex.Lock()
	s, exist := c.sessions[sid]
	if !exist {
		s = nil
	}
	c.sessionMutex.Unlock()
	return s
}

func (c *Client) closeRPCChannel(channel *rpcChannel) {
	c.clientsMutex.Lock()
	channel.Conn.Close()
	channel.running = false
	close(channel.ch)
	c.clients[channel.idx] = nil
	c.clientsMutex.Unlock()
}

func (c *Client) rpcChannelRead(channel *rpcChannel) {
	bufReader := bufio.NewReader(channel.Conn)
	lenBuffer := make([]byte, 4)
	var err error
	for channel.running {
		_, err = io.ReadAtLeast(bufReader, lenBuffer, len(lenBuffer))
		if nil != err {
			break
		}
		hlen := int32(0)
		binary.Read(bytes.NewBuffer(lenBuffer), binary.BigEndian, &hlen)
		b := make([]byte, int(hlen-4))
		_, err = io.ReadAtLeast(bufReader, b, len(b))
		if nil != err {
			break
		}
		var resp ResponsePacket
		err = resp.Decode(bytes.NewBuffer(b))
		if nil == err {
			s := c.getRPCSession(resp.IRequestId)
			if nil != s {
				s.ch <- &resp
			} else {
				log.Printf("Missing session:%d, maybe deleted by timeout task.", resp.IRequestId)
			}
		} else {
			log.Printf("Docode 'ResponsePacket' error:%v", err)
		}
	}
	c.closeRPCChannel(channel)
	if nil != err {
		log.Printf("RPCChannel:%d read close for reason:%v", channel.idx, err)
	}

}

func (c *Client) rpcChannelWrite(channel *rpcChannel) {
	var err error
	for channel.running {
		select {
		case packet := <-channel.ch:
			if nil != packet {
				var buf bytes.Buffer
				buf.Write(make([]byte, 4))
				packet.Encode(&buf)
				vlen := buf.Len()
				binary.BigEndian.PutUint32(buf.Bytes(), uint32(vlen))
				_, err = channel.Conn.Write(buf.Bytes())
				if nil != err {
					log.Printf("Failed to write rpc channel:%v", err)
					break
				}
				channel.counter++
			} else {
				return
			}
		}
	}
	c.closeRPCChannel(channel)
	if nil != err {
		log.Printf("RPCChannel:%d write close for reason:%v", channel.idx, err)
	}
}

func (c *Client) newRPCChannel(idx int) *rpcChannel {
	rc := new(rpcChannel)
	rc.idx = idx
	rc.running = true
	var err error
	endpoint := c.selectEndpoint()
	addr := fmt.Sprintf("%s:%d", endpoint.Host, endpoint.Port)
	rc.Conn, err = net.Dial("tcp", addr)
	if nil != err {
		log.Printf("Failed to connect server:%s for reason:%v", addr, err)
		return nil
	}
	rc.ch = make(chan TafStruct, 100)
	go c.rpcChannelWrite(rc)
	go c.rpcChannelRead(rc)
	c.clientsMutex.Lock()
	c.clients[idx] = rc
	c.clientsMutex.Unlock()
	return rc
}

func (c *Client) getRPCChannel() *rpcChannel {
	minCounter := int64(0)
	minIdx := 0
	for i, client := range c.clients {
		if nil == client {
			client = c.newRPCChannel(i)
			if nil == client {
				continue
			} else {
				return client
			}
		}
		if minCounter > client.counter {
			minCounter = client.counter
			minIdx = i
		}
	}
	return c.clients[minIdx]
}

func (c *Client) Invoke(ctype uint8, funcName string, req *bytes.Buffer, ctx map[string]string) (*ResponsePacket, error) {
	packet := RequestPacket{}
	packet.SBuffer = req.Bytes()
	packet.IVersion = 1
	packet.SServantName = c.servant
	packet.SFuncName = funcName
	packet.IRequestId = atomic.AddInt32(&c.sid, 1)
	packet.Context = ctx
	packet.IMessageType = int32(ctype)
	packet.ITimeout = 1000
	session := c.newRPCSession(packet.IRequestId)
	rpcConn := c.getRPCChannel()
	rpcConn.ch <- &packet
	var err error
	var resp *ResponsePacket
	select {
	case resp = <-session.ch:
		break
	case <-time.After(c.Timeout):
		err = ErrTafRPCTimeout
	}

	if nil == resp && nil == err {
		err = fmt.Errorf("No response recevied, maybe timeout")
	}
	c.closeRPCSession(packet.IRequestId)
	return resp, err
}

func NewClient(addr string, timeout time.Duration) *Client {
	c := &Client{}
	ss := strings.Split(addr, "@")
	if len(ss) == 2 {
		c.servant = ss[0]
		endpoints := strings.Split(ss[1], ":")
		for _, endpoint := range endpoints {
			e, err := parseEndpoint(endpoint)
			if nil != err {
				log.Printf("Invalid endpoint %s for reason:%v", endpoint, err)
			} else {
				c.endpoints = append(c.endpoints, e)
			}
		}
	} else {
		c.servant = addr
		if nil != DefaultNamingService {
			c.endpoints, _, _ = DefaultNamingService.FindObjectById(addr, nil)
		}
	}
	c.Timeout = timeout
	//c.Servant = servant

	c.clients = make([]*rpcChannel, 5)
	c.sessions = make(map[int32]*rpcSession)
	return c
}
