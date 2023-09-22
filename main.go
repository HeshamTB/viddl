package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"gitea.hbanafa.com/hesham/viddl/apkupdater"
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

type URLValidationCtx struct {
    URL string
    Valid bool
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
    remote := r.RemoteAddr
    realIP := r.Header.Get("X-Real-IP")
    if realIP != "" {
        remote = realIP
    }
    log.Printf(" %s %s %s %v", remote, methodString, r.URL, time.Since(tNow))
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
    isTLS := false
    tls := os.Getenv("TLS")
    if tls != "" {
       isTLS = true 
    }

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
        IsTLS: isTLS,
    }
}

func handleToAudio(ctx *Context, w http.ResponseWriter) http.ResponseWriter {
    
    log.Println("User requested audio")
    w.WriteHeader(400)

    jsonMsg, err := json.Marshal(
        apiMessageResponse{
            Message: "Audio only is not implemented",
        },
    )
    if err != nil {
        log.Println(err.Error())
        w.WriteHeader(500)
    }
    w.Write(jsonMsg) 

    return w
} 


func handleDownload(w http.ResponseWriter, r *http.Request) {
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
    toAudio := r.FormValue("toaudio")
    if toAudio == "on" {
        log.Println("User requested audio")
    }

    filenameChan := make(chan string)

    go GetContentFilename(userURL, filenameChan)
    downloadURL, err := getYoutubeDownloadURL(userURL)
    filename := <- filenameChan

    if err != nil {
        log.Println(err.Error())
        ctx.StatusCode = 500
        ctx.Err = &err
        err = templates.ExecuteTemplate(w,"download-result.html", ctx)
        return
    }

    ctx.DownloadURL = downloadURL
    ctx.DownloadURL = fmt.Sprintf("/download-direct?URL=%s&filename=%s",
        url.QueryEscape(ctx.DownloadURL), 
        url.QueryEscape(filename),
    )

    w.Header().Add("Hx-Redirect", ctx.DownloadURL)
}

func handleDirectDownload(w http.ResponseWriter, r *http.Request) {

    ctx := NewContext(r)
    if r.Method != "GET" {
        w.WriteHeader(400)
        return
    }


    userURL := strings.Trim(r.URL.Query().Get("URL"), "\n")
    filename := strings.Trim(r.URL.Query().Get("filename"), "\n")

    if userURL == "" {
        log.Println("Empty URL")
        w.WriteHeader(400)
        ctx.StatusCode = 400
        if err := templates.ExecuteTemplate(w,"download-result.html", ctx); err != nil {
            log.Println(err.Error())
        }
        return
    }
    ctx.DownloadURL = userURL

    req, err := http.NewRequest("GET", ctx.DownloadURL, nil)
    if err != nil {
        log.Println(err.Error())
        ctx.StatusCode = 500
        ctx.Err = &err
        if err := templates.ExecuteTemplate(w,"download-result.html", ctx); err != nil {
            log.Println(err.Error())
        }
        return
    }
    req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64; rv:109.0) Gecko/20100101 Firefox/117.0")

    client := http.Client{}

    dataRequest, err := client.Do(req)
    if err != nil {
        log.Println(err.Error())
        ctx.StatusCode = 500
        ctx.Err = &err
        if err := templates.ExecuteTemplate(w,"download-result.html", ctx); err != nil {
            log.Println(err.Error())
        }
        return
    }
    defer dataRequest.Body.Close()

    if dataRequest.StatusCode != 200 {
        log.Println("Failed to get content for URL", userURL)
        return
    }
    contentLength := dataRequest.Header.Get("Content-Length")
    if dataRequest.ContentLength == 0 {
        log.Println("Empty body from content url")
        w.WriteHeader(500)
        return
    }

    w.Header().Set(
        "Content-Disposition",
        fmt.Sprintf("attachment;filename=%s;", filename),
    )
    w.Header().Set("Content-Length", contentLength)
    w.WriteHeader(206)

    n, err := io.Copy(w, dataRequest.Body)
    if err != nil {
        log.Println(err.Error())
    }
    log.Printf("Copied %d bytes", n)
}

func handleRoot(w http.ResponseWriter, r *http.Request) {
    if r.URL.Path != "/" && r.URL.Path != "" {
        http.Redirect(w, r, "/", http.StatusPermanentRedirect)
        return
    }

    ctx := NewContext(r)
    formats := []DownloadFormats{}
    formats = append(formats, DownloadFormats{
        VideoRes: "720p",
        audioOnly: false,
        videoOnly: false,
    })
    err := templates.ExecuteTemplate(w, "download.html", ctx)
    if err != nil {
        log.Println(err.Error())
        w.WriteHeader(500)
        return
    }
}

func handleValidLink(w http.ResponseWriter, r *http.Request) {
    if r.Method != "POST" {
        w.WriteHeader(400)
        return
    }

    err := r.ParseForm()
    if err != nil {
        log.Println(err.Error())
        w.WriteHeader(400)
        return
    }

    url := r.FormValue("URL")

    ctx := URLValidationCtx{
        URL: url,
        Valid: isValidURL(url),
    }

    templates.ExecuteTemplate(w, "url-validation.html", ctx)
}



func init() {
    
    log.Println("[ init ] Starting...")
    templates = template.Must(template.ParseFS(TemplatesFS , "templates/*.html"))
    log.Println("[ init ] Templates Loaded")

}

func main() {

    updater, err := apkupdater.InitPkgUpdater()
    if err != nil {
        log.Println("Could not init Package Updater!\n", err.Error())
    }

    updater.Run()

    handler := http.NewServeMux()
    handler.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.FS(PublicFS))))
    handler.Handle("/assets/", http.FileServer(http.FS(AssetsFS)))

    handler.HandleFunc(
        "/robots.txt",
        func(w http.ResponseWriter, r *http.Request) {
            robotsFile, err := RobotsFS.ReadFile("static/robots.txt")
            if err != nil {
                log.Println(err.Error())
                w.WriteHeader(500)
                return
            }
            w.Write(robotsFile)
        },
    )

    handler.HandleFunc(
        "/sitemap.txt",
        func(w http.ResponseWriter, r *http.Request) {
            sitemapFile, err := AssetsFS.ReadFile("assets/sitemap.txt")
            if err != nil {
                log.Println(err.Error())
                w.WriteHeader(500)
                return
            }
            w.Write(sitemapFile)
        },
    )


    handler.HandleFunc("/", handleRoot)
    handler.HandleFunc("/download", handleDownload)
    handler.HandleFunc("/download-direct", handleDirectDownload)
    handler.HandleFunc("/valid-link", handleValidLink)


    wrappedHandler := NewLogger(handler)
    srv := http.Server{
        ReadTimeout: 1 * time.Minute,
        WriteTimeout: 1 * time.Minute,
        Addr: ":" + DEFAULT_HTTP_PORT,
        Handler: wrappedHandler,
    }
    
    log.Printf("Starting HTTP on %s", DEFAULT_HTTP_PORT)
    log.Fatalln(srv.ListenAndServe())
    log.Println("HTTP server stopped")
    updater.Stop()
}




