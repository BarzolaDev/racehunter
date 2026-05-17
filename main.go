package main

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"sync"
	"io"
	"strings"
)

func attack(url string, method string, body string, token string, wg *sync.WaitGroup, results chan string) {
	defer wg.Done()
	
	var req *http.Request
	var err error
	
	if body != "" {
		req, err = http.NewRequest(method, url, strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
	} else {
		req, err = http.NewRequest(method, url, nil)
	}
	
	if err != nil {
		results <- "error"
		return
	}

	if token != "" {
		req.Header.Set("Authorization", "Bearer " + token)
	}
	
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		results <- "error"
		return
	}
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(resp.Body)
	results <- string(respBody)
}

func main() {
	if len(os.Args) < 3 {
		fmt.Println("uso: racehunter <url> <threads> [method] [body] [token]")
		return
	}

	url := os.Args[1]
	threads, _ := strconv.Atoi(os.Args[2])
	method := "GET"
	body := ""
	token := ""
	
	if len(os.Args) >= 4 {
		method = os.Args[3]
	}
	if len(os.Args) >= 5 {
		body = os.Args[4]
	}
	if len(os.Args) >= 6 {
		token = os.Args[5]
	}

	results := make(chan string, threads)
	var wg sync.WaitGroup

	for i := 0; i < threads; i++ {
		wg.Add(1)
		go attack(url, method, body, token, &wg, results)
	}

	wg.Wait()
	close(results)

	responses := make(map[string]int)
	for r := range results {
		responses[r]++
	}

	if len(responses) > 1 {
		fmt.Println("[!] INCONSISTENCIA DETECTADA - posible race condition")
	} else {
		fmt.Println("[+] respuestas consistentes")
	}
}