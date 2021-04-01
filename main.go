package main

import (
	"math/rand"
	"net"
	"os"
	"strconv"
	"strings"

	"gopkg.in/mcuadros/go-syslog.v2"
)

// https://github.com/Aircoookie/WLED/wiki/UDP-Realtime-Control

func main() {
	numLeds, err := strconv.Atoi(os.Getenv("NUM_LEDS"))
	if err != nil {
		panic(err)
	}

	udpAddr, err := net.ResolveUDPAddr("udp4", os.Getenv("WLED_ADDR"))
	if err != nil {
		panic(err)
	}

	ledConn, err := net.DialUDP("udp", nil, udpAddr)
	if err != nil {
		panic(err)
	}

	channel := make(syslog.LogPartsChannel)
	handler := syslog.NewChannelHandler(channel)

	server := syslog.NewServer()
	server.SetFormat(syslog.Automatic)
	server.SetHandler(handler)
	err = server.ListenUDP(os.Getenv("SYSLOG_BIND"))
	if err != nil {
		panic(err)
	}
	err = server.Boot()
	if err != nil {
		panic(err)
	}

	go func(channel syslog.LogPartsChannel, ledConn *net.UDPConn) {
		for logParts := range channel {
			contentParts := strings.Split(logParts["content"].(string), ",")
			println(logParts["content"].(string))
			switch contentParts[6] {
			case "block":
				_, err = ledConn.Write([]byte{1, 0, byte(rand.Intn(numLeds)), 255, 0, 0})
				if err != nil {
					panic(err)
				}
			case "pass":
				switch contentParts[16] {
				case "icmp":
					_, err = ledConn.Write([]byte{1, 0, byte(rand.Intn(numLeds)), 0, 0, 255})
					if err != nil {
						panic(err)
					}
				default:
					_, err = ledConn.Write([]byte{1, 0, byte(rand.Intn(numLeds)), 0, 255, 0})
					if err != nil {
						panic(err)
					}
				}

			default:
				println(contentParts[6])
			}
		}
	}(channel, ledConn)

	server.Wait()
}
