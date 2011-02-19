package main

import (
    "net"
    "fmt"
    "os"
    "bufio"
    "bytes"
    "strings"
    "strconv"
)

type Client struct {
    host string
    port int
    db int
}

func bytesCommand(cmd string, args ...string) []byte {
    buf := bytes.NewBufferString(fmt.Sprintf("*%d\r\n$%d\r\n%s\r\n", len(args) + 1, len(cmd), cmd))
    for _, arg := range args {
        buf.WriteString(fmt.Sprintf("$%d\r\n%s\r\n", len(arg), arg))
    }    
    return buf.Bytes()
}

func log(args ...interface{}) {
    fmt.Printf("DEBUG: ")
    fmt.Println(args...)
}

func read(head *bufio.Reader) (interface{}, os.Error) {
    var res string
    var err os.Error

    for {
        res, err = head.ReadString('\n')
        if err != nil {
            return nil, err
        }
        break
    }
    res_type := res[0]
    res = strings.TrimSpace(res[1:])

    switch res_type {
        case '+':
            return res, nil
        case '-':
            return nil, os.NewError(res)
        case ':':
            n, err := strconv.Atoi64(res)
            log(n)
            return n, err
        case '$':
            l, _ := strconv.Atoi(res)
            l += 2 
            data := make([]byte, l)

            n, err := head.Read(data)
            if n != l || err != nil {
                if n != l {
                    err = os.NewError("Len mismatch")
                }
                return nil, err
            }

            log("bulk-len: " + strconv.Itoa(l))
            log("bulk-value: " + string(data))

            return data[:l - 2], nil
        case '*':
            l, _ := strconv.Atoi(string(res[0]))
            log("multi-bulk-len: " + strconv.Itoa(l))
            var data = make([][]byte, l)
            for i := 0; i < l; i++ {
                d, err := read(head)
                if err != nil {
                    log("returned with error")
                    return nil, err
                }
                data[i] = d.([]byte)
            }

            fmt.Printf("%q\n", data)
            return data, nil
    }
    return nil, os.NewError("Undefined redis response") 
}

func write(con net.Conn, cmd []byte) (*bufio.Reader, os.Error) {
    _, err := con.Write(cmd)
    if err != nil {
        return nil, os.NewError("Error writing cmd " + err.String())
    }
    
    return bufio.NewReader(con), nil
}

func (client *Client) send(cmd string, args...string) (data interface{}, err os.Error) {
    var addrString string = fmt.Sprintf("%s:%d", client.host, client.port)

    addr, err := net.ResolveTCPAddr(addrString)
    if err != nil {
        return nil, os.NewError("Error resolving Redis TCP addr")
    }

    con, err := net.DialTCP("tcp", nil, addr)
    if err != nil {
        return nil, os.NewError("Error connection to Redis at " + addr.String())
    }

    reader, err := write(con, bytesCommand(cmd, args...))
    if err != nil {
        return nil, err
    }

    data, err = read(reader) 
    con.Close()

    return
}

func main() {
    var client Client = Client{"127.0.0.1", 6379, 0} 

    // var enc_set []byte = bytesCommand("SET", "key", "hello")
    // fmt.Printf("%q\n", enc_set)

    // var enc_get []byte = bytesCommand("GET", "key")
    // fmt.Printf("%q\n", enc_get)

    // client.write(enc_set)
    // client.write(enc_get)
    //client.send("RPUSH", "keylist", "two")
    // client.write(bytesCommand("GET", "keylist"))
    // client.write(bytesCommand("GET", "nonexistant"))
    client.send("GET", "key")
    // client.send("SET", "key", "Hello")
    //client.send("LRANGE", "keylist", "0", "4")
    // client.send("KEYS", "*")
    client.send("EXISTS", "key")
}
