import { describe, expect, it } from "vitest";

import {
  backspace,
  formatNumber,
  initialState,
  MAX_EXPRESSION_LENGTH,
  press,
  showResult,
  type CalculatorState,
} from "@/lib/calculator";

function typeKeys(keys: string[], state: CalculatorState = initialState) {
  return keys.reduce(press, state);
}

describe("press", () => {
  it("appends keys to the expression", () => {
    expect(typeKeys(["(", "9", "8", "6", "×", "1", "2", ")", "+", "√", "4"]).expression).toBe(
      "(986×12)+√4",
    );
  });

  it("starts a new expression when a number follows a result", () => {
    const state = showResult({ expression: "2+2" }, 4);

    expect(press(state, "7")).toEqual({ expression: "7" });
  });

  it("continues from the result when an operator follows it", () => {
    const state = showResult({ expression: "2+2" }, 4);

    expect(press(state, "×")).toEqual({ expression: "4×" });
  });

  it("continues from the rounded result", () => {
    const state = showResult({ expression: "0.1+0.2" }, 0.30000000000000004);

    expect(press(state, "+")).toEqual({ expression: "0.3+" });
  });

  it("stops at the maximum length", () => {
    const full = { expression: "1".repeat(MAX_EXPRESSION_LENGTH) };

    expect(press(full, "2")).toEqual({
      ...full,
      error: "The expression can be at most 500 characters.",
    });
  });
});

describe("backspace", () => {
  it("removes the last character", () => {
    expect(backspace({ expression: "12+" })).toEqual({ expression: "12" });
  });

  it("removes a whole function key", () => {
    expect(backspace({ expression: "2+sin(" })).toEqual({ expression: "2+" });
  });

  it("does nothing on an empty expression", () => {
    expect(backspace(initialState)).toEqual(initialState);
  });

  it("returns to editing after a result", () => {
    expect(backspace(showResult({ expression: "2+2" }, 4))).toEqual({ expression: "2+2" });
  });
});

describe("formatNumber", () => {
  it("limits significant digits and groups thousands", () => {
    expect(formatNumber(1 / 3)).toBe("0.333333333333");
    expect(formatNumber(11834)).toBe("11,834");
  });
});
