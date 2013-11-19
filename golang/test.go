package main

import (
  "fmt"
  "net/http"
  //"io/ioutil"
  //"encoding/gob"
  //"bytes"
  "io"
  "time"
  "strconv"
  "os"
)

/*
func responseToString(resp *http.Response) string {
    status := string(resp.Status)
    statusCode := string(resp.StatusCode)
    proto := string(resp.Proto)
    protoMajor := string(resp.ProtoMajor)
    protoMinor := string(resp.ProtoMinor)
    header := resp.Header
    contentLength := string(resp.ContentLength)
    transferEncoding := resp.TransferEncoding
    close := resp.Close
    trailer := resp.Trailer
    defer resp.Body.Close()
    body, err := ioutil.ReadAll(resp.Body)

    return status+"##%%split%%##"+statusCode+"##%%split%%##"+proto+"##%%split%%##"+protoMajor+"##%%split%%##"+protoMinor+"##%%split%%##"+header+"##%%split%%##"+contentLength+"##%%split%%##"+transferEncoding+"##%%split%%##"+close+"##%%split%%##"+trailer+"##%%split%%##"+string(body)
 
}
*/

type MiniRequest struct{
    Status string
    StatusCode int
    Header http.Header
    Trailer http.Header
    Body string
    Close bool
}

type nopCloser struct { 
    io.Reader 
} 

func (nopCloser) Close() error { return nil } 


/*
func writeResponse(resp *http.Response) *http.Response {
    body, err := ioutil.ReadAll(resp.Body)
    miniRequest := MiniRequest{resp.Status, resp.StatusCode, resp.Header, resp.Trailer, string(body), resp.Close}
    m := new(bytes.Buffer) 
    g := gob.NewEncoder(m)
    err = g.Encode(miniRequest)
    //decode it
    dec := gob.NewDecoder(m)
    e := MiniRequest{}
    err = dec.Decode(&e)
    
    //create response
    response := http.Response{e.Status, e.StatusCode, "", 1, 0, e.Header, nopCloser{bytes.NewBufferString(e.Body)}, -1, nil, e.Close, e.Trailer, nil}
    return response
}
*/

func connect(i int, c chan int) {
    fmt.Println("new connection: "+strconv.Itoa(i))
    http.Get("http://www.bing.de/")
    fmt.Println("done connection: "+strconv.Itoa(i))
    c <- 1
}

func main() {
    c := make(chan int, 50)
    os.Setenv("HTTP_PROXY", "http://localhost:8080")

    timeOld := time.Now()
    for i := 0; i < 50; i++ {
        go connect(i, c)
    }

    for i := 0; i < 50; i++ {
      <-c
    }

    timeNow := time.Now().Sub(timeOld)
    fmt.Println("it took me %d", timeNow)
    
}

