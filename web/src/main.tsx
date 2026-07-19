import { StrictMode } from "react";
import { createRoot } from "react-dom/client";
import { DashboardPage } from "./features/dashboard/DashboardPage";
import "./styles.css";

function Edit() {
  return (
    <main>
      <h1>Edit dashboard</h1>
      <p>Edit placeholder</p>
      <a href="/">Back to dashboard</a>
    </main>
  );
}

function App() {
  return window.location.pathname === "/edit" ? <Edit /> : <Dashboard />;
}

function Dashboard() {
  return <DashboardPage />;
}

createRoot(document.getElementById("root")!).render(
  <StrictMode>
    <App />
  </StrictMode>,
);
