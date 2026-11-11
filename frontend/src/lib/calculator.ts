export const MAX_EXPRESSION_LENGTH = 500;

export enum ControlKey {
  Clear = "Clear",
  Backspace = "Backspace",
  Equals = "=",
}

export enum FunctionKey {
  Sin = "sin(",
  Cos = "cos(",
  Log = "log(",
  Exp = "exp(",
}

const functionKeys: string[] = Object.values(FunctionKey);

export type CalculatorState = {
  expression: string;
  result?: number;
  error?: string;
};

enum Message {
  TooLong = "The expression can be at most 500 characters.",
}

export const initialState: CalculatorState = { expression: "" };

const operators = new Set(["+", "−", "×", "÷", "^", "%"]);

export function press(state: CalculatorState, key: string): CalculatorState {
  if (state.result !== undefined) {
    const start = operators.has(key) ? toExpressionNumber(state.result) : "";
    return { expression: start + key };
  }
  if (state.expression.length + key.length > MAX_EXPRESSION_LENGTH) {
    return { expression: state.expression, error: Message.TooLong };
  }
  return { expression: state.expression + key };
}

export function backspace(state: CalculatorState): CalculatorState {
  if (state.result !== undefined) return { expression: state.expression };
  const last = functionKeys.find((key) => state.expression.endsWith(key));
  return { expression: state.expression.slice(0, -(last?.length ?? 1)) };
}

export function showResult(state: CalculatorState, result: number): CalculatorState {
  return { expression: state.expression, result };
}

function toExpressionNumber(value: number): string {
  return String(Number(value.toPrecision(12)));
}

const numberFormat = new Intl.NumberFormat("en-US", { maximumSignificantDigits: 12 });

export function formatNumber(value: number): string {
  return numberFormat.format(value);
}
