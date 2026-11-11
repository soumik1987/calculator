# Tasks

## Build order

If we were starting from scratch this is the order. Each step is small and ends with its tests passing.

1. Go module, internal/config with the defaults and env loading, a bare main, and a Makefile with run, test and check (vet, golangci-lint, govulncheck, tests with race).

2. Calculator package. Lexer, parser with the grammar at the top, the tree, the basic operators, the error sentinels and the finite check. Table tests for every example under Expressions, Results and Errors in requirements.md, messages checked word for word, 100% coverage.

3. The functions table with the root sign, "sqrt", sin, cos, log and exp, degrees with the exact multiples of 90, and the log error.

4. httpapi handler. Routes, content type check, strict decoding with the 4 KB cap, error mapping table, healthz. Handler tests with a fake Evaluator, including the 500 path and that nothing internal leaks.

5. Middleware for recover, security headers, CORS and request logging, then wire it in main with the server timeouts and graceful shutdown.

6. telemetry package and the otelhttp middleware, trace id header, error.type. Try it against a real collector once before calling it done.

7. Frontend keypad logic in lib/calculator.ts and its tests.

8. Frontend api client, the hook, the components and keyboard handling, with tests for the api client and the components.

9. Dockerfiles, compose, Caddyfile and the .env.example files.

10. READMEs for the root, backend, frontend and deploy. Setup, how to run, curl examples, the error table, design decisions.

All of that is done already.

## Open

Things where the code and the spec don't match yet. Remove a line once it's fixed.

CORS treats * in ALLOWED_ORIGINS as allow everything, but the spec says allowlist only. Drop the wildcard support.

Go through the tests against requirements.md and the What to test list in design.md. Add tests for anything that isn't covered, and rename any test whose name doesn't say what behaviour it checks.
