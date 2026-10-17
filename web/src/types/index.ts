export interface User {
  id: string;
  email: string;
  created_at: string;
  name: string;
  avatar_url: string;
  user_type: "guest" | "free" | "paid";
}

export interface AuthResponse {
  token: string;
  user: User;
}

export interface Property {
  id: string;
  name: string;
  address: string;
  created_at: string;
  updated_at: string;
  permission?: "admin" | "read";
}

export interface PropertyShare {
  user_id: string;
  email: string;
  name: string;
  permission: string;
}

export interface Booking {
  id: string;
  property_id: string;
  platform: "booking" | "airbnb";
  platform_reference: string;
  guest_name: string;
  check_in: string;
  check_out: string;
  nights: number;
  gross_amount: string;
  price: string;
  platform_fee: string;
  net_amount: string;
  currency: string;
  status: "confirmed" | "cancelled" | "completed";
  created_at: string;
}

export interface Expense {
  id: string;
  property_id: string;
  category: ExpenseCategory;
  description: string;
  amount: string;
  currency: string;
  date: string;
  end_date: string | null;
  recurring: boolean;
  is_percentage: boolean;
  percentage: string;
  created_at: string;
}

export type ExpenseCategory =
  | "electricity"
  | "heating"
  | "water"
  | "internet"
  | "maintenance"
  | "cleaning"
  | "other"
  | "partner_payout";

export interface ExpenseBreakdown {
  electricity: number;
  heating: number;
  water: number;
  internet: number;
  maintenance: number;
  cleaning: number;
  other: number;
  partner_payout: number;
  total: number;
}

export interface TaxRule {
  id: string;
  property_id: string;
  name: string;
  type: "percentage" | "fixed";
  percentage: number;
  fixed_amount: number;
  currency: string;
  created_at: string;
}

export interface TaxEstimate {
  name: string;
  amount: number;
}

export interface DashboardSummary {
  month: string;
  currency: string;
  total_income: number;
  platform_fees: number;
  net_income: number;
  expenses: ExpenseBreakdown;
  profit: number;
  taxes: TaxEstimate[];
  total_tax: number;
  profit_after_tax: number;
  occupancy_rate: number;
  estimated_occupancy_rate?: number;
  bookings_count: number;
  adr: number;
  revpar: number;
  avg_stay_length: number;
  avg_lead_time_days: number;
  cancellation_rate: number;
}

export interface Integration {
  id: string;
  user_id: string;
  platform: "booking" | "airbnb";
  credentials: Record<string, string>;
  enabled: boolean;
  last_synced_at: string | null;
  last_sync_error: string | null;
  created_at: string;
  updated_at: string;
}

export interface SyncResult {
  imported: number;
  updated: number;
  errors: number;
  message: string;
}

export interface PropertyMapping {
  id: string;
  integration_id: string;
  property_id: string;
  external_id: string;
  platform: string;
  created_at: string;
}

export interface ImportResult {
  imported: number;
  skipped: number;
  errors: number;
}

export interface ExtractedExpense {
  category: ExpenseCategory;
  description: string;
  amount: number;
  currency: string;
  date: string;
}

export interface SellerProfile {
  id: string;
  user_id: string;
  legal_name: string;
  street: string;
  city: string;
  county: string;
  postal_code: string;
  country_code: string;
  cui: string;
  cnp: string;
  trade_register_nr: string;
  vat_registered: boolean;
  created_at: string;
  updated_at: string;
}

export interface ANAFCompany {
  name: string;
  cui: string;
  street: string;
  city: string;
  county: string;
  postal_code: string;
  trade_register_nr: string;
  vat_registered: boolean;
}

export type InventoryCategory = "linens" | "kitchenware" | "furniture" | "electronics" | "toiletries" | "appliances" | "other";

export interface InventoryItem {
  id: string;
  property_id: string;
  name: string;
  category: InventoryCategory;
  quantity: number;
  acquired_date: string;
  lifetime_months: number;
  notes: string;
  created_at: string;
}

export interface Invoice {
  id: string;
  user_id: string;
  booking_id: string;
  invoice_number: number;
  invoice_series: string;
  issue_date: string;
  due_date: string;
  seller_legal_name: string;
  seller_cui: string;
  seller_cnp: string;
  seller_trade_register_nr: string;
  seller_street: string;
  seller_city: string;
  seller_county: string;
  seller_postal_code: string;
  seller_country_code: string;
  seller_vat_registered: boolean;
  buyer_legal_name: string;
  buyer_cui: string;
  buyer_trade_register_nr: string;
  buyer_street: string;
  buyer_city: string;
  buyer_county: string;
  buyer_postal_code: string;
  buyer_country_code: string;
  buyer_vat_registered: boolean;
  currency: string;
  net_amount: string;
  tax_amount: string;
  gross_amount: string;
  tax_percent: string;
  description: string;
  nights: number;
  xml_content: string;
  created_at: string;
}
