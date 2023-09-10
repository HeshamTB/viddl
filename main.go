package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"time"
)

const (
    DEFAULT_HTTP_PORT = "8080"
    DEFAULT_HTTPS_PORT = "4433"
)

type DownloadFormats struct {
    VideoRes string
    videoOnly bool
    audioOnly bool
}

type apiMessageResponse struct {
    Message string
}

type Logger struct {
    Handler http.Handler
}

func (l *Logger) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    
    tNow := time.Now().UTC()
    l.Handler.ServeHTTP(w, r)
    methodString := l.getMethodLogString(r.Method)
    log.Printf(" %s %s %s %v", r.RemoteAddr, methodString, r.URL, time.Since(tNow))
}

func (l *Logger) getMethodLogString(method string) string {
   colorBlue := "\033[34m"
   colorReset := "\033[0m"
   colorRed := "\033[31m"
   colorGreen := "\033[32m"
   colorYellow := "\033[33m"
   // colorPurple := "\033[35m"
   // colorCyan := "\033[36m"
   // colorWhite := "\033[37m"
    switch method {
    case "GET": return colorBlue + "GET" + colorReset
    case "POST": return colorGreen + "POST" + colorReset
    case "DELETE": return colorRed + "DELETE" + colorReset
    case "PUT": return colorYellow + "PUT" + colorReset
    default: return method
    }
}

func NewLogger(handler http.Handler) *Logger {
    return &Logger{Handler: handler}
}

func writeJSONResponse(w http.ResponseWriter, s string) http.ResponseWriter {
    w.WriteHeader(http.StatusBadRequest)
    w.Header().Set("Content-Type", "application/json")
    jsonResp, err := json.Marshal(apiMessageResponse{Message: s})
    if err != nil {
        w.WriteHeader(http.StatusInternalServerError)
        return w
    }
    _, err = w.Write(jsonResp)
    if err != nil {
        w.WriteHeader(http.StatusInternalServerError)
        return w
    }

    return w
    
}

func main() {

    handler := http.NewServeMux()

    handler.HandleFunc(
        "/download", 
        func(w http.ResponseWriter, r *http.Request) {
            if r.Method != "POST" {
                w.WriteHeader(400)
                return
            }
            err := r.ParseForm()
            if err != nil {
                w.WriteHeader(400)
                return
            }
            userURL := r.FormValue("URL")
            if userURL == "" {
                w = writeJSONResponse(w, "Provide URL as query")
                return
            }
            // Get URL from convx
            req, err := http.Get(fmt.Sprintf("http://localhost:80?url=%s", userURL))

            if err != nil {
                log.Println(err.Error())
                w.WriteHeader(500)
                return
            }
            if req.StatusCode != 200 {
                log.Printf("Got %v from convx\n", req)
                w.WriteHeader(500)
                return
            }
            body, err := io.ReadAll(req.Body)
            if err != nil {
                w.WriteHeader(500)
                log.Printf("Error while reading convx response body. \n%v", err.Error())
                return
            }
            downloadURL := string(body)
            log.Println("URL from convx", downloadURL)
            
            // Serve Button Template
            tmpl := template.Must(template.ParseFiles("templates/download-result.html"))
            err = tmpl.Execute(w, downloadURL)
            if err != nil {
                log.Println(err.Error())
                w.WriteHeader(500)
                return
            }
        },
    )
    handler.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        tmpl := template.Must(template.ParseFiles("templates/download.html"))
        formats := []DownloadFormats{}
        formats = append(formats, DownloadFormats{
            VideoRes: "720p",
            audioOnly: false,
            videoOnly: false,
        })
        err := tmpl.Execute(w, formats)
        if err != nil {
            log.Println(err.Error())
            w.WriteHeader(500)
            return
        }
    })

    wrappedHandler := NewLogger(handler)
    srv := http.Server{
        ReadTimeout: 10 * time.Second,
        WriteTimeout: 10 * time.Second,
        Addr: ":" + DEFAULT_HTTP_PORT,
        Handler: wrappedHandler,
    }
    
    log.Printf("Starting HTTP on %s", DEFAULT_HTTP_PORT)
    log.Fatalln(srv.ListenAndServe())
}




