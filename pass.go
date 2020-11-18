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

        for i := 0; i < 2; i++ {
                fmt.Print("Nickname> ")
                n, _ = r.ReadString('\n')
                n = strings.TrimSpace(n)
                if len(n) != 0 {
                        break
                }
                fmt.Println("Nickname can't be empty")
        }
        if len(n) == 0 {
                return
        }

        fmt.Print("Hint (optional)> ")
        h, _ = r.ReadString('\n')
        h = strings.TrimSpace(h)

        for i := 0; i < 3; i++ {
                fmt.Print("Enter Password> ")
                p, _ := terminal.ReadPassword(int(syscall.Stdin))
                //Hack to clear cursor after password read
                fmt.Print("\rEnter Password>                                      \r\n")
                if len(p) == 0 {
                        fmt.Println("Password can't be empty")
                        continue
                }

                fmt.Print("Repeat Password> ")
                p2, _ := terminal.ReadPassword(int(syscall.Stdin))
                //Hack to clear cursor after password read
                fmt.Print("\rRepeat Password>                                     \r\n")
                if !bytes.Equal(p, p2) {
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
                err = clipboard.WriteAll(string(p))
                if err != nil {
                        fmt.Println("Error pasting password into clipboard")
                }
                fmt.Println("Password is in clipboard")
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
                if err != nil {
                        fmt.Println("Error erasing password from clipboard")
                }
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

