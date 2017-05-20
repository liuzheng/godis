package main

import (
	"flag"
	"fmt"
	"errors"
	"runtime"
	"net"
	"os"
	"bufio"
	"strconv"
)

const (
	version = "0.1"
	ESC = []byte{byte(27)}
	project_name = "gedis"
)

var (
	host = flag.String("h", "127.0.0.1", "run host")
	port = flag.String("p", "6379", "listen port")
	printVersion = flag.Bool("version", false, "Print the version and exit")
	GracefulExit = errors.New("graceful exit")
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
		fmt.Println("Error listening:", err)
		os.Exit(1)
	}
	defer l.Close()
	fmt.Println("Listening on " + *host + ":" + *port)
	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting: ", err)
			os.Exit(1)
		}
		//logs an incoming message
		fmt.Printf("Received message %s -> %s \n", conn.RemoteAddr(), conn.LocalAddr())

		// Handle connections in a new goroutine.
		receive := make(chan []byte)
		go handleWrite(conn, receive)
		go handleRead(conn, receive)
	}

}

func handleWrite(conn net.Conn, receive chan []byte) {
	for {
		select {
		case rec := <-receive:

			if rec[0] == 27 {
				_, e := conn.Write([]byte("$-1\r\n"))
				if e != nil {
					fmt.Println("Error to send message because of ", e.Error())
				}
			} else {
				fmt.Println(string(rec))
			}
		}

	}
}
func handleRead(conn net.Conn, receive chan []byte) {
	for {
		reader := bufio.NewReader(conn)
		message, _, err := reader.ReadLine()
		if err != nil {
			fmt.Println(err)
			return
		}
		i, err := strconv.Atoi(string(message[1:]))
		if err != nil {
			fmt.Println(err)
			return
		}
		for j := 0; j < i * 2; j++ {
			message, _, err := reader.ReadLine()
			if err != nil {
				fmt.Println(err)
				return
			}
			receive <- message
		}
		// ESC
		receive <- ESC
	}

}

