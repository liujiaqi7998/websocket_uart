package main

import (
	"fmt"
	"net"
)

func getValidIP() (net.IP, error) {
	// 连接到一个公共地址，但不发送数据
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		fmt.Println("Error dialing:", err)
		return nil, fmt.Errorf("没有有效的ip地址")
	}
	defer func(conn net.Conn) {
		err := conn.Close()
		if err != nil {
			return
		}
	}(conn)

	// 获取本地出口IP
	localAddr := conn.LocalAddr().(*net.UDPAddr)

	return localAddr.IP, nil
}
