import type {
  Property,
  PropertyShare,
  Booking,
  Expense,
  DashboardSummary,
  ImportResult,
  ExtractedExpense,
  AuthResponse,
  User,
  TaxRule,
  Integration,
  SyncResult,
  PropertyMapping,
  SellerProfile,
  ANAFCompany,
  Invoice,
  InventoryItem,
} from "@/types";

const BASE = "/api";

const TOKEN_KEY = "auth_token";

export function getAuthToken(): string | null {
  return localStorage.getItem(TOKEN_KEY);
}

export function setAuthToken(token: string | null): void {
  if (token) {
    localStorage.setItem(TOKEN_KEY, token);
  } else {
    localStorage.removeItem(TOKEN_KEY);
  }
}

function authHeaders(): Record<string, string> {
  const token = getAuthToken();
  const headers: Record<string, string> = { "Content-Type": "application/json" };
  if (token) {
    headers["Authorization"] = `Bearer ${token}`;
  }
  return headers;
}

async function request<T>(path: string, options?: RequestInit): Promise<T> {
  const res = await fetch(`${BASE}${path}`, {
    headers: authHeaders(),
    ...options,
  });
  if (res.status === 401) {
    setAuthToken(null);
    window.location.href = "/login";
    throw new Error("Unauthorized");
  }
  if (!res.ok) {
    const body = await res.json().catch(() => ({}));
    throw new Error(body.error || `Request failed: ${res.status}`);
  }
  if (res.status === 204) return undefined as T;
  return res.json();
}

// Auth
export const register = (email: string, password: string) =>
  request<AuthResponse>("/auth/register", {
    method: "POST",
    body: JSON.stringify({ email, password }),
  });

export const login = (email: string, password: string) =>
  request<AuthResponse>("/auth/login", {
    method: "POST",
    body: JSON.stringify({ email, password }),
  });

export const getCurrentUser = () => request<User>("/auth/me");

export const guestLogin = () =>
  request<AuthResponse>("/auth/guest", { method: "POST" });

export const upgradeGuest = (email: string, password: string) =>
  request<AuthResponse>("/auth/upgrade", {
    method: "POST",
    body: JSON.stringify({ email, password }),
  });

// Properties
export const listProperties = () => request<Property[]>("/properties");

export const getProperty = (id: string) =>
  request<Property>(`/properties/${id}`);

export const createProperty = (data: { name: string; address: string }) =>
  request<Property>("/properties", {
    method: "POST",
    body: JSON.stringify(data),
  });

export const updateProperty = (
  id: string,
  data: { name: string; address: string }
) =>
  request<Property>(`/properties/${id}`, {
    method: "PUT",
    body: JSON.stringify(data),
  });

export const deleteProperty = (id: string) =>
  request<void>(`/properties/${id}`, { method: "DELETE" });

// Property Shares
export const listPropertyShares = (propertyId: string) =>
  request<PropertyShare[]>(`/properties/${propertyId}/shares`);

export const addPropertyShare = (propertyId: string, email: string) =>
  request<PropertyShare>(`/properties/${propertyId}/shares`, {
    method: "POST",
    body: JSON.stringify({ email }),
  });

export const removePropertyShare = (propertyId: string, userId: string) =>
  request<void>(`/properties/${propertyId}/shares/${userId}`, {
    method: "DELETE",
  });

// Bookings
export const listBookings = (
  propertyId: string,
  startDate: string,
  endDate: string,
  currency?: string
) =>
  request<Booking[]>(
    `/properties/${propertyId}/bookings?start=${startDate}&end=${endDate}${currency ? `&currency=${currency}` : ""}`
  );

export const createBooking = (
  propertyId: string,
  data: Omit<Booking, "id" | "property_id" | "created_at">
) =>
  request<Booking>(`/properties/${propertyId}/bookings`, {
    method: "POST",
    body: JSON.stringify(data),
  });

export const deleteBooking = (id: string) =>
  request<void>(`/bookings/${id}`, { method: "DELETE" });

// Expenses
export const listExpenses = (propertyId: string, startDate: string, endDate: string) =>
  request<Expense[]>(`/properties/${propertyId}/expenses?start=${startDate}&end=${endDate}`);

export const createExpense = (
  propertyId: string,
  data: Omit<Expense, "id" | "property_id" | "created_at">
) =>
  request<Expense>(`/properties/${propertyId}/expenses`, {
    method: "POST",
    body: JSON.stringify(data),
  });

export const updateExpense = (
  id: string,
  data: Omit<Expense, "id" | "property_id" | "created_at">
) =>
  request<Expense>(`/expenses/${id}`, {
    method: "PUT",
    body: JSON.stringify(data),
  });

export const deleteExpense = (id: string) =>
  request<void>(`/expenses/${id}`, { method: "DELETE" });

// Import
export const importCSV = (
  propertyId: string,
  platform: string,
  file: File,
  payoutsFile?: File
) => {
  const formData = new FormData();
  formData.append("file", file);
  formData.append("platform", platform);
  if (payoutsFile) {
    formData.append("payouts_file", payoutsFile);
  }

  const token = getAuthToken();
  const headers: Record<string, string> = {};
  if (token) {
    headers["Authorization"] = `Bearer ${token}`;
  }

  return fetch(`${BASE}/properties/${propertyId}/import/csv`, {
    method: "POST",
    headers,
    body: formData,
  }).then(async (res) => {
    if (res.status === 401) {
      setAuthToken(null);
      window.location.href = "/login";
      throw new Error("Unauthorized");
    }
    if (!res.ok) {
      const body = await res.json().catch(() => ({}));
      throw new Error(body.error || "Import failed");
    }
    return res.json() as Promise<ImportResult>;
  });
};

// Bill extraction
export const extractBillExpense = (propertyId: string, file: File) => {
  const formData = new FormData();
  formData.append("file", file);

  const token = getAuthToken();
  const headers: Record<string, string> = {};
  if (token) {
    headers["Authorization"] = `Bearer ${token}`;
  }

  return fetch(`${BASE}/properties/${propertyId}/expenses/extract`, {
    method: "POST",
    headers,
    body: formData,
  }).then(async (res) => {
    if (res.status === 401) {
      setAuthToken(null);
      window.location.href = "/login";
      throw new Error("Unauthorized");
    }
    if (!res.ok) {
      const body = await res.json().catch(() => ({}));
      throw new Error(body.error || "Extraction failed");
    }
    return res.json() as Promise<ExtractedExpense>;
  });
};

// Tax Rules
export const listTaxRules = (propertyId: string) =>
  request<TaxRule[]>(`/properties/${propertyId}/tax-rules`);

export const createTaxRule = (
  propertyId: string,
  data: Omit<TaxRule, "id" | "property_id" | "created_at">
) =>
  request<TaxRule>(`/properties/${propertyId}/tax-rules`, {
    method: "POST",
    body: JSON.stringify(data),
  });

export const updateTaxRule = (
  id: string,
  data: Omit<TaxRule, "id" | "property_id" | "created_at">
) =>
  request<TaxRule>(`/tax-rules/${id}`, {
    method: "PUT",
    body: JSON.stringify(data),
  });

export const deleteTaxRule = (id: string) =>
  request<void>(`/tax-rules/${id}`, { method: "DELETE" });

// Integrations
export const listIntegrations = () => request<Integration[]>("/integrations");

export const createIntegration = (data: {
  platform: string;
  credentials: Record<string, string>;
}) =>
  request<Integration>("/integrations", {
    method: "POST",
    body: JSON.stringify(data),
  });

export const updateIntegration = (
  id: string,
  data: { credentials?: Record<string, string>; enabled?: boolean }
) =>
  request<Integration>(`/integrations/${id}`, {
    method: "PUT",
    body: JSON.stringify(data),
  });

export const deleteIntegration = (id: string) =>
  request<void>(`/integrations/${id}`, { method: "DELETE" });

export const syncIntegration = (id: string) =>
  request<SyncResult>(`/integrations/${id}/sync`, { method: "POST" });

// Property Mappings
export const listPropertyMappings = (propertyId: string) =>
  request<PropertyMapping[]>(`/properties/${propertyId}/mappings`);

export const createPropertyMapping = (
  propertyId: string,
  data: { integration_id: string; external_id: string }
) =>
  request<PropertyMapping>(`/properties/${propertyId}/mappings`, {
    method: "POST",
    body: JSON.stringify(data),
  });

export const deletePropertyMapping = (propertyId: string, mappingId: string) =>
  request<void>(`/properties/${propertyId}/mappings/${mappingId}`, {
    method: "DELETE",
  });

// Seller Profile
export const getSellerProfile = () =>
  request<SellerProfile | null>("/seller-profile");

export const upsertSellerProfile = (
  data: Omit<SellerProfile, "id" | "user_id" | "created_at" | "updated_at">
) =>
  request<SellerProfile>("/seller-profile", {
    method: "PUT",
    body: JSON.stringify(data),
  });

// ANAF Lookup
export const lookupANAF = (cui: string) =>
  request<ANAFCompany>("/anaf/lookup", {
    method: "POST",
    body: JSON.stringify({ cui }),
  });

// Invoices
export const generateInvoice = (
  bookingId: string,
  data: {
    buyer_legal_name: string;
    buyer_cui: string;
    buyer_trade_register_nr: string;
    buyer_street: string;
    buyer_city: string;
    buyer_county: string;
    buyer_postal_code: string;
    buyer_country_code: string;
    buyer_vat_registered: boolean;
    net_amount: number;
    currency: string;
  }
) =>
  request<Invoice>(`/bookings/${bookingId}/invoice`, {
    method: "POST",
    body: JSON.stringify(data),
  });

export const getBookingInvoice = (bookingId: string) =>
  request<Invoice>(`/bookings/${bookingId}/invoice`);

export const listInvoices = () => request<Invoice[]>("/invoices");

export const downloadInvoiceXML = async (invoiceId: string) => {
  const res = await fetch(`${BASE}/invoices/${invoiceId}/xml`, {
    headers: authHeaders(),
  });
  if (!res.ok) throw new Error("Failed to download invoice XML");
  const blob = await res.blob();
  const url = URL.createObjectURL(blob);
  const a = document.createElement("a");
  a.href = url;
  a.download =
    res.headers.get("Content-Disposition")?.match(/filename="?(.+?)"?$/)?.[1] ??
    "invoice.xml";
  a.click();
  URL.revokeObjectURL(url);
};

export const deleteInvoice = (id: string) =>
  request<void>(`/invoices/${id}`, { method: "DELETE" });

// Inventory
export const listInventory = (propertyId: string) =>
  request<InventoryItem[]>(`/properties/${propertyId}/inventory`);

export const createInventoryItem = (
  propertyId: string,
  data: Omit<InventoryItem, "id" | "property_id" | "created_at">
) =>
  request<InventoryItem>(`/properties/${propertyId}/inventory`, {
    method: "POST",
    body: JSON.stringify(data),
  });

export const updateInventoryItem = (
  id: string,
  data: Omit<InventoryItem, "id" | "property_id" | "created_at">
) =>
  request<InventoryItem>(`/inventory/${id}`, {
    method: "PUT",
    body: JSON.stringify(data),
  });

export const deleteInventoryItem = (id: string) =>
  request<void>(`/inventory/${id}`, { method: "DELETE" });

// Dashboard
export const getDashboard = (startDate: string, endDate: string, currency?: string) =>
  request<DashboardSummary>(
    `/dashboard?start=${startDate}&end=${endDate}${currency ? `&currency=${currency}` : ""}`
  );

export const getPropertyDashboard = (
  propertyId: string,
  startDate: string,
  endDate: string,
  currency?: string
) =>
  request<DashboardSummary>(
    `/properties/${propertyId}/dashboard?start=${startDate}&end=${endDate}${currency ? `&currency=${currency}` : ""}`
  );
