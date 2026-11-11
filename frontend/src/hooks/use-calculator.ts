import { useMutation } from "@tanstack/react-query";

import { calculatorApi, type CalculationRequest } from "@/lib/api";

export function useCalculate() {
  return useMutation({
    mutationFn: ({ expression }: CalculationRequest) => calculatorApi.calculate(expression),
  });
}
