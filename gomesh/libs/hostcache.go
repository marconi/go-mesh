package libs

import (
	"bufio"
	"errors"
	"fmt"
	"os"

	"github.com/marconi/go-mesh/gomesh/utils"
)

const HOSTCACHE = "hostcache.txt"

type HostCache struct {
	set map[string]bool
}

func (hc *HostCache) Add(addr string) bool {
	if _, ok := hc.set[addr]; !ok {
		hc.set[addr] = true
		return true
	}
	return false
}

func (hc *HostCache) Items() []string {
	var addrs []string
	for addr, _ := range hc.set {
		addrs = append(addrs, addr)
	}
	return addrs
}

func (hc *HostCache) Save() {
	// save updated hostcache to file
	if err := utils.CreateIfNotExist(HOSTCACHE); err != nil {
		fmt.Println(err)
	} else {
		f, err := os.OpenFile(HOSTCACHE, os.O_WRONLY|os.O_TRUNC, 0600)
		if err != nil {
			fmt.Println("Unable to open hostcache: ", err)
		} else {
			defer f.Close()
			for _, addr := range hc.Items() {
				f.WriteString(addr + "\n")
			}
		}
	}
}

func (hc *HostCache) Delete(addr string) {
	delete(hc.set, addr)
	hc.Save()
}

func NewHostCache() (*HostCache, error) {
	f, err := os.Open(HOSTCACHE)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Error loading hostcache: ", err))
	}

	hc := HostCache{set: make(map[string]bool)}
	r := bufio.NewReader(f)
	line, isPrefix, err := r.ReadLine()
	for err == nil && !isPrefix {
		s := string(line)
		hc.Add(s)
		line, isPrefix, err = r.ReadLine()
	}
	return &hc, nil
}
