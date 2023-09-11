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

var templates *template.Template

// TODO: Change all this to have a unified context
type Context struct {
    request *http.Request
    Formats []DownloadFormats
    AppURL string
    StatusCode int
    Err *error
    IsTLS bool
    DownloadURL string
}

func NewContext(r *http.Request) *Context {
    return &Context{
        request: r,
        StatusCode: 200,
        Formats: []DownloadFormats{
            {
                VideoRes: "720p",
                videoOnly: false,
                audioOnly: false,
            },
        },
        AppURL: r.Host,
        IsTLS: false,
    }
}

var AppURL string = "http://localhost:8080"
type HTMLBaseData struct {
    Formats []DownloadFormats
    AppURL string
}

var appData HTMLBaseData = HTMLBaseData{
    AppURL: AppURL,
}

func init() {
    
    log.Println("[ init ] Starting...")
    templates = template.Must(template.ParseFS(TemplatesFS , "templates/*.html"))

}
func main() {

    handler := http.NewServeMux()
    handler.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.FS(PublicFS))))

    handler.HandleFunc(
        "/download", 
        func(w http.ResponseWriter, r *http.Request) {
            ctx := NewContext(r)
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
                ctx.StatusCode = 500
                ctx.Err = &err
                err = templates.ExecuteTemplate(w,"download-result.html", ctx)
                return
            }
            if req.StatusCode != 200 {
                log.Printf("Got %v from convx\n", req)
                ctx.StatusCode = 500
                ctx.Err = &err
                err = templates.ExecuteTemplate(w,"download-result.html", ctx)
                return
            }
            body, err := io.ReadAll(req.Body)
            if err != nil {
                log.Printf("Error while reading convx response body. \n%v", err.Error())
                ctx.StatusCode = 500
                ctx.Err = &err
                err = templates.ExecuteTemplate(w,"download-result.html", ctx)
                return
            }
            downloadURL := string(body)
            log.Println("URL from convx", downloadURL)
            
            ctx.DownloadURL = downloadURL
            err = templates.ExecuteTemplate(w,"download-result.html", ctx)
            if err != nil {
                log.Println(err.Error())
                ctx.StatusCode = 500
                ctx.Err = &err
                err = templates.ExecuteTemplate(w,"download-result.html", ctx)
                return
            }
        },
    )
    handler.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        ctx := NewContext(r)
        log.Println(ctx.AppURL)
        formats := []DownloadFormats{}
        formats = append(formats, DownloadFormats{
            VideoRes: "720p",
            audioOnly: false,
            videoOnly: false,
        })
        appData.Formats = formats
        err := templates.ExecuteTemplate(w, "download.html", ctx)
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




