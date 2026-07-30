import React from "react";
import ReactDOM from "react-dom/client";
import { BrowserRouter } from "react-router-dom";

import App from "./App";
import "./styles.css";
import "./reference-ui.css";
import "./typography.css";
import "./brand.css";
import "./project-market-enhancements.css";
import "./sandbox-reference.css";
import "./public-components-parity.css";
import "./landing-reference-parity.css";
import "./home-reference-parity.css";
import "./floating-copilot-orb.css";
import "./crm-reference-parity.css";
import "./v4-shell-consistency.css";
import "./unified-copilot-panel.css";
import "./membership-reference.css";
import "./ticket-center.css";

ReactDOM.createRoot(document.getElementById("root")!).render(
  <React.StrictMode>
    <BrowserRouter>
      <App />
    </BrowserRouter>
  </React.StrictMode>
);
