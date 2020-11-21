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
    login string
    hint string
    pass []byte
}

type Database struct {
        loaded bool
        filename string
        records []Record
}

var db Database

type Action func(r *bufio.Reader)

var commands = map[string]Action {
         "add":    passAdd,
         "list":   passList,
         "load":   passTodo,
         "save":   passTodo,
         "paste":  passPaste,
         "help":   passHelp,
         "delete": passTodo,
         "find":   passTodo,
}

var commands_help = map[string]string {
         "add":    "Add login/password pair",
         "list":   "List stored login/password pairs",
         "load":   "Load password database",
         "save":   "Save password database",
         "paste":  "Paste password into clipboard",
         "help":   "List available commands",
         "delete": "Delete login/password pair",
         "find":   "Find available login/password pairs by partial match",
         "quit":   "Exit program",
}



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
        fmt.Println("Not implemented yet")
}

func passHelp(r *bufio.Reader) {
        for k, v := range commands_help {
                fmt.Printf("%s:\t%s\n", k, v)
        }
}

func passList(r *bufio.Reader) {
        if len(db.records) == 0 {
                fmt.Println("No records found")
                } else {
                for i, v := range db.records {
                        fmt.Printf("%d. %s (%s)\n", i, v.nick, v.hint)
                }
        }
}

func passAdd(r *bufio.Reader) {
        n, err := mustString(r, "Nickname", 2)
        if (err != nil) {
                return
        }

        l, err := mustString(r, "Login", 2)
        if (err != nil) {
                return
        }

        fmt.Print("Hint (optional)> ")
        h, _ := r.ReadString('\n')
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
                        r := Record{n, l, h, p}
                        db.records = append(db.records, r)
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
        for _, v := range db.records {
                if v.nick == n {
                        return v.pass, nil
                }
        }
        return nil, fmt.Errorf("Record %s not found", n)
}

func mustString(r *bufio.Reader, prompt string, retries int) (string, error) {
        var n string

        for i := 0; i < retries; i++ {
                fmt.Printf("%s> ", prompt)
                n, _ = r.ReadString('\n')
                n = strings.TrimSpace(n)
                if len(n) != 0 {
                        break
                }
                fmt.Printf("%s can't be empty\n", prompt)
        }
        if len(n) == 0 {
                return "", fmt.Errorf("Empty input")
        } else {
                return n, nil
        }
}

func mustPath(r *bufio.Reader, retries int) (string, error) {
        var n string

        for i := 0; i < retries; i++ {
                fmt.Print("Enter filename> ")
                n, _ = r.ReadString('\n')
                n = strings.TrimSpace(n)
                if len(n) != 0 {
                        break
                }
                fmt.Println("Filename can't be empty")
        }
        if len(n) == 0 {
                return "", fmt.Errorf("Invalid filename")
        } else {
                return n, nil
        }
}

