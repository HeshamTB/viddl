package main

import (
	"net/http"
	"net/url"
	"os/exec"
	"strings"
)


// yt-dlp flags and options for all links
var ytdlpParams []string = []string{
    "--no-playlist",
}

// yt-dlp flags and options only for youtube links
var ytlinkParams []string = []string{
    "--get-url",
    "-f 22",
}

func getYoutubeDownloadURL(link string) (string, error) {
    
    var dlLink string
    params := make([]string, 0)
    params = append(params, ytdlpParams...)
    
    if isProbablyYT(link) {
        params = append(params, ytlinkParams...)
    }

    cmd := exec.Command("yt-dlp", link)
    result, err := cmd.Output()

    if err != nil {
        return "", err
    }

    dlLink = string(result)

    return dlLink, nil
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

