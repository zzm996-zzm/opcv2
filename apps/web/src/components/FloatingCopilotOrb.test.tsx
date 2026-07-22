import { fireEvent, render, screen } from "@testing-library/react";
import { MemoryRouter, Route, Routes, useLocation } from "react-router-dom";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import FloatingCopilotOrb from "./FloatingCopilotOrb";

function LocationProbe() {
  return <output aria-label="当前位置">{useLocation().pathname}</output>;
}

function renderOrb() {
  render(
    <MemoryRouter initialEntries={["/tasks"]}>
      <FloatingCopilotOrb />
      <Routes>
        <Route element={<LocationProbe />} path="*" />
      </Routes>
    </MemoryRouter>
  );
  return screen.getByRole("button", { name: "打开智活 Copilot" });
}

describe("FloatingCopilotOrb", () => {
  beforeEach(() => {
    vi.stubGlobal("PointerEvent", MouseEvent);
  });

  afterEach(() => {
    window.localStorage.clear();
    vi.unstubAllGlobals();
    vi.restoreAllMocks();
  });

  it("opens Copilot when clicked", () => {
    const orb = renderOrb();

    fireEvent.click(orb);

    expect(screen.getByLabelText("当前位置")).toHaveTextContent("/copilot");
  });

  it("drags inside the viewport and remembers its position", () => {
    const orb = renderOrb();
    vi.spyOn(orb, "getBoundingClientRect").mockReturnValue({
      bottom: 556,
      height: 76,
      left: 680,
      right: 756,
      top: 480,
      width: 76,
      x: 680,
      y: 480,
      toJSON: () => ({})
    });

    fireEvent.pointerDown(orb, { button: 0, clientX: 700, clientY: 500, pointerId: 1 });
    fireEvent.pointerMove(orb, { clientX: 400, clientY: 300, pointerId: 1 });
    fireEvent.pointerUp(orb, { clientX: 400, clientY: 300, pointerId: 1 });

    expect(orb).toHaveStyle({ left: "380px", top: "280px" });
    expect(JSON.parse(window.localStorage.getItem("opcv2:copilot-orb-position") || "null")).toEqual({
      left: 380,
      top: 280
    });
  });

  it("supports keyboard positioning", () => {
    const orb = renderOrb();
    vi.spyOn(orb, "getBoundingClientRect").mockReturnValue({
      bottom: 556,
      height: 76,
      left: 680,
      right: 756,
      top: 480,
      width: 76,
      x: 680,
      y: 480,
      toJSON: () => ({})
    });

    fireEvent.keyDown(orb, { key: "ArrowLeft", shiftKey: true });

    expect(orb).toHaveStyle({ left: "656px", top: "480px" });
  });
});
