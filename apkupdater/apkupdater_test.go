package apkupdater

import (
	"testing"
	"time"
)

func TestNewUpdater(t *testing.T) {
    interval := time.Second * 1
    packages := []string{"yt-dlp", "ffmpeg"}

    updater := NewAlpineLinuxPackageUpdater(
        interval,
        packages,
    )

    errs := make([]string, 0)
    for _, p := range packages {
        contains := false
        inner:
        for _, pp := range updater.Packages {
            if p == pp { 
                contains = true
                break inner
            }
        }
        if !contains {
            errs = append(errs, p + "Not found")
        }
    }

    if len(errs) != 0 {

        for _, e := range errs {
            t.Log(e)
        }
        t.Fail()
    }

}

func TestRunStop(t *testing.T) {
    interval := time.Second * 1
    packages := []string{"yt-dlp", "ffmpeg"}

    updater := NewAlpineLinuxPackageUpdater(
        interval,
        packages,
    )

    updater.Run()
    updater.Stop()
    // TODO: Wie teste ich das?
    
}
