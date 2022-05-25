/*
 * @Author: linin00
 * @Date: 2022-05-25 04:13:50
 * @LastEditTime: 2022-05-25 04:26:57
 * @LastEditors: linin00
 * @Description:
 * @FilePath: /dnsProxy/proxy/proxy_test.go
 * 天道酬勤
 */
package proxy

import (
	"net/rpc"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
)

func TestProxy(t *testing.T) {
	proxy := NewDnsProxy("localhost", 8081)
	proxy.Start()
	client, err := rpc.Dial("tcp", "localhost:50000")
	if err != nil {
		t.Fatal(err)
	}
	err = client.Call("DNSServer.Add", Record{
		Ip:     "10.10.10.1",
		Domain: "svc.default.nginx-linin1",
	}, new(struct{}))
	if err != nil {
		t.Fatal(err)
	}
	err = client.Call("DNSServer.Add", Record{
		Ip:     "10.10.10.2",
		Domain: "svc.default.nginx-linin2",
	}, new(struct{}))
	if err != nil {
		t.Fatal(err)
	}
	logrus.Infoln(proxy.records)
	err = client.Call("DNSServer.Delete", Record{
		Ip:     "10.10.10.3",
		Domain: "svc.default.nginx-linin3",
	}, new(struct{}))
	if err == nil {
		t.Fatal(err)
	}
	logrus.Infoln(proxy.records)
	err = client.Call("DNSServer.Delete", Record{
		Ip:     "10.10.10.2",
		Domain: "svc.default.nginx-linin2",
	}, new(struct{}))
	if err != nil {
		t.Fatal(err)
	}
	logrus.Infoln(proxy.records)
	time.Sleep(time.Second * 1)
}
