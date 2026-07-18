import { StrictMode } from "react";
import { createRoot } from "react-dom/client";
import "./styles.css";

function Dashboard() {
  return (
    <main>
      <h1>Finny</h1>
      <p>Dashboard placeholder</p>
      <a href="/edit">Edit dashboard</a>
    </main>
  );
}

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

createRoot(document.getElementById("root")!).render(
  <StrictMode>
    <App />
  </StrictMode>,
);
