# Calculator web app

The calculator screen. Built with React 19, TypeScript, TanStack Start and Tailwind. It
checks what the user types and shows results, but it never does the maths itself. Every
result comes from the calculator service.

## Run

Start the backend first, then:

```sh
cp .env.example .env
npx bun install
npm run dev
```

Open http://localhost:8080.

In development, requests to `/api` are forwarded to the backend on port 8081 (set in
`vite.config.ts`).

`VITE_API_URL` in `.env` is the backend's address. Leave it empty when the page and the API
share one address, as they do in development and in Docker. Anything in a `VITE_` variable
is visible to everyone who loads the page, so never put a secret there.

## Code

```
src/
  lib/calculator.ts        keypad logic (press, backspace, results) and number formatting
  lib/api.ts               the call to the backend
  hooks/use-calculator.ts  sends an expression when = is pressed
  components/calculator/   Calculator (screen and keyboard), Display, Keypad
  routes/index.tsx         the home page
```

## Design decisions

- **No maths in the browser.** The page builds the expression text and sends it on `=`.
  Results and errors come from the backend.
- **Keypad logic is pure functions** (`press`, `backspace`, `showResult`), so it is tested
  without React.
- **React Query `useMutation` for the call.** It gives pending and error state without
  hand-written effects.
- **The display rounds to 12 significant digits.** Continuing from a result uses the same
  rounded value.
- **The 500-character limit is also checked while typing.** The backend enforces it too,
  but checking early keeps long input from reaching the 4 KB body limit, whose message is
  technical.
- **`ApiError` keeps the status, error code and trace ID** for debugging. The UI shows only
  the message.

## Tests

```sh
npm test
npm run typecheck
npm run lint
```
