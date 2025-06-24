package padapter

// myping/ping.go

import (
	"crypto/tls"
	"fmt"
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
	OnFinish func(*orig.Statistics)
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
		_, err := p.runHttp()
		stats := &orig.Statistics{}
		if err == nil {

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
	p.orig.OnFinish = p.OnFinish
	p.orig.Run()
}

func (p *Pinger) runHttp() (time.Duration, error) {

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}
	_, err := client.Get(p.httpURL)
	if err != nil {
		fmt.Println(err)
		//return 0, err
	}

	return 0, nil // vrátíme 0, protože jsme neimplementovali měření času
	// If the ping was successful, return a dummy duration.
	// return 100 * time.Millisecond, nil
}
