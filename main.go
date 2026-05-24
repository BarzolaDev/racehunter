package main

import (
    "encoding/json"
    "fmt"
    "io"
    "net/http"
    "os"
    "strconv"
    "strings"
    "sync"
    "time"
)

type Result struct {
	Status int
	Body   string
}

// attack ejecuta UN solo request HTTP en paralelo
func attack(url string, method string, body string, token string, wg *sync.WaitGroup, results chan Result) {
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
		results <- Result{Status: 0, Body: "error"}
		return
	}

	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		results <- Result{Status: 0, Body: "error"}
		return
	}
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(resp.Body)
	results <- Result{Status: resp.StatusCode, Body: string(bodyBytes)}
}

// parseJSON intenta parsear el body como JSON y devuelve los campos clave
func parseJSON(body string) map[string]interface{} {
	var parsed map[string]interface{}
	if err := json.Unmarshal([]byte(body), &parsed); err != nil {
		return nil
	}
	return parsed
}

// diffJSON compara dos JSON y muestra los campos con valores diferentes
func diffJSON(a, b map[string]interface{}) []string {
	var diffs []string
	for key, valA := range a {
		if valB, ok := b[key]; ok {
			if fmt.Sprintf("%v", valA) != fmt.Sprintf("%v", valB) {
				diffs = append(diffs, fmt.Sprintf("  campo '%s': %v → %v", key, valA, valB))
			}
		}
	}
	return diffs
}

// enumerate itera IDs numéricos y reporta los que existen
func enumerate(baseUrl string, start int, end int) {
	fmt.Printf("[*] Enumerando %s del ID %d al %d...\n", baseUrl, start, end)
	found := 0

	for i := start; i <= end; i++ {
		url := fmt.Sprintf("%s/%d", baseUrl, i)
		resp, err := http.Get(url)
		if err != nil {
			continue
		}
		resp.Body.Close()

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

	if len(os.Args) >= 4 {
		method = os.Args[3]
	}
	if len(os.Args) >= 5 {
		body = os.Args[4]
	}
	if len(os.Args) >= 6 {
		token = os.Args[5]
	}

	results := make(chan Result, threads)
	var wg sync.WaitGroup

	for i := 0; i < threads; i++ {
		wg.Add(1)
		go attack(url, method, body, token, &wg, results)
	}

	wg.Wait()
	close(results)

	// Agrupa por status code
	statusCount := make(map[int]int)
	// Guarda un body por status para comparar
	statusBody := make(map[int]string)

	for r := range results {
		statusCount[r.Status]++
		if _, exists := statusBody[r.Status]; !exists {
			statusBody[r.Status] = r.Body
		}
	}

	if len(statusCount) > 1 {
		fmt.Println("[!] INCONSISTENCIA DETECTADA - posible race condition")
	} else {
		fmt.Println("[+] respuestas consistentes")
	}

	for status, count := range statusCount {
		fmt.Printf("  %d → %d veces\n", status, count)

		// Intenta mostrar el JSON de respuesta
		parsed := parseJSON(statusBody[status])
		if parsed != nil {
			fmt.Printf("  body: ")
			for key, val := range parsed {
				fmt.Printf("%s=%v ", key, val)
			}
			fmt.Println()
		}
	}

	// Si hay exactamente 2 status distintos, muestra el diff de los bodies
	if len(statusCount) == 2 {
		statuses := make([]int, 0)
		for s := range statusCount {
			statuses = append(statuses, s)
		}
		jsonA := parseJSON(statusBody[statuses[0]])
		jsonB := parseJSON(statusBody[statuses[1]])

		if jsonA != nil && jsonB != nil {
			diffs := diffJSON(jsonA, jsonB)
			if len(diffs) > 0 {
				fmt.Println("\n[!] CAMPOS INCONSISTENTES:")
				for _, d := range diffs {
					fmt.Println(d)
				}
			}
		}
	}
}