export type CalculationRequest = {
  expression: string;
};

export type Calculation = CalculationRequest & {
  result: number;
};

export enum ApiErrorCode {
  InvalidRequest = "invalid_request",
  InvalidExpression = "invalid_expression",
  RequestTooLarge = "request_too_large",
  UnsupportedMediaType = "unsupported_media_type",
  DivisionByZero = "division_by_zero",
  NegativeSquareRoot = "negative_square_root",
  NonPositiveLogarithm = "non_positive_logarithm",
  OutOfRange = "out_of_range",
  UndefinedResult = "undefined_result",
  Internal = "internal_error",
  Network = "network_error",
  Unknown = "unknown_error",
}

enum Endpoint {
  Calculations = "/api/v1/calculations",
}

enum HttpMethod {
  Post = "POST",
}

enum Header {
  ContentType = "Content-Type",
  TraceId = "X-Trace-ID",
}

enum ErrorMessage {
  Network = "Cannot reach the calculator service.",
  Unknown = "Something went wrong. Please try again.",
}

const jsonHeaders = { [Header.ContentType]: "application/json" };

export class ApiError extends Error {
  constructor(
    message: string,
    readonly status: number,
    readonly code: ApiErrorCode | string,
    readonly traceId?: string,
  ) {
    super(message);
    this.name = ApiError.name;
  }
}

const baseUrl = (import.meta.env.VITE_API_URL ?? "").replace(/\/$/, "");

type ErrorBody = { error?: { code?: unknown; message?: unknown } };

function fail(response: Response, body: unknown): never {
  const error = (body as ErrorBody | null)?.error;
  throw new ApiError(
    typeof error?.message === "string" ? error.message : ErrorMessage.Unknown,
    response.status,
    typeof error?.code === "string" ? error.code : ApiErrorCode.Unknown,
    response.headers.get(Header.TraceId) ?? undefined,
  );
}

function isCalculation(body: unknown): body is Calculation {
  return typeof (body as Calculation | null)?.result === "number";
}

export const calculatorApi = {
  async calculate(expression: string): Promise<Calculation> {
    let response: Response;
    try {
      response = await fetch(`${baseUrl}${Endpoint.Calculations}`, {
        method: HttpMethod.Post,
        headers: jsonHeaders,
        body: JSON.stringify({ expression } satisfies CalculationRequest),
      });
    } catch {
      throw new ApiError(ErrorMessage.Network, 0, ApiErrorCode.Network);
    }

    const body: unknown = await response.json().catch(() => null);
    if (!response.ok || !isCalculation(body)) fail(response, body);
    return body;
  },
};
