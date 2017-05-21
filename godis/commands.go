package godis

import (
    "strconv"
    "strings"
    "fmt"
)

var (
    COMMANDS = map[string]func([][]byte) []byte{
        "COMMAND":COMMAND,
        "INFO":INFO,
        "SET":SET,
        "GET":GET,
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
func SET(holeCMD [][]byte) []byte {
    err_msg := "-ERR wrong number of arguments for 'set' command\r\n"
    if strings.ToUpper(string(holeCMD[1])) == "SET" {
        if len(holeCMD) != 4 {
            return []byte(err_msg)
        }
        key := holeCMD[2]
        fmt.Println(key)
        value := holeCMD[3]
        fmt.Println(value)

        switch holeCMD[0][0] {
        case 1:
            return []byte("+OK\r\n")
        case 2:
            return []byte("+OK\r\n")
        }
    } else {
        return []byte("-ERR command, your command not SET\r\n")
    }
    return nil
}
func GET(holeCMD [][]byte) []byte {
    err_msg := "-ERR wrong number of arguments for 'get' command\r\n"
    if strings.ToUpper(string(holeCMD[1])) == "GET" {
        if len(holeCMD) != 3 {
            return []byte(err_msg)
        }
        key := holeCMD[2]
        fmt.Println(key)

        switch holeCMD[0][0] {
        case 1:
            return []byte("+OK\r\n")
        case 2:
            return []byte("+OK\r\n")
        }
    } else {
        return []byte("-ERR command, your command not GET\r\n")
    }
    return nil
}
