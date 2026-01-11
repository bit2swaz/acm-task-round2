package main

import (
	"io"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"time"
)

// config:dDefine backends and their weights here
var backends = []struct {
	URL    string
	Weight int
}{
	{"http://app-a:5678", 5}, // 50% chance (if both are 5)
	{"http://app-b:5678", 5}, // 50% chance
}

func main() {
	// seed the random number generator
	rand.Seed(time.Now().UnixNano())

	// the handler function
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// 1. select backend
		targetStr := pickBackend()
		targetURL, _ := url.Parse(targetStr)

		log.Printf("[Proxy] Routing %s -> %s", r.RemoteAddr, targetURL)

		// 2. create outgoing req
		// we create a new request to send to the backend container
		proxyReq, err := http.NewRequest(r.Method, targetURL.String(), r.Body)
		if err != nil {
			http.Error(w, "Bad Gateway", http.StatusBadGateway)
			return
		}

		// 3. copy headers (client -> backend)
		// we must forward headers like 'User-Agent', 'Accept', etc.
		for name, values := range r.Header {
			for _, value := range values {
				proxyReq.Header.Add(name, value)
			}
		}
		// add a custom header so the backend knows it passed through us
		proxyReq.Header.Set("X-Forwarded-By", "bit2swaz-Proxy")

		// 4. send req
		client := &http.Client{}
		resp, err := client.Do(proxyReq)
		if err != nil {
			http.Error(w, "Backend Unavailable", http.StatusServiceUnavailable)
			return
		}
		defer resp.Body.Close()

		// 5. copy headers (backend -> client)
		// now we take the backend's response headers and give them to the user
		for name, values := range resp.Header {
			for _, value := range values {
				w.Header().Add(name, value)
			}
		}

		// 6. set status code
		w.WriteHeader(resp.StatusCode)

		// 7. stream body (backend -> client)
		// efficiently copy bytes from backend response to user response writer
		io.Copy(w, resp.Body)
	})

	// start the server on port 80 inside the container
	log.Println("custom proxy listening on :80")
	log.Fatal(http.ListenAndServe(":80", nil))
}

// pickBackend implements the Weighted Random Algorithm
func pickBackend() string {
	totalWeight := 0
	for _, b := range backends {
		totalWeight += b.Weight
	}

	r := rand.Intn(totalWeight)

	currentSum := 0
	for _, b := range backends {
		currentSum += b.Weight
		if r < currentSum {
			return b.URL
		}
	}
	return backends[0].URL // fallback
}
