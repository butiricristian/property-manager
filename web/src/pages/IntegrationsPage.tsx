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
} from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Badge } from "@/components/ui/badge";
import {
  listIntegrations,
  createIntegration,
  updateIntegration,
  deleteIntegration,
  syncIntegration,
} from "@/api/client";
import type { Integration, SyncResult } from "@/types";
import { RefreshCw, Trash2, Pencil, Link2, HelpCircle, ChevronDown, ChevronUp } from "lucide-react";

interface PlatformConfig {
  key: "booking" | "airbnb";
  name: string;
  fields: { key: string; label: string; type: string }[];
  help: { steps: string[]; note: string };
}

const platforms: PlatformConfig[] = [
  {
    key: "booking",
    name: "Booking.com",
    fields: [
      { key: "client_id", label: "Client ID", type: "text" },
      { key: "client_secret", label: "Client Secret", type: "password" },
    ],
    help: {
      steps: [
        "Log in to the Booking.com Connectivity Partner portal.",
        "Navigate to your machine account settings to find your Client ID.",
        "Generate or retrieve your Client Secret from the same page.",
        "After connecting, link each property to its Hotel ID in the Properties page.",
      ],
      note: "Booking.com now uses token-based authentication. You need a machine account with client_id and client_secret from the Connectivity Partner portal.",
    },
  },
  {
    key: "airbnb",
    name: "Airbnb",
    fields: [
      { key: "client_id", label: "Client ID", type: "text" },
      { key: "client_secret", label: "Client Secret", type: "password" },
    ],
    help: {
      steps: [
        "Go to the Airbnb Developer Portal at developer.withairbnb.com and log in.",
        "Navigate to your application in the App Management section.",
        "Your Client ID (key) is displayed on the application details page.",
        "Your Client Secret was shown once when you first created the application. If you've lost it, you'll need to regenerate it from the app settings.",
      ],
      note: "Airbnb API access is granted only to approved partners. If you don't have a developer account, contact Airbnb's global support team to request access.",
    },
  },
];

export function IntegrationsPage() {
  const [integrations, setIntegrations] = useState<Integration[]>([]);
  const [loading, setLoading] = useState(true);
  const [syncing, setSyncing] = useState<string | null>(null);
  const [syncResult, setSyncResult] = useState<SyncResult | null>(null);
  const [connectForm, setConnectForm] = useState<Record<string, string>>({});
  const [connectingPlatform, setConnectingPlatform] = useState<string | null>(
    null
  );
  const [editDialogOpen, setEditDialogOpen] = useState(false);
  const [editingIntegration, setEditingIntegration] =
    useState<Integration | null>(null);
  const [editForm, setEditForm] = useState<Record<string, string>>({});
  const [showHelp, setShowHelp] = useState<Record<string, boolean>>({});

  const refresh = () => {
    setLoading(true);
    listIntegrations()
      .then(setIntegrations)
      .catch(console.error)
      .finally(() => setLoading(false));
  };

  useEffect(() => {
    refresh();
  }, []);

  const getIntegration = (platform: string) =>
    integrations.find((i) => i.platform === platform);

  const handleConnect = async (platform: PlatformConfig) => {
    setConnectingPlatform(platform.key);
    try {
      await createIntegration({
        platform: platform.key,
        credentials: connectForm,
      });
      setConnectForm({});
      refresh();
    } catch (e) {
      console.error(e);
    } finally {
      setConnectingPlatform(null);
    }
  };

  const handleDisconnect = async (id: string) => {
    if (!confirm("Disconnect this integration?")) return;
    await deleteIntegration(id);
    refresh();
  };

  const handleSync = async (id: string) => {
    setSyncing(id);
    setSyncResult(null);
    try {
      const result = await syncIntegration(id);
      setSyncResult(result);
      refresh();
    } catch (e) {
      console.error(e);
    } finally {
      setSyncing(null);
    }
  };

  const handleToggleEnabled = async (integration: Integration) => {
    await updateIntegration(integration.id, { enabled: !integration.enabled });
    refresh();
  };

  const openEditDialog = (integration: Integration, platform: PlatformConfig) => {
    setEditingIntegration(integration);
    const form: Record<string, string> = {};
    for (const field of platform.fields) {
      form[field.key] = "";
    }
    setEditForm(form);
    setEditDialogOpen(true);
  };

  const handleEditSave = async () => {
    if (!editingIntegration) return;
    await updateIntegration(editingIntegration.id, { credentials: editForm });
    setEditDialogOpen(false);
    setEditingIntegration(null);
    refresh();
  };

  const formatTime = (iso: string | null) => {
    if (!iso) return "Never";
    return new Date(iso).toLocaleString();
  };

  if (loading) {
    return <p className="text-muted-foreground">Loading...</p>;
  }

  return (
    <div className="space-y-6">
      <h1 className="text-3xl font-bold">Integrations</h1>

      {syncResult && (
        <div className="rounded-md border bg-muted/50 p-4 text-sm">
          <p>{syncResult.message}</p>
          {(syncResult.imported > 0 || syncResult.updated > 0) && (
            <p className="mt-1 text-muted-foreground">
              Imported: {syncResult.imported}, Updated: {syncResult.updated},
              Errors: {syncResult.errors}
            </p>
          )}
        </div>
      )}

      <div className="grid gap-6 md:grid-cols-2">
        {platforms.map((platform) => {
          const integration = getIntegration(platform.key);
          const isConnected = !!integration;

          return (
            <Card key={platform.key}>
              <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                <div className="flex items-center gap-3">
                  <div className="flex h-10 w-10 items-center justify-center rounded-lg bg-muted">
                    <Link2 className="h-5 w-5" />
                  </div>
                  <div>
                    <CardTitle className="text-lg">{platform.name}</CardTitle>
                  </div>
                </div>
                <Badge variant={isConnected ? "default" : "secondary"}>
                  {isConnected ? "Connected" : "Not Connected"}
                </Badge>
              </CardHeader>
              <CardContent>
                {isConnected ? (
                  <div className="space-y-4">
                    <div className="grid grid-cols-2 gap-2 text-sm">
                      <div>
                        <span className="text-muted-foreground">
                          Last synced:
                        </span>
                        <p>{formatTime(integration.last_synced_at)}</p>
                      </div>
                      <div>
                        <span className="text-muted-foreground">Status:</span>
                        <p>
                          {integration.enabled ? "Enabled" : "Disabled"}
                        </p>
                      </div>
                    </div>

                    {integration.last_sync_error && (
                      <p className="text-sm text-destructive">
                        {integration.last_sync_error}
                      </p>
                    )}

                    <div className="flex flex-wrap gap-2">
                      <Button
                        size="sm"
                        variant="outline"
                        onClick={() => handleSync(integration.id)}
                        disabled={syncing === integration.id}
                      >
                        <RefreshCw
                          className={`mr-2 h-4 w-4 ${syncing === integration.id ? "animate-spin" : ""}`}
                        />
                        {syncing === integration.id
                          ? "Syncing..."
                          : "Sync Now"}
                      </Button>
                      <Button
                        size="sm"
                        variant="outline"
                        onClick={() => handleToggleEnabled(integration)}
                      >
                        {integration.enabled ? "Disable" : "Enable"}
                      </Button>
                      <Button
                        size="sm"
                        variant="outline"
                        onClick={() => openEditDialog(integration, platform)}
                      >
                        <Pencil className="mr-2 h-4 w-4" />
                        Edit
                      </Button>
                      <Button
                        size="sm"
                        variant="outline"
                        onClick={() => handleDisconnect(integration.id)}
                      >
                        <Trash2 className="mr-2 h-4 w-4" />
                        Disconnect
                      </Button>
                    </div>
                  </div>
                ) : (
                  <div className="space-y-4">
                    <button
                      type="button"
                      className="flex w-full items-center gap-2 text-sm text-muted-foreground hover:text-foreground transition-colors"
                      onClick={() =>
                        setShowHelp((h) => ({
                          ...h,
                          [platform.key]: !h[platform.key],
                        }))
                      }
                    >
                      <HelpCircle className="h-4 w-4" />
                      <span>Where do I find these credentials?</span>
                      {showHelp[platform.key] ? (
                        <ChevronUp className="ml-auto h-4 w-4" />
                      ) : (
                        <ChevronDown className="ml-auto h-4 w-4" />
                      )}
                    </button>
                    {showHelp[platform.key] && (
                      <div className="rounded-md border bg-muted/50 p-3 text-sm space-y-2">
                        <ol className="list-decimal list-inside space-y-1 text-muted-foreground">
                          {platform.help.steps.map((step, i) => (
                            <li key={i}>{step}</li>
                          ))}
                        </ol>
                        <p className="text-xs text-muted-foreground/80 border-t pt-2">
                          {platform.help.note}
                        </p>
                      </div>
                    )}
                    {platform.fields.map((field) => (
                      <div key={field.key}>
                        <Label>{field.label}</Label>
                        <Input
                          type={field.type}
                          value={connectForm[field.key] || ""}
                          onChange={(e) =>
                            setConnectForm((f) => ({
                              ...f,
                              [field.key]: e.target.value,
                            }))
                          }
                          placeholder={field.label}
                        />
                      </div>
                    ))}
                    <Button
                      onClick={() => handleConnect(platform)}
                      disabled={connectingPlatform === platform.key}
                      className="w-full"
                    >
                      {connectingPlatform === platform.key
                        ? "Connecting..."
                        : `Connect ${platform.name}`}
                    </Button>
                  </div>
                )}
              </CardContent>
            </Card>
          );
        })}
      </div>

      <Dialog open={editDialogOpen} onOpenChange={setEditDialogOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Update Credentials</DialogTitle>
          </DialogHeader>
          <div className="space-y-4">
            {editingIntegration &&
              platforms
                .find((p) => p.key === editingIntegration.platform)
                ?.fields.map((field) => (
                  <div key={field.key}>
                    <Label>{field.label}</Label>
                    <Input
                      type={field.type}
                      value={editForm[field.key] || ""}
                      onChange={(e) =>
                        setEditForm((f) => ({
                          ...f,
                          [field.key]: e.target.value,
                        }))
                      }
                      placeholder={`New ${field.label}`}
                    />
                  </div>
                ))}
            <Button onClick={handleEditSave} className="w-full">
              Save
            </Button>
          </div>
        </DialogContent>
      </Dialog>
    </div>
  );
}
