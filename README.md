## Real Findings

### Rate limiter race condition
RaceHunter detected a race condition in the Redis-based 
rate limiter itself.

Under concurrent load, multiple requests passed the limiter 
before the counter updated.

The defense was vulnerable to the same attack 
it was designed to prevent.

Mitigation: Nginx as an additional layer.

### IDOR via numeric ID enumeration
enumerate confirmed all product IDs accessible 
without authentication.

Mitigation: UUID migration pending.

### Stock inconsistency under concurrency
attack confirmed race condition in stock endpoint 
before SELECT FOR UPDATE was implemented.

stock: 1 → -1 under 100 concurrent requests.