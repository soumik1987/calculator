import { useCallback, useEffect, useState } from "react";

import { useCalculate } from "@/hooks/use-calculator";
import { backspace, ControlKey, initialState, press, showResult } from "@/lib/calculator";

import { Display } from "./Display";
import { Keypad } from "./Keypad";

const keyboardKeys: Record<string, string> = {
  "*": "×",
  x: "×",
  "/": "÷",
  "-": "−",
  Enter: ControlKey.Equals,
  "=": ControlKey.Equals,
  Escape: ControlKey.Clear,
  Delete: ControlKey.Clear,
  Backspace: ControlKey.Backspace,
};

const typedKeys = new Set([
  "0",
  "1",
  "2",
  "3",
  "4",
  "5",
  "6",
  "7",
  "8",
  "9",
  ".",
  "+",
  "^",
  "%",
  "(",
  ")",
]);

export function Calculator() {
  const [state, setState] = useState(initialState);
  const calculate = useCalculate();

  const handleKey = useCallback(
    (key: string) => {
      if (key === ControlKey.Equals) {
        if (!state.expression || state.result !== undefined) return;
        calculate.mutate(
          { expression: state.expression },
          { onSuccess: (data) => setState((s) => showResult(s, data.result)) },
        );
        return;
      }
      calculate.reset();
      if (key === ControlKey.Clear) setState(initialState);
      else if (key === ControlKey.Backspace) setState(backspace);
      else setState((s) => press(s, key));
    },
    [calculate, state],
  );

  useEffect(() => {
    function onKeyDown(event: KeyboardEvent) {
      if (event.metaKey || event.ctrlKey || event.altKey) return;
      const key = keyboardKeys[event.key] ?? (typedKeys.has(event.key) ? event.key : undefined);
      if (!key) return;
      event.preventDefault();
      handleKey(key);
    }
    window.addEventListener("keydown", onKeyDown);
    return () => window.removeEventListener("keydown", onKeyDown);
  }, [handleKey]);

  return (
    <main className="flex min-h-svh items-center justify-center bg-background px-5 py-8 text-foreground">
      <section aria-labelledby="calculator-title" className="w-full max-w-sm">
        <h1 id="calculator-title" className="mb-5 text-center text-xl font-semibold">
          Sami Calculator
        </h1>
        <Display
          expression={state.expression}
          result={state.result}
          error={state.error ?? calculate.error?.message ?? ""}
        />
        <Keypad onKey={handleKey} disabled={calculate.isPending} />
      </section>
    </main>
  );
}
