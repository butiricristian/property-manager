import { useEffect, useMemo, useState } from "react";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import {
  getDashboard,
  getPropertyDashboard,
  listBookings,
} from "@/api/client";
import { useCurrency } from "@/context/CurrencyContext";
import { useProperty } from "@/context/PropertyContext";
import { useAuth } from "@/context/AuthContext";
import { DateRangeSelector } from "@/components/DateRangeSelector";
import { ImportBookingsDialog } from "@/components/ImportBookingsDialog";
import { Button } from "@/components/ui/button";
import type { DashboardSummary, Property } from "@/types";
import { AlertTriangle, Lock, Upload, X } from "lucide-react";
import {
  money,
  moneyShort,
  greeting,
  daysInRange,
  focusMonth,
  trendMonths,
  bookedDays,
  calendarCells,
  previousPeriodRange,
} from "@/lib/dashboard";

interface TrendPoint {
  label: string;
  value: number;
  current: boolean;
}

export function DashboardPage() {
  const { currency } = useCurrency();
  const { properties, propertyId, setPropertyId, dateRange, setDateRange } =
    useProperty();
  const { user, isGuest } = useAuth();

  const [viewMode, setViewMode] = useState<"all" | "property">("all");
  const [summary, setSummary] = useState<DashboardSummary | null>(null);
  const [trend, setTrend] = useState<TrendPoint[]>([]);
  const [bookedSet, setBookedSet] = useState<Set<number>>(new Set());
  const [prevTakehome, setPrevTakehome] = useState<number | null>(null);
  const [loading, setLoading] = useState(true);
  const [refreshKey, setRefreshKey] = useState(0);

  const focus = useMemo(() => focusMonth(dateRange.endDate), [dateRange.endDate]);

  useEffect(() => {
    let cancelled = false;
    setLoading(true);

    const summaryPromise =
      viewMode === "all"
        ? getDashboard(dateRange.startDate, dateRange.endDate, currency)
        : getPropertyDashboard(
            propertyId,
            dateRange.startDate,
            dateRange.endDate,
            currency
          );

    // Last 6 months of take-home for the trend chart.
    const months = trendMonths(focus, 6);
    const trendPromise = Promise.all(
      months.map((m) =>
        (viewMode === "all"
          ? getDashboard(m.startDate, m.endDate, currency)
          : getPropertyDashboard(propertyId, m.startDate, m.endDate, currency)
        )
          .then((s) => s.profit_after_tax)
          .catch(() => 0)
      )
    ).then((values) =>
      values.map((value, i) => ({
        label: months[i].label,
        value,
        current: i === months.length - 1,
      }))
    );

    // Bookings for the occupancy calendar. Bookings are filtered by check-out date,
    // so extend the window a month to catch stays that check in this month but
    // check out the next (their nights still belong on this month's calendar).
    const fStart = `${focus.year}-${String(focus.month + 1).padStart(2, "0")}-01`;
    const fEndDate = new Date(focus.year, focus.month + 2, 1);
    const fEnd = `${fEndDate.getFullYear()}-${String(fEndDate.getMonth() + 1).padStart(
      2,
      "0"
    )}-01`;
    const targetProperties =
      viewMode === "property"
        ? properties.filter((p) => p.id === propertyId)
        : properties;
    const bookingsPromise = Promise.all(
      targetProperties.map((p) =>
        listBookings(p.id, fStart, fEnd, currency).catch(() => [])
      )
    ).then((lists) => lists.flat());

    // Same-length window one month earlier, for the "vs last month" comparison.
    const prev = previousPeriodRange(dateRange.startDate, dateRange.endDate);
    const prevPromise = (
      viewMode === "all"
        ? getDashboard(prev.startDate, prev.endDate, currency)
        : getPropertyDashboard(propertyId, prev.startDate, prev.endDate, currency)
    )
      .then((s) => s.profit_after_tax)
      .catch(() => null);

    Promise.all([summaryPromise, trendPromise, bookingsPromise, prevPromise])
      .then(([s, t, bookings, prevValue]) => {
        if (cancelled) return;
        setSummary(s);
        setTrend(t);
        setBookedSet(bookedDays(bookings, focus.year, focus.month));
        setPrevTakehome(prevValue);
      })
      .catch((err) => {
        if (!cancelled) console.error(err);
      })
      .finally(() => {
        if (!cancelled) setLoading(false);
      });

    return () => {
      cancelled = true;
    };
  }, [dateRange, viewMode, propertyId, currency, properties, focus, refreshKey]);

  const ownerName =
    user?.name?.split(" ")[0] || user?.email?.split("@")[0] || "there";
  const selectedProperty = properties.find((p) => p.id === propertyId);
  const propertyName =
    viewMode === "all" ? "All properties" : selectedProperty?.name ?? "Property";
  const city = viewMode === "property" ? selectedProperty?.address : undefined;

  // Current, in-progress month → the take-home tile can show a run-rate projection.
  const periodDays = daysInRange(dateRange.startDate, dateRange.endDate);
  const [pmYear, pmMonth] = dateRange.startDate.split("-").map(Number);
  const inProgressMonth =
    dateRange.preset === "month" &&
    periodDays > 0 &&
    periodDays < new Date(pmYear, pmMonth, 0).getDate();

  const importProps: ImportProps = {
    properties,
    propertyId,
    setPropertyId: (id) => {
      setViewMode("property");
      setPropertyId(id);
    },
    onImported: () => setRefreshKey((k) => k + 1),
  };

  return (
    <div className="mx-auto flex max-w-[1240px] flex-col gap-[22px]">
      {/* Top bar */}
      <div className="flex flex-wrap items-end justify-between gap-4">
        <div className="flex flex-col gap-1">
          <div className="text-sm font-medium text-muted-foreground">
            {greeting()}, {ownerName} 👋
          </div>
          <div className="flex items-center gap-2.5">
            <div className="flex h-[34px] w-[34px] items-center justify-center rounded-[9px] bg-accent text-[17px]">
              🏠
            </div>
            <h1 className="font-heading text-[26px] font-bold leading-none tracking-[-0.02em]">
              {propertyName}
            </h1>
            {city && (
              <span className="mt-[3px] text-[13px] font-medium text-muted-foreground">
                · {city}
              </span>
            )}
          </div>
        </div>
        <div className="flex items-center gap-2.5">
          <DateRangeSelector value={dateRange} onChange={setDateRange} />
          <Select
            value={viewMode === "all" ? "all" : propertyId}
            onValueChange={(v) => {
              if (v === "all") setViewMode("all");
              else if (v) {
                setViewMode("property");
                setPropertyId(v);
              }
            }}
          >
            <SelectTrigger className="rounded-[11px] border border-border text-[13px]">
              <SelectValue placeholder="All Properties">
                {(value: string | null) => {
                  if (!value || value === "all") return "All Properties";
                  return properties.find((p) => p.id === value)?.name ?? value;
                }}
              </SelectValue>
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="all">All Properties</SelectItem>
              {properties.map((p) => (
                <SelectItem key={p.id} value={p.id}>
                  {p.name}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        </div>
      </div>

      {summary && (
        <LowDataBanner
          bookingsCount={summary.bookings_count}
          importProps={importProps}
        />
      )}

      {loading || !summary ? (
        <DashboardSkeleton />
      ) : (
        <DashboardContent
          summary={summary}
          trend={trend}
          bookedSet={bookedSet}
          focus={focus}
          days={periodDays}
          trendUp={takehomeChange(summary.profit_after_tax, prevTakehome)}
          isGuest={isGuest}
          inProgressMonth={inProgressMonth}
        />
      )}
    </div>
  );
}

interface ImportProps {
  properties: Property[];
  propertyId: string;
  setPropertyId: (id: string) => void;
  onImported: () => void;
}

const LOW_BOOKINGS_THRESHOLD = 3;
const DISMISS_KEY = "import-warning-dismissed";

/** Dismissible top-of-page nudge shown when the account has very few bookings. */
function LowDataBanner({
  bookingsCount,
  importProps,
}: {
  bookingsCount: number;
  importProps: ImportProps;
}) {
  const [dismissed, setDismissed] = useState(
    () => localStorage.getItem(DISMISS_KEY) === "1"
  );

  if (dismissed || bookingsCount >= LOW_BOOKINGS_THRESHOLD) return null;

  return (
    <div className="flex items-center gap-3 rounded-xl border border-[#EAD3A0] bg-[#FBF3E2] px-4 py-2.5 text-[13px] text-[#7A5A18]">
      <AlertTriangle className="size-4 shrink-0 text-[#C0603A]" />
      <span className="flex-1">
        Looks like you have very few bookings
        {bookingsCount > 0 ? ` (${bookingsCount})` : ""} — you might be missing data.
      </span>
      <ImportBookingsDialog
        properties={importProps.properties}
        propertyId={importProps.propertyId}
        setPropertyId={importProps.setPropertyId}
        onImported={importProps.onImported}
        triggerRender={
          <Button
            variant="outline"
            size="sm"
            className="gap-1.5 border-[#E3C98A] bg-white/60 text-[#7A5A18] hover:bg-white"
          />
        }
      >
        <Upload className="size-3.5" />
        Import from Booking.com / Airbnb
      </ImportBookingsDialog>
      <button
        type="button"
        aria-label="Dismiss"
        onClick={() => {
          localStorage.setItem(DISMISS_KEY, "1");
          setDismissed(true);
        }}
        className="rounded-md p-1 text-[#B78A2E] transition-colors hover:bg-[#F1E3BE] hover:text-[#7A5A18]"
      >
        <X className="size-4" />
      </button>
    </div>
  );
}

function takehomeChange(cur: number, prev: number | null): number | null {
  if (prev === null || prev <= 0) return null;
  return ((cur - prev) / prev) * 100;
}

interface ContentProps {
  summary: DashboardSummary;
  trend: TrendPoint[];
  bookedSet: Set<number>;
  focus: { year: number; month: number };
  days: number;
  trendUp: number | null;
  isGuest: boolean;
  inProgressMonth: boolean;
}

function DashboardContent({
  summary,
  trend,
  bookedSet,
  focus,
  days,
  trendUp,
  isGuest,
  inProgressMonth,
}: ContentProps) {
  const cur = summary.currency;
  const gross = summary.total_income;
  const expenses = summary.expenses.total;
  const tax = summary.total_tax;
  const takehome = summary.profit_after_tax;
  const profit = summary.profit;
  const fees = summary.platform_fees;
  const partnerPayout = summary.expenses.partner_payout;
  const hasPartner = partnerPayout > 0;
  // The payout is a % of net income and already sits inside expenses.total,
  // so `profit` is what's left for you after it — don't subtract it again.
  const partnerBase = summary.net_income;

  const occ = summary.occupancy_rate; // 0..1
  const bookedNights = Math.round(occ * days);
  const adr = summary.adr;
  const revpar = summary.revpar;
  const avgStay = summary.avg_stay_length;
  const leadDays = Math.round(Math.abs(summary.avg_lead_time_days));
  const cancRate = summary.cancellation_rate;
  const uplift = adr * 5;

  // ---- Money-flow segments (breakdown of gross income) ----
  const segments = [
    ...(fees > 0
      ? [{ label: "Platform fees", amount: fees, fill: "#D8C6B0", text: "#5F5646" }]
      : []),
    { label: "Expenses", amount: expenses, fill: "#E4B7A6", text: "#7A3A22" },
    ...(tax > 0
      ? [{ label: "Estimated tax", amount: tax, fill: "#EAD3A0", text: "#7A5A18" }]
      : []),
    {
      label: "You keep",
      amount: takehome,
      fill: "#A7D8B7",
      text: "#2C5E3E",
      amountColor: "#2C7A4C",
    },
  ];
  const segTotal = gross > 0 ? gross : segments.reduce((a, s) => a + s.amount, 0) || 1;

  // ---- Calendar ----
  const daysInFocusMonth = new Date(focus.year, focus.month + 1, 0).getDate();
  const cells = calendarCells(focus.year, focus.month, bookedSet);
  const openNights = daysInFocusMonth - bookedSet.size;
  const monthName = new Date(focus.year, focus.month, 1).toLocaleDateString(
    "en-US",
    { month: "long" }
  );

  // ---- Trend ----
  const trendMax = Math.max(1, ...trend.map((t) => t.value));
  const trendAvg =
    trend.length > 0 ? trend.reduce((a, t) => a + t.value, 0) / trend.length : 0;

  const lowOccupancy = occ < 0.6;

  // Run-rate projection of take-home to month-end (in-progress current month only).
  const daysInMonth = new Date(focus.year, focus.month + 1, 0).getDate();
  const estimateReady = inProgressMonth && days >= 5;
  const estMonthTakehome = days > 0 ? (takehome / days) * daysInMonth : 0;

  const blur = isGuest ? "blur-md select-none" : "";

  return (
    <>
      {/* Hero row */}
      <div className="grid gap-[18px] lg:grid-cols-[0.95fr_1.35fr]">
        {/* Take-home profit */}
        <div className="flex min-h-[240px] flex-col justify-between rounded-[20px] bg-primary p-[28px_30px] text-primary-foreground">
          <div className="flex items-start justify-between">
            <div className="text-sm font-semibold text-[#C9BEAF]">
              Take-home profit
            </div>
            {trendUp !== null && (
              <div
                className="rounded-[20px] px-2.5 py-[5px] text-[12.5px] font-bold"
                style={{
                  background:
                    trendUp >= 0
                      ? "rgba(127,209,154,0.16)"
                      : "rgba(212,120,90,0.18)",
                  color: trendUp >= 0 ? "#8FE0AB" : "#E9A98F",
                }}
              >
                {trendUp >= 0 ? "▲" : "▼"} {Math.abs(Math.round(trendUp))}% vs last
                month
              </div>
            )}
          </div>
          <div className="flex flex-col gap-1.5">
            <div
              className={`font-heading text-[52px] font-bold leading-[0.95] tracking-[-0.03em] ${blur}`}
            >
              {money(takehome, cur)}
            </div>
            <div className="text-[13.5px] leading-[1.45] text-[#B7AC9D]">
              What you actually keep this period — after {moneyShort(expenses, cur)}{" "}
              expenses{tax > 0 ? ` and ${moneyShort(tax, cur)} estimated tax` : ""}.
            </div>
            {inProgressMonth && (
              <div
                className="text-[12.5px] font-semibold"
                style={{ color: estimateReady ? "#8FE0AB" : "#B7AC9D" }}
              >
                {estimateReady ? (
                  <>
                    ≈ <span className={blur}>{moneyShort(estMonthTakehome, cur)}</span>{" "}
                    projected by month-end
                  </>
                ) : (
                  "Month-end projection in a few days"
                )}
              </div>
            )}
          </div>
        </div>

        {/* Where your money went */}
        <div className="flex flex-col gap-5 rounded-[20px] border bg-card p-[26px_28px]">
          <div className="flex items-baseline justify-between">
            <div className="font-heading text-[17px] font-semibold">
              Where your money went
            </div>
            <div className="text-[13px] text-muted-foreground">
              Gross income{" "}
              <b className="font-bold text-foreground">{money(gross, cur)}</b>
            </div>
          </div>

          {/* Segmented bar */}
          <div className="flex h-[46px] w-full gap-[3px]">
            {segments.map((s, i) => {
              const pct = (s.amount / segTotal) * 100;
              const radius =
                segments.length === 1
                  ? "8px"
                  : i === 0
                    ? "8px 4px 4px 8px"
                    : i === segments.length - 1
                      ? "4px 8px 8px 4px"
                      : "4px";
              return (
                <div
                  key={s.label}
                  className="flex items-center justify-center overflow-hidden text-[13px] font-bold"
                  style={{
                    width: `${Math.max(pct, 0)}%`,
                    background: s.fill,
                    color: s.text,
                    borderRadius: radius,
                  }}
                >
                  {pct >= 7 ? `${Math.round(pct)}%` : ""}
                </div>
              );
            })}
          </div>

          {/* Legend / waterfall */}
          <div
            className="grid gap-3.5"
            style={{
              gridTemplateColumns: `repeat(${segments.length}, minmax(0, 1fr))`,
            }}
          >
            {segments.map((s) => (
              <div key={s.label} className="flex flex-col gap-[3px]">
                <div className="flex items-center gap-[7px] text-[12.5px] font-semibold text-muted-foreground">
                  <span
                    className="h-[9px] w-[9px] rounded-[3px]"
                    style={{ background: s.fill }}
                  />
                  {s.label}
                </div>
                <div
                  className="font-heading text-[20px] font-bold tracking-[-0.01em]"
                  style={s.amountColor ? { color: s.amountColor } : undefined}
                >
                  {money(s.amount, cur)}
                </div>
              </div>
            ))}
          </div>

          {hasPartner && (
            <div className="flex items-center justify-between border-t border-dashed border-[#EAE2D6] pt-[13px] text-[12.5px] text-muted-foreground">
              <span>Partner split on net income ({money(partnerBase, cur)})</span>
              <span className={blur}>
                Your share after expenses{" "}
                <b className="font-bold text-foreground">{money(profit, cur)}</b>{" "}
                · Partner{" "}
                <b className="font-bold text-foreground">
                  {money(partnerPayout, cur)}
                </b>
              </span>
            </div>
          )}
        </div>
      </div>

      {/* Insights */}
      <SectionHeader>What to look at</SectionHeader>
      <div className="grid gap-4 lg:grid-cols-[1.3fr_1fr_1fr]">
        {/* Priority: occupancy — warm/terracotta only when it needs attention */}
        <div
          className={`flex flex-col gap-3 rounded-[18px] border p-[22px_22px_20px] ${
            lowOccupancy ? "border-[#F0D6C8] bg-[#FBF1EC]" : "bg-card"
          }`}
        >
          {lowOccupancy ? (
            <InsightTag bg="#C0603A" color="#FFFFFF">
              BIGGEST OPPORTUNITY
            </InsightTag>
          ) : (
            <InsightTag bg="#E4F0E6" color="#2C7A4C">
              LOOKING GOOD
            </InsightTag>
          )}
          {lowOccupancy ? (
            <>
              <div className="font-heading text-[18px] font-semibold leading-[1.25] tracking-[-0.01em]">
                Your place sat empty {days - bookedNights} of {days} nights.
              </div>
              <div className="text-[13.5px] leading-[1.5] text-[#6E6459]">
                Occupancy is just{" "}
                <b style={{ color: "#C0603A" }}>{Math.round(occ * 100)}%</b>. Your
                nightly rate is healthy — the gap is empty nights, not price. Filling{" "}
                <b>5 more nights</b> at your current rate would add about{" "}
                <b style={{ color: "#2C7A4C" }}>{moneyShort(uplift, cur)}</b> in gross
                income.
              </div>
              <div className="mt-0.5 flex flex-wrap gap-2">
                <button className="rounded-[9px] bg-primary px-[13px] py-2 text-[12.5px] font-semibold text-primary-foreground transition-opacity hover:opacity-90">
                  Lower minimum stay
                </button>
                <button className="rounded-[9px] border border-[#E7CFC0] bg-card px-[13px] py-2 text-[12.5px] font-semibold text-[#8A4E33] transition-colors hover:bg-[#FBF1EC]">
                  Add a weekly discount
                </button>
              </div>
            </>
          ) : (
            <>
              <div className="font-heading text-[18px] font-semibold leading-[1.25] tracking-[-0.01em]">
                You're nearly full — {bookedNights} of {days} nights booked.
              </div>
              <div className="text-[13.5px] leading-[1.5] text-[#6E6459]">
                Occupancy is a strong{" "}
                <b style={{ color: "#2C7A4C" }}>{Math.round(occ * 100)}%</b>. With
                demand this healthy, a small rate increase could lift take-home
                without costing you many nights.
              </div>
            </>
          )}
        </div>

        {/* Working well */}
        <div className="flex flex-col gap-2.5 rounded-[18px] border bg-card p-[22px]">
          <InsightTag bg="#E4F0E6" color="#2C7A4C">
            WORKING WELL
          </InsightTag>
          <div className="font-heading text-[16.5px] font-semibold leading-[1.3]">
            Your pricing &amp; guests are solid.
          </div>
          <div className="text-[13px] leading-[1.5] text-[#6E6459]">
            A <b>{moneyShort(adr, cur)}</b> average rate and{" "}
            <b>{Math.round(cancRate * 100)}% cancellations</b> mean the guests you do
            get are the right ones. Keep the rate — focus on volume.
          </div>
        </div>

        {/* Worth a try */}
        <div className="flex flex-col gap-2.5 rounded-[18px] border bg-card p-[22px]">
          <InsightTag bg="#F5EAD6" color="#97701F">
            WORTH A TRY
          </InsightTag>
          <div className="font-heading text-[16.5px] font-semibold leading-[1.3]">
            {avgStay < 2 ? "Stays are short & last-minute." : "Guests book close in."}
          </div>
          <div className="text-[13px] leading-[1.5] text-[#6E6459]">
            Guests stay <b>~{avgStay.toFixed(1)} night{avgStay >= 1.5 ? "s" : ""}</b>{" "}
            and book <b>{leadDays <= 7 ? "just days out" : `~${leadDays} days out`}</b>.
            A small discount for 3+ night stays could raise both occupancy and
            effort-per-booking.
          </div>
        </div>
      </div>

      {/* KPIs */}
      <SectionHeader>How the place is performing</SectionHeader>
      <div className="grid grid-cols-2 gap-3.5 md:grid-cols-3 xl:grid-cols-6">
        <KpiCard
          label="Occupancy"
          value={`${Math.round(occ * 100)}%`}
          note={`${bookedNights} of ${days} nights`}
          status={lowOccupancy ? "watch" : "good"}
        />
        <KpiCard
          label="Nightly rate (ADR)"
          value={moneyShort(adr, cur)}
          note="Healthy for the area"
          status="good"
        />
        <KpiCard
          label="RevPAR"
          value={money(revpar, cur)}
          note={lowOccupancy ? "Low — empty nights" : "Solid utilisation"}
          status={lowOccupancy ? "watch" : "good"}
        />
        <KpiCard
          label="Avg stay"
          value={`${avgStay.toFixed(1)} nt`}
          note={avgStay < 2 ? "Short bookings" : "Healthy stays"}
          status="neutral"
        />
        <KpiCard
          label="Booking lead"
          value={`~${leadDays} days`}
          note={leadDays < 7 ? "Last-minute guests" : "Guests plan ahead"}
          status="neutral"
          blur={blur}
        />
        <KpiCard
          label="Cancellations"
          value={`${Math.round(cancRate * 100)}%`}
          note={cancRate <= 0.05 ? "None — great" : "Keep an eye on this"}
          status={cancRate <= 0.05 ? "good" : "watch"}
          blur={blur}
        />
      </div>

      {/* Bottom row */}
      <div className="grid gap-[18px] lg:grid-cols-2">
        {/* Occupancy calendar */}
        <div className="flex flex-col gap-4 rounded-[20px] border bg-card p-[24px_26px]">
          <div className="flex items-baseline justify-between">
            <div className="font-heading text-[17px] font-semibold">
              {monthName} at a glance
            </div>
            <div className="text-[12.5px] text-muted-foreground">
              <b style={{ color: "#C0603A" }}>{openNights}</b> open nights left to fill
            </div>
          </div>
          <div className="grid grid-cols-7 gap-1.5">
            {["M", "T", "W", "T", "F", "S", "S"].map((d, i) => (
              <div
                key={i}
                className="text-center text-[10.5px] font-semibold text-[#B0A597]"
              >
                {d}
              </div>
            ))}
            {cells.map((c, i) =>
              c.blank ? (
                <div key={i} className="aspect-square" />
              ) : (
                <div
                  key={i}
                  className="flex aspect-square items-center justify-center rounded-lg text-[12px]"
                  style={
                    c.booked
                      ? { background: "#A7D8B7", color: "#235C39", fontWeight: 700 }
                      : { background: "#F5F0E8", color: "#C2B7A8", fontWeight: 600 }
                  }
                >
                  {c.day}
                </div>
              )
            )}
          </div>
          <div className="flex gap-4 text-[11.5px] text-muted-foreground">
            <div className="flex items-center gap-1.5">
              <span
                className="h-[11px] w-[11px] rounded-[4px]"
                style={{ background: "#A7D8B7" }}
              />
              Booked
            </div>
            <div className="flex items-center gap-1.5">
              <span
                className="h-[11px] w-[11px] rounded-[4px] border border-[#EAE2D6]"
                style={{ background: "#F5F0E8" }}
              />
              Open
            </div>
          </div>
        </div>

        {/* Take-home trend */}
        <div className="flex flex-col gap-4 rounded-[20px] border bg-card p-[24px_26px]">
          <div className="flex items-baseline justify-between">
            <div className="font-heading text-[17px] font-semibold">
              Take-home, last 6 months
            </div>
            <div className="text-[12.5px] text-muted-foreground">
              avg <b className="text-foreground">{moneyShort(trendAvg, cur)}</b>
            </div>
          </div>
          <div className="flex h-[168px] items-end justify-between gap-3 pt-2">
            {trend.map((m, i) => (
              <div
                key={i}
                className="flex h-full flex-1 flex-col items-center justify-end gap-2"
              >
                <div className="text-[11px] font-bold text-[#6E6459]">
                  {moneyShort(m.value, cur).replace(`${cur} `, "")}
                </div>
                <div
                  className="w-full max-w-[42px] rounded-t-lg"
                  style={{
                    height: `${(m.value / trendMax) * 100}%`,
                    minHeight: "4px",
                    borderRadius: "8px 8px 3px 3px",
                    background: m.current ? "#C0603A" : "#E3D8C7",
                  }}
                />
                <div className="text-[11px] font-semibold text-muted-foreground">
                  {m.label}
                </div>
              </div>
            ))}
          </div>
        </div>
      </div>
    </>
  );
}

function SectionHeader({ children }: { children: React.ReactNode }) {
  return (
    <div className="flex items-center gap-2">
      <div className="font-heading text-[15px] font-semibold">{children}</div>
      <div className="h-px flex-1 bg-[#E7DFD3]" />
    </div>
  );
}

function InsightTag({
  bg,
  color,
  children,
}: {
  bg: string;
  color: string;
  children: React.ReactNode;
}) {
  return (
    <span
      className="self-start rounded-md px-[9px] py-1 text-[10.5px] font-bold tracking-[0.06em]"
      style={{ background: bg, color }}
    >
      {children}
    </span>
  );
}

const dotColor: Record<"good" | "watch" | "neutral", string> = {
  good: "#4E9E6A",
  watch: "#C0603A",
  neutral: "#C7BCAD",
};

function KpiCard({
  label,
  value,
  note,
  status,
  blur = "",
}: {
  label: string;
  value: string;
  note: string;
  status: "good" | "watch" | "neutral";
  blur?: string;
}) {
  return (
    <div className="relative flex min-h-[132px] flex-col gap-2 rounded-2xl border bg-card p-[18px_16px]">
      <div className="text-[12px] font-semibold text-muted-foreground">{label}</div>
      <div
        className={`font-heading text-[26px] font-bold leading-none tracking-[-0.02em] ${blur}`}
      >
        {value}
      </div>
      <div className="mt-auto flex items-center gap-1.5">
        <span
          className="h-[7px] w-[7px] flex-none rounded-full"
          style={{ background: dotColor[status] }}
        />
        <span className="text-[11.5px] leading-[1.35] text-muted-foreground">
          {note}
        </span>
      </div>
      {blur && <GuestLockOverlay />}
    </div>
  );
}

function DashboardSkeleton() {
  return (
    <div className="flex animate-pulse flex-col gap-[22px]">
      <div className="grid gap-[18px] lg:grid-cols-[0.95fr_1.35fr]">
        <div className="h-[240px] rounded-[20px] bg-muted" />
        <div className="h-[240px] rounded-[20px] bg-muted" />
      </div>
      <div className="grid gap-4 lg:grid-cols-[1.3fr_1fr_1fr]">
        {[0, 1, 2].map((i) => (
          <div key={i} className="h-[180px] rounded-[18px] bg-muted" />
        ))}
      </div>
      <div className="grid grid-cols-2 gap-3.5 md:grid-cols-3 xl:grid-cols-6">
        {[0, 1, 2, 3, 4, 5].map((i) => (
          <div key={i} className="h-[132px] rounded-2xl bg-muted" />
        ))}
      </div>
      <div className="grid gap-[18px] lg:grid-cols-2">
        <div className="h-[280px] rounded-[20px] bg-muted" />
        <div className="h-[280px] rounded-[20px] bg-muted" />
      </div>
    </div>
  );
}

function GuestLockOverlay() {
  return (
    <div className="absolute inset-0 flex items-center justify-center rounded-2xl bg-card/60">
      <div className="flex items-center gap-1.5 text-xs text-muted-foreground">
        <Lock className="h-3.5 w-3.5" />
        <span>Create account to unlock</span>
      </div>
    </div>
  );
}
