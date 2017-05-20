package main

import (
	"flag"
	"errors"
	"runtime"
	"net"
	"os"
	"bufio"
	"strconv"
	"github.com/liuzheng712/godis/logger"
	"fmt"
	"io"
)

var log = logger.Logs()

const (
	version = "0.1"
	project_name = "godis"
)

var (
	host = flag.String("h", "127.0.0.1", "run host")
	port = flag.String("p", "6379", "listen port")
	printVersion = flag.Bool("version", false, "Print the version and exit")
	GracefulExit = errors.New("graceful exit")
	EOF = []byte{byte(26)}
	ESC = []byte{byte(27)}
	errEOF = errors.New("EOF")
)

func handler() error {
	flag.Parse()
	if *printVersion {
		fmt.Printf(project_name + " version %s (%s)", version, runtime.GOARCH)
		return GracefulExit
	}
	return nil

}
func main() {
	var err error
	err = handler()
	if err == GracefulExit {
		os.Exit(0)
	}
	var l net.Listener
	l, err = net.Listen("tcp", *host + ":" + *port)
	if err != nil {
		log.Fatal("Error listening:", err)
		os.Exit(1)
	}
	defer l.Close()
	log.Info("Listening on", *host + ":" + *port)
	for {
		conn, err := l.Accept()
		if err != nil {
			log.Fatal("Error accepting: ", err)
			os.Exit(1)
		}
		//logs an incoming message
		log.Infof("Start to receive message %s -> %s \n", conn.RemoteAddr(), conn.LocalAddr())

		// Handle connections in a new goroutine.
		receive := make(chan []byte)
		go handleWrite(conn, receive)
		go handleRead(conn, receive)
	}

}

func handleWrite(conn net.Conn, receive chan []byte) {
	writeFor:
	for {
		select {
		case rec := <-receive:

			if rec[0] == ESC[0] {
				_, e := conn.Write([]byte("$-1\r\n"))
				if e != nil {
					log.Error("Error to send message because of ", e.Error())
				}
			} else if rec[0] == EOF[0] {
				//log.Info("Close the write in ", conn.RemoteAddr())
				break writeFor
			} else {
				log.Debug(string(rec))
			}
		}

	}
}
func handleRead(conn net.Conn, receive chan []byte) {
	defer func() {
		log.Info("Close the connection in ", conn.RemoteAddr())
		conn.Close()
	}()
	readFor:
	for {
		reader := bufio.NewReader(conn)
		message, _, err := reader.ReadLine()
		if err != nil {
			if err == io.EOF {
				//log.Info("Close the read in ", conn.RemoteAddr())
				receive <- EOF
				break readFor
			}
			log.Error(err)

			return
		}
		i, err := strconv.Atoi(string(message[1:]))
		if err != nil {
			log.Error(err)
			return
		}
		for j := 0; j < i * 2; j++ {
			message, _, err := reader.ReadLine()
			if err != nil {
				log.Error(err)
				return
			}
			receive <- message
		}
		// ESC
		receive <- ESC
	}
}

