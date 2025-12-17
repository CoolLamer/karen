import React from "react";
import { createBrowserRouter } from "react-router-dom";
import { AppShellLayout } from "./ui/AppShellLayout";
import { CallInboxPage } from "./ui/CallInboxPage";
import { CallDetailPage } from "./ui/CallDetailPage";

export const router = createBrowserRouter([
  {
    path: "/",
    element: <AppShellLayout />,
    children: [
      { index: true, element: <CallInboxPage /> },
      { path: "calls/:providerCallId", element: <CallDetailPage /> },
    ],
  },
]);


