package main

import (
    "crypto/md5"
    "encoding/xml"
    "flag"
    "fmt"
    "io"
    "io/ioutil"
    "net/http"
    "net/url"
    "os"
    "sort"
    "strconv"
    "strings"
    "time"
)

var api_key *string
var api_secret *string
var session_key *string
var token *string

type LFM struct {
    Status string `xml:"status,attr"`
}

func main() {

    api_key = flag.String("api_key", "", "API Key")
    api_secret = flag.String("api_secret", "", "API Secret")
    session_key = flag.String("session", "", "Session")
    token = flag.String("token", "", "Token")

    artist := flag.String("artist", "", "Track artist")
    album := flag.String("album", "", "Track album")
    title := flag.String("title", "", "Track title")

    flag.Parse()

    u, _ := url.Parse("http://ws.audioscrobbler.com/2.0/")

    q := u.Query()
    q.Set("method", "track.scrobble")

    t := time.Now()
    ts := int(t.Unix())

    q.Set("token", *token)
    q.Set("api_key", *api_key)
    q.Set("sk", *session_key)
    q.Set("artist", *artist)
    q.Set("album", *album)
    q.Set("track", *title)
    q.Set("timestamp", strconv.Itoa(ts))

    s := make_api_sig(q)

    q.Set("api_sig", s)

    u.RawQuery = q.Encode()

    resp, err := http.Post(u.String(), "", strings.NewReader(q.Encode()))
    if err != nil {
        fmt.Println(err)
        os.Exit(1)
    }

    defer resp.Body.Close()
    body, _ := ioutil.ReadAll(resp.Body)

    v := LFM{}

    err = xml.Unmarshal([]byte(body), &v)
    if err != nil {
        fmt.Printf("error: %v", err)
        os.Exit(2)
    }

    if v.Status == "ok" {
        os.Exit(0)
    } else {
        os.Exit(3)
    }
}

func make_api_sig(q url.Values) string {

    mk := make([]string, len(q))
    i := 0
    for k, _ := range q {
        mk[i] = k
        i++
    }

    sort.Strings(mk)

    h := md5.New()

    var t string
    for _, v2 := range mk {
        t += v2
        t += q.Get(v2)
    }

    t += *api_secret

    io.WriteString(h, t)

    return fmt.Sprintf("%x", h.Sum(nil))
}
