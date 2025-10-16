package pingwrap

// myping/ping.go

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"strings"
	"time"

	orig "github.com/go-ping/ping"
	"github.com/hashicorp/go-hclog"
)

const (
	// jsrtodo: PingTypeHTTPS is not used, is relevant
	PingTypeHTTPS = "https"
	MaxRTT        = time.Second * 5 // nebo zkopíruj z původního balíčku
)

// logger to log messages
var logger = hclog.New(&hclog.LoggerOptions{
	Name:  "pingwrap",
	Level: hclog.LevelFromString("DEBUG"),
})

// Statistics  keep original statistics
type Statistics = orig.Statistics // použijeme původní Statistics, ale můžeme si ji i zjednodušit

// Pinger obaluje buď orig.Pinger, nebo HTTP „ping“
// Pinger wrapping original pinger in orig.Pinger
type Pinger struct {
	// add extended fields here
	orig     *orig.Pinger
	httpsURL string

	// keep original structure fields
	Count    int
	Timeout  time.Duration
	OnFinish func(*orig.Statistics)
}

// NewPinger přebírá stejný signaturu jako orig.NewPinger, ale přidáme volbou metody
func NewPinger(addr string) (*Pinger, error) {

	logger.Info("Creating new pinger", "addr", addr)

	// adresa url se pozná podle schématu http:// nebo https://
	if strings.HasPrefix(addr, "https://") {
		return &Pinger{
			httpsURL: addr,
			Count:    1,
			Timeout:  MaxRTT,
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

// jsrtodo: HasHttpsUrl is not used, is it relevant?
func (p *Pinger) HasHttpsUrl() bool {
	return p.httpsURL != ""
}

// Run se postará o obě varianty
func (p *Pinger) Run() {

	// HTTPS cesta – implementujeme sami
	if p.httpsURL != "" {
		start := time.Now()
		_, err := p.runHttps()
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

func (p *Pinger) runHttps() (time.Duration, error) {

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}
	_, err := client.Get(p.httpsURL)
	if err != nil {
		fmt.Println(err)
		//return 0, err
	}

	return 0, nil // vrátíme 0, protože jsme neimplementovali měření času
	// If the ping was successful, return a dummy duration.
	// return 100 * time.Millisecond, nil
}
