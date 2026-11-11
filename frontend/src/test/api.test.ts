import { afterEach, describe, expect, it, vi } from "vitest";

import { ApiError, ApiErrorCode, calculatorApi } from "@/lib/api";

function mockFetch(response: Response | Error) {
  const fetchMock =
    response instanceof Error
      ? vi.fn().mockRejectedValue(response)
      : vi.fn().mockResolvedValue(response);
  vi.stubGlobal("fetch", fetchMock);
  return fetchMock;
}

function json(body: unknown, status = 200) {
  return new Response(JSON.stringify(body), {
    status,
    headers: { "Content-Type": "application/json" },
  });
}

afterEach(() => {
  vi.unstubAllGlobals();
});

describe("calculatorApi", () => {
  it("posts an expression and returns the result", async () => {
    const calculation = { expression: "(986×(4+8))+2", result: 11834 };
    const fetchMock = mockFetch(json(calculation));

    await expect(calculatorApi.calculate("(986×(4+8))+2")).resolves.toEqual(calculation);

    const [url, init] = fetchMock.mock.calls[0] as [string, RequestInit];
    expect(url).toMatch(/\/api\/v1\/calculations$/);
    expect(init.method).toBe("POST");
    expect(new Headers(init.headers).get("Content-Type")).toBe("application/json");
    expect(JSON.parse(init.body as string)).toEqual({ expression: "(986×(4+8))+2" });
  });

  it("surfaces the backend error message", async () => {
    mockFetch(
      json(
        { error: { code: ApiErrorCode.DivisionByZero, message: "Cannot divide by zero." } },
        422,
      ),
    );

    const error = await calculatorApi.calculate("1/0").catch((e: unknown) => e);

    expect(error).toBeInstanceOf(ApiError);
    expect(error).toMatchObject({
      status: 422,
      code: ApiErrorCode.DivisionByZero,
      message: "Cannot divide by zero.",
    });
  });

  it("keeps the trace ID", async () => {
    mockFetch(
      new Response(JSON.stringify({ error: { code: ApiErrorCode.Internal, message: "x" } }), {
        status: 500,
        headers: { "X-Trace-ID": "4bf92f3577b34da6a3ce929d0e0e4736" },
      }),
    );

    await expect(calculatorApi.calculate("1+2")).rejects.toMatchObject({
      traceId: "4bf92f3577b34da6a3ce929d0e0e4736",
    });
  });

  it("uses a generic message when the error body is not JSON", async () => {
    mockFetch(new Response("bad gateway", { status: 502 }));

    await expect(calculatorApi.calculate("1+1")).rejects.toMatchObject({
      status: 502,
      code: ApiErrorCode.Unknown,
    });
  });

  it("rejects a successful response without a JSON body", async () => {
    mockFetch(new Response("<html></html>", { status: 200 }));

    await expect(calculatorApi.calculate("1+1")).rejects.toMatchObject({
      status: 200,
      code: ApiErrorCode.Unknown,
    });
  });

  it.each([{ expression: "1+1" }, { result: "2" }, { result: null }, [2], 2])(
    "rejects a successful response without a numeric result (%j)",
    async (body) => {
      mockFetch(json(body));

      await expect(calculatorApi.calculate("1+1")).rejects.toMatchObject({
        status: 200,
        code: ApiErrorCode.Unknown,
        message: "Something went wrong. Please try again.",
      });
    },
  );

  it("reports network failures", async () => {
    mockFetch(new TypeError("Failed to fetch"));

    await expect(calculatorApi.calculate("1+1")).rejects.toMatchObject({
      status: 0,
      code: ApiErrorCode.Network,
    });
  });
});
