import { DashboardPage } from "./features/dashboard/DashboardPage";
import { EditDashboardPage } from "./features/dashboard/EditDashboardPage";
import { SpendingPage } from "./features/spending/SpendingPage";

export function App() {
  if (window.location.pathname === "/edit") return <EditDashboardPage />;
  if (window.location.pathname === "/spending") return <SpendingPage />;
  return <DashboardPage />;
}
