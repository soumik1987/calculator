# Design

How we're building what's in requirements.md, and why.

## Layout

```
browser -> Caddy -> /api/*  -> backend   Go, calculator-api, 8081
                 -> the rest -> frontend  React app, 3000 (8080 in dev)
```

```
backend/
  cmd/server/          main, wires config, logger, telemetry, middleware and the server
  internal/calculator/ lexer.go, parser.go, expression.go, calculator.go
  internal/httpapi/    routes, JSON in and out, error mapping, middleware
  internal/telemetry/  tracer and meter providers, OTLP export
  internal/config/     reads the env
frontend/src/
  lib/calculator.ts        keypad state and number formatting
  lib/api.ts               the fetch call, ApiError, error codes
  hooks/use-calculator.ts  the react query mutation
  components/calculator/   Calculator, Display, Keypad
  routes/index.tsx         home page
deploy/                    docker-compose.yml, Caddyfile, .env.example
```

Go module is github.com/soumik1987/sami-calculator/backend, Go 1.26 with the toolchain pinned in go.mod. Bump the toolchain when govulncheck flags the stdlib.

## Backend

### Evaluating

The API only needs one thing from the calculator package, Evaluate(expression string) (float64, error). Nothing else is exported apart from the error values and SyntaxError.

Inside it's lexer, parser, then evaluate, and each one has its own kind of error. The lexer catches bad characters, unknown words and bad numbers. The parser catches bad structure. Evaluation catches the maths. Syntax errors are a *SyntaxError with the position already in the message, so the HTTP layer can pass the message straight through.

Because parsing finishes before any evaluation, "1/0+" fails as incomplete and never gets to the division.

The lexer has two maps. operators maps a symbol to a token kind: + - * x / ^ % ( ), plus the unicode minus, times and divide signs the keypad sends. functions maps a name to a Go function: the root sign, sqrt, sin, cos, log and exp. Lookup lowercases the text first, so names are case-insensitive. A run of letters is looked up as one word, so "tan" fails as an unknown word rather than three bad characters. Numbers are digits and dots, then an optional e or E with an optional sign and digits, parsed with strconv.ParseFloat. Positions count runes from 1.

Adding a function like tan should be one line in the functions map plus the function in calculator.go. Nothing in the grammar or the parser changes.

Hand-written recursive descent parser. Keep the grammar at the top of parser.go, lowest precedence first:

```
expression = term { ("+" | "-") term }
term       = unary { ("*" | "/") unary }
unary      = ("+" | "-") unary | power
power      = postfix [ "^" unary ]
postfix    = primary { "%" }
primary    = number | "(" expression ")" | function primary
```

power takes a unary on its right, which is what makes ^ go right to left and lets 2^-1 work. unary sits above power, which is why -2^2 is -4. A function takes a primary, not an expression, which is why "sqrt16^2" is 16 and "sqrt-4" is an error. Implied multiplication fails because after a primary the parser expects an operator and sees a ( or a function.

We don't need a separate nesting limit. The 500 character check runs before tokenizing and that already caps how deep it can recurse. Tests cover 500 characters of brackets and of minuses.

The tree is a node interface with an unexported eval method, with number, unary and binary nodes. Unexported so nothing outside the package can add node types. Each unary and binary node holds one of the functions from calculator.go, and every node's result goes through finite: NaN becomes ErrUndefinedResult, +Inf or -Inf becomes ErrOutOfRange, -0 becomes 0. Checking at every node, not just the end, means an overflow in the middle can't turn into something else later.

### The maths

All in calculator.go. One small function per operation with the signature func(a) (float64, error) or func(a, b) (float64, error). The errors are sentinels, matched with errors.Is.

divide fails with ErrDivisionByZero when b is 0. That covers 0/0 too.
power fails with ErrDivisionByZero when a is 0 and b is negative. Otherwise math.Pow, and (-8)^0.5 comes back NaN and finite turns it into ErrUndefinedResult.
squareRoot fails with ErrNegativeSquareRoot below 0.
logarithm fails with ErrNonPositiveLogarithm at 0 or below, -0 included. Otherwise math.Log10.
exponential is just math.Exp. exp(1000) overflows and finite catches it.
percent divides by 100, negate flips the sign.

sin and cos take degrees. Reduce the angle mod 360 before converting to radians, so big angles don't lose precision. If the angle mod 180 is 0, sin returns exactly 0. If the angle mod 180 is 90 or -90, cos returns exactly 0.

### HTTP

Standard http.ServeMux with method patterns, POST /api/v1/calculations and GET /healthz. Wrong method gets the mux's own 405.

NewHandler(evaluate Evaluator, logger *slog.Logger) http.Handler. Evaluator is type Evaluator func(expression string) (float64, error). main passes calculator.Evaluate, tests pass fakes, including one that returns a plain error to check the 500 path.

Request handling goes like this. Check Content-Type with mime.ParseMediaType so a charset parameter is fine. Wrap the body in http.MaxBytesReader at 4 KB. Decode with DisallowUnknownFields. Then try decoding again and make sure it hits EOF, so there's only one value. Each failure maps to its own message from requirements.md.

errors.go turns errors into responses. A *SyntaxError is 400 invalid_expression with its own message. The maths errors go through a small table of sentinel, status, code and message, so adding a new one is one row. Anything else gets logged with the trace id and returns 500 internal_error. Error codes are a typed ErrorCode string and messages are constants. They have to match requirements.md exactly. Every error response goes through one fail function, which also tags the span with error.type.

Responses are always JSON with Content-Type application/json.

Middleware, from the outside in: telemetry, logger, recover, security headers, CORS, then the routes. Telemetry is outermost so the logger and everything after it can see the trace id. Recover is inside the logger so a panic still gets its log line with a 500. Recover re-panics http.ErrAbortHandler since that one's meant to abort.

CORS adds Vary: Origin always and exposes X-Trace-ID. If the origin is on the list it's echoed back in Access-Control-Allow-Origin. OPTIONS gets a 204 with the allowed methods and headers and a max age of 600. The code currently treats * in ALLOWED_ORIGINS as allow everything, which the requirements don't want. It's on the open list in tasks.md.

The server is http.Server with ReadHeaderTimeout 5s, ReadTimeout 10s, WriteTimeout 10s, IdleTimeout 60s and MaxHeaderBytes 16 KB. On SIGINT or SIGTERM, Shutdown with a 10s deadline, then shut telemetry down with 5s to flush.

### Config

config.Load(getenv func(string) string) so tests can pass a map instead of touching the real env. Returns a Config with Addr, AllowedOrigins (split on commas, trimmed, empties dropped), LogLevel as slog.Level and OTLPEndpoint. A bad LOG_LEVEL is an error and main exits with code 1.

Logging is log/slog with the JSON handler to stdout.

### Telemetry

telemetry.Setup(ctx, export bool) returns a Providers struct with the tracer and meter providers, and we pass them in instead of setting them globally. The resource has service.name calculator-api and then reads the OTEL_ env vars, so OTEL_SERVICE_NAME overrides it. When export is false there are no exporters but spans still get real IDs, so every response has a trace id. When it's true, the OTLP/HTTP trace and metric exporters read the standard OTEL_EXPORTER_OTLP_ vars.

The tracing middleware is otelhttp. It names spans after the matched route, skips /healthz, uses the W3C trace context propagator so an incoming traceparent continues, sets X-Trace-ID on the response, and the handler sets error.type using the semconv key.

## Frontend

React 19, TypeScript, TanStack Start and Router, TanStack Query, Tailwind and shadcn buttons. Started in Lovable.

lib/calculator.ts holds CalculatorState, which is the expression, and maybe a result or an error. press, backspace and showResult take a state and return a new one, no React, so they're tested directly. press checks the 500 limit and the carry-on-from-result rule. backspace checks whether the expression ends with one of the function keys and removes the whole thing. formatNumber uses Intl.NumberFormat with 12 significant digits, and carrying on from a result uses the same rounding so what you see is what goes in.

lib/api.ts is the only place that calls fetch. calculatorApi.calculate posts the expression to the base url plus the endpoint. If fetch itself throws it's an ApiError with status 0 and code network_error. If the response isn't ok or doesn't have a numeric result, it reads the error body, and anything missing falls back to unknown_error and the generic message. ApiError keeps status, code, message and the X-Trace-ID header.

useCalculate is a react query useMutation around that. Calculator.tsx calls it on equals and uses isPending to disable the equals key. Pressing any other key resets the mutation so the old error disappears.

Calculator.tsx also listens for keydown on the window, ignores anything with meta, ctrl or alt, maps keys through a small table and passes them to the same handler as the buttons. Keypad is a 4 column grid built from a layout array, each key with a style (digit, operator, control, equals) and an optional label, so sin shows "sin" but types "sin(". Display shows the expression, then the error or the result, in an aria-live region.

Key values, function keys, key styles, error codes, endpoints, HTTP methods, headers and messages are all enums. No loose strings.

VITE_API_URL is the base url and stays empty when it's same origin. vite.config.ts proxies /api to localhost:8081 in dev. The dev server port is 8080, Lovable's config forces it.

Tests are vitest with Testing Library, and fetch is mocked for the api tests.

## Deploy

Backend Dockerfile is two stages. Build with golang alpine, CGO_ENABLED=0 and -trimpath, then copy the binary into gcr.io/distroless/static-debian12:nonroot and run as nonroot. Frontend is node:22-alpine, build then run the server output as the node user on 3000. Each image builds from its own folder.

Compose has three services, api, web and proxy. Only proxy publishes ports, 80 and 443. All restart unless stopped. Proxy depends on api and web but there are no healthchecks. The Caddyfile reverse proxies /api/* to api:8081 and everything else to web:3000, with zstd and gzip. deploy/.env holds SITE_ADDRESS, LOG_LEVEL and OTEL_EXPORTER_OTLP_ENDPOINT, which get passed through.

## Rules for the code

Go. Stick to the standard library. OpenTelemetry is the only exception so far and anything new needs a reason. One package per concern, so everything calculator related is in calculator, not split into models, services and controllers. Interfaces go on the consumer side, never next to the implementation, and only when there's a real second implementation or a test actually needs one. Ask me before adding one. If it's a single method a func type is enough. Sentinel errors for the expected failures, a typed error when it needs to carry data like SyntaxError does. No globals and no init magic, pass things in. make check has to pass, which is go vet, golangci-lint, govulncheck and the tests with -race.

TypeScript. Strict mode, no any, enums instead of string constants. Keep logic out of components so it can be tested without rendering. Don't edit routeTree.gen.ts, it's generated. npm test, typecheck and lint have to pass.

Tests. Write them first. Table tests in Go. Name tests after the behaviour, like "power goes right to left" or it("deletes sin( in one press"), so it's obvious which rule a test covers. Calculator package stays at 100% coverage. Error messages are checked word for word.

Docs. Spec changes first, then code, then READMEs. READMEs describe what's built, the spec describes what we want. Keep them short, and don't write comments that just repeat the code.

## What to test

Not a full list, just the cases I want to see covered at least.

Calculator. Every operator on its own and mixed. Precedence, like 2+3*4 and -2^2. Power going right to left. Brackets nested a few levels deep. Percent. Numbers with decimals, a leading dot and exponents. Upper and lower case function names. Spaces everywhere and nowhere. sin and cos at 0, 30, 90, 180, 270 and some big angle like 3690. log of 1, 10 and 1000, exp of 0 and 1.

Calculator errors. Every message in requirements.md, word for word, including the position. Divide by zero, 0/0 and 0 to a negative power. Root and log of negatives, and log of 0 and -0. Overflow from power and from exp. NaN from a negative to a fractional power. Something that overflows halfway through but would look fine at the end. 1/0+ coming back as incomplete. Exactly 500 characters working and 501 failing. 500 characters of nested brackets and of minuses not blowing up.

Results. -1*0 is 0 not -0. Underflow is 0. 0.1+0.2 gives the float answer.

HTTP. A good request end to end. Each bad body case with its message: empty, not JSON, an array, two objects, an extra field, a number instead of a string. Wrong content type and a content type with a charset. Over 4 KB. Each maths error to the right status and code. An unknown error gives a 500 and the real message isn't in the body. Wrong method on the route. healthz.

Middleware. A panic gives a 500 and the next request still works. Security headers are on every response. CORS for an allowed origin, a random origin, no origin, and a preflight. The trace id header is there, an incoming traceparent is kept, and error.type is set on failures. healthz isn't traced.

Config. Defaults when nothing is set, each var set, a list of origins with spaces and empty bits, and a bad LOG_LEVEL.

Frontend logic. Typing builds the expression. After a result an operator carries on and a digit starts over. Backspace on a normal character and on sin(. The 500 limit. Number formatting for big, small and long decimals.

Frontend api client. A good response. An error response passes the backend message through. A network failure. A response that isn't JSON or has no result. The trace id is kept on the error.

Frontend components. Clicking keys and typing on the keyboard both work. Equals shows the result. An error shows and goes away when you type. Equals is disabled while a request is running, and pressing it twice only sends once. Ctrl and Cmd shortcuts are left alone.

## Decisions

Backend does all the maths when you press equals, so there's one place that owns the rules and the errors. We lose live preview and need the backend for every result.

We send the typed text, not a parsed tree. Otherwise we'd be parsing in the browser too and keeping two parsers in sync.

Hand-written parser rather than a library. No dependency, errors with positions, and it handles the root sign, percent and the unicode times sign, which things like expr or govaluate don't. The downside is grammar changes are code changes.

400 when we can't read the input, 422 when the maths has no answer. A client can tell a bug from a user mistake.

Evaluator is a func type. One method and one implementation doesn't need an interface.

float64 everywhere because it's native in Go, JSON and JS. We accept the binary rounding. math/big would fix 0.1+0.2 but nobody's asked.

Degrees only, like a phone. A radians mode would mean state or a second set of functions.

Same origin through Caddy in prod and the Vite proxy in dev, so we normally don't need CORS.

No container healthchecks. Distroless has no shell or curl and nothing needs to poll. Caddy might 502 for a moment on startup. If that becomes a problem, add a healthcheck subcommand to the binary and use depends_on with service_healthy.

Spans get created even when we're not exporting, so there's always a trace id. Costs a bit of CPU per request.

The UI checks the 500 limit as well, so nobody hits the 4 KB body limit and its not so friendly message.

Stdlib plus OpenTelemetry makes the binary around 17 MB. Fine for now.
