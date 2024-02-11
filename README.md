# Web socket远程串口转发工具

本工具可以将虚拟串口的数据通过Web socket转发到远程芯片或服务器上

## 如何使用

1. 下载对应的版本 [Releases · liujiaqi7998/websocket_uart (github.com)](https://github.com/liujiaqi7998/websocket_uart/releases)

2. 使用控制台运行程序，带上启动参数


## 启动示例

windows:   `websocket_uart.exe --host="192.168.1.123" --com="COM2"`

## 启动参数

| 参数名           | 作用         | 类型    | 必选 | 默认值 |
| --------------- | ------ | ---- | --------------- | --------------- |
| host            | 设备IP地址或域名 | string | 是 | |
| uartCom         | 要连接的串口 | string | 是 | |
| uartBaud        | 波特率 80-5000000 | int    |      | 115200 |
| uartStopBit     | 停止位 1-2 | int    |      | 1 |
| uartDataBit     | 数据位 5-8 | int    |      | 8 |
| uartParityCheck | 校验模式：0-不校验 1-奇校验 2-偶检验 | int    |      | 0 |
| proxy           | 使用代理地址 | string |      | 空 |
| ssl             | 使用SSL连接（https） | bool   |      | 空 |
| ip              | 多张网卡使用此参数指定IP地址 | string |      | 空（自动） |

## 错误解决

**帮忙点个Star谢谢 Thanks♪(･ω･)ﾉ**

提出错误：[Issues · liujiaqi7998/websocket_uart (github.com)](https://github.com/liujiaqi7998/websocket_uart/issues)

#### 已知错误：

1. 过快发送信息回导致卡死
