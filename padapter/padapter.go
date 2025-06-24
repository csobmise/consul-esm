package padapter

// myping/ping.go

import (
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	orig "github.com/go-ping/ping"
)

const (
	PingTypeHTTP = "http"
	MaxRTT       = time.Second * 5 // nebo zkopíruj z původního balíčku
)

// Statistics – jen vybrané pole z orig. Statistics
type Statistics struct {
	PacketsRecv int
	MaxRtt      time.Duration
}

// Pinger obaluje buď orig.Pinger, nebo HTTP „ping“
type Pinger struct {
	orig    *orig.Pinger
	httpURL string

	// necháme stejná pole, co orig.Pinger exportuje
	Count    int
	Timeout  time.Duration
	OnFinish func(*Statistics)
}

// NewPinger přebírá stejný signaturu jako orig.NewPinger, ale přidáme volbou metody
func NewPinger(addr string) (*Pinger, error) {
	// adresa url se pozná podle schématu http:// nebo https://
	if strings.HasPrefix(addr, "http://") || strings.HasPrefix(addr, "https://") {
		return &Pinger{
			httpURL: addr,
			Count:   1,
			Timeout: MaxRTT,
		}, nil
	}
	// jinak zavoláme originál
	p, err := orig.NewPinger(addr)
	if err != nil {
		return nil, err
	}
	return &Pinger{
		orig:    p,
		Count:   p.Count,
		Timeout: p.Timeout,
	}, nil
}

func (p *Pinger) SetPrivileged(v bool) {
	if p.orig != nil {
		p.orig.SetPrivileged(v)
	}
}

func (p *Pinger) HasHttpUrl() bool {
	return p.httpURL != ""
}

// Run se postará o obě varianty
func (p *Pinger) Run() {

	if p.httpURL != "" {
		start := time.Now()
		resp, err := http.Get(p.httpURL)
		stats := &Statistics{}
		if err == nil {
			io.Copy(io.Discard, resp.Body)
			resp.Body.Close()
			if p.Count > 0 {
				stats.PacketsRecv = 1
				stats.MaxRtt = time.Since(start)
			}
		}
		if p.OnFinish != nil {
			p.OnFinish(stats)
		}
		return
	}
	// ICMP/UDP cesta – jen přepošleme data do orig.Pinger
	p.orig.Count = p.Count
	p.orig.Timeout = p.Timeout
	p.orig.OnFinish = func(os *orig.Statistics) {
		if p.OnFinish != nil {
			p.OnFinish(&Statistics{
				PacketsRecv: os.PacketsRecv,
				MaxRtt:      os.MaxRtt,
			})
		}
	}
	p.orig.Run()
}

func (p *Pinger) runHttp() {

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}
	_, err := client.Get(p.httpURL)
	if err != nil {
		fmt.Println(err)
		//return 0, err
	}

	return
	// If the ping was successful, return a dummy duration.
	// return 100 * time.Millisecond, nil
}
