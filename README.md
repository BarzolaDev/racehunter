# RaceHunter

CLI tool written in Go to detect race conditions and enumeration vulnerabilities in HTTP endpoints.

## Why this exists

Race conditions are one of the most overlooked vulnerabilities in web applications. They don't show up in static analysis or basic testing — they only appear under real concurrency.

This tool was built after demonstrating race conditions in a real inventory API:
- Python: built the vulnerable system and defended it with SELECT FOR UPDATE
- Java: demonstrated the attack with 1000 concurrent threads
- Go: built the tool to detect it automatically in any system

## Commands

### attack — Race condition detection

Sends N concurrent requests to the same endpoint and detects inconsistent responses.
Now parses JSON response bodies to show exactly which fields are inconsistent across concurrent requests.

```bash
go build -o racehunter main.go

# GET
./racehunter https://target.com/endpoint 10

# POST with body
./racehunter https://target.com/checkout 10 POST '{"product_id":123}'

# Authenticated endpoint
./racehunter https://target.com/stock 100 POST '{"quantity":-1}' YOUR_TOKEN
```

**Output:**
```
[!] INCONSISTENCIA DETECTADA - posible race condition
  200 → 95 veces
  body: stock=1
  400 → 5 veces
  body: stock=-1

[!] CAMPOS INCONSISTENTES:
  campo 'stock': 1 → -1
```

### enumerate — Numeric ID enumeration (IDOR)

Iterates numeric IDs and reports which resources exist without authentication.

```bash
./racehunter enumerate https://target.com/products 1 100
```

**Output against inventory-api (numeric IDs):**
```
[*] Enumerando http://localhost:8000/products del ID 1 al 10...
[!] ID 1 existe → http://localhost:8000/products/1
[!] ID 2 existe → http://localhost:8000/products/2
[!] VULNERABLE — IDs numéricos enumerables. Migrar a UUID.
```

**Fix:** migrating IDs to UUID eliminates brute-force enumeration.  
Covered under OWASP A01 — Broken Access Control.

## How it works

**attack:**
1. Launches N goroutines simultaneously against the target endpoint
2. Collects all responses via a channel
3. Parses JSON response bodies and compares field values across responses
4. If responses are inconsistent — race condition detected with exact field diff

**enumerate:**
1. Iterates numeric IDs sequentially
2. Reports every ID that returns 200
3. Flags the system as vulnerable if any exist

## Disclaimer

Use only on systems you own or have explicit authorization to test — bug bounty programs with defined scope. Unauthorized use against third-party systems is illegal.