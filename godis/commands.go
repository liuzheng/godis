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
        "LPUSH":LPUSH,
        "LRANGE":LRANGE,
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

        return []byte(ok_msg1)
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

var db = 0

func LPUSH(holeCMD [][]byte) []byte {
    err_msg := "-ERR wrong number of arguments for 'lpush' command\r\n"
    if strings.ToUpper(string(holeCMD[1])) == "LPUSH" {
        if len(holeCMD) != 4 {
            return []byte(err_msg)
        }
        key := holeCMD[2]
        value := holeCMD[3]
        if q, ok := memDB[db][string(key)]; ok {
            if q.(*Queue).T != "list" {
                return []byte("-ERR key type\r\n")
            } else {
                return []byte(":" + strconv.Itoa(q.(*Queue).ListLpush(value)) + "\r\n")
            }
        } else {
            q := NewQueue(string(key), 0)
            memDB[db][string(key)] = q
            return []byte(":" + strconv.Itoa(q.ListLpush(value)) + "\r\n")
        }

    } else {
        return []byte("-ERR command, your command not LPUSH\r\n")
    }
    return nil
}
func LRANGE(holeCMD [][]byte) []byte {
    err_msg := "-ERR wrong number of arguments for 'lpush' command\r\n"
    if strings.ToUpper(string(holeCMD[1])) == "LRANGE" {
        if len(holeCMD) != 5 {
            return []byte(err_msg)
        }
        key := holeCMD[2]
        start, err := strconv.Atoi(string(holeCMD[3]))
        if err != nil {
            return []byte("-ERR value is not an integer or out of range\r\n")
        }
        if start < 0 {
            return []byte("*0\r\n")
        }

        stop, err := strconv.Atoi(string(holeCMD[4]))
        if err != nil {
            return []byte("-ERR value is not an integer or out of range\r\n")
        }
        if stop < start {
            return []byte("*0\r\n")
        }
fmt.Println(start,stop)
        if q, ok := memDB[db][string(key)]; ok {
            if q.(*Queue).T != "list" {
                return []byte("-ERR key type\r\n")
            } else {
                q.(*Queue).ListLrange(start, stop)
                fmt.Println(q)
                return []byte(ok_msg1)
            }
        } else {
            return []byte("*0\r\n")
        }

    } else {
        return []byte("-ERR command, your command not LRANGE\r\n")
    }
    return nil
}
