import { useEffect, useRef, useState } from "react";
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
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Badge } from "@/components/ui/badge";
import {
  listExpenses,
  createExpense,
  updateExpense,
  deleteExpense,
  extractBillExpense,
} from "@/api/client";
import { useProperty } from "@/context/PropertyContext";
import { useAuth } from "@/context/AuthContext";
import { DateRangeSelector } from "@/components/DateRangeSelector";
import type { Expense, ExpenseCategory } from "@/types";
import { Plus, Pencil, Copy, Trash2, Upload, Loader2 } from "lucide-react";

const CATEGORIES: ExpenseCategory[] = [
  "electricity",
  "heating",
  "internet",
  "water",
  "maintenance",
  "cleaning",
  "partner_payout",
  "other",
];

const CATEGORY_COLORS: Record<ExpenseCategory, string> = {
  electricity: "#C08A1E",
  heating: "#C0603A",
  internet: "#3B6FB0",
  water: "#2E8B84",
  cleaning: "#5C8A4E",
  partner_payout: "#7B6FB0",
  maintenance: "#8A8075",
  other: "#8A8075",
};

// Most-used utility categories surfaced in the quick-add bar, in order.
const QUICK_CATEGORIES: ExpenseCategory[] = [
  "electricity",
  "heating",
  "internet",
  "water",
];


// Most recent non-percentage amount per category (with its date so newer wins).
type LastAmountMap = Record<string, { amount: number; date: string }>;

function lastAmountsKey(propertyId: string): string {
  return `expense-last-amount:${propertyId}`;
}

function persistLastAmounts(propertyId: string, map: LastAmountMap) {
  try {
    localStorage.setItem(lastAmountsKey(propertyId), JSON.stringify(map));
  } catch {
    /* ignore */
  }
}

interface ExpenseForm {
  category: ExpenseCategory;
  description: string;
  amount: string;
  currency: string;
  date: string;
  endDate: string;
  recurring: boolean;
  isPercentage: boolean;
  percentage: string;
}

const emptyForm: ExpenseForm = {
  category: "electricity",
  description: "",
  amount: "",
  currency: "EUR",
  date: new Date().toISOString().slice(0, 10),
  endDate: "",
  recurring: false,
  isPercentage: false,
  percentage: "",
};

const dateOnly = (d: string) => String(d).slice(0, 10);

export function ExpensesPage() {
  const { properties, propertyId, setPropertyId, dateRange, setDateRange } = useProperty();
  const { isGuest } = useAuth();
  const [expenses, setExpenses] = useState<Expense[]>([]);
  const [loading, setLoading] = useState(false);
  const [formOpen, setFormOpen] = useState(false);
  const [editingId, setEditingId] = useState<string | null>(null);
  const [duplicating, setDuplicating] = useState(false);
  const [form, setForm] = useState<ExpenseForm>(emptyForm);
  const [extracting, setExtracting] = useState(false);
  const [lastAmounts, setLastAmounts] = useState<LastAmountMap>({});
  const [focusAmount, setFocusAmount] = useState(false);
  const fileInputRef = useRef<HTMLInputElement>(null);
  const amountInputRef = useRef<HTMLInputElement>(null);

  useEffect(() => {
    if (!propertyId) return;
    setLoading(true);
    listExpenses(propertyId, dateRange.startDate, dateRange.endDate)
      .then(setExpenses)
      .catch(console.error)
      .finally(() => setLoading(false));
  }, [propertyId, dateRange]);

  // Chip hints: seed from persisted values, then refresh from loaded history.
  useEffect(() => {
    if (!propertyId) return;
    try {
      const raw = localStorage.getItem(lastAmountsKey(propertyId));
      setLastAmounts(raw ? JSON.parse(raw) : {});
    } catch {
      setLastAmounts({});
    }
  }, [propertyId]);

  useEffect(() => {
    if (!propertyId || expenses.length === 0) return;
    setLastAmounts((prev) => {
      const next = { ...prev };
      for (const e of expenses) {
        if (e.is_percentage) continue;
        const d = String(e.date).slice(0, 10);
        const cur = next[e.category];
        if (!cur || d > cur.date) {
          next[e.category] = { amount: parseFloat(e.amount), date: d };
        }
      }
      persistLastAmounts(propertyId, next);
      return next;
    });
  }, [expenses, propertyId]);

  // Auto-focus the amount field when the modal is opened from a quick-add chip.
  useEffect(() => {
    if (!formOpen || !focusAmount) return;
    const t = setTimeout(() => {
      amountInputRef.current?.focus();
      setFocusAmount(false);
    }, 60);
    return () => clearTimeout(t);
  }, [formOpen, focusAmount]);

  const refresh = () => {
    if (!propertyId) return;
    listExpenses(propertyId, dateRange.startDate, dateRange.endDate).then(setExpenses).catch(console.error);
  };

  const handleSubmit = async () => {
    if (!propertyId) return;

    // Validate the relevant value is a positive number; otherwise keep the modal open.
    const numericValue = parseFloat(form.isPercentage ? form.percentage : form.amount);
    if (!(numericValue > 0)) {
      amountInputRef.current?.focus();
      return;
    }

    const data = {
      category: form.category,
      description: form.description,
      amount: form.isPercentage ? 0 : parseFloat(form.amount),
      currency: form.currency,
      date: form.date,
      end_date: form.recurring ? form.endDate : "",
      recurring: form.recurring,
      is_percentage: form.isPercentage,
      percentage: form.isPercentage ? parseFloat(form.percentage) : 0,
    };

    if (editingId) {
      await updateExpense(editingId, data as never);
      refresh();
    } else {
      const created = await createExpense(propertyId, data as never);
      setExpenses((prev) => [created, ...prev]);
      if (!form.isPercentage) {
        setLastAmounts((prev) => {
          const cur = prev[form.category];
          if (cur && cur.date > form.date) return prev;
          const next = {
            ...prev,
            [form.category]: { amount: parseFloat(form.amount), date: form.date },
          };
          persistLastAmounts(propertyId, next);
          return next;
        });
      }
    }
    setFormOpen(false);
    setEditingId(null);
    setDuplicating(false);
    setForm(emptyForm);
  };

  // Open the Add-Expense modal pre-filled for a quick-add category (RON, today, focus amount).
  const handleQuickAdd = (category: ExpenseCategory) => {
    setEditingId(null);
    setDuplicating(false);
    setForm({
      category,
      description: "",
      amount: "",
      currency: "RON",
      date: new Date().toISOString().slice(0, 10),
      endDate: "",
      recurring: false,
      isPercentage: false,
      percentage: "",
    });
    setFocusAmount(true);
    setFormOpen(true);
  };

  const submitOnEnter = (e: React.KeyboardEvent) => {
    if (e.key === "Enter") {
      e.preventDefault();
      handleSubmit();
    }
  };

  const handleEdit = (expense: Expense) => {
    setEditingId(expense.id);
    setDuplicating(false);
    setForm({
      category: expense.category,
      description: expense.description,
      amount: expense.is_percentage ? "" : expense.amount,
      currency: expense.currency,
      date: dateOnly(expense.date),
      endDate: expense.end_date ? dateOnly(expense.end_date) : "",
      recurring: expense.recurring,
      isPercentage: expense.is_percentage,
      percentage: expense.is_percentage ? expense.percentage : "",
    });
    setFormOpen(true);
  };

  // Duplicate an expense as a brand-new one, pre-dated to the next month so the
  // user just tweaks the amount before saving.
  const handleDuplicate = (expense: Expense) => {
    setEditingId(null);
    setDuplicating(true);
    setForm({
      category: expense.category,
      description: expense.description,
      amount: expense.is_percentage ? "" : expense.amount,
      currency: expense.currency,
      date: nextMonthDate(dateOnly(expense.date)),
      // The copy starts next month, so an inherited end date would already be past.
      endDate: "",
      recurring: expense.recurring,
      isPercentage: expense.is_percentage,
      percentage: expense.is_percentage ? expense.percentage : "",
    });
    setFormOpen(true);
  };

  const handleDelete = async (id: string) => {
    if (!confirm("Delete this expense?")) return;
    await deleteExpense(id);
    setExpenses((prev) => prev.filter((e) => e.id !== id));
  };

  const openNew = () => {
    setEditingId(null);
    setDuplicating(false);
    setForm(emptyForm);
    setFormOpen(true);
  };

  const handleFileUpload = async (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (!file || !propertyId) return;
    e.target.value = "";

    setExtracting(true);
    try {
      const result = await extractBillExpense(propertyId, file);
      setEditingId(null);
      setDuplicating(false);
      setForm({
        category: result.category,
        description: result.description,
        amount: String(result.amount),
        currency: result.currency || "EUR",
        date: result.date,
        endDate: "",
        recurring: false,
        isPercentage: false,
        percentage: "",
      });
      setFormOpen(true);
    } catch (err) {
      alert(err instanceof Error ? err.message : "Extraction failed");
    } finally {
      setExtracting(false);
    }
  };

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-3xl font-bold">Expenses</h1>
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
          <input
            ref={fileInputRef}
            type="file"
            accept=".pdf"
            className="hidden"
            onChange={handleFileUpload}
          />
          <Button
            variant="outline"
            onClick={() => fileInputRef.current?.click()}
            disabled={extracting || !propertyId || isGuest}
            title={isGuest ? "Create an account to use AI bill extraction" : undefined}
          >
            {extracting ? (
              <Loader2 className="mr-2 h-4 w-4 animate-spin" />
            ) : (
              <Upload className="mr-2 h-4 w-4" />
            )}
            {extracting ? "Extracting..." : "Upload Bill"}
          </Button>
          <Dialog open={formOpen} onOpenChange={setFormOpen}>
            <DialogTrigger
              render={<Button />}
              onClick={openNew}
            >
              <Plus className="mr-2 h-4 w-4" />
              Add Expense
            </DialogTrigger>
            <DialogContent className="w-[460px] sm:max-w-[460px]">
              <DialogHeader>
                <DialogTitle>
                  {editingId
                    ? "Edit expense"
                    : duplicating
                      ? "Duplicate expense"
                      : "Add expense"}
                </DialogTitle>
              </DialogHeader>

              <div className="space-y-4">
                <div className="space-y-[7px]">
                  <Label>Category</Label>
                  <Select
                    value={form.category}
                    onValueChange={(v) =>
                      v && setForm((f) => ({
                        ...f,
                        category: v as ExpenseCategory,
                      }))
                    }
                  >
                    <SelectTrigger className="w-full">
                      <SelectValue>
                        {(value: string | null) => value ? formatCategory(value) : "Select category"}
                      </SelectValue>
                    </SelectTrigger>
                    <SelectContent>
                      {CATEGORIES.map((c) => (
                        <SelectItem key={c} value={c}>
                          {formatCategory(c)}
                        </SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                </div>

                <div className="space-y-[7px]">
                  <Label>Description</Label>
                  <Input
                    placeholder="e.g. Enel monthly bill"
                    value={form.description}
                    onChange={(e) =>
                      setForm((f) => ({ ...f, description: e.target.value }))
                    }
                  />
                </div>

                <label className="flex items-center gap-2.5 text-[13.5px] font-semibold text-[#2A241F]">
                  <input
                    type="checkbox"
                    className="size-4 accent-[#C0603A]"
                    checked={form.isPercentage}
                    onChange={(e) =>
                      setForm((f) => ({ ...f, isPercentage: e.target.checked }))
                    }
                  />
                  Percentage of net income
                </label>

                <div className="space-y-[7px]">
                  <Label>{form.isPercentage ? "Percentage" : "Amount"}</Label>
                  <div className="flex gap-2">
                    <div className="flex h-11 flex-1 items-center gap-2 rounded-lg border-[1.5px] border-input bg-card px-3.5 transition-colors focus-within:border-ring">
                      {!form.isPercentage && (
                        <span className="text-[15px] font-medium text-muted-foreground">
                          {form.currency}
                        </span>
                      )}
                      <input
                        ref={amountInputRef}
                        inputMode="decimal"
                        value={form.isPercentage ? form.percentage : form.amount}
                        placeholder={
                          form.isPercentage
                            ? "0"
                            : lastAmounts[form.category]
                              ? lastAmounts[form.category].amount.toFixed(2)
                              : "0.00"
                        }
                        onChange={(e) => {
                          const v = e.target.value.replace(/[^\d.]/g, "");
                          setForm((f) =>
                            f.isPercentage
                              ? { ...f, percentage: v }
                              : { ...f, amount: v }
                          );
                        }}
                        onKeyDown={submitOnEnter}
                        className="w-full min-w-0 bg-transparent font-heading text-[17px] font-semibold text-foreground outline-none placeholder:font-normal placeholder:text-placeholder"
                      />
                      {form.isPercentage && (
                        <span className="text-[15px] font-medium text-muted-foreground">
                          %
                        </span>
                      )}
                    </div>
                    {!form.isPercentage && (
                      <Select
                        value={form.currency}
                        onValueChange={(v) =>
                          v && setForm((f) => ({ ...f, currency: v }))
                        }
                      >
                        <SelectTrigger className="w-32 shrink-0">
                          <SelectValue />
                        </SelectTrigger>
                        <SelectContent>
                          <SelectItem value="RON">RON</SelectItem>
                          <SelectItem value="EUR">EUR</SelectItem>
                          <SelectItem value="USD">USD</SelectItem>
                        </SelectContent>
                      </Select>
                    )}
                  </div>
                </div>

                <div className="space-y-[7px]">
                  <Label>{form.recurring ? "Starts" : "Date"}</Label>
                  <Input
                    type="date"
                    value={form.date}
                    onChange={(e) =>
                      setForm((f) => ({ ...f, date: e.target.value }))
                    }
                  />
                </div>

                <label className="flex items-center gap-2.5 text-[13.5px] font-semibold text-[#2A241F]">
                  <input
                    type="checkbox"
                    className="size-4 accent-[#C0603A]"
                    checked={form.recurring}
                    onChange={(e) =>
                      setForm((f) => ({ ...f, recurring: e.target.checked }))
                    }
                  />
                  Recurring (monthly)
                </label>

                {form.recurring && (
                  <div className="space-y-[7px]">
                    <Label>Ends after (optional)</Label>
                    <Input
                      type="date"
                      min={form.date}
                      value={form.endDate}
                      onChange={(e) =>
                        setForm((f) => ({ ...f, endDate: e.target.value }))
                      }
                    />
                    <p className="text-xs text-[#9A9086]">
                      Last day this applies — leave empty to run indefinitely. To
                      change a rate, end this one and add a replacement starting the
                      next day.
                    </p>
                  </div>
                )}

                <Button
                  onClick={handleSubmit}
                  className="mt-1 h-[50px] w-full rounded-[13px] text-[15px] font-semibold"
                >
                  {editingId ? "Update" : "Create"}
                </Button>
              </div>
            </DialogContent>
          </Dialog>
        </div>
      </div>

      {propertyId && (
        <div className="flex flex-col gap-3.5 rounded-2xl border bg-card p-[16px_18px]">
          <div className="flex flex-wrap items-baseline gap-x-2 gap-y-1">
            <span className="font-heading text-sm font-semibold">Quick add</span>
            <span className="text-xs text-[#9A9086]">
              Tap a category — the form opens pre-filled with today's date, ready for
              the amount.
            </span>
          </div>
          <div className="flex flex-wrap gap-2.5">
            {QUICK_CATEGORIES.map((cat) => {
              const last = lastAmounts[cat]?.amount;
              return (
                <button
                  key={cat}
                  type="button"
                  onClick={() => handleQuickAdd(cat)}
                  className="flex items-center gap-2.5 rounded-xl border-[1.5px] border-[#EDE6DB] bg-[#FBF9F5] px-[15px] py-2.5 transition-colors hover:border-[#D8C4B4]"
                >
                  <span
                    className="size-2.5 shrink-0 rounded-full"
                    style={{ background: CATEGORY_COLORS[cat] }}
                  />
                  <span className="flex flex-col items-start leading-tight">
                    <span className="text-[13.5px] font-semibold text-[#2A241F]">
                      {formatCategory(cat)}
                    </span>
                    {last != null && (
                      <span className="text-[11px] text-[#9A9086]">
                        ≈ last RON {last.toFixed(2)}
                      </span>
                    )}
                  </span>
                </button>
              );
            })}
          </div>
        </div>
      )}

      {loading ? (
        <p className="text-muted-foreground">Loading...</p>
      ) : expenses.length === 0 ? (
        <p className="text-muted-foreground">
          No expenses found for this month.
        </p>
      ) : (
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>Date</TableHead>
              <TableHead>Category</TableHead>
              <TableHead>Description</TableHead>
              <TableHead>Amount</TableHead>
              <TableHead>Recurring</TableHead>
              <TableHead></TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {expenses.map((e) => (
              <TableRow key={e.id}>
                <TableCell>{formatDate(e.date)}</TableCell>
                <TableCell>
                  <Badge variant="outline">
                    {formatCategory(e.category)}
                  </Badge>
                </TableCell>
                <TableCell>{e.description}</TableCell>
                <TableCell className="font-medium">
                  {e.is_percentage
                    ? `${parseFloat(e.percentage)}%`
                    : formatAmount(e.amount, e.currency)}
                </TableCell>
                <TableCell>
                  {e.recurring
                    ? e.end_date
                      ? `Until ${formatDate(e.end_date)}`
                      : "Yes"
                    : "No"}
                </TableCell>
                <TableCell>
                  <div className="flex gap-1">
                    <Button
                      variant="ghost"
                      size="icon"
                      onClick={() => handleEdit(e)}
                    >
                      <Pencil className="h-4 w-4" />
                    </Button>
                    <Button
                      variant="ghost"
                      size="icon"
                      onClick={() => handleDuplicate(e)}
                      title="Duplicate to next month"
                    >
                      <Copy className="h-4 w-4" />
                    </Button>
                    <Button
                      variant="ghost"
                      size="icon"
                      onClick={() => handleDelete(e.id)}
                    >
                      <Trash2 className="h-4 w-4" />
                    </Button>
                  </div>
                </TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      )}
    </div>
  );
}

function formatDate(d: string): string {
  return new Date(d).toLocaleDateString("en-GB");
}

// Same day of the following month, clamped to that month's last day.
function nextMonthDate(d: string): string {
  const [y, m, day] = d.slice(0, 10).split("-").map(Number);
  const lastDay = new Date(y, m + 1, 0).getDate();
  const result = new Date(y, m, Math.min(day, lastDay));
  return `${result.getFullYear()}-${String(result.getMonth() + 1).padStart(
    2,
    "0"
  )}-${String(result.getDate()).padStart(2, "0")}`;
}

function formatCategory(c: string): string {
  return c
    .split("_")
    .map((w) => w.charAt(0).toUpperCase() + w.slice(1))
    .join(" ");
}

function formatAmount(v: string, currency = "EUR"): string {
  const n = typeof v === "string" ? parseFloat(v) : v;
  return new Intl.NumberFormat("en-EU", {
    style: "currency",
    currency,
  }).format(n);
}
