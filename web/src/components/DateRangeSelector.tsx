import { useState } from "react";
import { Popover } from "@base-ui/react/popover";
import { ChevronLeft, ChevronRight } from "lucide-react";
import { clampEndToToday } from "@/lib/dashboard";
import type { DateRange } from "@/context/PropertyContext";

interface Props {
  value: DateRange;
  onChange: (range: DateRange) => void;
}

function pad(n: number): string {
  return String(n).padStart(2, "0");
}

/** A single-month range (preset "month") for the given year / 0-based month. */
function monthRange(year: number, month0: number): DateRange {
  const start = new Date(year, month0, 1);
  const next = new Date(year, month0 + 1, 1);
  const startDate = `${start.getFullYear()}-${pad(start.getMonth() + 1)}-01`;
  const endDate = `${next.getFullYear()}-${pad(next.getMonth() + 1)}-01`;
  return {
    startDate,
    // In-progress month: compute stats month-to-date, not to the month's end.
    endDate: clampEndToToday(startDate, endDate),
    preset: "month",
  };
}

/** The month the navigator should display: the selected month, or the current month. */
function displayedMonth(value: DateRange): { year: number; month: number } {
  if (value.preset === "month") {
    const [y, m] = value.startDate.split("-").map(Number);
    return { year: y, month: m - 1 };
  }
  const now = new Date();
  return { year: now.getFullYear(), month: now.getMonth() };
}

function computeRangePreset(preset: "ytd" | "last-year"): DateRange {
  const now = new Date();
  const y = now.getFullYear();
  const m = now.getMonth();
  if (preset === "ytd") {
    const next = new Date(y, m + 1, 1);
    const endDate = `${next.getFullYear()}-${pad(next.getMonth() + 1)}-01`;
    return {
      startDate: `${y}-01-01`,
      endDate: clampEndToToday(`${y}-01-01`, endDate),
      preset,
    };
  }
  // last-year
  return { startDate: `${y - 1}-01-01`, endDate: `${y}-01-01`, preset };
}

function monthFromDate(dateStr: string): string {
  return dateStr.slice(0, 7);
}

function startOfMonth(monthStr: string): string {
  return `${monthStr}-01`;
}

function endOfMonthExclusive(monthStr: string): string {
  const [y, m] = monthStr.split("-").map(Number);
  const next = new Date(y, m, 1);
  return `${next.getFullYear()}-${pad(next.getMonth() + 1)}-01`;
}

function formatMonthLabel(dateStr: string): string {
  const d = new Date(dateStr + "T00:00:00");
  return d.toLocaleDateString("en-US", { month: "short", year: "numeric" });
}

function formatMonthYear(year: number, month0: number): string {
  return new Date(year, month0, 1).toLocaleDateString("en-US", {
    month: "long",
    year: "numeric",
  });
}

export function DateRangeSelector({ value, onChange }: Props) {
  const [customStart, setCustomStart] = useState(() => monthFromDate(value.startDate));
  const [customEnd, setCustomEnd] = useState(() => {
    const [y, m] = value.endDate.split("-").map(Number);
    const prev = new Date(y, m - 2, 1);
    return `${prev.getFullYear()}-${pad(prev.getMonth() + 1)}`;
  });
  const [popoverOpen, setPopoverOpen] = useState(false);
  const [monthPickerOpen, setMonthPickerOpen] = useState(false);

  const dm = displayedMonth(value);
  const monthActive = value.preset === "month";

  const stepMonth = (delta: number) => {
    const d = new Date(dm.year, dm.month + delta, 1);
    onChange(monthRange(d.getFullYear(), d.getMonth()));
  };

  const jumpToMonth = (monthStr: string) => {
    const [y, m] = monthStr.split("-").map(Number);
    if (!y || !m) return;
    onChange(monthRange(y, m - 1));
    setMonthPickerOpen(false);
  };

  const pill = (active: boolean) =>
    `inline-flex h-9 items-center rounded-lg px-3.5 text-[13px] transition-colors ${
      active
        ? "bg-primary font-semibold text-primary-foreground"
        : "font-medium text-muted-foreground hover:bg-muted"
    }`;

  const handleCustomStartChange = (monthStr: string) => {
    setCustomStart(monthStr);
    if (monthStr && customEnd && monthStr <= customEnd) {
      onChange({
        startDate: startOfMonth(monthStr),
        endDate: endOfMonthExclusive(customEnd),
        preset: "custom",
      });
    }
  };

  const handleCustomEndChange = (monthStr: string) => {
    setCustomEnd(monthStr);
    if (customStart && monthStr && customStart <= monthStr) {
      onChange({
        startDate: startOfMonth(customStart),
        endDate: endOfMonthExclusive(monthStr),
        preset: "custom",
      });
    }
  };

  const customLabel =
    customStart === customEnd
      ? formatMonthLabel(customStart)
      : `${formatMonthLabel(customStart)} – ${formatMonthLabel(customEnd)}`;

  return (
    <div className="flex items-center gap-2">
      <div className="inline-flex items-center gap-0.5 rounded-[11px] border bg-card p-[3px]">
        {/* Month navigator */}
        <button
          type="button"
          aria-label="Previous month"
          onClick={() => stepMonth(-1)}
          className="flex h-9 w-9 items-center justify-center rounded-lg text-muted-foreground transition-colors hover:bg-muted"
        >
          <ChevronLeft className="h-4 w-4" />
        </button>
        <Popover.Root open={monthPickerOpen} onOpenChange={setMonthPickerOpen}>
          <Popover.Trigger className={`${pill(monthActive)} min-w-[104px] justify-center`}>
            {formatMonthYear(dm.year, dm.month)}
          </Popover.Trigger>
          <Popover.Portal>
            <Popover.Positioner sideOffset={6} side="bottom" align="center">
              <Popover.Popup className="z-50 rounded-lg border bg-popover p-3 text-popover-foreground shadow-md outline-none animate-in fade-in-0 zoom-in-95">
                <div className="flex flex-col gap-2">
                  <label className="text-xs font-medium text-muted-foreground">
                    Jump to month
                  </label>
                  <input
                    type="month"
                    value={`${dm.year}-${pad(dm.month + 1)}`}
                    onChange={(e) => jumpToMonth(e.target.value)}
                    className="rounded-md border px-2 py-1.5 text-sm"
                  />
                  <button
                    type="button"
                    onClick={() => {
                      const now = new Date();
                      onChange(monthRange(now.getFullYear(), now.getMonth()));
                      setMonthPickerOpen(false);
                    }}
                    className="text-left text-xs font-medium text-primary hover:underline"
                  >
                    Jump to this month
                  </button>
                </div>
              </Popover.Popup>
            </Popover.Positioner>
          </Popover.Portal>
        </Popover.Root>
        <button
          type="button"
          aria-label="Next month"
          onClick={() => stepMonth(1)}
          className="flex h-9 w-9 items-center justify-center rounded-lg text-muted-foreground transition-colors hover:bg-muted"
        >
          <ChevronRight className="h-4 w-4" />
        </button>

        <div className="mx-1 h-5 w-px bg-border" />

        {/* Range presets */}
        <button
          type="button"
          onClick={() => onChange(computeRangePreset("ytd"))}
          className={pill(value.preset === "ytd")}
        >
          YTD
        </button>
        <button
          type="button"
          onClick={() => onChange(computeRangePreset("last-year"))}
          className={pill(value.preset === "last-year")}
        >
          Last Year
        </button>
        <Popover.Root open={popoverOpen} onOpenChange={setPopoverOpen}>
          <Popover.Trigger
            className={pill(value.preset === "custom")}
            onClick={() => {
              if (value.preset !== "custom") {
                onChange({
                  startDate: startOfMonth(customStart),
                  endDate: endOfMonthExclusive(customEnd),
                  preset: "custom",
                });
              }
            }}
          >
            {value.preset === "custom" ? customLabel : "Custom"}
          </Popover.Trigger>
          <Popover.Portal>
            <Popover.Positioner sideOffset={6} side="bottom" align="end">
              <Popover.Popup className="z-50 rounded-lg border bg-popover p-4 text-popover-foreground shadow-md outline-none animate-in fade-in-0 zoom-in-95">
                <div className="flex flex-col gap-3">
                  <div className="flex items-center gap-3">
                    <div className="flex flex-col gap-1">
                      <label className="text-xs font-medium text-muted-foreground">From</label>
                      <input
                        type="month"
                        value={customStart}
                        onChange={(e) => handleCustomStartChange(e.target.value)}
                        className="rounded-md border px-2 py-1.5 text-sm"
                      />
                    </div>
                    <span className="mt-5 text-sm text-muted-foreground">to</span>
                    <div className="flex flex-col gap-1">
                      <label className="text-xs font-medium text-muted-foreground">To</label>
                      <input
                        type="month"
                        value={customEnd}
                        onChange={(e) => handleCustomEndChange(e.target.value)}
                        className="rounded-md border px-2 py-1.5 text-sm"
                      />
                    </div>
                  </div>
                </div>
              </Popover.Popup>
            </Popover.Positioner>
          </Popover.Portal>
        </Popover.Root>
      </div>
    </div>
  );
}
