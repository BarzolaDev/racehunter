# RaceHunter

CLI tool written in Go to detect race conditions in HTTP endpoints.

Sends N concurrent requests to the same endpoint and detects inconsistent responses — the signature of a race condition vulnerability.

## Why this exists

Race conditions are one of the most overlooked vulnerabilities in web applications. They don't show up in static analysis or basic testing — they only appear under real concurrency.

This tool was built after demonstrating race conditions in a real inventory API:
- Python: built the vulnerable system and defended it with SELECT FOR UPDATE
- Java: demonstrated the attack with 1000 concurrent threads
- Go: built the tool to detect it automatically in any system

## Usage

```bash
go build -o racehunter main.go

# GET
./racehunter https://target.com/endpoint 10

# POST with body
./racehunter https://target.com/checkout 10 POST '{"product_id":123}'
```

## How it works

1. Launches N goroutines simultaneously against the target endpoint
2. Collects all responses via a channel
3. Groups responses by content
4. If responses are inconsistent — race condition detected

## Output
[!] INCONSISTENCIA DETECTADA - posible race condition
[+] respuestas consistentes

## Disclaimer

Use only on systems you own or have explicit authorization to test — bug bounty programs with defined scope. Unauthorized use against third-party systems is illegal.