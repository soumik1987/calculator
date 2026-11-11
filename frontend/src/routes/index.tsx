import { createFileRoute } from "@tanstack/react-router";

import { Calculator } from "@/components/calculator/Calculator";

export const Route = createFileRoute("/")({
  head: () => ({
    meta: [
      { title: "Sami Calculator" },
      { name: "description", content: "A simple calculator backed by a Go microservice." },
    ],
  }),
  component: Calculator,
});
