import React from "react";
import ReactDOM from "react-dom/client";
import { MantineProvider } from "@mantine/core";
import { RouterProvider } from "react-router-dom";
import { router } from "./router";
import { AuthProvider } from "./AuthContext";
import { zvednuTheme } from "./theme";
import { useHotjar } from "./useHotjar";

import "@mantine/core/styles.css";

function App() {
  useHotjar();

  return (
    <MantineProvider theme={zvednuTheme} defaultColorScheme="light">
      <AuthProvider>
        <RouterProvider router={router} />
      </AuthProvider>
    </MantineProvider>
  );
}

ReactDOM.createRoot(document.getElementById("root")!).render(
  <React.StrictMode>
    <App />
  </React.StrictMode>
);
