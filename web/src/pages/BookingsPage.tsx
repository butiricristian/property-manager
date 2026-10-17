import { useEffect, useState } from "react";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { Button } from "@/components/ui/button";
import {
  listBookings,
  deleteBooking,
  getBookingInvoice,
  downloadInvoiceXML,
} from "@/api/client";
import { useCurrency } from "@/context/CurrencyContext";
import { useProperty } from "@/context/PropertyContext";
import { useAuth } from "@/context/AuthContext";
import { DateRangeSelector } from "@/components/DateRangeSelector";
import { GenerateInvoiceDialog } from "@/components/GenerateInvoiceDialog";
import { ImportBookingsDialog } from "@/components/ImportBookingsDialog";
import { daysInRange } from "@/lib/dashboard";
import type { Booking, Invoice } from "@/types";
import { Upload, Trash2, FileText, Download } from "lucide-react";

export function BookingsPage() {
  const { currency } = useCurrency();
  const { properties, propertyId, setPropertyId, dateRange, setDateRange } = useProperty();
  const { isGuest } = useAuth();
  const [bookings, setBookings] = useState<Booking[]>([]);
  const [loading, setLoading] = useState(false);
  const [invoiceMap, setInvoiceMap] = useState<Record<string, Invoice>>({});
  const [invoiceDialogBooking, setInvoiceDialogBooking] = useState<Booking | null>(null);

  const loadBookings = () => {
    if (!propertyId) return;
    setLoading(true);
    listBookings(propertyId, dateRange.startDate, dateRange.endDate, currency)
      .then((bks) => {
        setBookings(bks);
        const map: Record<string, Invoice> = {};
        Promise.all(
          bks.map((b) =>
            getBookingInvoice(b.id)
              .then((inv) => {
                map[b.id] = inv;
              })
              .catch(() => {})
          )
        ).then(() => setInvoiceMap(map));
      })
      .catch(console.error)
      .finally(() => setLoading(false));
  };

  useEffect(loadBookings, [propertyId, dateRange, currency]);

  const refreshInvoices = () => {
    const map: Record<string, Invoice> = {};
    Promise.all(
      bookings.map((b) =>
        getBookingInvoice(b.id)
          .then((inv) => {
            map[b.id] = inv;
          })
          .catch(() => {})
      )
    ).then(() => setInvoiceMap(map));
  };

  const handleDelete = async (id: string) => {
    if (!confirm("Delete this booking?")) return;
    await deleteBooking(id);
    setBookings((prev) => prev.filter((b) => b.id !== id));
  };

  // Summary aggregates — confirmed (non-cancelled) bookings only.
  const active = bookings.filter((b) => b.status !== "cancelled");
  const confirmedCount = active.length;
  const cancelledCount = bookings.length - confirmedCount;
  const nightsBooked = active.reduce((a, b) => a + b.nights, 0);
  const daysInPeriod = daysInRange(dateRange.startDate, dateRange.endDate);
  const occupancy = daysInPeriod > 0 ? Math.min(nightsBooked / daysInPeriod, 1) * 100 : 0;
  const netRevenue = active.reduce((a, b) => a + parseFloat(b.net_amount), 0);
  const grossSum = active.reduce((a, b) => a + parseFloat(b.gross_amount), 0);
  const avgNightlyRate = confirmedCount > 0 ? grossSum / confirmedCount : 0;
  const avgStay = confirmedCount > 0 ? nightsBooked / confirmedCount : 0;

  // Run-rate projection for the current, in-progress month. The range is clamped
  // to today, so daysInPeriod = days elapsed; extrapolate net revenue to full month.
  const [sy, sm] = dateRange.startDate.split("-").map(Number);
  const daysInMonth = new Date(sy, sm, 0).getDate();
  const inProgressMonth =
    dateRange.preset === "month" && daysInPeriod > 0 && daysInPeriod < daysInMonth;
  const MIN_DAYS_FOR_ESTIMATE = 5;
  const estimateReady = inProgressMonth && daysInPeriod >= MIN_DAYS_FOR_ESTIMATE;
  const estMonthRevenue = daysInPeriod > 0 ? (netRevenue / daysInPeriod) * daysInMonth : 0;

  const summary = [
    { label: "Confirmed bookings", value: String(confirmedCount), note: `${cancelledCount} cancelled` },
    {
      label: "Nights booked",
      value: `${nightsBooked} / ${daysInPeriod}`,
      note: `${Math.round(occupancy)}% occupancy`,
    },
    { label: "Net revenue", value: formatAmount(String(netRevenue), currency), note: "after platform fees" },
    { label: "Avg nightly rate", value: formatAmount(String(avgNightlyRate), currency), note: "per booking, gross" },
    { label: "Avg stay length", value: `${avgStay.toFixed(1)} nts`, note: "nights per booking" },
    ...(inProgressMonth
      ? [
          {
            label: "Est. month revenue",
            value: estimateReady ? formatAmount(String(estMonthRevenue), currency) : "N/A",
            note: estimateReady
              ? `projected from ${daysInPeriod} of ${daysInMonth} days`
              : "not enough data yet",
          },
        ]
      : []),
  ];

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-3xl font-bold">Bookings</h1>
        <div className="flex gap-3">
          <DateRangeSelector value={dateRange} onChange={setDateRange} />
          <Select value={propertyId} onValueChange={(v) => v && setPropertyId(v)}>
            <SelectTrigger className="w-48">
              <SelectValue placeholder="Select property">
                {(value: string | null) => {
                  if (!value) return "Select property";
                  return properties.find((p) => p.id === value)?.name ?? value;
                }}
              </SelectValue>
            </SelectTrigger>
            <SelectContent>
              {properties.map((p) => (
                <SelectItem key={p.id} value={p.id}>
                  {p.name}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
          <ImportBookingsDialog
            properties={properties}
            propertyId={propertyId}
            onImported={loadBookings}
            triggerRender={<Button />}
          >
            <Upload className="mr-2 h-4 w-4" />
            Import File
          </ImportBookingsDialog>
        </div>
      </div>

      {loading ? (
        <p className="text-muted-foreground">Loading...</p>
      ) : bookings.length === 0 ? (
        <p className="text-muted-foreground">No bookings found for this month.</p>
      ) : (
        <>
          {/* Summary strip */}
          <div
            className={`grid grid-cols-2 overflow-hidden rounded-[16px] border bg-card sm:grid-cols-3 ${
              inProgressMonth ? "lg:grid-cols-6" : "lg:grid-cols-5"
            }`}
          >
            {summary.map((s) => (
              <div
                key={s.label}
                className="flex flex-col gap-1.5 border-divider px-5 py-4 [&:not(:last-child)]:border-r"
              >
                <div className="text-[11px] font-bold tracking-[0.04em] text-warm-600 uppercase">
                  {s.label}
                </div>
                <div className="font-heading text-[22px] font-bold tracking-[-0.01em]">
                  {s.value}
                </div>
                <div className="text-[11.5px] text-warm-600">{s.note}</div>
              </div>
            ))}
          </div>

          {/* Table */}
          <Table className="min-w-[1080px]">
            <TableHeader>
              <TableRow>
                <TableHead>Guest</TableHead>
                <TableHead>Platform</TableHead>
                <TableHead>Check-in</TableHead>
                <TableHead>Check-out</TableHead>
                <TableHead className="text-right">Nights</TableHead>
                <TableHead className="text-right">Gross</TableHead>
                <TableHead className="text-right">Price</TableHead>
                <TableHead className="text-right">Fee</TableHead>
                <TableHead className="text-right">Net</TableHead>
                <TableHead>Status</TableHead>
                <TableHead />
              </TableRow>
            </TableHeader>
            <TableBody>
              {bookings.map((b) => {
                const cancelled = b.status === "cancelled";
                return (
                  <TableRow
                    key={b.id}
                    className={cancelled ? "opacity-[0.62] hover:bg-[#FAF7F2]" : undefined}
                  >
                    <TableCell
                      className={`font-heading text-[14.5px] font-semibold ${
                        cancelled ? "text-warm-700 line-through" : "text-foreground"
                      }`}
                    >
                      {b.guest_name}
                    </TableCell>
                    <TableCell>
                      <PlatformBadge platform={b.platform} />
                    </TableCell>
                    <TableCell className="text-[13.5px] text-warm-800">
                      {formatDate(b.check_in)}
                    </TableCell>
                    <TableCell className="text-[13.5px] text-warm-800">
                      {formatDate(b.check_out)}
                    </TableCell>
                    <TableCell className="text-right text-[13.5px] text-warm-800">
                      {b.nights}
                    </TableCell>
                    <TableCell className="text-right text-[13.5px] text-warm-600">
                      {formatAmount(b.gross_amount, b.currency)}
                    </TableCell>
                    <TableCell className="text-right text-[13.5px] text-warm-600">
                      {formatAmount(b.price, b.currency)}
                    </TableCell>
                    <TableCell className="text-right text-[13.5px] text-terracotta">
                      −{formatAmount(b.platform_fee, b.currency)}
                    </TableCell>
                    <TableCell
                      className="text-right text-[14.5px] font-bold"
                      style={{ color: cancelled ? "#A99485" : "#2C7A4C" }}
                    >
                      {formatAmount(b.net_amount, b.currency)}
                    </TableCell>
                    <TableCell>
                      <StatusBadge status={b.status} />
                    </TableCell>
                    <TableCell>
                      <div className="flex items-center justify-end text-warm-600">
                        {isGuest ? (
                          <span title="Create an account to generate invoices">
                            <FileText className="h-4 w-4 text-muted-foreground/50" />
                          </span>
                        ) : invoiceMap[b.id] ? (
                          <>
                            <Button
                              variant="ghost"
                              size="icon"
                              title="Download Invoice"
                              onClick={() => downloadInvoiceXML(invoiceMap[b.id].id)}
                            >
                              <Download className="h-4 w-4" />
                            </Button>
                            <Button
                              variant="ghost"
                              size="icon"
                              title="Regenerate Invoice"
                              onClick={() => setInvoiceDialogBooking(b)}
                            >
                              <FileText className="h-4 w-4" />
                            </Button>
                          </>
                        ) : (
                          <Button
                            variant="ghost"
                            size="icon"
                            title="Generate Invoice"
                            onClick={() => setInvoiceDialogBooking(b)}
                          >
                            <FileText className="h-4 w-4" />
                          </Button>
                        )}
                        <Button
                          variant="ghost"
                          size="icon"
                          title="Delete booking"
                          onClick={() => handleDelete(b.id)}
                        >
                          <Trash2 className="h-4 w-4" />
                        </Button>
                      </div>
                    </TableCell>
                  </TableRow>
                );
              })}
            </TableBody>
          </Table>
        </>
      )}

      {invoiceDialogBooking && (
        <GenerateInvoiceDialog
          booking={invoiceDialogBooking}
          open={!!invoiceDialogBooking}
          onOpenChange={(open) => {
            if (!open) setInvoiceDialogBooking(null);
          }}
          onGenerated={refreshInvoices}
        />
      )}
    </div>
  );
}

function PlatformBadge({ platform }: { platform: string }) {
  const airbnb = platform === "airbnb";
  const cfg = airbnb
    ? { bg: "#FDEDE9", fg: "#C0603A", label: "Airbnb" }
    : { bg: "#E9F0FB", fg: "#2C5FA8", label: "Booking" };
  return (
    <span
      className="inline-flex items-center gap-1.5 rounded-[20px] px-3 py-[5px] text-[12px] font-bold"
      style={{ background: cfg.bg, color: cfg.fg }}
    >
      <span className="size-[7px] rounded-full" style={{ background: cfg.fg }} />
      {cfg.label}
    </span>
  );
}

function StatusBadge({ status }: { status: string }) {
  const cancelled = status === "cancelled";
  const cfg = cancelled
    ? { bg: "#F5EDE6", fg: "#A99485", dot: "#C2AF9E", label: "Cancelled" }
    : {
        bg: "#E7F2E9",
        fg: "#2C7A4C",
        dot: "#4E9E6A",
        label: status === "completed" ? "Completed" : "Confirmed",
      };
  return (
    <span
      className="inline-flex items-center gap-1.5 rounded-[20px] px-3 py-[5px] text-[12px] font-bold"
      style={{ background: cfg.bg, color: cfg.fg }}
    >
      <span className="size-[7px] rounded-full" style={{ background: cfg.dot }} />
      {cfg.label}
    </span>
  );
}

function formatDate(d: string): string {
  return new Date(d).toLocaleDateString("en-GB");
}

function formatAmount(v: string, currency = "EUR"): string {
  const n = typeof v === "string" ? parseFloat(v) : v;
  return new Intl.NumberFormat("en-EU", {
    style: "currency",
    currency,
  }).format(n);
}
