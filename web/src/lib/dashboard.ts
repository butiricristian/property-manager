import type { Booking } from "@/types";

/** Format a value as `RON 1,234.56` (currency code prefix, matching the design). */
export function money(value: number, currency = "EUR"): string {
  return `${currency} ${value.toLocaleString("en-US", {
    minimumFractionDigits: 2,
    maximumFractionDigits: 2,
  })}`;
}

/** Format a value as a rounded `RON 1,234` (no decimals). */
export function moneyShort(value: number, currency = "EUR"): string {
  return `${currency} ${Math.round(value).toLocaleString("en-US")}`;
}

export function greeting(date = new Date()): string {
  const h = date.getHours();
  if (h < 12) return "Good morning";
  if (h < 18) return "Good afternoon";
  return "Good evening";
}

/**
 * Clamp a range's exclusive end to the start of tomorrow when the range is
 * still in progress (today falls inside it). This makes an in-progress month —
 * or YTD — compute stats month-to-date rather than diluting rate metrics
 * (occupancy, RevPAR) across days that haven't happened yet. Past and fully
 * future ranges are returned unchanged.
 */
export function clampEndToToday(startDate: string, endDate: string): string {
  const now = new Date();
  const tomorrow = new Date(now.getFullYear(), now.getMonth(), now.getDate() + 1);
  const start = new Date(startDate + "T00:00:00");
  const end = new Date(endDate + "T00:00:00");
  if (tomorrow > start && tomorrow < end) {
    return `${tomorrow.getFullYear()}-${String(tomorrow.getMonth() + 1).padStart(
      2,
      "0"
    )}-${String(tomorrow.getDate()).padStart(2, "0")}`;
  }
  return endDate;
}

function shiftMonthISO(dateStr: string, delta: number): string {
  const [y, m, d] = dateStr.split("-").map(Number);
  const shifted = new Date(y, m - 1 + delta, d);
  return `${shifted.getFullYear()}-${String(shifted.getMonth() + 1).padStart(
    2,
    "0"
  )}-${String(shifted.getDate()).padStart(2, "0")}`;
}

/**
 * The equivalent window one month earlier, for period-over-period comparison.
 * For an in-progress month (month-to-date range) this yields the *same day span*
 * of the previous month rather than the whole month, so the "vs last month" chip
 * compares like-for-like. The window is never allowed to bleed into the current
 * period's start month.
 */
export function previousPeriodRange(
  startDate: string,
  endDate: string
): { startDate: string; endDate: string } {
  const shiftedEnd = shiftMonthISO(endDate, -1);
  return {
    startDate: shiftMonthISO(startDate, -1),
    endDate: shiftedEnd < startDate ? shiftedEnd : startDate,
  };
}

/** Number of whole days in a [start, end) exclusive range. */
export function daysInRange(startDate: string, endDate: string): number {
  const start = new Date(startDate + "T00:00:00");
  const end = new Date(endDate + "T00:00:00");
  return Math.max(0, Math.round((end.getTime() - start.getTime()) / 86400000));
}

/** The month to focus the calendar/trend on: the last full month inside the range. */
export function focusMonth(endDateExclusive: string): { year: number; month: number } {
  const end = new Date(endDateExclusive + "T00:00:00");
  end.setDate(end.getDate() - 1);
  return { year: end.getFullYear(), month: end.getMonth() };
}

export interface TrendMonth {
  year: number;
  month: number; // 0-based
  label: string; // e.g. "Jul"
  startDate: string;
  endDate: string; // exclusive
}

/** The `count` months ending at (and including) the focus month, oldest first. */
export function trendMonths(focus: { year: number; month: number }, count = 6): TrendMonth[] {
  const months: TrendMonth[] = [];
  for (let i = count - 1; i >= 0; i--) {
    const d = new Date(focus.year, focus.month - i, 1);
    const next = new Date(focus.year, focus.month - i + 1, 1);
    months.push({
      year: d.getFullYear(),
      month: d.getMonth(),
      label: d.toLocaleDateString("en-US", { month: "short" }),
      startDate: iso(d),
      // The current month's bar reflects take-home month-to-date, matching the hero.
      endDate: clampEndToToday(iso(d), iso(next)),
    });
  }
  return months;
}

function iso(d: Date): string {
  return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, "0")}-${String(
    d.getDate()
  ).padStart(2, "0")}`;
}

/** Days of the focus month that have at least one active (non-cancelled) booking night. */
export function bookedDays(bookings: Booking[], year: number, month: number): Set<number> {
  const booked = new Set<number>();
  for (const b of bookings) {
    if (b.status === "cancelled") continue;
    const checkIn = new Date(b.check_in + "T00:00:00");
    const checkOut = new Date(b.check_out + "T00:00:00");
    for (let d = new Date(checkIn); d < checkOut; d.setDate(d.getDate() + 1)) {
      if (d.getFullYear() === year && d.getMonth() === month) {
        booked.add(d.getDate());
      }
    }
  }
  return booked;
}

export type CalendarCell = { blank: true } | { blank: false; day: number; booked: boolean };

/** Monday-first grid cells for a month, with leading blanks to align the 1st. */
export function calendarCells(year: number, month: number, booked: Set<number>): CalendarCell[] {
  const first = new Date(year, month, 1);
  const daysInMonth = new Date(year, month + 1, 0).getDate();
  const leading = (first.getDay() + 6) % 7; // Monday = 0
  const cells: CalendarCell[] = [];
  for (let i = 0; i < leading; i++) cells.push({ blank: true });
  for (let d = 1; d <= daysInMonth; d++) {
    cells.push({ blank: false, day: d, booked: booked.has(d) });
  }
  return cells;
}
