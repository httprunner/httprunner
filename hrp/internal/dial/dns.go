package dial

import (
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/miekg/dns"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"

	"github.com/httprunner/httprunner/v4/hrp/internal/builtin"
	"github.com/httprunner/httprunner/v4/hrp/internal/env"
	"github.com/httprunner/httprunner/v4/hrp/internal/json"
)

const (
	httpDnsUrl   = "https://dig.bdurl.net/q"
	googleDnsUrl = "https://dns.google/resolve"
)

const (
	DnsSourceTypeLocal = iota
	DnsSourceTypeHttp
	DnsSourceTypeGoogle
)

const (
	DnsRecordTypeA     = 1
	DnsRecordTypeAAAA  = 28
	DnsRecordTypeCNAME = 5
)

var dnsHttpClient = &http.Client{
	Timeout: 5 * time.Minute,
}

type DnsOptions struct {
	DnsSourceType int
	DnsRecordType int
	DnsServer     string
	SaveTests     bool
}

type DnsResult struct {
	DnsList       []string `json:"dnsList"`
	DnsSource     int      `json:"dnsType"`
	DnsRecordType int      `json:"dnsRecordType"`
	DnsServer     string   `json:"dnsServer,omitempty"`
	Ttl           int      `json:"ttl"`
	Suc           bool     `json:"suc"`
	ErrMsg        string   `json:"errMsg"`
}

type googleDnsResp struct {
	Answer []googleDnsAnswer `json:"Answer"`
}

type httpDnsResp struct {
	Ips []string `json:"ips"`
	Ttl int      `json:"ttl"`
}

type googleDnsAnswer struct {
	Name string `json:"name"`
	Type int    `json:"type"`
	TTL  int    `json:"TTL"`
	Data string `json:"data"`
}

func ParseIP(s string) (net.IP, int) {
	ip := net.ParseIP(s)
	if ip == nil {
		return nil, 0
	}
	for i := 0; i < len(s); i++ {
		switch s[i] {
		case '.':
			return ip, 4
		case ':':
			return ip, 6
		}
	}
	return nil, 0
}

func localDns(src string, dnsRecordType int, dnsServer string) (dnsResult DnsResult, err error) {
	dnsResult.DnsSource = DnsSourceTypeLocal
	dnsResult.DnsRecordType = dnsRecordType

	if dnsServer == "" {
		config, _ := dns.ClientConfigFromFile("/etc/resolv.conf")
		dnsServer = config.Servers[0]
	} else {
		dnsResult.DnsServer = dnsServer
	}

	_, ipType := ParseIP(dnsServer)
	if ipType == 4 {
		dnsServer += ":53"
	}

	c := dns.Client{
		Timeout: 5 * time.Second,
	}
	m := dns.Msg{}

	m.SetQuestion(src+".", uint16(dnsRecordType))
	r, _, err := c.Exchange(&m, dnsServer)
	if err != nil {
		return
	}
	for _, ans := range r.Answer {
		switch dnsRecordType {
		case DnsRecordTypeA:
			record, isType := ans.(*dns.A)
			if isType {
				dnsResult.Ttl = int(record.Hdr.Ttl)
				dnsResult.DnsList = append(dnsResult.DnsList, record.A.String())
			}
		case DnsRecordTypeAAAA:
			record, isType := ans.(*dns.AAAA)
			if isType {
				dnsResult.Ttl = int(record.Hdr.Ttl)
				dnsResult.DnsList = append(dnsResult.DnsList, record.AAAA.String())
			}
		case DnsRecordTypeCNAME:
			record, isType := ans.(*dns.CNAME)
			if isType {
				dnsResult.Ttl = int(record.Hdr.Ttl)
				dnsResult.DnsList = append(dnsResult.DnsList, record.Target)
			}
		}
	}
	return
}

func httpDns(url string, dnsRecordType int) (dnsResult DnsResult, err error) {
	target := httpDnsUrl + "?host=" + url
	if dnsRecordType == DnsRecordTypeAAAA {
		target += "&aid=13&f=2"
	}
	resp, err := dnsHttpClient.Get(target)

	dnsResult.DnsSource = DnsSourceTypeHttp
	dnsResult.DnsRecordType = dnsRecordType

	if err != nil {
		return
	}
	defer resp.Body.Close()
	var buf []byte
	buf, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}
	var result httpDnsResp
	err = json.Unmarshal(buf, &result)
	if err != nil {
		return
	}
	dnsResult.DnsList = result.Ips
	dnsResult.Ttl = result.Ttl
	return
}

func googleDns(url string, dnsRecordType int) (dnsResult DnsResult, err error) {
	resp, err := dnsHttpClient.Get(googleDnsUrl + "?name=" + url + "&type=" + strconv.Itoa(dnsRecordType))

	dnsResult.DnsSource = DnsSourceTypeGoogle
	dnsResult.DnsRecordType = dnsRecordType

	if err != nil {
		return
	}
	defer resp.Body.Close()
	var buf []byte
	buf, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}
	var result googleDnsResp
	err = json.Unmarshal(buf, &result)
	if err != nil {
		return
	}
	if len(result.Answer) == 0 {
		return
	}
	for _, answer := range result.Answer {
		if answer.Type == dnsRecordType {
			dnsResult.Ttl = answer.TTL
			dnsResult.DnsList = append(dnsResult.DnsList, answer.Data)
		}
	}
	return
}

func DoDns(dnsOptions *DnsOptions, args []string) (err error) {
	if len(args) != 1 {
		return errors.New("there should be one argument")
	}

	var dnsResult DnsResult
	defer func() {
		if dnsOptions.SaveTests {
			dnsResultName := fmt.Sprintf("dns_result_%v.json", env.StartTimeStr)
			dnsResultPath := filepath.Join(env.RootDir, dnsResultName)
			err = builtin.Dump2JSON(dnsResult, dnsResultPath)
			if err != nil {
				log.Error().Err(err).Msg("save dns resolution result failed")
			}
		}
	}()

	dnsTarget := args[0]

	parsedURL, err := url.Parse(dnsTarget)
	if err == nil && parsedURL.Host != "" {
		log.Info().Msgf("parse input url %v and extract host %v", dnsTarget, parsedURL.Host)
		dnsTarget = strings.Split(parsedURL.Host, ":")[0]
	}
	log.Info().Msgf("resolve DNS for %v", dnsTarget)
	dnsRecordType := dnsOptions.DnsRecordType
	dnsServer := dnsOptions.DnsServer
	switch dnsOptions.DnsSourceType {
	case DnsSourceTypeLocal:
		dnsResult, err = localDns(dnsTarget, dnsRecordType, dnsServer)
	case DnsSourceTypeHttp:
		dnsResult, err = httpDns(dnsTarget, dnsRecordType)
	case DnsSourceTypeGoogle:
		dnsResult, err = googleDns(dnsTarget, dnsRecordType)
	}
	if err != nil {
		dnsResult.Suc = false
		dnsResult.ErrMsg = err.Error()
		log.Error().Err(err).Msgf("fail to do DNS for %s", dnsTarget)
	} else {
		dnsResult.Suc = true
		dnsResult.ErrMsg = ""
		fmt.Printf("\nDNS resolution done, result IP list: %v\n", dnsResult.DnsList)
	}
	return
}
