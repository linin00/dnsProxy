/*
 * @Author: linin00
 * @Date: 2022-05-25 03:55:22
 * @LastEditTime: 2022-05-25 04:24:06
 * @LastEditors: linin00
 * @Description:
 * @FilePath: /dnsProxy/proxy/proxy.go
 * 天道酬勤
 */
package proxy

import (
	"errors"
	"net"
	"net/rpc"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/sirupsen/logrus"
)

type Record struct {
	Ip     string
	Domain string
}
type dnsProxy struct {
	addr    *net.UDPAddr
	con     *net.UDPConn
	records map[string]string
}

func NewDnsProxy(ip string, port int) *dnsProxy {
	return &dnsProxy{
		addr: &net.UDPAddr{
			IP:   net.ParseIP(ip),
			Port: port,
		},
		records: map[string]string{},
	}
}

func (p *dnsProxy) Start() error {
	go func() {
		con, err := net.ListenUDP("udp", p.addr)
		if err != nil {
			logrus.Fatalln(err)
		}
		p.con = con
		for {
			buf := make([]byte, 1024)
			_, addr, err := p.con.ReadFrom(buf)
			if err != nil {
				logrus.Fatalln(err)
			}
			packet := gopacket.NewPacket(buf, layers.LayerTypeDNS, gopacket.Default)
			dnsPacket := packet.Layer(layers.LayerTypeDNS)
			request, _ := dnsPacket.(*layers.DNS)
			go p.process(addr, request)
		}
	}()
	err := rpc.RegisterName("DNSServer", p)
	if err != nil {
		panic(err)
	}
	listener, err := net.Listen("tcp", ":50000")
	if err != nil {
		panic(err)
	}
	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				panic(err)
			}
			logrus.Infoln("Accept conn from", conn.RemoteAddr())
			go func() {
				rpc.ServeConn(conn)
				defer conn.Close()
			}()
		}
	}()
	return nil
}

func (p *dnsProxy) process(addr net.Addr, request *layers.DNS) error {
	var dnsAnswer layers.DNSResourceRecord
	var (
		ip  string
		ok  bool
		err error
	)
	ip, ok = p.records[string(request.Questions[0].Name)]
	if !ok {
		ip, err = p.forward(string(request.Questions[0].Name))
		if err != nil {
			logrus.Errorln(err)
		}
	}
	res, _, _ := net.ParseCIDR(ip + "/24")
	dnsAnswer.Type = layers.DNSTypeA
	dnsAnswer.IP = res
	dnsAnswer.Name = request.Questions[0].Name
	dnsAnswer.Class = layers.DNSClassIN

	replyMess := request
	replyMess.QR = true
	replyMess.ANCount = 1
	replyMess.OpCode = layers.DNSOpCodeNotify
	replyMess.AA = true
	replyMess.Answers = append(replyMess.Answers, dnsAnswer)
	replyMess.ResponseCode = layers.DNSResponseCodeNoErr
	buf := gopacket.NewSerializeBuffer()
	opts := gopacket.SerializeOptions{}
	err = replyMess.SerializeTo(buf, opts)
	if err != nil {
		logrus.Fatalln(err)
	}
	p.con.WriteTo(buf.Bytes(), addr)
	return nil
}

func (p *dnsProxy) forward(ip string) (string, error) {
	res, err := net.LookupIP(ip)
	if err != nil {
		return "", errors.New("forward: can not get IP address")
	} else if len(res) == 0 {
		return "", errors.New("forward: get 0 record")
	}
	return res[0].String(), nil
}

func (p *dnsProxy) Add(r Record, reply *struct{}) error {
	p.records[r.Domain] = r.Ip
	return nil
}

func (p *dnsProxy) Delete(r Record, reply *struct{}) error {
	if _, ok := p.records[r.Domain]; ok {
		delete(p.records, r.Domain)
		return nil
	} else {
		return errors.New("record does not exist")
	}
}
