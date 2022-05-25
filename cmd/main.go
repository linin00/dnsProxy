/*
 * @Author: linin00
 * @Date: 2022-05-25 01:59:27
 * @LastEditTime: 2022-05-25 03:59:49
 * @LastEditors: linin00
 * @Description: dnsProxy
 * @FilePath: /dnsProxy/cmd/main.go
 * 天道酬勤
 */
package main

import "dnsProxy/proxy"

func main() {
	proxy := proxy.NewDnsProxy("localhost", 8080)
	proxy.Start()
	quit := make(chan struct{})
	quit <- struct{}{}
}
