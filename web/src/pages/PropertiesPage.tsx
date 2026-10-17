import { useEffect, useState } from "react";
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
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
import {
  listProperties,
  createProperty,
  updateProperty,
  deleteProperty,
  listPropertyMappings,
  createPropertyMapping,
  deletePropertyMapping,
  listIntegrations,
  listPropertyShares,
  addPropertyShare,
  removePropertyShare,
} from "@/api/client";
import { useAuth } from "@/context/AuthContext";
import type {
  Property,
  PropertyMapping,
  Integration,
  PropertyShare,
} from "@/types";
import {
  Plus,
  Pencil,
  Trash2,
  Building2,
  ChevronDown,
  ChevronUp,
  Link2,
  X,
  Users,
} from "lucide-react";

interface PropertyForm {
  name: string;
  address: string;
}

const emptyForm: PropertyForm = { name: "", address: "" };

const platformNames: Record<string, string> = {
  booking: "Booking.com",
  airbnb: "Airbnb",
};

export function PropertiesPage() {
  const { isGuest } = useAuth();
  const [properties, setProperties] = useState<Property[]>([]);
  const [loading, setLoading] = useState(true);
  const [formOpen, setFormOpen] = useState(false);
  const [editingId, setEditingId] = useState<string | null>(null);
  const [form, setForm] = useState<PropertyForm>(emptyForm);

  // Mapping state
  const [expandedMappings, setExpandedMappings] = useState<
    Record<string, boolean>
  >({});
  const [mappings, setMappings] = useState<Record<string, PropertyMapping[]>>(
    {}
  );
  const [integrations, setIntegrations] = useState<Integration[]>([]);
  const [linkForm, setLinkForm] = useState<{
    propertyId: string;
    integrationId: string;
    externalId: string;
  } | null>(null);

  // Share state (applies to the property currently being edited)
  const [shares, setShares] = useState<PropertyShare[]>([]);
  const [shareEmail, setShareEmail] = useState("");
  const [shareError, setShareError] = useState<string | null>(null);

  const refresh = () => {
    setLoading(true);
    listProperties()
      .then(setProperties)
      .catch(console.error)
      .finally(() => setLoading(false));
  };

  useEffect(() => {
    refresh();
    listIntegrations().then(setIntegrations).catch(console.error);
  }, []);

  const loadMappings = (propertyId: string) => {
    listPropertyMappings(propertyId)
      .then((m) => setMappings((prev) => ({ ...prev, [propertyId]: m })))
      .catch(console.error);
  };

  const toggleMappings = (propertyId: string) => {
    const isExpanding = !expandedMappings[propertyId];
    setExpandedMappings((prev) => ({
      ...prev,
      [propertyId]: isExpanding,
    }));
    if (isExpanding && !mappings[propertyId]) {
      loadMappings(propertyId);
    }
  };

  const handleSubmit = async () => {
    if (!form.name.trim()) return;

    if (editingId) {
      await updateProperty(editingId, form);
    } else {
      await createProperty(form);
    }
    setFormOpen(false);
    setEditingId(null);
    setForm(emptyForm);
    refresh();
  };

  const loadShares = (propertyId: string) => {
    listPropertyShares(propertyId)
      .then(setShares)
      .catch((e) => {
        console.error(e);
        setShares([]);
      });
  };

  const handleEdit = (property: Property) => {
    setEditingId(property.id);
    setForm({ name: property.name, address: property.address });
    setShares([]);
    setShareEmail("");
    setShareError(null);
    setFormOpen(true);
    if (property.permission === "admin" || !property.permission) {
      loadShares(property.id);
    }
  };

  const handleAddShare = async () => {
    if (!editingId || !shareEmail.trim()) return;
    setShareError(null);
    try {
      await addPropertyShare(editingId, shareEmail.trim());
      setShareEmail("");
      loadShares(editingId);
    } catch (e) {
      setShareError(
        e instanceof Error ? e.message : "Failed to share property"
      );
    }
  };

  const handleRemoveShare = async (userId: string) => {
    if (!editingId) return;
    try {
      await removePropertyShare(editingId, userId);
      loadShares(editingId);
    } catch (e) {
      console.error(e);
    }
  };

  const handleDelete = async (id: string) => {
    if (
      !confirm(
        "Delete this property? This will also delete all associated bookings and expenses."
      )
    )
      return;
    await deleteProperty(id);
    refresh();
  };

  const openNew = () => {
    setEditingId(null);
    setForm(emptyForm);
    setFormOpen(true);
  };

  const handleCreateMapping = async (propertyId: string) => {
    if (!linkForm || !linkForm.integrationId || !linkForm.externalId.trim())
      return;
    try {
      await createPropertyMapping(propertyId, {
        integration_id: linkForm.integrationId,
        external_id: linkForm.externalId.trim(),
      });
      setLinkForm(null);
      loadMappings(propertyId);
    } catch (e) {
      console.error(e);
    }
  };

  const handleDeleteMapping = async (
    propertyId: string,
    mappingId: string
  ) => {
    if (!confirm("Remove this platform link?")) return;
    try {
      await deletePropertyMapping(propertyId, mappingId);
      loadMappings(propertyId);
    } catch (e) {
      console.error(e);
    }
  };

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-3xl font-bold">Properties</h1>
        <Dialog open={formOpen} onOpenChange={setFormOpen}>
          <DialogTrigger
            render={<Button />}
            onClick={openNew}
            disabled={isGuest && properties.length >= 1}
            title={isGuest && properties.length >= 1 ? "Create an account to add more properties" : undefined}
          >
            <Plus className="mr-2 h-4 w-4" />
            Add Property
          </DialogTrigger>
          <DialogContent>
            <DialogHeader>
              <DialogTitle>
                {editingId ? "Edit Property" : "Add Property"}
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
                  placeholder="e.g. Beach Apartment"
                />
              </div>
              <div>
                <Label>Address</Label>
                <Input
                  value={form.address}
                  onChange={(e) =>
                    setForm((f) => ({ ...f, address: e.target.value }))
                  }
                  placeholder="e.g. 123 Ocean Drive"
                />
              </div>
              <Button onClick={handleSubmit} className="w-full">
                {editingId ? "Update" : "Create"}
              </Button>

              {editingId && (
                <div className="border-t pt-4 space-y-3">
                  <div className="flex items-center gap-2">
                    <Users className="h-4 w-4" />
                    <h3 className="text-sm font-medium">Shared with</h3>
                  </div>
                  {shares.length === 0 ? (
                    <p className="text-xs text-muted-foreground">
                      Not shared with anyone yet.
                    </p>
                  ) : (
                    <div className="space-y-2">
                      {shares.map((s) => (
                        <div
                          key={s.user_id}
                          className="flex items-center justify-between rounded-md border bg-muted/50 px-3 py-2 text-sm"
                        >
                          <div className="flex flex-col">
                            <span className="font-medium">
                              {s.name || s.email}
                            </span>
                            {s.name && (
                              <span className="text-xs text-muted-foreground">
                                {s.email}
                              </span>
                            )}
                          </div>
                          <Button
                            variant="ghost"
                            size="icon"
                            className="h-6 w-6"
                            onClick={() => handleRemoveShare(s.user_id)}
                          >
                            <Trash2 className="h-3.5 w-3.5" />
                          </Button>
                        </div>
                      ))}
                    </div>
                  )}
                  <div className="space-y-2">
                    <Label className="text-xs">Invite by email</Label>
                    <div className="flex gap-2">
                      <Input
                        type="email"
                        placeholder="user@example.com"
                        value={shareEmail}
                        onChange={(e) => setShareEmail(e.target.value)}
                        disabled={isGuest}
                      />
                      <Button
                        onClick={handleAddShare}
                        disabled={isGuest || !shareEmail.trim()}
                      >
                        Invite
                      </Button>
                    </div>
                    {isGuest && (
                      <p className="text-xs text-muted-foreground">
                        Create an account to share properties.
                      </p>
                    )}
                    {shareError && (
                      <p className="text-xs text-destructive">{shareError}</p>
                    )}
                  </div>
                </div>
              )}
            </div>
          </DialogContent>
        </Dialog>
      </div>

      {loading ? (
        <p className="text-muted-foreground">Loading...</p>
      ) : properties.length === 0 ? (
        <div className="text-center py-12">
          <Building2 className="mx-auto h-12 w-12 text-muted-foreground" />
          <h3 className="mt-4 text-lg font-semibold">No properties yet</h3>
          <p className="text-muted-foreground">
            Add your first property to get started.
          </p>
        </div>
      ) : (
        <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
          {properties.map((p) => {
            const isOwner = !p.permission || p.permission === "admin";
            return (
            <Card key={p.id}>
              <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                <div className="flex items-center gap-2">
                  <CardTitle className="text-lg font-medium">
                    {p.name}
                  </CardTitle>
                  {!isOwner && (
                    <span className="inline-flex items-center gap-1 rounded-full bg-muted px-2 py-0.5 text-xs font-medium text-muted-foreground">
                      <Users className="h-3 w-3" />
                      Shared
                    </span>
                  )}
                </div>
                {isOwner && (
                  <div className="flex gap-1">
                    <Button
                      variant="ghost"
                      size="icon"
                      onClick={() => handleEdit(p)}
                    >
                      <Pencil className="h-4 w-4" />
                    </Button>
                    <Button
                      variant="ghost"
                      size="icon"
                      onClick={() => handleDelete(p.id)}
                    >
                      <Trash2 className="h-4 w-4" />
                    </Button>
                  </div>
                )}
              </CardHeader>
              <CardContent className="space-y-3">
                <p className="text-sm text-muted-foreground">
                  {p.address || "No address"}
                </p>

                {/* Platform Mappings */}
                <div className="border-t pt-3">
                  <button
                    type="button"
                    className="flex w-full items-center gap-2 text-sm font-medium text-muted-foreground hover:text-foreground transition-colors"
                    onClick={() => toggleMappings(p.id)}
                  >
                    <Link2 className="h-4 w-4" />
                    <span>Platform Mappings</span>
                    {expandedMappings[p.id] ? (
                      <ChevronUp className="ml-auto h-4 w-4" />
                    ) : (
                      <ChevronDown className="ml-auto h-4 w-4" />
                    )}
                  </button>

                  {expandedMappings[p.id] && (
                    <div className="mt-3 space-y-3">
                      {/* Existing mappings */}
                      {(mappings[p.id] || []).length > 0 ? (
                        <div className="space-y-2">
                          {mappings[p.id].map((m) => (
                            <div
                              key={m.id}
                              className="flex items-center justify-between rounded-md border bg-muted/50 px-3 py-2 text-sm"
                            >
                              <div className="flex items-center gap-2">
                                <Link2 className="h-3.5 w-3.5 text-muted-foreground" />
                                <span className="font-medium">
                                  {platformNames[m.platform] || m.platform}
                                </span>
                                <span className="text-muted-foreground">
                                  {m.external_id}
                                </span>
                              </div>
                              <Button
                                variant="ghost"
                                size="icon"
                                className="h-6 w-6"
                                onClick={() =>
                                  handleDeleteMapping(p.id, m.id)
                                }
                              >
                                <X className="h-3.5 w-3.5" />
                              </Button>
                            </div>
                          ))}
                        </div>
                      ) : (
                        <p className="text-xs text-muted-foreground">
                          No platforms linked yet.
                        </p>
                      )}

                      {/* Add mapping form */}
                      {integrations.length > 0 && (
                        <>
                          {linkForm?.propertyId === p.id ? (
                            <div className="space-y-2 rounded-md border p-3">
                              <div>
                                <Label className="text-xs">Integration</Label>
                                <Select
                                  value={linkForm.integrationId}
                                  onValueChange={(val) =>
                                    setLinkForm((f) =>
                                      f
                                        ? {
                                            ...f,
                                            integrationId: val ?? "",
                                          }
                                        : null
                                    )
                                  }
                                >
                                  <SelectTrigger className="w-full">
                                    <SelectValue>
                                      {(value: string | null) => {
                                        if (!value) return "Select integration";
                                        const integ = integrations.find(
                                          (i) => i.id === value
                                        );
                                        return integ
                                          ? platformNames[integ.platform] ||
                                              integ.platform
                                          : value;
                                      }}
                                    </SelectValue>
                                  </SelectTrigger>
                                  <SelectContent>
                                    {integrations.map((integ) => (
                                      <SelectItem
                                        key={integ.id}
                                        value={integ.id}
                                      >
                                        {platformNames[integ.platform] ||
                                          integ.platform}
                                      </SelectItem>
                                    ))}
                                  </SelectContent>
                                </Select>
                              </div>
                              <div>
                                <Label className="text-xs">
                                  External ID (e.g. Hotel ID)
                                </Label>
                                <Input
                                  value={linkForm.externalId}
                                  onChange={(e) =>
                                    setLinkForm((f) =>
                                      f
                                        ? {
                                            ...f,
                                            externalId: e.target.value,
                                          }
                                        : null
                                    )
                                  }
                                  placeholder="e.g. 123456"
                                />
                              </div>
                              <div className="flex gap-2">
                                <Button
                                  size="sm"
                                  onClick={() => handleCreateMapping(p.id)}
                                  disabled={
                                    !linkForm.integrationId ||
                                    !linkForm.externalId.trim()
                                  }
                                >
                                  Save
                                </Button>
                                <Button
                                  size="sm"
                                  variant="outline"
                                  onClick={() => setLinkForm(null)}
                                >
                                  Cancel
                                </Button>
                              </div>
                            </div>
                          ) : (
                            <Button
                              size="sm"
                              variant="outline"
                              className="w-full"
                              onClick={() =>
                                setLinkForm({
                                  propertyId: p.id,
                                  integrationId: "",
                                  externalId: "",
                                })
                              }
                            >
                              <Plus className="mr-2 h-3.5 w-3.5" />
                              Link Platform
                            </Button>
                          )}
                        </>
                      )}
                    </div>
                  )}
                </div>
              </CardContent>
            </Card>
            );
          })}
        </div>
      )}
    </div>
  );
}
