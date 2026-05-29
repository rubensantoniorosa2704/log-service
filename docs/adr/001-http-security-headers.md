# ADR-001: HTTP Security Headers Standard

**Status:** Accepted  
**Date:** 2026-05-29  
**Author:** log-service team

---

## Context

The log-service exposes HTTP endpoints for log ingestion (`POST /api/v1/logs`) and real-time streaming via Server-Sent Events (`GET /api/v1/events/{applicationID}`). As a service that receives structured application data and streams it to consumers, HTTP responses must include security headers to instruct clients and browsers on safe handling of the content.

During a code review of the HTTP layer, two issues were identified:

1. No security headers were being set on any response.
2. CORS headers were defined in two places: the Chi middleware in `routes.go` and directly in the SSE server configuration in `sse/server.go`, creating duplication and a potential for inconsistency.

---

## Decision

All HTTP security headers will be applied through a **single global middleware** registered in `routes.go`, alongside the existing Chi middlewares. This is the only place where security headers are defined — no handler or sub-component (including the SSE server) should set its own security or CORS headers.

### Standards Reference

This decision follows the [OWASP HTTP Headers Cheat Sheet](https://cheatsheetseries.owasp.org/cheatsheets/HTTP_Headers_Cheat_Sheet.html) and the relevant IETF specifications:

| Header | Specification |
|---|---|
| `X-Content-Type-Options` | [WHATWG Fetch Standard](https://fetch.spec.whatwg.org/#x-content-type-options-header) |
| `X-Frame-Options` | [RFC 7034](https://datatracker.ietf.org/doc/html/rfc7034) |
| `Referrer-Policy` | [W3C Referrer Policy](https://www.w3.org/TR/referrer-policy/) |
| `Content-Security-Policy` | [W3C CSP Level 3](https://www.w3.org/TR/CSP3/) |
| `Strict-Transport-Security` | [RFC 6797](https://datatracker.ietf.org/doc/html/rfc6797) |
| `Cache-Control` | [RFC 9111](https://datatracker.ietf.org/doc/html/rfc9111) |
| CORS (`Access-Control-*`) | [W3C CORS / Fetch Standard](https://fetch.spec.whatwg.org/#http-cors-protocol) |

---

## Headers Applied

### `X-Content-Type-Options: nosniff`

Prevents browsers from MIME-sniffing a response away from the declared `Content-Type`. This mitigates MIME confusion attacks where a browser might interpret a JSON response as executable content.

### `X-Frame-Options: DENY`

Prevents the service responses from being embedded in `<frame>`, `<iframe>`, or `<object>` elements. This mitigates clickjacking attacks. Note: for modern browsers, `Content-Security-Policy: frame-ancestors 'none'` is the preferred mechanism; `X-Frame-Options` is retained for compatibility with older clients.

### `Referrer-Policy: strict-origin-when-cross-origin`

Controls how much referrer information is included in requests. With this policy, the full URL is sent for same-origin requests, and only the origin is sent for cross-origin requests. No referrer is sent for requests to less-secure destinations (HTTP from HTTPS). This is the OWASP-recommended default.

### `Content-Security-Policy: default-src 'none'`

Since all API endpoints return JSON or SSE streams (not HTML pages that load scripts or assets), a restrictive CSP is appropriate. `default-src 'none'` blocks all resource loading, which is correct for a pure API service. The Swagger UI routes (`/swagger/*`, `/docs/*`) are excluded from this policy as they require loading scripts and styles.

### `Cache-Control: no-store`

Prevents any caching of API responses. Log data may contain sensitive application information and must not be stored in browser or proxy caches. Per [RFC 9111 §5.2.2.5](https://datatracker.ietf.org/doc/html/rfc9111#section-5.2.2.5), `no-store` instructs caches not to store any part of the request or response.

> **SSE exception:** The SSE endpoint (`/api/v1/events/{applicationID}`) requires `Cache-Control: no-cache` (not `no-store`) to maintain the streaming connection correctly. This is handled by the SSE library and must not be overridden by the global middleware.

---

## CORS Consolidation

CORS headers (`Access-Control-Allow-Origin`, `Access-Control-Allow-Methods`, etc.) were previously duplicated between the Chi CORS middleware and the SSE server's `s.Headers` map. This creates a risk of inconsistency and makes it harder to audit the effective policy.

**Decision:** CORS is handled exclusively by the Chi `cors.Handler` middleware in `routes.go`. The `s.Headers` map in `sse/server.go` is removed. The SSE library will inherit the CORS headers set by the middleware.

---

## Consequences

- All responses from the service will carry the defined security headers.
- There is a single place to audit and update the security header policy.
- The SSE server no longer manages its own headers, reducing duplication.
- The Swagger UI routes are intentionally excluded from the restrictive CSP to remain functional.
- Future handlers added to the router automatically inherit the security headers without any additional work.

---

## References

- [OWASP HTTP Headers Cheat Sheet](https://cheatsheetseries.owasp.org/cheatsheets/HTTP_Headers_Cheat_Sheet.html)
- [RFC 6797 — HTTP Strict Transport Security (HSTS)](https://datatracker.ietf.org/doc/html/rfc6797)
- [RFC 7034 — HTTP Header Field X-Frame-Options](https://datatracker.ietf.org/doc/html/rfc7034)
- [RFC 9111 — HTTP Caching](https://datatracker.ietf.org/doc/html/rfc9111)
- [W3C Content Security Policy Level 3](https://www.w3.org/TR/CSP3/)
- [W3C Referrer Policy](https://www.w3.org/TR/referrer-policy/)
- [WHATWG Fetch Standard — CORS](https://fetch.spec.whatwg.org/#http-cors-protocol)
