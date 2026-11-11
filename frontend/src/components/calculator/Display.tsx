import { formatNumber } from "@/lib/calculator";

type Props = {
  expression: string;
  result: number | undefined;
  error: string;
};

export function Display({ expression, result, error }: Props) {
  return (
    <div className="flex min-h-36 flex-col justify-end rounded-2xl bg-muted px-5 py-4 text-right">
      <p
        data-testid="expression"
        className={`break-all font-mono tabular-nums ${result === undefined ? "text-3xl" : "text-lg text-muted-foreground"}`}
      >
        {expression}
      </p>
      <div aria-live="polite" className="min-h-10">
        {error ? (
          <p role="alert" className="mt-2 text-sm font-medium text-destructive">
            {error}
          </p>
        ) : result !== undefined ? (
          <p
            data-testid="result"
            className="mt-1 break-all font-mono text-4xl font-semibold tabular-nums"
          >
            {formatNumber(result)}
          </p>
        ) : null}
      </div>
    </div>
  );
}
