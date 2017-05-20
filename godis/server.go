package godis

import (
	"errors"
	"strconv"
	"net"
	"io"
	"os"
	"flag"
	"runtime"
	"fmt"
	"bufio"
	"github.com/op/go-logging"
	"github.com/liuzheng712/godis/logger"
)

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
	err error
)

type Redis struct {
	log  *logging.Logger
	host string
	port string
}

func handler() error {
	flag.Parse()
	if *printVersion {
		fmt.Printf(project_name + " version %s (%s)", version, runtime.GOARCH)
		return GracefulExit
	}
	return nil

}
func New() (*Redis, error) {
	var err error
	err = handler()
	if err == GracefulExit {
		os.Exit(0)
	}
	log, err := logger.Logs("/tmp/godis.log")
	if err != nil {
		return nil, err
	}
	return &Redis{log:log, host: *host, port: *port}, nil
}
func (g *Redis)Run() {
	var l net.Listener
	l, err = net.Listen("tcp", g.host + ":" + g.port)
	if err != nil {
		g.log.Fatal("Error listening:", err)
		os.Exit(1)
	}
	defer l.Close()
	g.log.Info("Listening on", g.host + ":" + g.port)
	for {
		conn, err := l.Accept()
		if err != nil {
			g.log.Fatal("Error accepting: ", err)
			os.Exit(1)
		}
		//logs an incoming message
		g.log.Infof("Start to receive message %s -> %s \n", conn.RemoteAddr(), conn.LocalAddr())

		// Handle connections in a new goroutine.
		receive := make(chan []byte)
		go g.handleWrite(conn, receive)
		go g.handleRead(conn, receive)
	}
}
func (g *Redis) handleWrite(conn net.Conn, receive chan []byte) {
	writeFor:
	for {
		select {
		case rec := <-receive:

			if rec[0] == ESC[0] {
				_, e := conn.Write([]byte("$-1\r\n"))
				if e != nil {
					g.log.Error("Error to send message because of ", e.Error())
				}
			} else if rec[0] == EOF[0] {
				//log.Info("Close the write in ", conn.RemoteAddr())
				break writeFor
			} else {
				g.log.Debug(string(rec))
			}
		}

	}
}
func (g *Redis)handleRead(conn net.Conn, receive chan []byte) {
	defer func() {
		g.log.Info("Close the connection in ", conn.RemoteAddr())
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
			g.log.Error(err)

			return
		}
		i, err := strconv.Atoi(string(message[1:]))
		if err != nil {
			g.log.Error(err)
			return
		}
		for j := 0; j < i * 2; j++ {
			message, _, err := reader.ReadLine()
			if err != nil {
				g.log.Error(err)
				return
			}
			receive <- message
		}
		// ESC
		receive <- ESC
	}
}

