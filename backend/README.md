# Calculator service

The Go service that evaluates expressions. It keeps no data between requests. Apart from
OpenTelemetry, it uses only the standard library.

## Run

```sh
cp .env.example .env
make run
```

It listens on http://localhost:8081. Settings:

- `PORT`: default `8081`
- `ALLOWED_ORIGINS`: web addresses that browsers may call from, comma-separated. Default
  `http://localhost:8080`
- `LOG_LEVEL`: `debug`, `info`, `warn` or `error`. Default `info`
- `OTEL_EXPORTER_OTLP_ENDPOINT`: OTLP/HTTP collector, e.g. `http://localhost:4318`. Empty
  means nothing is exported. The other standard `OTEL_*` variables (`OTEL_SERVICE_NAME`,
  `OTEL_EXPORTER_OTLP_HEADERS`, ...) work too

## API

```
POST /api/v1/calculations
Content-Type: application/json

{"expression": "(986×(4+8))+2"}
```

```json
{ "expression": "(986×(4+8))+2", "result": 11834 }
```

The expression rules are in the [main README](../README.md#calculation-rules). Besides the
keypad symbols, it also accepts `*`, `/`, `x`, `-` and `sqrt`. The body must be JSON, under
4 KB, with only the `expression` field.

`GET /healthz` returns `{"status":"ok"}`.

### Telemetry

Each request gets an OpenTelemetry span named after its route, plus the standard HTTP server
metrics. A caller's W3C `traceparent` header is continued. The trace ID is returned in
`X-Trace-ID` and written on the request's log line, so a response, its log and its trace
can be matched. Failed requests carry `error.type` with the error code. `/healthz` is not
traced.

### Errors

```json
{ "error": { "code": "division_by_zero", "message": "Cannot divide by zero." } }
```

| Status | Code                     | Cause                                                   |
| ------ | ------------------------ | ------------------------------------------------------- |
| 400    | `invalid_request`        | Body is empty, isn't a JSON object, or has wrong fields |
| 400    | `invalid_expression`     | Expression can't be read; the message says where        |
| 413    | `request_too_large`      | Body is over 4 KB                                       |
| 415    | `unsupported_media_type` | `Content-Type` isn't `application/json`                 |
| 422    | `division_by_zero`       | Dividing by 0, or 0 to a negative power                 |
| 422    | `negative_square_root`   | Square root of a negative number                        |
| 422    | `non_positive_logarithm` | `log` of zero or a negative number                      |
| 422    | `undefined_result`       | No real answer, e.g. `(-8)^0.5`                         |
| 422    | `out_of_range`           | Result too large, e.g. `10^1000`                        |
| 500    | `internal_error`         | A bug; details go to the log only                       |

## Code

```
cmd/server/          starts the server and stops it cleanly on SIGINT or SIGTERM
internal/calculator/
  lexer.go           text → tokens
  parser.go          tokens → tree, by precedence (the grammar is at the top)
  expression.go      evaluates the tree
  calculator.go      one function per operation, and the error rules
internal/httpapi/    route, JSON, errors, middleware (tracing, logging, recovery, CORS)
internal/telemetry/  OpenTelemetry tracer and meter providers, OTLP export
internal/config/     reads the settings
```

Each node in the tree holds one of the functions from `calculator.go`. Every result is
checked the same way: NaN becomes `undefined_result`, infinity becomes `out_of_range`, and
-0 becomes 0. `(986×(4+8))+2` becomes:

```
      +
     / \
    ×   2
   / \
 986   +
      / \
     4   8
```

## Design decisions

- **Lexer, parser, then evaluation.** Each step reports one kind of error: characters,
  structure, maths. The whole expression is read before any maths, so `1/0 +` is
  "incomplete" (400), not "divide by zero" (422).
- **Hand-written recursive-descent parser.** No dependency, errors carry positions, and it
  supports `√ % ×`. Libraries such as `expr` or `govaluate` lack both. Cost: grammar
  changes are code changes.
- **400 for input that can't be read, 422 for maths with no answer.** Clients can tell a bug
  from a user mistake.
- **Functions are one table** mapping a name to a Go function (`√`, `sin`, ...). The parser
  treats them all like `√`, so adding one changes no grammar.
- **Degrees, not radians**, as on a phone. Multiples of 90° are special-cased so `sin(180)`
  is 0, not 1.2e-16. Cost: no radians mode.
- **`1e5` notation is accepted.** The frontend sends large results back in that form, e.g.
  `1e+21`.
- **`httpapi.NewHandler` takes `Evaluate` as a function.** Tests pass fakes, and a single
  method doesn't need an interface.
- **Standard library only, except OpenTelemetry.** Cost: the binary is about 17 MB.
- **Spans are created even when nothing is exported.** Every response gets a trace ID. Cost:
  a little CPU per request.

## Tests

```sh
make test    # tests with the race detector
make check   # go vet, linters, vulnerability scan, tests
```
