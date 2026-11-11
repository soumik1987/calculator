## UI prompt


Build a simple frontend for a calculator. A separate Go service does the maths, the frontend just builds the expression, sends it when you hit equals and shows whatever comes back. Should feel like a phone calculator and work well on mobile. React, TypeScript, Tailwind.


## Repo prompts

Iteration 1

I want to build a full-stack calculator, spec first. I've put my raw notes in the spec folder, requirements, design and tasks, plus a rough list of the test cases I care about. Go through them and format them properly so they read like a polished spec, but don't drop anything or add things I didn't ask for. If something is missing or contradicts itself, ask me. React and TypeScript frontend, Go microservice called calculator-api for the backend that does all the maths, basic and advanced. Main goes in cmd/server and the rest under internal. Standard library only for now, no frameworks. The spec is the source of truth, when I change it the code gets updated to match. Don't write any code yet, we'll start on that once the spec is done.

Iteration 2

Reviewed the spec. Drop the history feature, we don't need it. The error handling needs more thought. Every error should give the user a clear message and a code, and nothing internal should leak out. Unexpected stuff just gets logged and returns a 500. A panic shouldn't take the server down either. Put the exact messages in the requirements.

Iteration 3

The design section is too generic. Interfaces go on the consumer side, not next to the implementation, and don't add them everywhere. Check with me before you introduce one. For a single function a func type is enough. Keep the Go side structured the Go way, one package per concern, so all the calculator stuff lives in calculator. Not the Django style models, services, controllers split. Use SOLID where it makes sense, not just for show. Put all that in the rules for the code.

Iteration 4

How would the spec handle something like (986x(4+8))+2 with a root in it? Honestly having two modes is overkill. Make it work like a phone calculator, you type the whole expression and it gets evaluated on equals. So one endpoint that takes the expression as JSON and the backend parses it. Normal precedence rules, brackets first, and put the grammar in the design. Cover the edge cases too, divide by zero, negative roots, out of range should be an error and not infinity, and syntax errors should say where the problem is. Bad input and maths that has no answer should be different status codes.

Iteration 5

Add the security and ops side to the spec. A sensible limit on input length, cap the body size and be strict about what you accept, no unknown fields. The usual security headers. Keep it same origin, and if CORS is ever needed it should be an allowlist from env, not a wildcard. Sensible timeouts and a graceful shutdown. Separate backend, frontend and deploy folders, each with its own README, gitignore and env example. Never commit real env values. Small non-root Docker images and a compose file with Caddy in front.

Iteration 6

Spec looks good. Build it in the order in tasks.md, one step at a time, test first. Set up gofmt, golangci-lint and govulncheck behind a make check. I'd like the calculator package fully covered. Show me the tests passing after each step, and if the spec is unclear ask me instead of guessing.

Iteration 7

Reviewed it. Remove anything redundant and make sure parse errors are handled everywhere. Also why is the healthcheck in main? If the code does something the spec doesn't mention, either it goes in the spec or it goes.

Iteration 8

Spec change, I added OpenTelemetry. Only export when an endpoint is configured, and use the trace id as the request id so I can match a response to its logs. No global otel state. Make sure it actually works against a collector before you call it done.

Iteration 9

Spec change, the healthcheck subcommand and the docker healthchecks aren't worth it. Remove them from both. Keep the healthz endpoint though.

Iteration 10

Spec change, added sin, cos, log and exp. Degrees like a phone. sin(180) should come out as 0, not some tiny number. Adding another function later should be easy.

Iteration 11

Run govulncheck again and upgrade anything it flags. Put it in the spec that it has to stay clean.

Iteration 12

The READMEs read too much like AI. Rewrite them from the spec like a product doc, what numbers we support, the operations, edge cases with the exact messages, how to set it up and run it, a few curl examples, design decisions and assumptions. Keep it short and don't repeat things. Keep the spec, code and READMEs in sync from now on.



// UI prompts to claude

Hook it up to the API. Post the expression to /api/v1/calculations and show the result, or the error message the backend sends. If the backend is down say so. Base URL from VITE_API_URL, and proxy /api to the backend in dev. Nothing secret in VITE vars, it all ends up in the browser.


Clean up the structure a bit. Keypad logic in plain functions so it's testable without React, all the fetch stuff in one place. Use enums instead of string constants, and no any.


Make it a bit more robust. After a result, typing a number starts over and an operator carries on from the result. Don't allow double submits. Keyboard should work too.


Add sin, cos, log and exp keys. Backspace should delete the whole sin( in one go.


Add tests for the logic, the api client and the components. Everything should pass before we call it done.
