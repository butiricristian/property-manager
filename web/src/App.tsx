import { BrowserRouter, Routes, Route } from "react-router-dom";
import { Layout } from "@/components/Layout";
import { CurrencyProvider } from "@/context/CurrencyContext";
import { PropertyProvider } from "@/context/PropertyContext";
import { AuthProvider } from "@/context/AuthContext";
import { ProtectedRoute } from "@/components/ProtectedRoute";
import { DashboardPage } from "@/pages/DashboardPage";
import { BookingsPage } from "@/pages/BookingsPage";
import { ExpensesPage } from "@/pages/ExpensesPage";
import { TaxRulesPage } from "@/pages/TaxRulesPage";
import { PropertiesPage } from "@/pages/PropertiesPage";
// TODO: re-enable once Booking.com machine account credentials are set up
// import { IntegrationsPage } from "@/pages/IntegrationsPage";
import { InventoryPage } from "@/pages/InventoryPage";
import { SellerProfilePage } from "@/pages/SellerProfilePage";
import { LoginPage } from "@/pages/LoginPage";
import { RegisterPage } from "@/pages/RegisterPage";
import { OAuthCallbackPage } from "@/pages/OAuthCallbackPage";
import { PrivacyPolicyPage } from "@/pages/PrivacyPolicyPage";
import { TermsPage } from "@/pages/TermsPage";

function App() {
  return (
    <BrowserRouter>
      <AuthProvider>
        <CurrencyProvider>
          <Routes>
            <Route path="/login" element={<LoginPage />} />
            <Route path="/register" element={<RegisterPage />} />
            <Route path="/privacy" element={<PrivacyPolicyPage />} />
            <Route path="/terms" element={<TermsPage />} />
            <Route path="/auth/callback" element={<OAuthCallbackPage />} />
            <Route
              element={
                <ProtectedRoute>
                  <PropertyProvider>
                    <Layout />
                  </PropertyProvider>
                </ProtectedRoute>
              }
            >
              <Route path="/" element={<DashboardPage />} />
              <Route path="/bookings" element={<BookingsPage />} />
              <Route path="/expenses" element={<ExpensesPage />} />
              <Route path="/tax-rules" element={<TaxRulesPage />} />
              <Route path="/inventory" element={<InventoryPage />} />
              <Route path="/properties" element={<PropertiesPage />} />
              <Route path="/billing-profile" element={<SellerProfilePage />} />
              {/* TODO: re-enable once Booking.com machine account credentials are set up */}
              {/* <Route path="/integrations" element={<IntegrationsPage />} */}
            </Route>
          </Routes>
        </CurrencyProvider>
      </AuthProvider>
    </BrowserRouter>
  );
}

export default App;
