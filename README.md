# Sami Calculator

A calculator that works like the one on a phone. The web page has a display and a keypad.
When you press `=`, the expression goes to a Go service, which works it out and sends back
the result or an error.

## Features

- Keypad: digits, `.`, `+ − × ÷`, `^` (power), `√` (square root), `%`, brackets, `C` (clear), `⌫` (delete) and `=`.
- Scientific keys: `sin`, `cos`, `log` (base 10) and `exp` (eˣ). Each types the name and `(`.
- The keyboard works too. `*` and `/` type `×` and `÷`, Enter is `=`, and Esc is `C`.
- After a result, a number starts a new calculation, and an operator continues from the result.
- Errors appear in the display. Typing again clears them.
- Nothing is saved, and there are no accounts.

## Calculation rules

- Precedence, highest first: brackets, functions (`√ sin cos log exp`) and `%`, `^`, a
  leading minus, `× ÷`, `+ −`. So `2 + 3 × 4` = 14 and `-2 ^ 2` = -4.
- `^` groups from the right (`2 ^ 3 ^ 2` = 512). Everything else groups from the left.
- A function applies to the number or bracket right after it: `√16 ^ 2` = 16. Use `√(-4)`,
  not `√-4`.
- Angles are in degrees: `sin(30)` = 0.5. Multiples of 90° are exact: `sin(180)` = 0.
- `%` divides by 100: `200 × 10%` = 20.
- Multiplication must be written out: `2(3)` is an error.
- The limit is 500 characters.

## Numbers

Numbers are 64-bit floating point (IEEE 754): `float64` in Go, `number` in TypeScript.

- The largest value is about 1.8 × 10³⁰⁸. Whole numbers are exact up to 2⁵³.
- Some decimals aren't exact: `0.1 + 0.2` = 0.30000000000000004.
- The display rounds to 12 significant digits. The API returns the full value.

## Errors and edge cases

| Input                        | Display shows                                       |
| ---------------------------- | --------------------------------------------------- |
| `5 ÷ 0`, `0 ÷ 0`, `0 ^ -1`   | Cannot divide by zero.                              |
| `√(-4)`                      | Square root needs a number that is zero or greater. |
| `log(0)`, `log(-1)`          | Logarithm needs a number greater than zero.         |
| `(-8) ^ 0.5`                 | The result is not a real number.                    |
| `10 ^ 1000`, `exp(1000)`     | The result is outside the supported number range.   |
| More than 500 characters     | The expression can be at most 500 characters.       |
| `2 +`                        | The expression is incomplete.                       |
| `2 + 3)`                     | Unexpected ")" at position 4.                       |
| `(2 + 3`                     | Missing ")" for the "(" at position 1.              |
| `10 ^ -200 × 10 ^ -200`      | 0 (too small to store)                              |
| `0.1 + 0.2`                  | 0.3                                                 |
| `-1 × 0`                     | 0, never -0                                         |
| Backend not running          | Cannot reach the calculator service.                |

## Structure

```
browser ──► Caddy ──┬── /api/* ──► backend   Go service, does the maths
                    └── rest   ──► frontend  React app, display and keypad
```

- [backend/](backend/): the Go service and its API
- [frontend/](frontend/): the React app
- [deploy/](deploy/): Docker Compose and Caddy

## Design decisions

- **The backend does all the maths, when `=` is pressed.** One place owns the rules and
  errors. Cost: no live preview, and every result needs the backend.
- **The request is the typed text: `{"expression": "..."}`.** The frontend stays a keypad.
  Sending a parsed tree instead would move parsing into the browser.
- **`float64` throughout.** Native to Go, JSON and JavaScript. Cost: binary rounding.
- **Stateless service.** Scale by adding copies. Cost: no history.
- **Same origin** (Vite proxy in development, Caddy in production), so normal use needs no
  CORS setup.
- **Distroless backend image, no Docker health checks.** No shell in the image, and nothing
  polling. Cost: Caddy may return 502 for a moment at startup.

## Future improvements

- **Container health checks**, if startup order or health in `docker ps` matters. Distroless
  has no `curl`, so the binary would need a `server healthcheck` subcommand, and Compose
  would use `depends_on: service_healthy`. Kubernetes can probe `GET /healthz` without it.
- **More functions** (`tan`, `ln`, a radians mode): one entry in the lexer's function table
  and a key each.
- **Exact decimals** with `math/big`, if `0.1 + 0.2` must equal 0.3.

## Run locally

You need Go 1.26+ and Node 20.19+. In two terminals:

```sh
cd backend && cp .env.example .env && make run
cd frontend && cp .env.example .env && npx bun install && npm run dev
```

Open http://localhost:8082.

