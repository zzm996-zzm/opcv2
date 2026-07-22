import { useEffect, useRef, useState, type KeyboardEvent, type PointerEvent } from "react";
import { createPortal } from "react-dom";
import { useNavigate } from "react-router-dom";

const STORAGE_KEY = "opcv2:copilot-orb-position";
const ORB_SIZE = 76;
const VIEWPORT_MARGIN = 8;

type OrbPosition = {
  left: number;
  top: number;
};

type DragState = OrbPosition & {
  moved: boolean;
  pointerId: number;
  startX: number;
  startY: number;
};

type FloatingCopilotOrbProps = {
  className?: string;
  onActivate?: () => void;
};

function FloatingCopilotOrb({ className = "", onActivate }: FloatingCopilotOrbProps) {
  const navigate = useNavigate();
  const [position, setPosition] = useState<OrbPosition | null>(() => readStoredPosition());
  const [dragging, setDragging] = useState(false);
  const dragRef = useRef<DragState | null>(null);
  const blockClickRef = useRef(false);

  useEffect(() => {
    function keepInsideViewport() {
      setPosition((current) => current ? constrainPosition(current) : current);
    }

    window.addEventListener("resize", keepInsideViewport);
    return () => window.removeEventListener("resize", keepInsideViewport);
  }, []);

  function activate() {
    if (blockClickRef.current) {
      blockClickRef.current = false;
      return;
    }
    if (onActivate) {
      onActivate();
      return;
    }
    navigate("/copilot");
  }

  function beginDrag(event: PointerEvent<HTMLButtonElement>) {
    if (event.button !== 0) return;
    const rect = event.currentTarget.getBoundingClientRect();
    event.currentTarget.setPointerCapture?.(event.pointerId);
    dragRef.current = {
      left: rect.left,
      top: rect.top,
      moved: false,
      pointerId: event.pointerId,
      startX: event.clientX,
      startY: event.clientY
    };
    blockClickRef.current = false;
    setDragging(true);
  }

  function moveOrb(event: PointerEvent<HTMLButtonElement>) {
    const drag = dragRef.current;
    if (!drag || drag.pointerId !== event.pointerId) return;
    const deltaX = event.clientX - drag.startX;
    const deltaY = event.clientY - drag.startY;
    if (Math.abs(deltaX) + Math.abs(deltaY) > 4) drag.moved = true;
    const next = constrainPosition({ left: drag.left + deltaX, top: drag.top + deltaY });
    setPosition(next);
    if (drag.moved) storePosition(next);
  }

  function endDrag(event: PointerEvent<HTMLButtonElement>) {
    const drag = dragRef.current;
    if (!drag || drag.pointerId !== event.pointerId) return;
    event.currentTarget.releasePointerCapture?.(event.pointerId);
    dragRef.current = null;
    blockClickRef.current = drag.moved;
    window.setTimeout(() => {
      blockClickRef.current = false;
    }, 0);
    setDragging(false);
    setPosition((current) => {
      if (drag.moved && current) storePosition(current);
      return current;
    });
  }

  function moveWithKeyboard(event: KeyboardEvent<HTMLButtonElement>) {
    const directions: Record<string, [number, number]> = {
      ArrowDown: [0, 1],
      ArrowLeft: [-1, 0],
      ArrowRight: [1, 0],
      ArrowUp: [0, -1]
    };
    const direction = directions[event.key];
    if (!direction) return;
    event.preventDefault();
    const rect = event.currentTarget.getBoundingClientRect();
    const current = position ?? { left: rect.left, top: rect.top };
    const step = event.shiftKey ? 24 : 8;
    const next = constrainPosition({
      left: current.left + direction[0] * step,
      top: current.top + direction[1] * step
    });
    setPosition(next);
    storePosition(next);
  }

  return createPortal(
    <button
      aria-label="打开智活 Copilot"
      className={`global-copilot-orb ${dragging ? "dragging" : ""} ${className}`.trim()}
      onClick={activate}
      onKeyDown={moveWithKeyboard}
      onPointerCancel={endDrag}
      onPointerDown={beginDrag}
      onPointerMove={moveOrb}
      onPointerUp={endDrag}
      style={position ? { bottom: "auto", left: position.left, right: "auto", top: position.top } : undefined}
      title="拖动调整位置，点击打开智活 Copilot"
      type="button"
    >
      <span className="v4-logo" aria-hidden="true" />
    </button>,
    document.body
  );
}

function constrainPosition(position: OrbPosition): OrbPosition {
  const maxLeft = Math.max(VIEWPORT_MARGIN, window.innerWidth - ORB_SIZE - VIEWPORT_MARGIN);
  const maxTop = Math.max(VIEWPORT_MARGIN, window.innerHeight - ORB_SIZE - VIEWPORT_MARGIN);
  return {
    left: Math.min(Math.max(position.left, VIEWPORT_MARGIN), maxLeft),
    top: Math.min(Math.max(position.top, VIEWPORT_MARGIN), maxTop)
  };
}

function readStoredPosition(): OrbPosition | null {
  if (typeof window === "undefined") return null;
  try {
    const stored = window.localStorage.getItem(STORAGE_KEY);
    if (!stored) return null;
    const value = JSON.parse(stored) as Partial<OrbPosition>;
    if (typeof value.left !== "number" || typeof value.top !== "number") return null;
    if (!Number.isFinite(value.left) || !Number.isFinite(value.top)) return null;
    return constrainPosition({ left: value.left, top: value.top });
  } catch {
    return null;
  }
}

function storePosition(position: OrbPosition) {
  try {
    window.localStorage.setItem(STORAGE_KEY, JSON.stringify(position));
  } catch {
    // Dragging remains available when storage is blocked by the browser.
  }
}

export default FloatingCopilotOrb;
