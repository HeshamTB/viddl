package apkupdater

import (
	"fmt"
	"log"
	"sync"
	"time"
)



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
    u.apkLock.Lock()
    log.Println("Updating packages...")
    u.apkLock.Unlock()
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
    log.Println("Started apline linux package updater")
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
) AlpineLinuxPackageUpdate {
        return AlpineLinuxPackageUpdate{
            Interval: interval,
            Packages: packages,
            doneChan: make(chan string),
            apkLock: &sync.Mutex{},
    }
}

