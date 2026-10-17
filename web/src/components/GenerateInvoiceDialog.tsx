import { useEffect, useState } from "react";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Separator } from "@/components/ui/separator";
import {
  lookupANAF,
  generateInvoice,
  downloadInvoiceXML,
  getSellerProfile,
} from "@/api/client";
import type { Booking } from "@/types";
import { RO_COUNTIES } from "@/lib/counties";
import { Search, Loader2 } from "lucide-react";

interface Props {
  booking: Booking;
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onGenerated: () => void;
}

export function GenerateInvoiceDialog({
  booking,
  open,
  onOpenChange,
  onGenerated,
}: Props) {
  const [lookupCUI, setLookupCUI] = useState("");
  const [lookingUp, setLookingUp] = useState(false);
  const [generating, setGenerating] = useState(false);
  const [error, setError] = useState("");
  const [sellerVatRegistered, setSellerVatRegistered] = useState(false);

  const [buyer, setBuyer] = useState({
    legal_name: "",
    cui: "",
    trade_register_nr: "",
    street: "",
    city: "",
    county: "",
    postal_code: "",
    country_code: "RO",
    vat_registered: false,
  });

  useEffect(() => {
    if (open) {
      getSellerProfile()
        .then((profile) => {
          if (profile) {
            setSellerVatRegistered(profile.vat_registered);
          }
        })
        .catch(() => {});
    }
  }, [open]);

  const invoiceBase = parseFloat(booking.gross_amount) || 0;
  const taxPercent = sellerVatRegistered ? 19 : 0;
  const taxAmount = invoiceBase * taxPercent / 100;
  const invoiceTotal = invoiceBase + taxAmount;

  const updateBuyer = (field: string, value: string | boolean) =>
    setBuyer((prev) => ({ ...prev, [field]: value }));

  const handleLookup = async () => {
    if (!lookupCUI.trim()) return;
    setLookingUp(true);
    setError("");
    try {
      const company = await lookupANAF(lookupCUI.trim());
      setBuyer({
        legal_name: company.name,
        cui: company.cui,
        trade_register_nr: company.trade_register_nr,
        street: company.street,
        city: company.city,
        county: company.county,
        postal_code: company.postal_code,
        country_code: "RO",
        vat_registered: company.vat_registered,
      });
      setLookupCUI(company.cui);
    } catch (e) {
      setError(e instanceof Error ? e.message : "Lookup failed");
    } finally {
      setLookingUp(false);
    }
  };

  const handleGenerate = async () => {
    if (!buyer.legal_name) {
      setError("Buyer legal name is required");
      return;
    }
    setGenerating(true);
    setError("");
    try {
      const invoice = await generateInvoice(booking.id, {
        buyer_legal_name: buyer.legal_name,
        buyer_cui: buyer.cui,
        buyer_trade_register_nr: buyer.trade_register_nr,
        buyer_street: buyer.street,
        buyer_city: buyer.city,
        buyer_county: buyer.county,
        buyer_postal_code: buyer.postal_code,
        buyer_country_code: buyer.country_code,
        buyer_vat_registered: buyer.vat_registered,
        net_amount: invoiceBase,
        currency: booking.currency,
      });
      await downloadInvoiceXML(invoice.id);
      onGenerated();
      onOpenChange(false);
    } catch (e) {
      setError(e instanceof Error ? e.message : "Generation failed");
    } finally {
      setGenerating(false);
    }
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-lg max-h-[90vh] overflow-y-auto">
        <DialogHeader>
          <DialogTitle>Generate Invoice</DialogTitle>
        </DialogHeader>

        <div className="space-y-4">
          {/* Booking info */}
          <div className="rounded-md border bg-muted/50 p-3 text-sm space-y-1">
            <p>
              <span className="font-medium">Guest:</span> {booking.guest_name}
            </p>
            <p>
              <span className="font-medium">Period:</span>{" "}
              {new Date(booking.check_in).toLocaleDateString("en-GB")} -{" "}
              {new Date(booking.check_out).toLocaleDateString("en-GB")} (
              {booking.nights} nights)
            </p>
            <p>
              <span className="font-medium">Gross Amount:</span>{" "}
              {formatAmount(invoiceBase, booking.currency)}
            </p>
          </div>

          <Separator />

          {/* ANAF Lookup */}
          <div>
            <Label>Buyer CUI Lookup</Label>
            <div className="flex gap-2 mt-1">
              <Input
                value={lookupCUI}
                onChange={(e) => setLookupCUI(e.target.value)}
                placeholder="e.g. RO12345678"
                onKeyDown={(e) => e.key === "Enter" && handleLookup()}
              />
              <Button
                variant="outline"
                onClick={handleLookup}
                disabled={lookingUp}
              >
                {lookingUp ? (
                  <Loader2 className="h-4 w-4 animate-spin" />
                ) : (
                  <Search className="h-4 w-4" />
                )}
              </Button>
            </div>
          </div>

          <Separator />

          {/* Buyer details */}
          <div className="grid gap-3 sm:grid-cols-2">
            <div className="sm:col-span-2">
              <Label>Legal Name *</Label>
              <Input
                value={buyer.legal_name}
                onChange={(e) => updateBuyer("legal_name", e.target.value)}
              />
            </div>
            <div>
              <Label>CUI</Label>
              <Input
                value={buyer.cui}
                onChange={(e) => updateBuyer("cui", e.target.value)}
              />
            </div>
            <div>
              <Label>Trade Register Nr</Label>
              <Input
                value={buyer.trade_register_nr}
                onChange={(e) =>
                  updateBuyer("trade_register_nr", e.target.value)
                }
              />
            </div>
            <div className="sm:col-span-2">
              <Label>Street</Label>
              <Input
                value={buyer.street}
                onChange={(e) => updateBuyer("street", e.target.value)}
              />
            </div>
            <div>
              <Label>{buyer.county === "B" ? "Sector" : "City"}</Label>
              {buyer.county === "B" ? (
                <Select
                  value={buyer.city}
                  onValueChange={(v) => v && updateBuyer("city", v)}
                >
                  <SelectTrigger>
                    <SelectValue placeholder="Select sector">
                      {(value: string | null) => value || "Select sector"}
                    </SelectValue>
                  </SelectTrigger>
                  <SelectContent>
                    {["SECTOR1", "SECTOR2", "SECTOR3", "SECTOR4", "SECTOR5", "SECTOR6"].map((s) => (
                      <SelectItem key={s} value={s}>
                        {s}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              ) : (
                <Input
                  value={buyer.city}
                  onChange={(e) => updateBuyer("city", e.target.value)}
                />
              )}
            </div>
            <div>
              <Label>County</Label>
              <Select
                value={buyer.county}
                onValueChange={(v) => v && updateBuyer("county", v)}
              >
                <SelectTrigger>
                  <SelectValue placeholder="Select county">
                    {(value: string | null) => {
                      if (!value) return "Select county";
                      const c = RO_COUNTIES.find((c) => c.code === value);
                      return c ? `${c.name} (${c.code})` : value;
                    }}
                  </SelectValue>
                </SelectTrigger>
                <SelectContent>
                  {RO_COUNTIES.map((c) => (
                    <SelectItem key={c.code} value={c.code}>
                      {c.name} ({c.code})
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
            <div>
              <Label>Postal Code</Label>
              <Input
                value={buyer.postal_code}
                onChange={(e) => updateBuyer("postal_code", e.target.value)}
              />
            </div>
            <div className="flex items-center gap-2">
              <input
                type="checkbox"
                id="buyer_vat"
                checked={buyer.vat_registered}
                onChange={(e) =>
                  updateBuyer("vat_registered", e.target.checked)
                }
                className="h-4 w-4 rounded border"
              />
              <Label htmlFor="buyer_vat">VAT Registered</Label>
            </div>
          </div>

          <Separator />

          {/* Summary */}
          <div className="rounded-md border p-3 text-sm space-y-1">
            <p className="font-medium">Invoice Summary</p>
            <div className="flex justify-between">
              <span>Base (Gross Amount):</span>
              <span>{formatAmount(invoiceBase, booking.currency)}</span>
            </div>
            <div className="flex justify-between">
              <span>Tax ({taxPercent}%):</span>
              <span>{formatAmount(taxAmount, booking.currency)}</span>
            </div>
            <div className="flex justify-between font-semibold">
              <span>Total:</span>
              <span>{formatAmount(invoiceTotal, booking.currency)}</span>
            </div>
            <p className="text-xs text-muted-foreground mt-1">
              {sellerVatRegistered
                ? "19% TVA applied (VAT registered seller)."
                : "No TVA applied (seller not VAT registered)."}
            </p>
          </div>

          {error && <p className="text-sm text-destructive">{error}</p>}

          <Button
            className="w-full"
            onClick={handleGenerate}
            disabled={generating}
          >
            {generating ? "Generating..." : "Generate Invoice"}
          </Button>
        </div>
      </DialogContent>
    </Dialog>
  );
}

function formatAmount(v: number, currency: string): string {
  return new Intl.NumberFormat("en-EU", {
    style: "currency",
    currency,
  }).format(v);
}
