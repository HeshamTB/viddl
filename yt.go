package main

import (
	"net/http"
	"net/url"
	"os/exec"
)


func getYoutubeDownloadURL(link string) (string, error) {
    
    var dlLink string
    cmd := exec.Command("yt-dlp", "--get-url", "-f 22", link)
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

    
    httpClient := &http.Client{}
    resp, err := httpClient.Get(data)

    if err != nil {
        return false
    }
    
    if resp.StatusCode != 200 {
        return false
    }

    return true
}

