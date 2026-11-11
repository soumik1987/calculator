import { Delete } from "lucide-react";

import { Button } from "@/components/ui/button";
import { ControlKey, FunctionKey } from "@/lib/calculator";

enum KeyStyle {
  Digit = "digit",
  Operator = "operator",
  Control = "control",
  Equals = "equals",
}

type KeyDef = {
  key: string;
  style: KeyStyle;
  label?: string;
  wide?: boolean;
};

const layout: KeyDef[] = [
  { key: ControlKey.Clear, style: KeyStyle.Control },
  { key: "(", style: KeyStyle.Control },
  { key: ")", style: KeyStyle.Control },
  { key: ControlKey.Backspace, style: KeyStyle.Control },
  { key: FunctionKey.Sin, label: "sin", style: KeyStyle.Operator },
  { key: FunctionKey.Cos, label: "cos", style: KeyStyle.Operator },
  { key: FunctionKey.Log, label: "log", style: KeyStyle.Operator },
  { key: FunctionKey.Exp, label: "exp", style: KeyStyle.Operator },
  { key: "√", style: KeyStyle.Operator },
  { key: "^", style: KeyStyle.Operator },
  { key: "%", style: KeyStyle.Operator },
  { key: "÷", style: KeyStyle.Operator },
  { key: "7", style: KeyStyle.Digit },
  { key: "8", style: KeyStyle.Digit },
  { key: "9", style: KeyStyle.Digit },
  { key: "×", style: KeyStyle.Operator },
  { key: "4", style: KeyStyle.Digit },
  { key: "5", style: KeyStyle.Digit },
  { key: "6", style: KeyStyle.Digit },
  { key: "−", style: KeyStyle.Operator },
  { key: "1", style: KeyStyle.Digit },
  { key: "2", style: KeyStyle.Digit },
  { key: "3", style: KeyStyle.Digit },
  { key: "+", style: KeyStyle.Operator },
  { key: "0", style: KeyStyle.Digit, wide: true },
  { key: ".", style: KeyStyle.Digit },
  { key: ControlKey.Equals, style: KeyStyle.Equals },
];

const styles: Record<KeyStyle, string> = {
  [KeyStyle.Digit]: "bg-card text-foreground hover:bg-accent border border-border",
  [KeyStyle.Operator]: "bg-secondary text-secondary-foreground hover:bg-secondary/80",
  [KeyStyle.Control]: "bg-muted text-foreground hover:bg-muted/70",
  [KeyStyle.Equals]: "bg-primary text-primary-foreground hover:bg-primary/90",
};

type Props = {
  onKey: (key: string) => void;
  disabled: boolean;
};

export function Keypad({ onKey, disabled }: Props) {
  return (
    <div className="mt-5 grid grid-cols-4 gap-3">
      {layout.map(({ key, style, label = key, wide }) => (
        <Button
          key={key}
          type="button"
          aria-label={label}
          disabled={disabled && key === ControlKey.Equals}
          onClick={() => onKey(key)}
          className={`h-16 rounded-2xl font-mono text-2xl shadow-none ${styles[style]} ${wide ? "col-span-2" : ""}`}
        >
          {key === ControlKey.Clear ? (
            "C"
          ) : key === ControlKey.Backspace ? (
            <Delete aria-hidden="true" />
          ) : (
            label
          )}
        </Button>
      ))}
    </div>
  );
}
