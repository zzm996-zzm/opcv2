import react from "@vitejs/plugin-react";
import { defineConfig } from "vitest/config";

export default defineConfig({
  plugins: [react()],
  server: {
    allowedHosts: [".ngrok-free.app", ".ngrok.app", ".ngrok.io"],
    proxy: {
      "/api": process.env.VITE_API_PROXY_TARGET ?? "http://localhost:8080",
      "/health": process.env.VITE_API_PROXY_TARGET ?? "http://localhost:8080"
    }
  },
  test: {
    environment: "jsdom",
    setupFiles: "./src/test/setup.ts"
  }
});
