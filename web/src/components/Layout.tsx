import { NavLink, Outlet, Link } from "react-router-dom";
import {
  LayoutDashboard,
  CalendarDays,
  Receipt,
  Building2,
  Landmark,
  FileText,
  Package,
  LogOut,
} from "lucide-react";
import { Button } from "@/components/ui/button";
import { useAuth } from "@/context/AuthContext";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { useCurrency } from "@/context/CurrencyContext";
import { GuestBanner } from "@/components/GuestBanner";

const navItems = [
  { to: "/", icon: LayoutDashboard, label: "Dashboard" },
  { to: "/bookings", icon: CalendarDays, label: "Bookings" },
  { to: "/expenses", icon: Receipt, label: "Expenses" },
  { to: "/tax-rules", icon: Landmark, label: "Tax Rules" },
  { to: "/inventory", icon: Package, label: "Inventory" },
  { to: "/properties", icon: Building2, label: "Properties" },
  { to: "/billing-profile", icon: FileText, label: "Billing" },
  // TODO: re-enable once Booking.com machine account credentials are set up
  // { to: "/integrations", icon: Link2, label: "Integrations" },
];

const currencies = ["EUR", "USD", "GBP", "RON"];

export function Layout() {
  const { currency, setCurrency } = useCurrency();
  const { user, logout, isGuest } = useAuth();

  return (
    <div className="flex h-screen flex-col bg-background">
      {isGuest && <GuestBanner />}
      <div className="flex flex-1 overflow-hidden">
      <aside className="w-64 border-r bg-card p-4 flex flex-col">
        <h1 className="text-xl font-bold px-3 py-4">Property Manager</h1>
        <nav className="flex flex-col gap-1 flex-1">
          {navItems.map(({ to, icon: Icon, label }) => (
            <NavLink
              key={to}
              to={to}
              className={({ isActive }) =>
                `flex items-center gap-3 rounded-md px-3 py-2 text-sm font-medium transition-colors ${
                  isActive
                    ? "bg-primary text-primary-foreground"
                    : "text-muted-foreground hover:bg-accent hover:text-accent-foreground"
                }`
              }
              end={to === "/"}
            >
              <Icon className="h-4 w-4" />
              {label}
            </NavLink>
          ))}
        </nav>
        <div className="px-3 py-2">
          <label className="text-xs font-medium text-muted-foreground mb-1 block">
            Currency
          </label>
          <Select value={currency} onValueChange={(v) => setCurrency(v ?? "EUR")}>
            <SelectTrigger className="w-full">
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              {currencies.map((c) => (
                <SelectItem key={c} value={c}>
                  {c}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        </div>
        {user && (
          <div className="border-t px-3 py-3 flex items-center gap-2">
            {user.avatar_url && (
              <img
                src={user.avatar_url}
                alt=""
                className="h-6 w-6 rounded-full shrink-0"
                referrerPolicy="no-referrer"
              />
            )}
            <span className="text-xs text-muted-foreground truncate flex-1">
              {user.name || user.email}
            </span>
            <Button variant="ghost" size="icon" onClick={logout} title="Logout">
              <LogOut className="h-4 w-4" />
            </Button>
          </div>
        )}
        <div className="px-3 py-2 text-xs text-muted-foreground">
          <Link to="/privacy" className="hover:underline">
            Privacy
          </Link>
          {" | "}
          <Link to="/terms" className="hover:underline">
            Terms
          </Link>
        </div>
      </aside>
      <main className="flex-1 overflow-y-auto p-6">
        <Outlet />
      </main>
      </div>
    </div>
  );
}
