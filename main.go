package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

const (
    DEFAULT_HTTP_PORT = "8080"
    DEFAULT_HTTPS_PORT = "4433"
)

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
            userURL := r.URL.Query().Get("url")
            w.Header().Set("Content-Type", "application/json")
            if userURL == "" {
                w = writeJSONResponse(w, "Provide URL as query")
                return
            }
            w = writeJSONResponse(w, fmt.Sprintf("You sent %s ", userURL))},
    )

    wrappedHandler := NewLogger(handler)
    srv := http.Server{
        ReadTimeout: 3 * time.Second,
        WriteTimeout: 10 * time.Second,
        Addr: ":" + DEFAULT_HTTP_PORT,
        Handler: wrappedHandler,
    }
    
    log.Printf("Starting HTTP on %s", DEFAULT_HTTP_PORT)
    log.Fatalln(srv.ListenAndServe())
}




