package godis

import (
	"strconv"
)

var (
	COMMANDS = map[string]func([][]byte) []byte{
		"COMMAND":COMMAND,
		"INFO":INFO,
	}
)

func COMMAND(holeCMD [][]byte) []byte {
	return []byte("$-1\r\n")
}

func INFO(holeCMD [][]byte) []byte {
	info := "Godis(" + version + ") Copyright liuzheng712@gmail.com\r\n"

	switch holeCMD[0][0] {
	case 1:
		return []byte("*1\r\n$" + strconv.Itoa(len(info)) + "\r\n" + info + "\r\n")
	case 2:
		return []byte( info + "\r\n")
	}
	return nil
}
