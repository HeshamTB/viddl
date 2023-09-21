package apkupdater

import (
	"fmt"
	"log"
	"os/exec"
	"sync"
	"time"
)


// Create a new updater. Invoke Run().
func InitPkgUpdater() (PackageUpdater, error) {
    
    alpine := isAlpine()

    if alpine {
        log.Println("Detected Alpine. Setting up Alpine Package Updater.")
         return NewAlpineLinuxPackageUpdater(
            time.Hour * 24,
            []string{"yt-dlp", "ffmpeg"},
        ), nil
    }

    log.Println("OS not known. Using dummy Package Updater")
    return NewDummyPackageUpdater(), nil
}

// Returns true if the os running the application
// is using Alpine Linux. If an error occurs it
// returns false
func isAlpine() bool {

    cmd := exec.Command("grep", "Alpine", "/etc/os-release")
    _, err := cmd.Output()
    if err != nil || cmd.ProcessState.ExitCode() != 0 { return false }
    
    return true
}

/* 
Provides a package updater to run along side
Application. It updates a set of packages at 
runtime. Run() starts a goroutine and returns.
Stop() signals stop to the running goroutine.
*/
type PackageUpdater interface {
    Update()
    Run()
    Stop()
    ValidPackages() bool
}

type AlpineLinuxPackageUpdate struct {
    Interval time.Duration
    Packages []string
    doneChan chan string
    apkLock *sync.Mutex
}

func (u *AlpineLinuxPackageUpdate) Update() {
    
    log.Println("Updating packages...")
    cmd := exec.Command("apk", u.Packages...)
    u.apkLock.Lock()
    stdout, err := cmd.Output()
    u.apkLock.Unlock()
    if err != nil {
        log.Println(err.Error())
    }
    if cmd.ProcessState.ExitCode() != 1 {
        log.Printf("Update command returned exit status %d\n", cmd.ProcessState.ExitCode())
        log.Println(string(stdout))
    }
    log.Println("Done updating packges")
}

func (u *AlpineLinuxPackageUpdate) Run() {

    go func() {
        mainloop:
        for { 
            select {
            case <- u.doneChan:
                fmt.Println("Stopping updater...")
                break mainloop
            case <- time.After(u.Interval):
                u.Update()
            }
        }
        fmt.Println("Updater stopped")
    }()
    log.Println("Started Alpine linux package updater")
}

func (u *AlpineLinuxPackageUpdate) ValidPackages() bool {
    return true
}

func (u *AlpineLinuxPackageUpdate) Stop() {
    u.doneChan <- ""
}

func NewAlpineLinuxPackageUpdater (
    interval time.Duration,
    packages []string,
) PackageUpdater {
        return &AlpineLinuxPackageUpdate{
            Interval: interval,
            Packages: packages,
            doneChan: make(chan string),
            apkLock: &sync.Mutex{},
    }
}

type DummyPackageUpdater struct {}
func (u *DummyPackageUpdater) Run() {}
func (u *DummyPackageUpdater) Stop() {}
func (u *DummyPackageUpdater) ValidPackages() bool { return true }
func (u *DummyPackageUpdater) Update() {}

func NewDummyPackageUpdater() PackageUpdater {
    return &DummyPackageUpdater{}
}

