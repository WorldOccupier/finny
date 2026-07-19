import { DashboardPage } from "./features/dashboard/DashboardPage";
import { EditDashboardPage } from "./features/dashboard/EditDashboardPage";

export function App() {
  return window.location.pathname === "/edit" ? <EditDashboardPage /> : <DashboardPage />;
}

