package main

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"github.com/go-resty/resty/v2"
	"log"
	"net"
	"net/http"
	"net/url"
	"strconv"
)

// Payload 定义要发送的JSON数据结构
type Payload struct {
	Baud   string `json:"baud"`
	Config string `json:"config"`
}

// Response 定义响应的JSON数据结构
type Response struct {
	Type  int    `json:"type"`
	Level string `json:"level"`
	Title string `json:"title"`
	Body  string `json:"body"`
}

func uartConfigInt(uartStopBit int, uartDataBit int, uartParityCheck int) (int, error) {
	result := 0b1000000000000000000000000000
	result += (uartStopBit - 1) << 5
	result += (uartDataBit - 1) << 2
	result += uartParityCheck
	if result < 134217744 || result > 134217791 {
		return 0, errors.New("请检测串口配置，串口配置计算结果超出范围：" + strconv.Itoa(result))
	}
	return result, nil
}

func setUart(host string, proxy string, Baud int, Config int, dialer *net.Dialer) error {

	if Baud < 80 || Baud > 5000000 {
		return errors.New("波特率超出可用范围80-5000000")
	}
	if Config < 134217744 || Config > 134217791 {
		return errors.New("串口配置错误，请检测停止位，检验位等")
	}

	// 初始化数据
	data := Payload{
		Baud:   strconv.Itoa(Baud),
		Config: strconv.Itoa(Config),
	}

	// 将数据编码为JSON格式
	jsonData, err := json.Marshal(data)
	if err != nil {
		return errors.New("编码json发送包错误：" + err.Error())
	}

	// 设置请求的URL
	Scheme := "http"
	if ssl {
		Scheme = "https"
	}
	setURL := url.URL{Scheme: Scheme, Host: host, Path: "/api/uart/set"}

	// 创建POST请求
	client := resty.New()

	// 设置客户端使用自定义的Dialer
	client.SetTransport(&http.Transport{
		DialContext: dialer.DialContext,
	})

	// 是否启用代理模式
	if len(proxy) > 0 {
		client.SetProxy(proxy)
	}

	// 禁用ssl证书检查
	client.SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true})

	// 构建post请求
	resp, err := client.R().
		SetHeader("Content-Type", "application/json").
		SetHeader("User-Agent", "qi-ws-uart").
		SetHeader("Connection", "keep-alive").
		SetBody(string(jsonData)).
		Post(setURL.String())

	if err != nil {
		return errors.New("POST请求发生错误：" + err.Error())
	}
	if resp.StatusCode() != 200 {
		return errors.New("POST请求发生错误：错误码" + strconv.Itoa(resp.StatusCode()))
	}

	// 解析响应体
	var apiResponse Response
	if err := json.Unmarshal(resp.Body(), &apiResponse); err != nil {
		return errors.New("POST接收发生错误：" + err.Error())
	}

	// 打印title和body
	log.Printf("[打开串口结果] [%s]: %s", apiResponse.Title, apiResponse.Body)

	// 根据type和level判断执行结果
	if apiResponse.Type == 1 {
		if apiResponse.Level == "1" {
			log.Println("串口打开成功")
		} else {
			log.Fatal("串口打开失败")
		}
	}

	return nil
}
