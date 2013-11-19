package main
import (
    "github.com/elazarl/goproxy"
    "log"
    "net/http"
    "fmt"
    "io"
    "io/ioutil"
    "bytes"
    "encoding/gob"
    "os"
    "strings"
    "errors"
    "crypto/md5"
)


//dp implementation
//trueReqCondition
func TrueReqConditionFunc() goproxy.ReqConditionFunc {
	return func(req *http.Request, ctx *goproxy.ProxyCtx) bool {
    return true
  }
}

//dp implementation
//true for resp
func TrueRespCondition() goproxy.RespCondition {
      return goproxy.RespConditionFunc(func(resp *http.Response, ctx *goproxy.ProxyCtx) bool {
                return true
        })
}

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

func createFullFilePath(req *http.Request) string {
   
    h := md5.New()
    io.WriteString(h, req.URL.String())
    hashPath := fmt.Sprintf("%x", h.Sum(nil))


    return "tmp/"+string(hashPath)
}


func writeResponse(resp *http.Response) *http.Response{
    if resp.Request.Method == "POST" || resp.Request.Method == "PUT" {
        return resp
    }

    body, err := ioutil.ReadAll(resp.Body)
    miniRequest := MiniRequest{resp.Status, resp.StatusCode, resp.Header, resp.Trailer, string(body), resp.Close}

    m := new(bytes.Buffer) 
    g := gob.NewEncoder(m)
    err = g.Encode(miniRequest)
    
    //save it
    fullFile := createFullFilePath(resp.Request)
    err = ioutil.WriteFile(fullFile, m.Bytes(), 0777)
    fmt.Println(err)

    fmt.Println("writeResponse")
    //todo this should not be neccessary
    newBody := nopCloser{bytes.NewBufferString(string(body))}
    resp.Body = newBody
    return resp
}

func readCache(req *http.Request) (*http.Response, error) {
    if req.Method == "POST" || req.Method == "PUT" {
        return nil, errors.New("We dont cache post")
    }


    //read it
    fullFile := createFullFilePath(req)
    b, err := ioutil.ReadFile(fullFile)
    if err == nil {
      p := bytes.NewBuffer(b)

      //decode it
      dec := gob.NewDecoder(p)
      e := MiniRequest{}
      err = dec.Decode(&e)
      if err != nil {
          panic(err)
      } 
      response := &http.Response{e.Status, e.StatusCode, "", 1, 0, e.Header, nopCloser{bytes.NewBufferString(e.Body)}, -1, nil, e.Close, e.Trailer, nil}
      return response, nil 
    } else {
      return nil, errors.New("file not found")
    }

}



func main() {
    os.MkdirAll("tmp", 0777)
    proxy := goproxy.NewProxyHttpServer()
    proxy.OnRequest(TrueReqConditionFunc()).DoFunc(
        func(req *http.Request,ctx *goproxy.ProxyCtx)(*http.Request,*http.Response) {

            url := req.URL.String()
            if(strings.Contains(url, ".css") ||
                strings.Contains(url, ".png") ||
                strings.Contains(url, ".jpg") ||
                strings.Contains(url, ".jpeg") ||
                strings.Contains(url, ".gif") ||
                strings.Contains(url, ".swf") ||
                strings.Contains(url, ".flv") ||
                strings.Contains(url, ".f4v") ||
                strings.Contains(url, "facebook") ||
                strings.Contains(url, "google-analytics") ||
                strings.Contains(url, "plusone.js") ||
                strings.Contains(url, "googleads") ||
                strings.Contains(url, "google.com")) {
                    return req, goproxy.NewResponse(req,
                    goproxy.ContentTypeText, http.StatusForbidden,
                    "Don't waste your time!")
            }

            resp, err := readCache(req)
            if err == nil {
                fmt.Println("Cache hit")
                return req, resp
            } else {
                fmt.Println("Cache miss")
                return req, nil 
            }
            
        })

    proxy.OnResponse(TrueRespCondition()).DoFunc(
        func(resp *http.Response, ctx *goproxy.ProxyCtx) *http.Response {
          if resp.Request != nil {
              writeResponse(resp)
          }
          return resp
        })

    proxy.Verbose = false
    log.Fatal(http.ListenAndServe(":8080", proxy))

}


