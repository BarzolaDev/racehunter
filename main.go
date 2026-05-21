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

// attack ejecuta UN solo request HTTP en paralelo
func attack(url string, method string, body string, token string, wg *sync.WaitGroup, results chan string) {
	defer wg.Done() // avisa que terminó cuando la función salga

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
		req.Header.Set("Authorization", "Bearer "+token)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		results <- "error"
		return
	}
	defer resp.Body.Close()
	io.Copy(io.Discard, resp.Body)
	results <- fmt.Sprintf("%d", resp.StatusCode)
}

// enumerate itera IDs numéricos y reporta los que existen
// Demuestra que IDs incrementales son enumerables sin autenticación
func enumerate(baseUrl string, start int, end int) {
	fmt.Printf("[*] Enumerando %s del ID %d al %d...\n", baseUrl, start, end)
	found := 0

	for i := start; i <= end; i++ {
		// Construye la URL con el ID numérico
		url := fmt.Sprintf("%s/%d", baseUrl, i)
		resp, err := http.Get(url)
		if err != nil {
			continue
		}
		resp.Body.Close()

		// Si responde 200 → el recurso existe y es visible
		if resp.StatusCode == 200 {
			fmt.Printf("  [!] ID %d existe → %s\n", i, url)
			found++
		}
	}

	fmt.Printf("\n[*] %d recursos encontrados\n", found)
	if found > 0 {
		fmt.Println("[!] VULNERABLE — IDs numéricos enumerables. Migrar a UUID.")
	} else {
		fmt.Println("[+] Sin resultados en el rango dado.")
	}
}

func main() {
	// Comando enumerate — demuestra IDOR por IDs numéricos
	// uso: racehunter enumerate <baseUrl> <start> <end>
	if len(os.Args) >= 2 && os.Args[1] == "enumerate" {
		if len(os.Args) < 5 {
			fmt.Println("uso: racehunter enumerate <baseUrl> <start> <end>")
			fmt.Println("ejemplo: racehunter enumerate https://api.com/products 1 100")
			return
		}
		baseUrl := os.Args[2]
		start, _ := strconv.Atoi(os.Args[3])
		end, _ := strconv.Atoi(os.Args[4])
		enumerate(baseUrl, start, end)
		return
	}

	// Comando default — ataque de race condition
	// uso: racehunter <url> <threads> [method] [body] [token]
	if len(os.Args) < 3 {
		fmt.Println("uso: racehunter <url> <threads> [method] [body] [token]")
		fmt.Println("     racehunter enumerate <baseUrl> <start> <end>")
		return
	}

	url := os.Args[1]
	threads, _ := strconv.Atoi(os.Args[2])
	method := "GET"
	body := ""
	token := ""

	if len(os.Args) >= 4 { method = os.Args[3] }
	if len(os.Args) >= 5 { body = os.Args[4] }
	if len(os.Args) >= 6 { token = os.Args[5] }

	results := make(chan string, threads)
	var wg sync.WaitGroup

	// Lanza N goroutines simultáneas — el ataque real
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
	for status, count := range responses {
		fmt.Printf("  %s → %d veces\n", status, count)
	}
}