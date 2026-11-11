# Requirements

What the calculator should do. How we build it is in design.md.

## Expressions

It works like a phone calculator. You type the whole thing, something like "(986*(4+8))+sqrt(16)", and it gets worked out when you hit equals. One expression, one result.

Numbers can be whole, decimal or have an exponent. 12, 0.5, .5, 1e5 and 2.5E-3 should all work. The exponent matters because the UI sends big results back in that form when you carry on from them, like 1e+21.

Operators are plus, minus, times, divide, power with ^, percent, and brackets. Percent just divides by 100, so 200*10% is 20. The keypad shows and sends proper math signs for minus, times and divide (the unicode ones, not the keyboard ones). The API takes those, and also plain -, *, x and / since that's what people type on a keyboard.

Functions are square root (the root sign or "sqrt"), sin, cos, log base 10 and exp (e to the x). Case doesn't matter, SIN(30) is fine. A function applies to whatever number or bracket comes right after it, so "sqrt16^2" is (sqrt 16)^2 which is 16. "sqrt-4" is a syntax error, for a negative you write "sqrt(-4)".

Normal precedence. Brackets first, then functions and percent, then power, then a leading plus or minus, then times and divide, then plus and minus. So 2+3*4 is 14, and -2^2 is -4 because the power happens before the minus.

Power goes right to left, so 2^3^2 is 2^9 which is 512. Everything else left to right, 8/4/2 is 1.

sin and cos take degrees, sin(30) is 0.5. Multiples of 90 have to come out exact, sin(180) is 0 and cos(90) is 0, not 1.2e-16 or whatever floating point gives.

No implied multiplication. 2(3) and 2sin(1) are errors, you have to write the times sign.

Spaces don't matter, anywhere.

Max 500 characters.

## Results

We use normal 64 bit floats on both ends, float64 in Go and number in TypeScript. Largest is around 1.8e308 and whole numbers are exact up to 2^53. Some decimals aren't exact, 0.1+0.2 comes back from the API as 0.30000000000000004 and we're ok with that. The display rounds it to 0.3 anyway.

Never send back infinity or NaN. Infinity is out_of_range and NaN is undefined_result, see the errors below.

-0 should just be 0, so -1*0 is 0. Something too small to store is also 0, e.g. 10^-200 * 10^-200.

## Errors

Each error has a code that stays the same and a message we can put straight on screen. The UI doesn't rewrite messages, so they need to read well as they are.

Maths that has no answer is a 422:

5/0, 0/0 or 0^-1 is division_by_zero, "Cannot divide by zero."
sqrt(-4) is negative_square_root, "Square root needs a number that is zero or greater."
log(0), log(-1) or log(-0) is non_positive_logarithm, "Logarithm needs a number greater than zero."
(-8)^0.5 is undefined_result, "The result is not a real number."
10^1000 or exp(1000) is out_of_range, "The result is outside the supported number range."

If we can't read the expression it's a 400 invalid_expression, and the message should say where the problem is. Positions are characters, not bytes, starting at 1.

Empty or only spaces gives "Enter an expression."
Over 500 characters gives "The expression is longer than 500 characters."
2+ or sin with nothing after it gives "The expression is incomplete."
2+3) gives "Unexpected ")" at position 4."
(2+3 gives "Missing ")" for the "(" at position 1."
1.2.3 gives "Invalid number "1.2.3" at position 1."
tan(1) gives "Unknown word "tan" at position 1."
2#3 gives "Unexpected character "#" at position 2."

Read the whole expression before doing any maths. So "1/0+" is incomplete (400), not divide by zero (422).

Anything we didn't expect is a 500 internal_error with "Something went wrong. Please try again." The real error only goes in the log, with the trace id.

## API

One endpoint for calculating. POST /api/v1/calculations with {"expression": "(986*(4+8))+sqrt(16)"} gives back 200 {"expression": "(986*(4+8))+sqrt(16)", "result": 11836}. The expression is echoed back as sent.

Errors always look the same, {"error": {"code": "division_by_zero", "message": "Cannot divide by zero."}}, with the status from the tables above.

Be strict with the body. Content-Type has to be application/json, otherwise 415 unsupported_media_type with "Content-Type must be application/json." Everything below is 400 invalid_request, with a message saying what's wrong:

empty body, "Request body is empty."
not JSON, "Request body is not valid JSON."
JSON but not an object, "Request body must be a JSON object."
more than one value, "Request body must contain a single JSON object."
an extra field, "Unknown field "foo"."
expression not a string, "Field "expression" must be a string."

Body max 4 KB, over that is 413 request_too_large with "Request body must be at most 4096 bytes."

GET /healthz returns 200 {"status":"ok"}. Nothing else, it doesn't check anything.

No state, nothing is kept between requests. No history either.

## UI

Single screen, display on top and keypad below. Should feel right on a phone first, and still look fine on a desktop.

Keypad has digits, dot, plus, minus, times, divide, power, square root, percent, brackets, sin cos log exp, C to clear, backspace and equals. sin types "sin(" and the others the same, so you only have to close the bracket.

Keyboard should work. Digits, dot, + ^ % ( ) type themselves. "*" and "x" give times, "/" gives divide, "-" gives minus. Enter or = evaluates, Esc or Delete clears, Backspace deletes. Don't grab anything with Ctrl, Cmd or Alt, people still need copy and refresh.

Backspace right after sin( removes the whole sin( in one go, not just the bracket. Same for cos( log( and exp(.

Show the expression, and after equals the result under it rounded to 12 significant digits. The API value isn't rounded, only the display.

After a result, an operator (plus, minus, times, divide, power or percent) carries on from the shown value, so the rounded number goes into the new expression. Anything else starts fresh.

Errors show in the display with whatever message the backend sent. They go away as soon as you type. The trace id, status and code are kept on the error object for debugging but not shown.

No double submits. Equals does nothing while a request is running, on an empty expression, or when a result is already showing.

Stop at 500 characters and say "The expression can be at most 500 characters." The backend checks it too, this is just so nobody hits the 4 KB limit and its more technical message.

If the backend can't be reached show "Cannot reach the calculator service." If it sends back something we don't understand show "Something went wrong. Please try again."

No maths in the browser at all, not even a preview. Every result comes from the API.

## Security

Nothing internal leaks out in errors. No Go error text, no stack traces.

A panic in a handler shouldn't take the server down. Log it and return the normal 500.

Every API response gets X-Content-Type-Options nosniff, Content-Security-Policy "default-src 'none'; frame-ancestors 'none'", Cache-Control no-store and Referrer-Policy no-referrer. It only ever returns JSON so this can be locked right down.

Same origin by default, so normally no CORS at all. If the web app ever lives on another domain, ALLOWED_ORIGINS is a comma separated allowlist. Origins that aren't on it just don't get the header. Preflight returns 204 and allows GET, POST, OPTIONS with Content-Type, cached for 10 minutes. X-Trace-ID is exposed so the browser can read it.

Never commit real env values. .env stays in gitignore and only .env.example goes in. Nothing secret in VITE vars, they end up in the browser bundle.

No container runs as root. The backend image has no shell or package manager.

Server timeouts so slow clients can't hold connections forever, and a cap on header size.

govulncheck has to come back clean, and anything it flags gets upgraded.

## Observability

One JSON log line per request with method, path, status, duration and trace id. Log level from LOG_LEVEL.

Each request gets an OpenTelemetry span named after its route, plus the standard HTTP server metrics. If the caller sends a W3C traceparent, continue that trace. Don't trace /healthz, it's just noise.

Send the trace id back in X-Trace-ID so a response can be matched to its log line and its trace.

Failed requests get error.type on the span set to the error code, like division_by_zero.

Only export, over OTLP/HTTP, when OTEL_EXPORTER_OTLP_ENDPOINT is set. The other standard OTEL vars like OTEL_SERVICE_NAME and OTEL_EXPORTER_OTLP_HEADERS should just work. Service name defaults to calculator-api. No global otel state.

## Running it

Config comes from env. PORT defaults to 8081, ALLOWED_ORIGINS to http://localhost:8080, LOG_LEVEL to info (debug, info, warn or error), and the OTLP endpoint to empty. A wrong LOG_LEVEL should stop startup with a clear message rather than quietly fall back.

Shut down cleanly on SIGINT and SIGTERM. Finish the requests in flight, then flush telemetry.

Compose runs the api, the web app and Caddy, and only Caddy is exposed, on 80 and 443. /api goes to the api, everything else to the web app. SITE_ADDRESS defaults to :80. Put a domain in it and Caddy gets and renews the HTTPS certificate on its own.

No docker healthchecks. If a container dies it gets restarted. /healthz is there if a load balancer or Kubernetes needs to probe.

It should be fine to run several copies of the api, e.g. docker compose up --scale api=3.

Locally the backend runs on 8081 and the frontend on 8080, with /api proxied through to the backend, so it's same origin in dev too. You need Go 1.26+ and Node 20.19+.

If the web app is ever served from a different domain, build it with VITE_API_URL pointing at the api and add the web app's address to ALLOWED_ORIGINS.

## Not doing

History, logins, rate limiting, a radians mode, live preview while typing, exact decimals, other languages. It's assumed to be one user at a time on a trusted network, and a current browser.
