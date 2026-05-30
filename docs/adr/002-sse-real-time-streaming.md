# ADR-002: Server-Sent Events for Real-Time Log Streaming

**Status:** Accepted  
**Date:** 2026-05-30  
**Author:** log-service team

---

## Context

log-service is a centralized log broker: applications send structured log entries via `POST /api/v1/logs`, the service persists them to MongoDB, and consumers (dashboards, monitoring tools) receive those logs in real time.

The initial implementation used the `r3labs/sse/v2` library to handle the SSE transport layer. While functional, this introduced a coupling problem: the library expects a `?stream=<name>` query parameter to identify the channel, which conflicts with the RESTful route design (`/api/v1/events/{applicationID}`). The current workaround in `routes.go` manually rewrites the request URL to inject the query parameter before delegating to the library's handler — a clear sign that the abstraction is working against the design.

This ADR documents the decision to replace `r3labs/sse` with a native Go SSE implementation, and reflects on the broader architectural choice of using SSE over alternatives such as WebSockets or a message broker like Kafka.

---

## Decision

### 1. Replace `r3labs/sse` with a native implementation

SSE is a simple protocol defined in the [HTML Living Standard](https://html.spec.whatwg.org/multipage/server-sent-events.html). A message is a plain-text frame:

```
data: {"level":"ERROR","message":"..."}\n\n
```

Go's `net/http` provides everything needed to implement this natively: `http.Flusher` for streaming, `http.ResponseWriter` for writing frames, and `r.Context().Done()` for detecting client disconnection. There is no justification for a third-party library dependency here.

Removing `r3labs/sse` eliminates the URL-rewriting workaround, gives full control over the channel lifecycle, and makes the streaming behavior explicit and auditable within the codebase.

### 2. Retain SSE as the streaming transport

SSE is the appropriate transport for this service's use case. The reasoning is detailed in the alternatives section below.

---

## Alternatives Considered

### WebSockets

WebSockets provide full-duplex communication. For log streaming, the consumer never sends data back to the server — the channel is unidirectional by nature. WebSockets would add protocol complexity (upgrade handshake, framing, ping/pong) with no benefit over SSE for this use case.

**Rejected:** unnecessary complexity for a unidirectional stream.

### Kafka (or another message broker)

Kafka is a distributed log designed for high-throughput, durable, replayable event streams between services. It would be appropriate if:

- Multiple independent services needed to consume the same log events.
- Log replay (consuming past events) was a requirement.
- The ingestion volume exceeded what a single service instance could handle.

None of these conditions apply to the current scope. Introducing Kafka would require operating a broker cluster, managing consumer groups, and handling offset commits — significant operational overhead for a service whose primary value is simplicity and low latency delivery to a dashboard consumer.

A relevant trade-off to acknowledge: Kafka provides durability at the transport layer (events are persisted in the broker). SSE does not — if no consumer is connected when a log is ingested, the event is not delivered via SSE (though it is still persisted in MongoDB). This is an acceptable trade-off for the current use case, where the SSE stream is intended for live monitoring, not guaranteed delivery.

**Rejected for current scope.** Kafka remains a valid future evolution if the service grows into a multi-consumer pipeline.

### Polling (`GET /api/v1/logs?since=<timestamp>`)

Simple and stateless, but requires the client to manage timing, generates unnecessary requests when there is no new data, and introduces latency proportional to the polling interval.

**Rejected:** SSE provides lower latency with less client complexity.

---

## MongoDB as the Persistence Layer

MongoDB was chosen for its flexible document schema, which accommodates the `tags` and `metadata` fields without requiring schema migrations as log structures evolve across different applications.

Known trade-offs that should be addressed for production readiness:

- **Index strategy:** queries and SSE fan-out both benefit from a compound index on `(application_id, timestamp)`. Without it, reads degrade as the collection grows.
- **TTL index:** log data has a natural expiry. A TTL index on `timestamp` prevents unbounded collection growth without requiring a separate cleanup job.
- **Write volume:** MongoDB handles moderate write throughput well, but under high ingestion rates, write concern and connection pool configuration become relevant.

These are operational concerns, not architectural blockers. The schema flexibility and operational simplicity of MongoDB remain appropriate for this service's scope.

---

## Consequences

- The `r3labs/sse/v2` dependency is removed from `go.mod`.
- The URL-rewriting workaround in `routes.go` is eliminated.
- The SSE channel lifecycle (creation, fan-out, cleanup on disconnect) is managed explicitly in `internal/infrastructure/http/sse/server.go`.
- The route handler becomes a clean delegation: `cfg.SSEServer.Subscribe(w, chi.URLParam(r, "applicationID"))`.
- If Kafka adoption becomes necessary in the future, the `Publish` interface on the SSE server can be adapted to consume from a Kafka topic without changes to the HTTP layer.

---

## References

- [HTML Living Standard — Server-Sent Events](https://html.spec.whatwg.org/multipage/server-sent-events.html)
- [MDN — Using server-sent events](https://developer.mozilla.org/en-US/docs/Web/API/Server-sent_events/Using_server-sent_events)
- [r3labs/sse — GitHub](https://github.com/r3labs/sse)
- [Apache Kafka — Introduction](https://kafka.apache.org/intro)
- [MongoDB TTL Indexes](https://www.mongodb.com/docs/manual/core/index-ttl/)
