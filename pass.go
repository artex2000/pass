package main

import (
    "fmt"
    "time"
    "os"
    "bufio"
    "io/ioutil"
    "bytes"
    "strings"
    "syscall"
    "path/filepath"
    "crypto/sha256"
    "crypto/aes"
    "crypto/cipher"
    "encoding/base64"
    "golang.org/x/crypto/ssh/terminal"
    "github.com/artex2000/pass/clipboard"
)

type Record struct {
    nick  string
    login string
    hint  string
    pass  string //base64 encoded password 
}

type Database struct {
        filename string
        sha      []byte
        iv       []byte
        key      []byte
        records  []Record
        existing bool
}

var db Database

type Action func(r *bufio.Reader)

var commands = map[string]Action {
         "add":    passAdd,
         "init":   passInit,
         "list":   passList,
         "load":   passLoad,
         "save":   passSave,
         "paste":  passPaste,
         "help":   passHelp,
         "delete": passTodo,
         "find":   passTodo,
}

var commands_help = map[string]string {
         "add":    "Add login/password pair",
         "init":    "Init new database",
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

func passInit(r *bufio.Reader) {
        //get filename
        fn, err := mustPath(r, 2)
        if err != nil {
                return
        }

        //get pass phrase
        fmt.Print("Enter Pass phrase> ")
        p, _ := terminal.ReadPassword(int(syscall.Stdin))
        //Hack to clear cursor after password read
        //TODO: set correct number of white spaces for proper clearing
        fmt.Print("\rEnter Passphrase>                                      \r\n")
        if len(p) == 0 {
                fmt.Println("Pass phrase can't be empty")
                return
        }

        fn, err = filepath.Abs(fn)
        if err != nil {
                return
        }

        db.filename = fn
        db.key, db.iv = passToKey(p)
        db.existing = false
}

func passLoad(r *bufio.Reader) {
        failed := false
        fn, err := mustPath(r, 2)
        if err != nil {
                return
        }

        fn, err = filepath.Abs(fn)
        if err != nil {
                fmt.Printf("Error opening file %s\n", err)
                return
        }

        data, err := ioutil.ReadFile(fn)
        if err != nil {
                fmt.Printf("Error reading file %s\n", err)
                return
        }

        //get pass phrase
        fmt.Print("Enter Pass phrase> ")
        p, _ := terminal.ReadPassword(int(syscall.Stdin))
        //Hack to clear cursor after password read
        //TODO: set correct number of white spaces for proper clearing
        fmt.Print("\rEnter Passphrase>                                      \r\n")
        if len(p) == 0 {
                fmt.Println("Pass phrase can't be empty")
                return
        }

        db.key, db.iv = passToKey(p)
        db.filename = fn
        db.existing = true

        defer func() {
                if failed {
                        db.filename = ""
                        db.key = nil
                        db.iv = nil
                        db.existing = false
                }
        }()

        content, err := cryptFile(data)
        if err != nil {
                failed = true
                fmt.Printf("Error decode file: %s\n", err)
                return
        }

        //Here we will try to verify file hash
        //1. First string should be sha256 hash encoded in base64
        // - string length must be EncodedLen(32)
        idx := base64.StdEncoding.EncodedLen(32)
        sha_enc := content[0:idx]
        sha := make([]byte, 32)
        n, err := base64.StdEncoding.Decode(sha, sha_enc)
        if err != nil || n != 32 {
                failed = true
                fmt.Println("Invalid file format")
                return
        }
        record := content[(idx + 2):]
        sha_calc := sha256.Sum256(record)
        if !bytes.Equal(sha, sha_calc[:]) {
                failed = true
                fmt.Println("File hash doesn't match")
                return
        }

        lines := strings.Split(string(record), "\r\n")
        for i := 1; i < len(lines); i++ {
                t := lines[i]
                p := strings.SplitN(t, ":", 4)
                r := Record{p[0], p[1], p[2], p[3]} 
                db.records = append(db.records, r)
        }
}

func passSave(r *bufio.Reader) {
        if len(db.records) == 0 {
                return
        }

        c     := serializeDb()
        sha   := sha256.Sum256([]byte(c))
        sha_t := base64.StdEncoding.EncodeToString(sha[:])
        out   := sha_t + "\r\n" + c

        data, err := cryptFile([]byte(out))
        if err != nil {
                fmt.Printf("Error encoding file %s\n", err)
                return
        }

        f, err := os.Create(db.filename)
        if err != nil {
                fmt.Printf("Error open file %s\n", err)
                return
        }
        defer f.Close()
        _, err = f.Write(data)
        if err != nil {
                fmt.Printf("Error writing file %s\n", err)
        }
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
                for _, v := range db.records {
                        fmt.Printf("[%s]:\tlogin: %s\t-- %s\n", v.nick, v.login, v.hint)
                }
        }
}

func passAdd(r *bufio.Reader) {
        n, err := mustString(r, "Nickname", 2, true)
        if (err != nil) {
                return
        }

        l, err := mustString(r, "Login", 2, false)
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
                        r := Record{n, l, h, base64.StdEncoding.EncodeToString(p)}
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
        } else {
                fmt.Printf("Error %s\n", err)
        }
}

func findPass(n string) ([]byte, error) {
        for _, v := range db.records {
                if v.nick == n {
                        r, err := base64.StdEncoding.DecodeString(v.pass)
                        if err != nil {
                                return nil, err
                        }
                        return r, nil
                }
        }
        return nil, fmt.Errorf("Record %s not found", n)
}

func mustString(r *bufio.Reader, prompt string, retries int, unique bool) (string, error) {
        var n string
        entered := false

        for i := 0; i < retries; i++ {
                fmt.Printf("%s> ", prompt)
                n, _ = r.ReadString('\n')
                n = strings.TrimSpace(n)
                if len(n) == 0 {
                        fmt.Printf("%s can't be empty\n", prompt)
                } else if unique {
                        _, err := findPass(n)
                        if err == nil {
                                fmt.Printf("%s nickname already present\n", n)
                        } else {
                                entered = true
                                break
                        }
                } else {
                        entered = true
                        break
                }
        }
        if entered {
                return n, nil
        } else {
                return "", fmt.Errorf("Invalid input")
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

func passToKey(pass []byte) (key, iv []byte) {
        const BLOCK_SIZE = 16
        sha := sha256.Sum256([]byte(pass))
        key = sha[0:BLOCK_SIZE]
        iv  = sha[BLOCK_SIZE:len(sha)]
        return
}

func serializeDb() string {
        var txt []string
        for _, v := range db.records {
                s := fmt.Sprintf("%s:%s:%s:%s\n", v.nick, v.login, v.hint, v.pass)
                txt = append(txt, s)
        }
        return strings.Join(txt, "\r\n")
}

func cryptFile(data []byte) ([]byte, error) {
        out := make([]byte, len(data))
        block, err := aes.NewCipher(db.key)
        if err != nil {
                return nil, err
        }

        stream := cipher.NewOFB(block, db.iv)
        stream.XORKeyStream(out, data)
        return out, nil
}
