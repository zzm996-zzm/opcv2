import { createContext, useCallback, useContext, useLayoutEffect, useMemo, useState, type ReactNode } from "react";

type CopilotPanelVisibilityValue = {
  closePanel: () => void;
  hasPanel: boolean;
  isPanelOpen: boolean;
  openPanel: () => void;
  registerPanel: () => () => void;
};

const CopilotPanelVisibilityContext = createContext<CopilotPanelVisibilityValue | null>(null);

export function CopilotPanelVisibilityProvider({ children }: { children: ReactNode }) {
  const [isPanelOpen, setIsPanelOpen] = useState(true);
  const [panelCount, setPanelCount] = useState(0);
  const closePanel = useCallback(() => setIsPanelOpen(false), []);
  const openPanel = useCallback(() => setIsPanelOpen(true), []);
  const registerPanel = useCallback(() => {
    setPanelCount((count) => count + 1);
    return () => setPanelCount((count) => Math.max(0, count - 1));
  }, []);
  const value = useMemo(() => ({
    closePanel,
    hasPanel: panelCount > 0,
    isPanelOpen,
    openPanel,
    registerPanel
  }), [closePanel, isPanelOpen, openPanel, panelCount, registerPanel]);

  return (
    <CopilotPanelVisibilityContext.Provider value={value}>
      {children}
    </CopilotPanelVisibilityContext.Provider>
  );
}

// eslint-disable-next-line react-refresh/only-export-components
export function useCopilotPanelVisibility() {
  return useContext(CopilotPanelVisibilityContext);
}

// eslint-disable-next-line react-refresh/only-export-components
export function useRegisteredCopilotPanel(initiallyOpen = true) {
  const visibility = useCopilotPanelVisibility();
  const [locallyOpen, setLocallyOpen] = useState(initiallyOpen);
  const registerPanel = visibility?.registerPanel;

  useLayoutEffect(() => registerPanel?.(), [registerPanel]);

  return {
    closePanel: visibility?.closePanel ?? (() => setLocallyOpen(false)),
    hasSharedController: Boolean(visibility),
    isPanelOpen: visibility?.isPanelOpen ?? locallyOpen,
    openPanel: visibility?.openPanel ?? (() => setLocallyOpen(true))
  };
}
