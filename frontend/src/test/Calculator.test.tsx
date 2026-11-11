import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { fireEvent, render, screen } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { Calculator } from "@/components/calculator/Calculator";
import { ApiError, ApiErrorCode, calculatorApi } from "@/lib/api";

vi.mock("@/lib/api", async (importOriginal) => {
  const actual = await importOriginal<typeof import("@/lib/api")>();
  return { ...actual, calculatorApi: { calculate: vi.fn() } };
});

const api = vi.mocked(calculatorApi);

function renderCalculator() {
  const client = new QueryClient();
  return render(
    <QueryClientProvider client={client}>
      <Calculator />
    </QueryClientProvider>,
  );
}

function tap(...keys: string[]) {
  for (const key of keys) fireEvent.click(screen.getByRole("button", { name: key }));
}

function expression() {
  return screen.getByTestId("expression").textContent;
}

beforeEach(() => {
  vi.resetAllMocks();
});

describe("Calculator", () => {
  it("builds the expression from the keypad", () => {
    renderCalculator();

    tap("(", "9", "8", "6", "×", "(", "4", "+", "8", ")", ")", "+", "2");

    expect(expression()).toBe("(986×(4+8))+2");
  });

  it("inserts functions with an opening bracket", () => {
    renderCalculator();

    tap("sin", "9", "0", ")", "+", "cos", "0", ")", "+", "log", "1", ")", "+", "exp", "0", ")");

    expect(expression()).toBe("sin(90)+cos(0)+log(1)+exp(0)");
  });

  it("sends one POST on equals and shows the result", async () => {
    api.calculate.mockResolvedValue({ expression: "(986×(4+8))+2", result: 11834 });
    renderCalculator();

    tap("(", "9", "8", "6", "×", "(", "4", "+", "8", ")", ")", "+", "2", "=");

    expect(await screen.findByTestId("result")).toHaveTextContent("11,834");
    expect(api.calculate).toHaveBeenCalledTimes(1);
    expect(api.calculate).toHaveBeenCalledWith("(986×(4+8))+2");
  });

  it("does nothing on equals with an empty display", () => {
    renderCalculator();

    tap("=");

    expect(api.calculate).not.toHaveBeenCalled();
  });

  it("shows errors from the backend", async () => {
    api.calculate.mockRejectedValue(
      new ApiError("Cannot divide by zero.", 422, ApiErrorCode.DivisionByZero),
    );
    renderCalculator();

    tap("1", "÷", "0", "=");

    expect(await screen.findByRole("alert")).toHaveTextContent("Cannot divide by zero.");
  });

  it("shows an error when the result is out of range", async () => {
    api.calculate.mockRejectedValue(
      new ApiError(
        "The result is outside the supported number range.",
        422,
        ApiErrorCode.OutOfRange,
      ),
    );
    renderCalculator();

    tap("1", "0", "^", "1", "0", "0", "0", "=");

    expect(await screen.findByRole("alert")).toHaveTextContent(
      "The result is outside the supported number range.",
    );
  });

  it("shows an error when the expression is too long", () => {
    renderCalculator();

    for (let i = 0; i < 501; i++) fireEvent.keyDown(window, { key: "1" });

    expect(expression()).toHaveLength(500);
    expect(screen.getByRole("alert")).toHaveTextContent(
      "The expression can be at most 500 characters.",
    );
  });

  it("clears the error when the user keeps typing", async () => {
    api.calculate.mockRejectedValue(
      new ApiError("The expression is incomplete.", 400, ApiErrorCode.InvalidExpression),
    );
    renderCalculator();

    tap("2", "+", "=");
    await screen.findByRole("alert");
    tap("3");

    expect(screen.queryByRole("alert")).not.toBeInTheDocument();
    expect(expression()).toBe("2+3");
  });

  it("continues from the result with an operator", async () => {
    api.calculate.mockResolvedValue({ expression: "2+2", result: 4 });
    renderCalculator();

    tap("2", "+", "2", "=");
    await screen.findByTestId("result");
    tap("×", "3");

    expect(expression()).toBe("4×3");
  });

  it("deletes with backspace and clears with C", () => {
    renderCalculator();

    tap("1", "2", "3", "Backspace");
    expect(expression()).toBe("12");

    tap("Clear");
    expect(expression()).toBe("");
  });

  it("accepts keyboard input", async () => {
    api.calculate.mockResolvedValue({ expression: "9÷4", result: 2.25 });
    renderCalculator();

    for (const key of ["9", "/", "4"]) fireEvent.keyDown(window, { key });
    expect(expression()).toBe("9÷4");

    fireEvent.keyDown(window, { key: "Enter" });

    expect(await screen.findByTestId("result")).toHaveTextContent("2.25");
    expect(api.calculate).toHaveBeenCalledWith("9÷4");
  });

  it("maps keyboard shortcuts for backspace and clear", () => {
    renderCalculator();

    for (const key of ["1", "*", "2", "-", "3"]) fireEvent.keyDown(window, { key });
    expect(expression()).toBe("1×2−3");

    fireEvent.keyDown(window, { key: "Backspace" });
    expect(expression()).toBe("1×2−");

    fireEvent.keyDown(window, { key: "Escape" });
    expect(expression()).toBe("");
  });
});
