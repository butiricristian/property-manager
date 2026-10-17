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
  listTaxRules,
  createTaxRule,
  updateTaxRule,
  deleteTaxRule,
} from "@/api/client";
import { useProperty } from "@/context/PropertyContext";
import type { TaxRule } from "@/types";
import { Plus, Pencil, Trash2 } from "lucide-react";

interface TaxRuleForm {
  name: string;
  type: "percentage" | "fixed";
  percentage: string;
  fixed_amount: string;
  currency: string;
}

const emptyForm: TaxRuleForm = {
  name: "",
  type: "percentage",
  percentage: "",
  fixed_amount: "",
  currency: "EUR",
};

export function TaxRulesPage() {
  const { properties, propertyId, setPropertyId } = useProperty();
  const [rules, setRules] = useState<TaxRule[]>([]);
  const [loading, setLoading] = useState(false);
  const [formOpen, setFormOpen] = useState(false);
  const [editingId, setEditingId] = useState<string | null>(null);
  const [form, setForm] = useState<TaxRuleForm>(emptyForm);

  useEffect(() => {
    if (!propertyId) return;
    setLoading(true);
    listTaxRules(propertyId)
      .then(setRules)
      .catch(console.error)
      .finally(() => setLoading(false));
  }, [propertyId]);

  const refresh = () => {
    if (!propertyId) return;
    listTaxRules(propertyId).then(setRules).catch(console.error);
  };

  const handleSubmit = async () => {
    if (!propertyId) return;
    const data = {
      name: form.name,
      type: form.type,
      percentage: form.type === "percentage" ? parseFloat(form.percentage) || 0 : 0,
      fixed_amount: form.type === "fixed" ? parseFloat(form.fixed_amount) || 0 : 0,
      currency: form.currency,
    };

    if (editingId) {
      await updateTaxRule(editingId, data as never);
    } else {
      await createTaxRule(propertyId, data as never);
    }
    setFormOpen(false);
    setEditingId(null);
    setForm(emptyForm);
    refresh();
  };

  const handleEdit = (rule: TaxRule) => {
    setEditingId(rule.id);
    setForm({
      name: rule.name,
      type: rule.type,
      percentage: String(rule.percentage),
      fixed_amount: String(rule.fixed_amount),
      currency: rule.currency,
    });
    setFormOpen(true);
  };

  const handleDelete = async (id: string) => {
    if (!confirm("Delete this tax rule?")) return;
    await deleteTaxRule(id);
    setRules((prev) => prev.filter((r) => r.id !== id));
  };

  const openNew = () => {
    setEditingId(null);
    setForm(emptyForm);
    setFormOpen(true);
  };

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-3xl font-bold">Tax Rules</h1>
        <div className="flex gap-3">
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
          <Dialog open={formOpen} onOpenChange={setFormOpen}>
            <DialogTrigger
              render={<Button />}
              onClick={openNew}
            >
              <Plus className="mr-2 h-4 w-4" />
              Add Tax Rule
            </DialogTrigger>
            <DialogContent>
              <DialogHeader>
                <DialogTitle>
                  {editingId ? "Edit Tax Rule" : "Add Tax Rule"}
                </DialogTitle>
              </DialogHeader>
              <div className="space-y-4">
                <div>
                  <Label>Name</Label>
                  <Input
                    value={form.name}
                    onChange={(e) =>
                      setForm((f) => ({ ...f, name: e.target.value }))
                    }
                    placeholder="e.g. Income Tax, Tourism Tax"
                  />
                </div>
                <div>
                  <Label>Type</Label>
                  <Select
                    value={form.type}
                    onValueChange={(v) =>
                      v && setForm((f) => ({
                        ...f,
                        type: v as "percentage" | "fixed",
                      }))
                    }
                  >
                    <SelectTrigger>
                      <SelectValue>
                        {(value: string | null) => {
                          if (value === "percentage") return "Percentage of Profit";
                          if (value === "fixed") return "Fixed Annual Amount";
                          return value;
                        }}
                      </SelectValue>
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value="percentage">Percentage of Profit</SelectItem>
                      <SelectItem value="fixed">Fixed Annual Amount</SelectItem>
                    </SelectContent>
                  </Select>
                </div>
                {form.type === "percentage" ? (
                  <div>
                    <Label>Percentage (%)</Label>
                    <Input
                      type="number"
                      step="0.01"
                      min="0"
                      max="100"
                      value={form.percentage}
                      onChange={(e) =>
                        setForm((f) => ({ ...f, percentage: e.target.value }))
                      }
                    />
                  </div>
                ) : (
                  <div className="grid grid-cols-2 gap-2">
                    <div>
                      <Label>Annual Amount</Label>
                      <Input
                        type="number"
                        step="0.01"
                        value={form.fixed_amount}
                        onChange={(e) =>
                          setForm((f) => ({ ...f, fixed_amount: e.target.value }))
                        }
                      />
                    </div>
                    <div>
                      <Label>Currency</Label>
                      <Select
                        value={form.currency}
                        onValueChange={(v) =>
                          v && setForm((f) => ({ ...f, currency: v }))
                        }
                      >
                        <SelectTrigger>
                          <SelectValue />
                        </SelectTrigger>
                        <SelectContent>
                          <SelectItem value="EUR">EUR</SelectItem>
                          <SelectItem value="USD">USD</SelectItem>
                          <SelectItem value="GBP">GBP</SelectItem>
                          <SelectItem value="RON">RON</SelectItem>
                        </SelectContent>
                      </Select>
                    </div>
                  </div>
                )}
                <Button onClick={handleSubmit} className="w-full">
                  {editingId ? "Update" : "Create"}
                </Button>
              </div>
            </DialogContent>
          </Dialog>
        </div>
      </div>

      {loading ? (
        <p className="text-muted-foreground">Loading...</p>
      ) : rules.length === 0 ? (
        <p className="text-muted-foreground">
          No tax rules configured for this property.
        </p>
      ) : (
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>Name</TableHead>
              <TableHead>Type</TableHead>
              <TableHead>Value</TableHead>
              <TableHead></TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {rules.map((r) => (
              <TableRow key={r.id}>
                <TableCell className="font-medium">{r.name}</TableCell>
                <TableCell>
                  <Badge variant="outline">
                    {r.type === "percentage" ? "Percentage" : "Fixed"}
                  </Badge>
                </TableCell>
                <TableCell>
                  {r.type === "percentage"
                    ? `${r.percentage}%`
                    : `${formatAmount(r.fixed_amount, r.currency)}/year`}
                </TableCell>
                <TableCell>
                  <div className="flex gap-1">
                    <Button
                      variant="ghost"
                      size="icon"
                      onClick={() => handleEdit(r)}
                    >
                      <Pencil className="h-4 w-4" />
                    </Button>
                    <Button
                      variant="ghost"
                      size="icon"
                      onClick={() => handleDelete(r.id)}
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

function formatAmount(v: number, currency = "EUR"): string {
  return new Intl.NumberFormat("en-EU", {
    style: "currency",
    currency,
  }).format(v);
}
