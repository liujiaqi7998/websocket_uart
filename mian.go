package main

import (
	"flag"
	"fmt"
	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
	"github.com/tarm/serial"
	"net/url"
	"os"
	"time"
)

var (
	mode            int
	host            string
	uartCom         string
	uartBaud        int
	uartStopBit     int
	uartDataBit     int
	uartParityCheck int
)

func init() {
	flag.IntVar(&mode, "mode", 1, "模式选择：1-连接到已有串口 2-新建虚拟串口")
	flag.IntVar(&uartBaud, "baud", 115200, "波特率 80-5000000")
	flag.IntVar(&uartStopBit, "stop", 1, "停止位 1-2")
	flag.IntVar(&uartDataBit, "data", 8, "数据位 5-8")
	flag.IntVar(&uartParityCheck, "check", 0, "校验模式：0-不校验 1-奇校验 2-偶检验")
	flag.StringVar(&host, "host", "", "设备IP地址或域名")
	flag.StringVar(&uartCom, "com", "", "要连接的串口")
	// 设置将日志输出到标准输出（默认的输出为stderr，标准错误）
	// 日志消息输出可以是任意的io.writer类型
	log.SetOutput(os.Stdout)
	// 设置日志级别为warn以上
	log.SetLevel(log.DebugLevel)
}

func main() {
	fmt.Println("--- 欢迎使用 感为科技 远程串口转发工具 v0.1.0 ---")

	flag.Parse()

	if len(host) == 0 {
		log.Fatal("参数错误：设备IP地址或域名不能为空")
	}

	if mode < 1 || mode > 2 {
		log.Fatal("参数错误：mode参数无效，无效的模式，需要格式 1-连接到已有串口 2-新建虚拟串口，默认是1")
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

	if mode == 1 && len(uartCom) == 0 {
		log.Fatal("参数错误：com参数无效，连接到已有串口模式下串口不能为空")
	}

	// 定义WebSocket服务器URL
	u := url.URL{Scheme: "ws", Host: host, Path: "/api/uart"}

	// 创建带有60秒超时的WebSocket连接
	dialer := websocket.Dialer{
		HandshakeTimeout: 60 * time.Second,
	}
	c, _, err := dialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("连接WebSocket失败:", err)
	}
	defer func(c *websocket.Conn) {
		err := c.Close()
		if err != nil {
			log.Fatal("连接WebSocket失败:", err)
		}
	}(c)

	// 打开串口
	config := &serial.Config{Name: "COM3", Baud: uartBaud}
	s, err := serial.OpenPort(config)
	if err != nil {
		log.Fatal("打开串口失败:", err)
	}
	defer func(s *serial.Port) {
		err := s.Close()
		if err != nil {
			log.Fatal("打开串口失败:", err)
		}
	}(s)

	// 从WebSocket接收消息并发送到串口
	go func() {
		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				log.Println("读取WebSocket失败:", err)
				return
			}
			_, err = s.Write(message)
			if err != nil {
				log.Println("写入串口失败:", err)
				return
			}
		}
	}()

	// 从串口接收消息并发送到WebSocket
	go func() {
		buf := make([]byte, 128)
		for {
			n, err := s.Read(buf)
			if err != nil {
				log.Println("读取串口失败:", err)
				return
			}
			err = c.WriteMessage(websocket.BinaryMessage, buf[:n])
			if err != nil {
				log.Println("写入WebSocket失败:", err)
				return
			}
		}
	}()

	// 阻塞主线程，防止程序退出
	select {}
}
