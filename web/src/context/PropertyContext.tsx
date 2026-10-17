import { createContext, useContext, useEffect, useState, type ReactNode } from "react";
import { listProperties } from "@/api/client";
import { clampEndToToday } from "@/lib/dashboard";
import type { Property } from "@/types";

export interface DateRange {
  startDate: string; // YYYY-MM-DD
  endDate: string;   // YYYY-MM-DD
  preset: "month" | "ytd" | "last-year" | "custom";
}

function getCurrentMonthRange(): DateRange {
  const now = new Date();
  const y = now.getFullYear();
  const m = now.getMonth();
  const startDate = `${y}-${String(m + 1).padStart(2, "0")}-01`;
  const next = new Date(y, m + 1, 1);
  const endDate = `${next.getFullYear()}-${String(next.getMonth() + 1).padStart(2, "0")}-01`;
  // Current month defaults to a month-to-date range so stats aren't diluted.
  return { startDate, endDate: clampEndToToday(startDate, endDate), preset: "month" };
}

interface PropertyContextValue {
  properties: Property[];
  propertyId: string;
  setPropertyId: (id: string) => void;
  dateRange: DateRange;
  setDateRange: (range: DateRange) => void;
}

const PropertyContext = createContext<PropertyContextValue | undefined>(undefined);

export function PropertyProvider({ children }: { children: ReactNode }) {
  const [properties, setProperties] = useState<Property[]>([]);
  const [propertyId, setPropertyIdState] = useState(
    () => localStorage.getItem("selected-property") || ""
  );
  const [dateRange, setDateRange] = useState<DateRange>(getCurrentMonthRange);

  useEffect(() => {
    listProperties()
      .then((props) => {
        setProperties(props);
        if (props.length > 0) {
          const stored = localStorage.getItem("selected-property");
          const valid = stored && props.some((p) => p.id === stored);
          if (!valid) {
            setPropertyIdState(props[0].id);
            localStorage.setItem("selected-property", props[0].id);
          }
        }
      })
      .catch(console.error);
  }, []);

  const setPropertyId = (id: string) => {
    setPropertyIdState(id);
    localStorage.setItem("selected-property", id);
  };

  return (
    <PropertyContext value={{ properties, propertyId, setPropertyId, dateRange, setDateRange }}>
      {children}
    </PropertyContext>
  );
}

export function useProperty() {
  const ctx = useContext(PropertyContext);
  if (!ctx) throw new Error("useProperty must be used within PropertyProvider");
  return ctx;
}
