package main

import (
    "flag"
    "fmt"
    "os/exec"
    "regexp"
    "strconv"
    "strings"
    "time"
)

type CmusStatus struct {
    status   string
    artist   string
    album    string
    title    string
    duration int
    position int
}

var api_key *string
var api_secret *string
var session_key *string
var token *string

var queue CmusStatus

func cmus_remote_status(ch chan CmusStatus) {
    output, err := exec.Command("cmus-remote", "-Q").Output()

    if err != nil {
        fmt.Println("Exec: failed to run command", "cmus-remote")
        return
    }

    var status CmusStatus

    items := strings.Split(string(output), "\n")

    for _, v := range items {

        if m, _ := regexp.MatchString("status", v); m == true {
            re, _ := regexp.Compile("^status ")
            status.status = string(re.ReplaceAll([]byte(v), []byte("")))
        } else if m, _ := regexp.MatchString("tag artist", v); m == true {
            re, _ := regexp.Compile("^tag artist ")
            status.artist = string(re.ReplaceAll([]byte(v), []byte("")))
        } else if m, _ := regexp.MatchString("tag album", v); m == true {
            re, _ := regexp.Compile("^tag album ")
            status.album = string(re.ReplaceAll([]byte(v), []byte("")))
        } else if m, _ := regexp.MatchString("tag title", v); m == true {
            re, _ := regexp.Compile("^tag title ")
            status.title = string(re.ReplaceAll([]byte(v), []byte("")))
        } else if m, _ := regexp.MatchString("duration", v); m == true {
            re, _ := regexp.Compile("^duration ")
            status.duration, _ = strconv.Atoi(string(re.ReplaceAll([]byte(v), []byte(""))))
        } else if m, _ := regexp.MatchString("position", v); m == true {
            re, _ := regexp.Compile("^position ")
            status.position, _ = strconv.Atoi(string(re.ReplaceAll([]byte(v), []byte(""))))
        }
    }

    ch <- status
}

func main() {

    api_key = flag.String("api_key", "", "API Key")
    api_secret = flag.String("api_secret", "", "API Secret")
    session_key = flag.String("session", "", "Session")
    token = flag.String("token", "", "Token")

    flag.Parse()

    ch := make(chan CmusStatus)

    for {
        go cmus_remote_status(ch)

        time.Sleep(time.Duration(1) * time.Second)

        select {
        case s := <-ch:

            if s.status == "playing" && queue.artist != s.artist && queue.album != s.album && queue.title != s.title && s.duration > 20 && (float64(s.position)/float64(s.duration) > 0.3) {
                fmt.Println("Scrobbling:", s.artist, " - ", s.album, " - ", s.title, " (", s.position, "/", s.duration, ")")

                c := exec.Command("goscrobble", "-artist", s.artist, "-album", s.album, "-title", s.title, "-api_key", *api_key, "-api_secret", *api_secret, "-session", *session_key, "-token", *token)

                err := c.Start()

                if err != nil {
                    fmt.Println("Exec: failed to run command goscrobble")
                }

                c.Wait()

                queue = s
            }
        }
    }

}
