import { useEffect, useState } from "react";
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
  CardDescription,
} from "@/components/ui/card";
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
import { getSellerProfile, upsertSellerProfile } from "@/api/client";
import { RO_COUNTIES } from "@/lib/counties";
import { useAuth } from "@/context/AuthContext";
import { UpgradeDialog } from "@/components/UpgradeDialog";

export function SellerProfilePage() {
  const { isGuest } = useAuth();
  const [upgradeOpen, setUpgradeOpen] = useState(false);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [message, setMessage] = useState("");

  const [form, setForm] = useState({
    legal_name: "",
    cui: "",
    cnp: "",
    trade_register_nr: "",
    vat_registered: false,
    street: "",
    city: "",
    county: "",
    postal_code: "",
    country_code: "RO",
  });

  useEffect(() => {
    getSellerProfile()
      .then((profile) => {
        if (profile) {
          setForm({
            legal_name: profile.legal_name,
            cui: profile.cui,
            cnp: profile.cnp,
            trade_register_nr: profile.trade_register_nr,
            vat_registered: profile.vat_registered,
            street: profile.street,
            city: profile.city,
            county: profile.county,
            postal_code: profile.postal_code,
            country_code: profile.country_code || "RO",
          });
        }
      })
      .catch(console.error)
      .finally(() => setLoading(false));
  }, []);

  const handleSave = async () => {
    setSaving(true);
    setMessage("");
    try {
      await upsertSellerProfile(form);
      setMessage("Profile saved successfully.");
    } catch (e) {
      setMessage(`Error: ${e instanceof Error ? e.message : "Unknown"}`);
    } finally {
      setSaving(false);
    }
  };

  const update = (field: string, value: string | boolean) =>
    setForm((prev) => ({ ...prev, [field]: value }));

  if (loading && !isGuest) return <p className="text-muted-foreground">Loading...</p>;

  if (isGuest) {
    return (
      <div className="space-y-6">
        <h1 className="text-3xl font-bold">Billing Profile</h1>
        <Card>
          <CardContent className="py-12 text-center">
            <p className="text-muted-foreground mb-4">
              Create an account to set up your billing profile and generate invoices.
            </p>
            <Button onClick={() => setUpgradeOpen(true)}>Create an account</Button>
          </CardContent>
        </Card>
        <UpgradeDialog open={upgradeOpen} onOpenChange={setUpgradeOpen} />
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <h1 className="text-3xl font-bold">Billing Profile</h1>
      <p className="text-muted-foreground">
        Set up your seller / billing details for e-factura invoice generation.
      </p>

      <Card>
        <CardHeader>
          <CardTitle>Legal Entity</CardTitle>
          <CardDescription>
            Your legal name and registration details (PF, PFA, or SRL).
          </CardDescription>
        </CardHeader>
        <CardContent>
          <div className="grid gap-4 sm:grid-cols-2">
            <div>
              <Label>Legal Name *</Label>
              <Input
                value={form.legal_name}
                onChange={(e) => update("legal_name", e.target.value)}
                placeholder="e.g. SC Example SRL"
              />
            </div>
            <div>
              <Label>CUI / CIF</Label>
              <Input
                value={form.cui}
                onChange={(e) => update("cui", e.target.value)}
                placeholder="e.g. 12345678"
              />
            </div>
            <div>
              <Label>CNP</Label>
              <Input
                value={form.cnp}
                onChange={(e) => update("cnp", e.target.value)}
                placeholder="For PF / PFA"
              />
            </div>
            <div>
              <Label>Trade Register Nr</Label>
              <Input
                value={form.trade_register_nr}
                onChange={(e) => update("trade_register_nr", e.target.value)}
                placeholder="e.g. J40/1234/2020"
              />
            </div>
            <div className="flex items-center gap-2 sm:col-span-2">
              <input
                type="checkbox"
                id="vat_registered"
                checked={form.vat_registered}
                onChange={(e) => update("vat_registered", e.target.checked)}
                className="h-4 w-4 rounded border"
              />
              <Label htmlFor="vat_registered">
                VAT Registered (Platitor de TVA)
              </Label>
            </div>
          </div>
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>Address</CardTitle>
          <CardDescription>
            Your registered business address.
          </CardDescription>
        </CardHeader>
        <CardContent>
          <div className="grid gap-4 sm:grid-cols-2">
            <div className="sm:col-span-2">
              <Label>Street</Label>
              <Input
                value={form.street}
                onChange={(e) => update("street", e.target.value)}
                placeholder="Street address"
              />
            </div>
            <div>
              <Label>{form.county === "B" ? "Sector" : "City"}</Label>
              {form.county === "B" ? (
                <Select
                  value={form.city}
                  onValueChange={(v) => v && update("city", v)}
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
                  value={form.city}
                  onChange={(e) => update("city", e.target.value)}
                />
              )}
            </div>
            <div>
              <Label>County</Label>
              <Select
                value={form.county}
                onValueChange={(v) => v && update("county", v)}
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
                value={form.postal_code}
                onChange={(e) => update("postal_code", e.target.value)}
              />
            </div>
            <div>
              <Label>Country Code</Label>
              <Input
                value={form.country_code}
                onChange={(e) => update("country_code", e.target.value)}
              />
            </div>
          </div>
        </CardContent>
      </Card>

      {message && (
        <p
          className={`text-sm ${message.startsWith("Error") ? "text-destructive" : "text-muted-foreground"}`}
        >
          {message}
        </p>
      )}

      <Button onClick={handleSave} disabled={saving}>
        {saving ? "Saving..." : "Save Profile"}
      </Button>
    </div>
  );
}
