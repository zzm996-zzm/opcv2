import "@testing-library/jest-dom/vitest";
import { cleanup } from "@testing-library/react";
import { afterEach } from "vitest";

import { authSession } from "../lib/authSession";

afterEach(() => {
  cleanup();
  authSession.clear();
});
