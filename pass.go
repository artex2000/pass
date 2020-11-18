package main

import (
    "fmt"
    "time"
    "os"
    "bufio"
    "bytes"
    "strings"
    "syscall"
    "golang.org/x/crypto/ssh/terminal"
    "github.com/artex2000/pass/clipboard"
)

type Record struct {
    nick string
    hint string
    pass []byte
}

type Action func(r *bufio.Reader)

var commands = map[string]Action {
         "add":    passAdd,
         "list":   passList,
         "load":   passTodo,
         "save":   passTodo,
         "paste":  passPaste,
         "help":   passTodo,
         "delete": passTodo,
         "find":   passTodo,
}

var db []Record

func main() {
        r := bufio.NewReader(os.Stdin)
        db = make([]Record, 0)
        for {
                fmt.Print("Pass> ")
                c, _ := r.ReadString('\n')
                c = strings.TrimSpace(c)
                if c == "quit" {
                        break
                } else if c == "" {
                        continue
                } else {
                        if p, ok := commands[c]; !ok {
                                fmt.Printf("Unknown command: %s\n", c)
                        } else {
                                p(r)
                        }
                }
        }
}

func passTodo(r *bufio.Reader) {
}

func passList(r *bufio.Reader) {
        if len(db) == 0 {
                fmt.Println("No records found")
                } else {
                for i, v := range db {
                        fmt.Printf("%d. %s (%s)\n", i, v.nick, v.hint)
                }
        }
}

func passAdd(r *bufio.Reader) {
        var n, h string
        fmt.Print("Nickname> ")
        n, _ = r.ReadString('\n')
        n = strings.TrimSpace(n)
        fmt.Print("Hint (optional)> ")
        h, _ = r.ReadString('\n')
        h = strings.TrimSpace(h)
        for i := 0; i < 3; i++ {
                fmt.Print("Enter Password> ")
                p, _ := terminal.ReadPassword(int(syscall.Stdin))
                fmt.Print("\r                                                     \r\n")
                fmt.Print("Repeat Password> ")
                p2, _ := terminal.ReadPassword(int(syscall.Stdin))
                fmt.Print("\r                                                     \r\n")
                if len(p) == 0 {
                        fmt.Println("Password can't be empty")
                } else if !bytes.Equal(p, p2) {
                        fmt.Println("Passwords don't match")
                } else {
                        r := Record{n, h, p}
                        db = append(db, r)
                        break
                }
        }
}

func passPaste(r *bufio.Reader) {
        fmt.Print("Nickname> ")
        n, _ := r.ReadString('\n')
        n = strings.TrimSpace(n)
        p, err := findPass(n)
        if err == nil {
                fmt.Println("Password is in clipboard")
                err = clipboard.WriteAll(string(p))
                st := time.Now()
                c := time.Tick(time.Second)
                for range c {
                        el := int(time.Since(st).Milliseconds() / 1000)
                        if el >= 10 {
                                fmt.Print("                                        \r")
                                break
                        }
                        fmt.Printf("Clipboard with clear in %ds\r", 10 - el)
                }
                err = clipboard.ClearAll()
        }
}


func findPass(n string) ([]byte, error) {
        for _, v := range db {
                if v.nick == n {
                        return v.pass, nil
                }
        }
        return nil, fmt.Errorf("Record %s not found", n)
}

