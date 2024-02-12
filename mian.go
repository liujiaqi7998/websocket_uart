package main

import (
	"flag"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/tarm/serial"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"time"
)

var (
	host            string
	uartCom         string
	uartBaud        int
	uartStopBit     int
	uartDataBit     int
	uartParityCheck int
	proxy           string
	ssl             bool
	ip              string
)

func init() {
	flag.IntVar(&uartBaud, "baud", 115200, "波特率 80-5000000")
	flag.IntVar(&uartStopBit, "stop", 1, "停止位 1-2")
	flag.IntVar(&uartDataBit, "data", 8, "数据位 5-8")
	flag.IntVar(&uartParityCheck, "check", 0, "校验模式：0-不校验 1-奇校验 2-偶检验")
	flag.StringVar(&host, "host", "", "设备IP地址或域名")
	flag.StringVar(&uartCom, "com", "", "要连接的串口")
	flag.StringVar(&proxy, "proxy", "", "代理地址")
	flag.BoolVar(&ssl, "ssl", false, "使用SSL连接（https）")
	flag.StringVar(&ip, "ip", "", "多张网卡使用此参数指定IP地址")
	// 设置将日志输出到标准输出（默认的输出为stderr，标准错误）
	// 日志消息输出可以是任意的io.writer类型
	log.SetOutput(os.Stdout)
}

func main() {
	fmt.Println("--- 欢迎使用 小奇远程串口转发工具 v0.1.0 ---")

	flag.Parse()

	if len(host) == 0 {
		log.Fatal("参数错误：设备IP地址或域名不能为空")
	}

	if uartBaud < 80 || uartBaud > 5000000 {
		log.Fatal("参数错误：baud参数无效，无效的波特率，需要格式 80-5000000 范围之间，默认是115200")
	}

	if uartStopBit < 1 || uartStopBit > 2 {
		log.Fatal("参数错误：stop参数无效，无效的停止位，需要格式 1-2 范围之间，默认是1")
	}

	if uartDataBit < 5 || uartDataBit > 8 {
		log.Fatal("参数错误：data参数无效，无效的数据位，需要格式 5-8 范围之间，默认是8")
	}

	if uartParityCheck < 0 || uartParityCheck > 2 {
		log.Fatal("参数错误：check参数无效，无效的校验模式，需要格式 0-不校验 1-奇校验 2-偶检验，默认是0")
	}

	if len(uartCom) == 0 {
		log.Fatal("参数错误：com参数无效，连接到已有串口模式下串口不能为空")
	}

	// 本地IP地址，应该是你希望使用的网卡的IP地址
	localIp := ip
	if !(len(ip) > 0) {
		ip, err := getValidIP()
		if err != nil {
			log.Fatalf("网络连接错误：" + err.Error())
		}
		localIp = string(ip)
	}

	// 创建一个net.Dialer，绑定到特定的本地IP地址
	dialer := &net.Dialer{
		LocalAddr: &net.TCPAddr{IP: net.IP(localIp)},
		Timeout:   30 * time.Second,
	}

	log.Println("本机IP是:", net.IP(localIp))
	log.Println("连接到：", host)
	log.Println("波特率：", uartBaud)
	log.Println("数据位：", uartDataBit)
	log.Println("停止位：", uartStopBit)
	log.Println("校验位：", uartParityCheck)
	if len(proxy) > 0 {
		log.Println("使用代理：", proxy)
	}
	fmt.Println("--------------------------------------------")
	result, err := uartConfigInt(uartStopBit, uartDataBit, uartParityCheck)
	if err != nil {
		log.Fatalf("[加载串口配置时错误] %v", err)
	}

	err = setUart(host, proxy, uartBaud, result, dialer)
	if err != nil {
		log.Fatalf("[打开远程串口时发生错误] %v", err)
	}

	// 定义WebSocket服务器URL

	// 设置请求的URL
	Scheme := "ws"
	if ssl {
		Scheme = "wss"
	}
	wsURL := url.URL{Scheme: Scheme, Host: host, Path: "/api/uart"}

	// 配置串口
	c := &serial.Config{Name: uartCom, Baud: uartBaud}
	s, err := serial.OpenPort(c)
	if err != nil {
		log.Fatalf("[打开本机串口错误] %v", err)
	}
	defer func(s *serial.Port) {
		err := s.Close()
		if err != nil {
			log.Fatalf("[关闭本机串口错误] %v", err)
		}
	}(s)

	WSDialer := websocket.Dialer{
		NetDial: dialer.Dial,
	}

	// 配置WebSocket的Dialer，包括代理设置
	// 是否启用代理模式
	if len(proxy) > 0 {
		WSDialer.Proxy = func(request *http.Request) (*url.URL, error) {
			// 返回代理服务器的URL
			return url.Parse(proxy)
		}
	}

	// 连接到WebSocket服务器
	ws, _, err := WSDialer.Dial(wsURL.String(), nil)
	if err != nil {
		log.Fatalf("[打开websocket错误] %v", err)
	}

	defer func(ws *websocket.Conn) {
		err := ws.Close()
		if err != nil {
			log.Fatalf("[关闭websocket错误] %v", err)
		}
	}(ws)

	// 开启goroutine来读取WebSocket服务器发送的数据并写入串口
	go func() {
		for {
			_, message, err := ws.ReadMessage()
			if err != nil {
				log.Printf("[websocket读取错误] %v", err)
				return
			}
			_, err = s.Write(message)
			if err != nil {
				log.Printf("[串口写入错误] %v", err)
				return
			}
			// 防止CPU占用过高，适当休眠
			time.Sleep(10 * time.Microsecond)
		}
	}()

	// 读取串口数据并发送到WebSocket服务器
	for {
		buf := make([]byte, 128)
		n, err := s.Read(buf)
		if err != nil {
			log.Printf("[串口读取错误] %v", err)
			continue
		}
		if n > 0 {
			err = ws.WriteMessage(websocket.TextMessage, buf[:n])
			if err != nil {
				log.Printf("[websocket写入错误] %v", err)
				return
			}
		}
		// 防止CPU占用过高，适当休眠
		time.Sleep(10 * time.Microsecond)
	}
}
