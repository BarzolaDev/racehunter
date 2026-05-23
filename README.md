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

Every ID exposed. Every resource mapped.
No authentication required.

This finding revealed a deeper problem:
the API was returning more information than necessary.
If they can enumerate it, they can extract it.

Mitigation: UUID migration pending.
Longer term: return only what the client needs. Nothing more.