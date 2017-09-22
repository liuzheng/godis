package godis

import (
	"errors"
	"strconv"
	"net"
	"os"
	"flag"
	"runtime"
	"strings"
	"fmt"
	log "github.com/liuzheng712/golog"
	"bytes"
	"io"
)

const (
	version      = "0.1"
	project_name = "godis"
	ok_msg1      = "+OK\r\n"
	ok_msg2      = "+OK\r\n"
)

type MSG struct {
	command *[][]byte
	fun     func([][]byte) []byte
	ch      *chan []byte
}
type DBtype map[int]map[string]chan MSG

var memDB DBtype

var (
	host         = flag.String("h", "127.0.0.1", "run host")
	port         = flag.String("p", "6379", "listen port")
	mode         = flag.String("m", "standalone", "standalone/cluster ")
	printVersion = flag.Bool("version", false, "Print the version and exit")
	GracefulExit = errors.New("graceful exit")
	EOF          = []byte{byte(26)}
	ESC          = []byte{byte(27)}
	errEOF       = errors.New("EOF")
	err          error

	ascii_logo = "\n" +
		"                _._                                                  \n" +
		"           _.-``__ ''-._                                             \n" +
		"      _.-``    `.  `_.  ''-._           Godis %v ( %v %v ) \n" +
		"  .-`` .-```.  ```\\/    _.,_ ''-._                                   \n" +
		" (    '      ,       .-`  | `,    )     Running in %v mode\n" +
		" |`-._`-...-` __...-.``-._|'` _.-'|     Port: %v\n" +
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
	host string
	port string
	mode string
}

func handler() error {
	flag.Parse()
	if *printVersion {
		fmt.Printf(project_name+" version %s (%s)", version, runtime.GOARCH)
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

	if err != nil {
		return nil, err
	}

	return &Redis{host: *host, port: *port, mode: *mode}, nil
}
func (g *Redis) Run() {
	var l net.Listener
	l, err = net.Listen("tcp", g.host+":"+g.port)
	if err != nil {
		log.Error("Run", "Error listening:", err)
		os.Exit(1)
	}

	memDB = make(DBtype)
	memDB[db] = make(map[string]chan MSG)

	go func() {
		for i, db := range memDB {
			for key, ch := range db {
				log.Debug("", "%v,%v,%v", i, key, ch)
			}
		}
	}()

	log.Info("Run", ascii_logo, version, runtime.Version(), runtime.GOARCH, g.mode, g.port, os.Getppid())
	for {
		conn, err := l.Accept()
		if err != nil {
			log.Error("Run", "Error accepting: %v", err)
			os.Exit(1)
		}
		//logs an incoming message
		log.Info("Run", "Start to receive message %s -> %s \n", conn.RemoteAddr(), conn.LocalAddr())

		// Handle connections in a new goroutine.
		receive := make(chan [][]byte)
		go g.handleWrite(conn, receive)
		go g.handleRead(conn, receive)
		defer func() {
			log.Info("defer", "Close the connection in %v", conn.RemoteAddr())
			conn.Close()
		}()
	}
}
func (g *Redis) handleWrite(conn net.Conn, receive chan [][]byte) {
	for {
		holeCMD := <-receive
		l := len(holeCMD)
		switch {
		case l < 2:
			continue
		default:
			if fun, ok := COMMANDS[strings.ToUpper(string(holeCMD[1]))]; ok {
				conn.Write(fun(holeCMD))
			} else {
				conn.Write([]byte("-ERR unknown command '"))
				conn.Write(holeCMD[1])
				conn.Write([]byte("'\r\n"))
			}
		}
	}
}
func (g *Redis) handleRead(conn net.Conn, receive chan [][]byte) {
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
			i, err := strconv.Atoi(lens.String())
			if err != nil {
				log.Error("handleRead", "%v", err)
				return
			}
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
						log.Error("handleRead", "%v", err)
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
	}
}
