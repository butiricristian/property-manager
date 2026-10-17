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
  listInventory,
  createInventoryItem,
  updateInventoryItem,
  deleteInventoryItem,
} from "@/api/client";
import { useProperty } from "@/context/PropertyContext";
import type { InventoryItem, InventoryCategory } from "@/types";
import { Plus, Pencil, Trash2 } from "lucide-react";

const CATEGORIES: InventoryCategory[] = [
  "linens",
  "kitchenware",
  "furniture",
  "electronics",
  "toiletries",
  "appliances",
  "other",
];

interface InventoryForm {
  name: string;
  category: InventoryCategory;
  quantity: string;
  acquired_date: string;
  lifetime_months: string;
  notes: string;
}

const emptyForm: InventoryForm = {
  name: "",
  category: "linens",
  quantity: "1",
  acquired_date: new Date().toISOString().slice(0, 10),
  lifetime_months: "12",
  notes: "",
};

function getStatus(acquiredDate: string, lifetimeMonths: number): { label: string; variant: "default" | "secondary" | "destructive" } {
  const acquired = new Date(acquiredDate);
  const replaceBy = new Date(acquired);
  replaceBy.setMonth(replaceBy.getMonth() + lifetimeMonths);
  const now = new Date();
  const diffMs = replaceBy.getTime() - now.getTime();
  const diffMonths = diffMs / (1000 * 60 * 60 * 24 * 30.44);

  if (diffMonths < 0) return { label: "Overdue", variant: "destructive" };
  if (diffMonths <= 3) return { label: "Replace Soon", variant: "secondary" };
  return { label: "Good", variant: "default" };
}

function getReplaceByDate(acquiredDate: string, lifetimeMonths: number): string {
  const acquired = new Date(acquiredDate);
  acquired.setMonth(acquired.getMonth() + lifetimeMonths);
  return acquired.toLocaleDateString("en-GB");
}

export function InventoryPage() {
  const { properties, propertyId, setPropertyId } = useProperty();
  const [items, setItems] = useState<InventoryItem[]>([]);
  const [loading, setLoading] = useState(false);
  const [formOpen, setFormOpen] = useState(false);
  const [editingId, setEditingId] = useState<string | null>(null);
  const [form, setForm] = useState<InventoryForm>(emptyForm);

  useEffect(() => {
    if (!propertyId) return;
    setLoading(true);
    listInventory(propertyId)
      .then(setItems)
      .catch(console.error)
      .finally(() => setLoading(false));
  }, [propertyId]);

  const refresh = () => {
    if (!propertyId) return;
    listInventory(propertyId).then(setItems).catch(console.error);
  };

  const handleSubmit = async () => {
    if (!propertyId) return;
    const data = {
      name: form.name,
      category: form.category,
      quantity: parseInt(form.quantity) || 1,
      acquired_date: form.acquired_date,
      lifetime_months: parseInt(form.lifetime_months) || 12,
      notes: form.notes,
    };

    if (editingId) {
      await updateInventoryItem(editingId, data as never);
    } else {
      await createInventoryItem(propertyId, data as never);
    }
    setFormOpen(false);
    setEditingId(null);
    setForm(emptyForm);
    refresh();
  };

  const handleEdit = (item: InventoryItem) => {
    setEditingId(item.id);
    setForm({
      name: item.name,
      category: item.category,
      quantity: String(item.quantity),
      acquired_date: typeof item.acquired_date === "string" ? item.acquired_date.slice(0, 10) : item.acquired_date,
      lifetime_months: String(item.lifetime_months),
      notes: item.notes,
    });
    setFormOpen(true);
  };

  const handleDelete = async (id: string) => {
    if (!confirm("Delete this inventory item?")) return;
    await deleteInventoryItem(id);
    setItems((prev) => prev.filter((i) => i.id !== id));
  };

  const openNew = () => {
    setEditingId(null);
    setForm(emptyForm);
    setFormOpen(true);
  };

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-3xl font-bold">Inventory</h1>
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
              Add Item
            </DialogTrigger>
            <DialogContent>
              <DialogHeader>
                <DialogTitle>
                  {editingId ? "Edit Inventory Item" : "Add Inventory Item"}
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
                    placeholder="e.g. Bath towels"
                  />
                </div>
                <div>
                  <Label>Category</Label>
                  <Select
                    value={form.category}
                    onValueChange={(v) =>
                      v && setForm((f) => ({
                        ...f,
                        category: v as InventoryCategory,
                      }))
                    }
                  >
                    <SelectTrigger>
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
                <div className="grid grid-cols-2 gap-2">
                  <div>
                    <Label>Quantity</Label>
                    <Input
                      type="number"
                      min="1"
                      value={form.quantity}
                      onChange={(e) =>
                        setForm((f) => ({ ...f, quantity: e.target.value }))
                      }
                    />
                  </div>
                  <div>
                    <Label>Lifetime (months)</Label>
                    <Input
                      type="number"
                      min="1"
                      value={form.lifetime_months}
                      onChange={(e) =>
                        setForm((f) => ({ ...f, lifetime_months: e.target.value }))
                      }
                    />
                  </div>
                </div>
                <div>
                  <Label>Acquired Date</Label>
                  <Input
                    type="date"
                    value={form.acquired_date}
                    onChange={(e) =>
                      setForm((f) => ({ ...f, acquired_date: e.target.value }))
                    }
                  />
                </div>
                <div>
                  <Label>Notes</Label>
                  <Input
                    value={form.notes}
                    onChange={(e) =>
                      setForm((f) => ({ ...f, notes: e.target.value }))
                    }
                    placeholder="Optional notes"
                  />
                </div>
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
      ) : items.length === 0 ? (
        <p className="text-muted-foreground">
          No inventory items found for this property.
        </p>
      ) : (
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>Name</TableHead>
              <TableHead>Category</TableHead>
              <TableHead>Qty</TableHead>
              <TableHead>Acquired</TableHead>
              <TableHead>Replace By</TableHead>
              <TableHead>Status</TableHead>
              <TableHead></TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {items.map((item) => {
              const status = getStatus(item.acquired_date, item.lifetime_months);
              return (
                <TableRow key={item.id}>
                  <TableCell className="font-medium">{item.name}</TableCell>
                  <TableCell>
                    <Badge variant="outline">
                      {formatCategory(item.category)}
                    </Badge>
                  </TableCell>
                  <TableCell>{item.quantity}</TableCell>
                  <TableCell>{formatDate(item.acquired_date)}</TableCell>
                  <TableCell>{getReplaceByDate(item.acquired_date, item.lifetime_months)}</TableCell>
                  <TableCell>
                    <Badge variant={status.variant}>
                      {status.label}
                    </Badge>
                  </TableCell>
                  <TableCell>
                    <div className="flex gap-1">
                      <Button
                        variant="ghost"
                        size="icon"
                        onClick={() => handleEdit(item)}
                      >
                        <Pencil className="h-4 w-4" />
                      </Button>
                      <Button
                        variant="ghost"
                        size="icon"
                        onClick={() => handleDelete(item.id)}
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
      )}
    </div>
  );
}

function formatDate(d: string): string {
  return new Date(d).toLocaleDateString("en-GB");
}

function formatCategory(c: string): string {
  return c
    .split("_")
    .map((w) => w.charAt(0).toUpperCase() + w.slice(1))
    .join(" ");
}
