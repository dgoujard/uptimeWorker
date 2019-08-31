package services

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptrace"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

type uptimeService struct {
	maxRedirects int
}

func CreateUptimeService() *uptimeService {
	return &uptimeService{
		maxRedirects:10,
	}
}


func (u *uptimeService)CheckSite(site *SiteBdd){
	//TODO en cas d'erreur dans la réponse utilisé une fonction dans le cloud pour savoir si c'est local ou pas le probléme
	//TODO log des temps de réponse dans influxdb
	url := parseURL(site.Url)
	var reqHeaders = headers{}
	//reqHeaders = append(reqHeaders, "headername:value")
	var redirectsFollowed = 0
	var timeout = 500 * time.Millisecond
	fmt.Println(visit(url, reqHeaders,&redirectsFollowed,u.maxRedirects,timeout,"",""))
	os.Exit(0)
}

func parseURL(uri string) *url.URL {
	if !strings.Contains(uri, "://") && !strings.HasPrefix(uri, "//") {
		uri = "//" + uri
	}

	url, err := url.Parse(uri)
	if err != nil {
		log.Fatalf("could not parse url %q: %v", uri, err)
	}

	if url.Scheme == "" {
		url.Scheme = "http"
		if !strings.HasSuffix(url.Host, ":80") {
			url.Scheme += "s"
		}
	}
	return url
}


// visit visits a url and times the interaction.
// If the response is a 30x, visit follows the redirect.
func visit(url *url.URL, httpHeaders headers, redirectsFollowed *int, maxRedirects int, timeoutAccepted time.Duration,responseMustContain string, responseMustNotCountain string) (visitErr error, httpCode int, totalTime time.Duration) {
	req := newRequest(http.MethodGet, url, "",httpHeaders) //changer la méthode ?

	var t0, t1, t2, t3, t4, t5, t6 time.Time

	trace := &httptrace.ClientTrace{
		DNSStart: func(_ httptrace.DNSStartInfo) { t0 = time.Now() },
		DNSDone:  func(_ httptrace.DNSDoneInfo) { t1 = time.Now() },
		ConnectStart: func(_, _ string) {
			if t1.IsZero() {
				// connecting to IP
				t1 = time.Now()
			}
		},
		ConnectDone: func(net, addr string, err error) {
			if err != nil {
				log.Fatalf("unable to connect to host %v: %v", addr, err)
			}
			t2 = time.Now()

		},
		GotConn:              func(_ httptrace.GotConnInfo) { t3 = time.Now() },
		GotFirstResponseByte: func() { t4 = time.Now() },
		TLSHandshakeStart:    func() { t5 = time.Now() },
		TLSHandshakeDone:     func(_ tls.ConnectionState, _ error) { t6 = time.Now() },
	}
	req = req.WithContext(httptrace.WithClientTrace(context.Background(), trace))

	tr := &http.Transport{
		Proxy:                 http.ProxyFromEnvironment,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}
/*
peut etre pas nécessaire car pas d'envoi de certificat client
	switch url.Scheme {
	case "https":
		host, _, err := net.SplitHostPort(req.Host)
		if err != nil {
			host = req.Host
		}

		tr.TLSClientConfig = &tls.Config{
			ServerName:         host,
			InsecureSkipVerify: false,
			Certificates:       nil,
		}

		// Because we create a custom TLSClientConfig, we have to opt-in to HTTP/2.
		// See https://github.com/golang/go/issues/14275
		err = http2.ConfigureTransport(tr)
		if err != nil {
			log.Fatalf("failed to prepare transport for HTTP/2: %v", err)
		}
	}*/

	client := &http.Client{
		Transport: tr,
		Timeout: timeoutAccepted,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			// always refuse to follow redirects, visit does that
			// manually if required.
			return http.ErrUseLastResponse
		},
	}

	resp, err := client.Do(req)
	if err != nil {
		t7Err := time.Now() // after read body
		if t0.IsZero() {
			// we skipped DNS
			t0 = t1
		}
		totalTime = t7Err.Sub(t0)
		return err, 0,totalTime
	}
	defer resp.Body.Close()

	t7 := time.Now() // after read body
	if t0.IsZero() {
		// we skipped DNS
		t0 = t1
	}
	totalTime = t7.Sub(t0)
	httpCode = resp.StatusCode


	// print status line and headers

	/*
	names := make([]string, 0, len(resp.Header))
	for k := range resp.Header {
		names = append(names, k)
	}
	sort.Sort(headers(names))
	for _, k := range names {
		fmt.Println(k," ",strings.Join(resp.Header[k], ","))
	}*/
/*
	if bodyMsg != "" {
		printf("\n%s\n", bodyMsg)
	}

	colorize := func(s string) string {
		v := strings.Split(s, "\n")
		v[0] = grayscale(16)(v[0])
		return strings.Join(v, "\n")
	}

	fmt.Println()

	switch url.Scheme {
	case "https":
		printf(colorize(httpsTemplate),
			fmta(t1.Sub(t0)), // dns lookup
			fmta(t2.Sub(t1)), // tcp connection
			fmta(t6.Sub(t5)), // tls handshake
			fmta(t4.Sub(t3)), // server processing
			fmta(t7.Sub(t4)), // content transfer
			fmtb(t1.Sub(t0)), // namelookup
			fmtb(t2.Sub(t0)), // connect
			fmtb(t3.Sub(t0)), // pretransfer
			fmtb(t4.Sub(t0)), // starttransfer
			fmtb(t7.Sub(t0)), // total
		)
	case "http":
		printf(colorize(httpTemplate),
			fmta(t1.Sub(t0)), // dns lookup
			fmta(t3.Sub(t1)), // tcp connection
			fmta(t4.Sub(t3)), // server processing
			fmta(t7.Sub(t4)), // content transfer
			fmtb(t1.Sub(t0)), // namelookup
			fmtb(t3.Sub(t0)), // connect
			fmtb(t4.Sub(t0)), // starttransfer
			fmtb(t7.Sub(t0)), // total
		)
	}*/

	if isRedirect(resp) {
		loc, err := resp.Location()
		if err != nil {
			if err == http.ErrNoLocation {
				// 30x but no Location to follow, give up.
				return visitErr, httpCode, totalTime
			}
			log.Fatalf("unable to follow redirect: %v", err)
		}

		*redirectsFollowed = *redirectsFollowed+1
		if *redirectsFollowed > maxRedirects {
			//log.Fatalf("maximum number of redirects (%d) followed", maxRedirects)
			return errors.New("maximum number of redirects ("+strconv.Itoa(maxRedirects)+") followed"), httpCode, totalTime
		}

		return visit(loc,httpHeaders,redirectsFollowed, maxRedirects,timeoutAccepted,responseMustContain,responseMustNotCountain)
	}
	visitErr = readResponseBody(req, resp,responseMustContain,responseMustNotCountain)
	return visitErr, httpCode,totalTime

}

func isRedirect(resp *http.Response) bool {
	return resp.StatusCode > 299 && resp.StatusCode < 400
}


func newRequest(method string, url *url.URL, body string,httpHeaders headers) *http.Request {
	req, err := http.NewRequest(method, url.String(), createBody(body))
	if err != nil {
		log.Fatalf("unable to create request: %v", err)
	}
	for _, h := range httpHeaders {
		k, v := headerKeyValue(h)
		if strings.EqualFold(k, "host") {
			req.Host = v
			continue
		}
		req.Header.Add(k, v)
	}
	return req
}

func createBody(body string) io.Reader {
	if strings.HasPrefix(body, "@") {
		filename := body[1:]
		f, err := os.Open(filename)
		if err != nil {
			log.Fatalf("failed to open data file %s: %v", filename, err)
		}
		return f
	}
	return strings.NewReader(body)
}


// readResponseBody consumes the body of the response.
// readResponseBody returns an informational message about the
// disposition of the response body's contents.
func readResponseBody(req *http.Request, resp *http.Response,mustContainString string, mustNotCountainString string) error {
	if isRedirect(resp) || req.Method == http.MethodHead {
		return nil
	}
	//TODO lecture du body pour le controle si contient bien la chaine de caracyére controlé

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	bodyString := string(body)
	if len(mustContainString) > 0 && !strings.Contains(bodyString, mustContainString) {
		return errors.New("La réponse ne contient pas "+mustContainString)
	}
	if len(mustNotCountainString) > 0 && strings.Contains(bodyString, mustNotCountainString) {
		return errors.New("La réponse  contient "+mustContainString)
	}

	return nil
}

func headerKeyValue(h string) (string, string) {
	i := strings.Index(h, ":")
	if i == -1 {
		log.Fatalf("Header '%s' has invalid format, missing ':'", h)
	}
	return strings.TrimRight(h[:i], " "), strings.TrimLeft(h[i:], " :")
}

type headers []string
func (h headers) String() string {
	var o []string
	for _, v := range h {
		o = append(o, "-H "+v)
	}
	return strings.Join(o, " ")
}

func (h *headers) Set(v string) error {
	*h = append(*h, v)
	return nil
}

func (h headers) Len() int      { return len(h) }
func (h headers) Swap(i, j int) { h[i], h[j] = h[j], h[i] }
func (h headers) Less(i, j int) bool {
	a, b := h[i], h[j]

	// server always sorts at the top
	if a == "Server" {
		return true
	}
	if b == "Server" {
		return false
	}

	endtoend := func(n string) bool {
		// https://www.w3.org/Protocols/rfc2616/rfc2616-sec13.html#sec13.5.1
		switch n {
		case "Connection",
			"Keep-Alive",
			"Proxy-Authenticate",
			"Proxy-Authorization",
			"TE",
			"Trailers",
			"Transfer-Encoding",
			"Upgrade":
			return false
		default:
			return true
		}
	}

	x, y := endtoend(a), endtoend(b)
	if x == y {
		// both are of the same class
		return a < b
	}
	return x
}