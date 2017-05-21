package godis

import (
	"errors"
	"strconv"
	"net"
	//"io"
	"os"
	"flag"
	"runtime"
	"strings"
	"fmt"
	//"bufio"
	"github.com/op/go-logging"
	"github.com/liuzheng712/godis/logger"
	"bytes"
	"io"
)

const (
	version = "0.1"
	project_name = "godis"
)

var (
	host = flag.String("h", "127.0.0.1", "run host")
	port = flag.String("p", "6379", "listen port")
	mode = flag.String("m", "standalone", "standalone/cluster ")
	printVersion = flag.Bool("version", false, "Print the version and exit")
	GracefulExit = errors.New("graceful exit")
	EOF = []byte{byte(26)}
	ESC = []byte{byte(27)}
	errEOF = errors.New("EOF")
	err error

	ascii_logo = "\n" +
		"                _._                                                  \n" +
		"           _.-``__ ''-._                                             \n" +
		"      _.-``    `.  `_.  ''-._           Godis %s ( %s %s ) \n" +
		"  .-`` .-```.  ```\\/    _.,_ ''-._                                   \n" +
		" (    '      ,       .-`  | `,    )     Running in %s mode\n" +
		" |`-._`-...-` __...-.``-._|'` _.-'|     Port: %s\n" +
		" |    `-._   `._    /     _.-'    |     PID: %v\n" +
		"  `-._    `-._  `-./  _.-'    _.-'                                   \n" +
		" |`-._`-._    `-.__.-'    _.-'_.-'|                                  \n" +
		" |    `-._`-._        _.-'_.-'    |           http://godis.io        \n" +
		"  `-._    `-._`-.__.-'_.-'    _.-'                                   \n" +
		" |`-._`-._    `-.__.-'    _.-'_.-'|                                  \n" +
		" |    `-._`-._        _.-'_.-'    |                                  \n" +
		"  `-._    `-._`-.__.-'_.-'    _.-'                                   \n" +
		"      `-._    `-.__.-'    _.-'                                       \n" +
		"          `-._        _.-'                                           \n" +
		"              `-.__.-'                                               \n\n";
)

type Redis struct {
	log  *logging.Logger
	host string
	port string
	mode string
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
	return &Redis{log:log, host: *host, port: *port, mode:*mode}, nil
}
func (g *Redis)Run() {
	var l net.Listener
	l, err = net.Listen("tcp", g.host + ":" + g.port)
	if err != nil {
		g.log.Fatal("Error listening:", err)
		os.Exit(1)
	}
	defer l.Close()
	g.log.Infof(ascii_logo, version, runtime.Version(), runtime.GOARCH, g.mode, g.port, os.Getppid())
	for {
		conn, err := l.Accept()
		if err != nil {
			g.log.Fatal("Error accepting: ", err)
			os.Exit(1)
		}
		//logs an incoming message
		g.log.Infof("Start to receive message %s -> %s \n", conn.RemoteAddr(), conn.LocalAddr())

		// Handle connections in a new goroutine.
		receive := make(chan [][]byte)
		go g.handleWrite(conn, receive)
		go g.handleRead(conn, receive)
	}
}
func (g *Redis) handleWrite(conn net.Conn, receive chan [][]byte) {
	for {
		holeCMD := <-receive
		g.log.Debug(holeCMD)
		if len(holeCMD) < 2 {
			continue
		}
		if fun, ok := COMMANDS[strings.ToUpper(string(holeCMD[1]))]; ok {
			conn.Write(fun(holeCMD))
		} else {
			switch holeCMD[0][0] {
			case 1:
				conn.Write([]byte("$-1\r\n"))
			case 2:
				conn.Write([]byte(""))
			}
		}

	}
	//cmd := make(chan []byte)
	//cmdresult := make(chan []byte)
	//go g.analizeCommand(cmd, cmdresult)
	//writeFor:
	//for {
	//	select {
	//	case rec := <-receive:
	//		cmd <- rec
	//		if rec[0] == ESC[0] {
	//			//_, e := conn.Write([]byte("$-1\r\n"))
	//			_, e := conn.Write(<-cmdresult)
	//			if e != nil {
	//				g.log.Error("Error to send message because of ", e.Error())
	//			}
	//		} else if rec[0] == EOF[0] {
	//			//log.Info("Close the write in ", conn.RemoteAddr())
	//			break writeFor
	//		} else {
	//			g.log.Debug(string(rec))
	//		}
	//	}
	//
	//}
}
func (g *Redis)handleRead(conn net.Conn, receive chan [][]byte) {
	defer func() {
		g.log.Info("Close the connection in ", conn.RemoteAddr())
		conn.Close()
	}()
	readFor:
	for {
		onebyte := make([]byte, 1)
		twobyte := make([]byte, 2)
		holeCMD := [][]byte{}
		_, err = conn.Read(onebyte)
		if err == io.EOF {
			break readFor
		}
		if onebyte[0] == 42 {
			holeCMD = append(holeCMD, []byte{1})
			// *
			var lens bytes.Buffer
			for {
				_, _ = conn.Read(onebyte)

				if onebyte[0] == 13 {
					// \r
					break
				}
				lens.Write(onebyte)
			}
			conn.Read(onebyte)
			//g.log.Debug(lens.String())
			i, err := strconv.Atoi(lens.String())
			if err != nil {
				g.log.Error(err)
				return
			}
			//g.log.Debug(i)
			for j := 0; j < i; j++ {
				_, _ = conn.Read(onebyte)
				if onebyte[0] == 36 {
					// $
					lens.Reset()
					for {
						_, _ = conn.Read(onebyte)
						if onebyte[0] == 13 {
							// \r
							conn.Read(onebyte)
							break
						}
						lens.Write(onebyte)
					}
					cmdLen, err := strconv.Atoi(lens.String())
					if err != nil {
						g.log.Error(err)
						return
					}
					//g.log.Debug(cmdLen)
					cmd := make([]byte, cmdLen)
					_, err = conn.Read(cmd)
					//g.log.Debug(cmd)
					holeCMD = append(holeCMD, []byte(cmd))
					//g.log.Debug(holeCMD)
					conn.Read(twobyte)
				}
			}

		} else if onebyte[0] == 13 || onebyte[0] == 10 {
			continue
		} else {
			var cmd bytes.Buffer
			cmd.Write(onebyte)
			holeCMD = append(holeCMD, []byte{2})
			// TODO: split command with spaces(To: liuzheng712@gmail.com)
			for {
				_, err = conn.Read(onebyte)
				if err == io.EOF {
					break readFor
				}
				if onebyte[0] == 13 {
					conn.Read(onebyte)
					break
				} else if onebyte[0] > 20 {
					cmd.Write(onebyte)
				}

			}
			holeCMD = append(holeCMD, cmd.Bytes())

		}
		receive <- holeCMD
		//g.log.Debug("holeCMD", holeCMD)
	}
	//readFor:
	//for {
	//	reader := bufio.NewReader(conn)
	//	message, _, err := reader.ReadLine()
	//	if err != nil {
	//		if err == io.EOF {
	//			//log.Info("Close the read in ", conn.RemoteAddr())
	//			receive <- EOF
	//			break readFor
	//		}
	//		g.log.Error(err)
	//
	//		return
	//	}
	//	i, err := strconv.Atoi(string(message[1:]))
	//	if err != nil {
	//		g.log.Error(err)
	//		return
	//	}
	//	for j := 0; j < i * 2; j++ {
	//		message, _, err := reader.ReadLine()
	//		if err != nil {
	//			g.log.Error(err)
	//			return
	//		}
	//		receive <- message
	//	}
	//	// ESC
	//	receive <- ESC
	//}
}
//func (g *Redis)analizeCommand(cmd, cmdresult chan []byte) {
//	analizeFor:
//	for {
//		select {
//		case rec := <-cmd:
//			if rec[0] == ESC[0] {
//				cmdresult <- []byte("$-1\r\n")
//			} else if rec[0] == EOF[0] {
//				//log.Info("Close the write in ", conn.RemoteAddr())
//				break analizeFor
//			} else {
//				if fun, ok := COMMANDS[strings.ToUpper(string(rec))]; ok {
//					ss := fun()
//					g.log.Debug(ss)
//
//					//cmdresult <- ss
//				}
//			}
//		}
//
//	}
//}
