package main

import (
	"net/http"
	"net/url"
	"os/exec"
	"strings"
)


// yt-dlp flags and options for all links
var ytdlpParams []string = []string{
    "--get-url",
    "--no-playlist",
}

// yt-dlp flags and options only for youtube links
var ytlinkParams []string = []string{
    "-f 22",
}

func getYoutubeDownloadURL(link string) (string, error) {
    
    var dlLink string
    params := make([]string, 0)
    params = append(params, ytdlpParams...)
    
    if isProbablyYT(link) {
        params = append(params, ytlinkParams...)
    }

    params = append(params, link)

    cmd := exec.Command("yt-dlp", params...)
    result, err := cmd.Output()

    if err != nil {
        return "", err
    }

    dlLink = string(result)

    return dlLink, nil
}

// Get the content filename with the extension. If not possible,
// and empty string is sent to c
func GetContentFilename(link string, c chan string) {

    var filename string

    params := make([]string, 0)
    params = append(params, "--no-playlist", "--get-title")

    if isProbablyYT(link) {
        params = append(params, ytlinkParams...)
    }

    params = append(params, link)

    cmd := exec.Command("yt-dlp", params...)
    result, err := cmd.Output()

    if err != nil {
        c <- ""
    }

    filename = string(result)

    c <- filename

}

func isValidURL(data string) bool {

    _, err := url.ParseRequestURI(data)
    
    if err != nil {
        return false
    }

    
    httpClient := &http.Client{
        CheckRedirect: func(req *http.Request, via []*http.Request) error {
            return http.ErrUseLastResponse
        },
    }
    resp, err := httpClient.Get(data)

    if err != nil {
        return false
    }
    
    if resp.StatusCode != 200 {
        return false
    }

    return true
}

func isProbablyYT(link string) bool {
    return strings.Contains(link, "youtube")
}

