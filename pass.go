package main

import (
    "fmt"
    "os"
    "bufio"
    "bytes"
    "strings"
    "syscall"
    "golang.org/x/crypto/ssh/terminal"
)

type Record struct {
    nick string
    hint string
    pass []byte
}

func main() {
    r := bufio.NewReader(os.Stdin)
    db := make([]Record, 0)
    for {
        fmt.Print("Pass> ")
        c, _ := r.ReadString('\n')
        c = strings.TrimSpace(c)
        if c == "add" {
            var n, h string
            fmt.Print("Nickname> ")
            n, _ = r.ReadString('\n')
            n = strings.TrimSpace(n)
            fmt.Print("Hint> ")
            h, _ = r.ReadString('\n')
            h = strings.TrimSpace(h)
            for i := 0; i < 3; i++ {
                fmt.Print("Password> ")
                p, _ := terminal.ReadPassword(int(syscall.Stdin))
                fmt.Print("\nRepeat> ")
                p2, _ := terminal.ReadPassword(int(syscall.Stdin))
                if len(p) == 0 {
                    fmt.Println("\nPassword can't be empty")
                } else if !bytes.Equal(p, p2) {
                    fmt.Println("\nPasswords don't match")
                } else {
                    fmt.Print("\n")
                    r := Record{n, h, p}
                    db = append(db, r)
                    break
                }
            }
            continue
        } else if c == "list" {
            if len(db) == 0 {
                fmt.Println("No records found")
            } else {
                for i, v := range db {
                    fmt.Printf("%d. %s (%s)\n", i, v.nick, v.hint)
                }
            }
            continue
        } else if c == "quit" {
            break
        } else {
            fmt.Printf("Unknown command: %s\n", c)
        }
    }
}

