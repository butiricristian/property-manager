import { useEffect, useRef, useState, type ReactElement, type ReactNode } from "react";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
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
import { importCSV } from "@/api/client";
import type { Property } from "@/types";

interface Props {
  properties: Property[];
  /** Default property to import into. */
  propertyId: string;
  /** When provided (and there is more than one property), a target-property picker is shown. */
  setPropertyId?: (id: string) => void;
  /** Called after a successful import so the caller can refresh its data. */
  onImported?: () => void;
  /** Element rendered as the dialog trigger (e.g. a <Button /> or plain <button />). */
  triggerRender: ReactElement;
  /** Content placed inside the trigger element. */
  children: ReactNode;
}

export function ImportBookingsDialog({
  properties,
  propertyId,
  setPropertyId,
  onImported,
  triggerRender,
  children,
}: Props) {
  const [open, setOpen] = useState(false);
  const [platform, setPlatform] = useState("booking");
  const [message, setMessage] = useState("");
  const [importing, setImporting] = useState(false);
  const [altImport, setAltImport] = useState(false);
  const [targetId, setTargetId] = useState(propertyId);
  const fileRef = useRef<HTMLInputElement>(null);
  const payoutsFileRef = useRef<HTMLInputElement>(null);

  useEffect(() => {
    setTargetId(propertyId);
  }, [propertyId]);

  const showPropertyPicker = !!setPropertyId && properties.length > 1;

  const handleImport = async () => {
    const file = fileRef.current?.files?.[0];
    if (!file) {
      setMessage("Please select a file first.");
      return;
    }
    if (!targetId) {
      setMessage("Please select a property first.");
      return;
    }

    setImporting(true);
    setMessage("");
    try {
      const payoutsFile = payoutsFileRef.current?.files?.[0];
      const result = await importCSV(targetId, platform, file, payoutsFile);
      setMessage(
        `Imported: ${result.imported}, Skipped: ${result.skipped}, Errors: ${result.errors}`
      );
      onImported?.();
    } catch (e) {
      setMessage(`Error: ${e instanceof Error ? e.message : "Unknown"}`);
    } finally {
      setImporting(false);
    }
  };

  return (
    <Dialog
      open={open}
      onOpenChange={(o) => {
        setOpen(o);
        if (!o) {
          setAltImport(false);
          setMessage("");
        }
      }}
    >
      <DialogTrigger render={triggerRender}>{children}</DialogTrigger>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Import Bookings</DialogTitle>
        </DialogHeader>
        <div className="space-y-4">
          {showPropertyPicker && (
            <div>
              <Label>Property</Label>
              <Select
                value={targetId}
                onValueChange={(v) => {
                  if (!v) return;
                  setTargetId(v);
                  setPropertyId?.(v);
                }}
              >
                <SelectTrigger>
                  <SelectValue>
                    {(value: string | null) =>
                      properties.find((p) => p.id === value)?.name ?? "Select property"
                    }
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
            </div>
          )}

          <div>
            <Label>Platform</Label>
            <Select value={platform} onValueChange={(v) => v && setPlatform(v)}>
              <SelectTrigger>
                <SelectValue>
                  {(value: string | null) => {
                    if (value === "booking") return "Booking.com";
                    if (value === "airbnb") return "Airbnb";
                    return value;
                  }}
                </SelectValue>
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="booking">Booking.com</SelectItem>
                <SelectItem value="airbnb">Airbnb</SelectItem>
              </SelectContent>
            </Select>
          </div>

          <div className="space-y-2 rounded-md border bg-muted/50 p-3 text-sm text-muted-foreground">
            {platform === "booking" ? (
              altImport ? (
                <>
                  <p className="font-medium text-foreground">
                    Alternative: Reservations + Payouts
                  </p>
                  <ol className="list-inside list-decimal space-y-1">
                    <li>
                      Log in to the{" "}
                      <span className="font-medium">Booking.com Extranet</span>
                    </li>
                    <li>
                      Export <span className="font-medium">Reservations</span> from the
                      reservations page
                    </li>
                    <li>
                      Export <span className="font-medium">Payouts</span> from{" "}
                      <span className="font-medium">Finance &gt; Payments</span>
                    </li>
                  </ol>
                  <p className="text-xs">The payouts file includes city tax in the amounts.</p>
                </>
              ) : (
                <>
                  <p className="font-medium text-foreground">
                    How to export from Booking.com
                  </p>
                  <ol className="list-inside list-decimal space-y-1">
                    <li>
                      Log in to the{" "}
                      <span className="font-medium">Booking.com Extranet</span>
                    </li>
                    <li>
                      Go to{" "}
                      <span className="font-medium">
                        Finance &gt; Reservations statements
                      </span>
                    </li>
                    <li>Select the desired date range</li>
                    <li>
                      Click <span className="font-medium">Export</span> to download (CSV)
                    </li>
                  </ol>
                  <p className="text-xs">
                    Expected columns: Reservation number, Guest name, Arrival, Departure,
                    Final amount, Commission amount, Status, Currency
                  </p>
                </>
              )
            ) : (
              <>
                <p className="font-medium text-foreground">How to export from Airbnb</p>
                <ol className="list-inside list-decimal space-y-1">
                  <li>
                    Log in to <span className="font-medium">Airbnb</span> and go to your
                    hosting dashboard
                  </li>
                  <li>
                    Navigate to{" "}
                    <span className="font-medium">Earnings &gt; Transaction History</span>
                  </li>
                  <li>Set the desired date range filter</li>
                  <li>
                    Click <span className="font-medium">Download CSV</span>
                  </li>
                </ol>
                <p className="text-xs">
                  Expected columns: Confirmation Code, Guest, Start Date, End Date, Amount,
                  Host Fee, Status
                </p>
              </>
            )}
          </div>

          {platform === "booking" && altImport ? (
            <>
              <div>
                <Label>Reservations file</Label>
                <Input type="file" accept=".csv,.xls,.xlsx" ref={fileRef} />
              </div>
              <div>
                <Label>Payouts file</Label>
                <Input type="file" accept=".csv,.xls,.xlsx" ref={payoutsFileRef} />
              </div>
            </>
          ) : (
            <div>
              <Label>File</Label>
              <Input type="file" accept=".csv,.xls,.xlsx" ref={fileRef} />
            </div>
          )}

          {platform === "booking" && (
            <p className="text-xs text-muted-foreground">
              {altImport ? (
                <>
                  Want to use a single file?{" "}
                  <button
                    type="button"
                    className="underline hover:text-foreground"
                    onClick={() => setAltImport(false)}
                  >
                    Go back to reservations statement import
                  </button>
                </>
              ) : (
                <>
                  Reservations statement page is unavailable?{" "}
                  <button
                    type="button"
                    className="underline hover:text-foreground"
                    onClick={() => setAltImport(true)}
                  >
                    Click here to try an alternative
                  </button>
                </>
              )}
            </p>
          )}

          {message && <p className="text-sm text-muted-foreground">{message}</p>}

          <Button onClick={handleImport} className="w-full" disabled={importing}>
            {importing ? "Importing..." : "Import"}
          </Button>
        </div>
      </DialogContent>
    </Dialog>
  );
}
