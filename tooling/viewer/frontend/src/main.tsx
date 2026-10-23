import { CSPProvider } from "@base-ui/react/csp-provider";
import { StrictMode } from "react";
import { createRoot } from "react-dom/client";
import { AccessClient, takeCode } from "./api";
import { ViewerApplication } from "./application";
import { ViewerController } from "./controller";
import "./styles.css";

const container = document.getElementById("root");
if (!container) throw new Error("Missing application root");
const root = createRoot(container);
const code = takeCode();
const controller = new ViewerController(new AccessClient());
document.documentElement.lang = navigator.language.startsWith("es")
  ? "es"
  : "en";
if (
  import.meta.env.DEV &&
  new URLSearchParams(location.search).has("preview")
) {
  void import("./dev/preview").then(({ DevelopmentPreview }) =>
    root.render(
      <StrictMode>
        <CSPProvider disableStyleElements>
          <DevelopmentPreview />
        </CSPProvider>
      </StrictMode>,
    ),
  );
} else {
  void controller.open(code);
  root.render(
    <StrictMode>
      <CSPProvider disableStyleElements>
        <ViewerApplication controller={controller} />
      </CSPProvider>
    </StrictMode>,
  );
}
