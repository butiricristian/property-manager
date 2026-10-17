import { useState } from "react";
import { Button } from "@/components/ui/button";
import { UpgradeDialog } from "@/components/UpgradeDialog";

export function GuestBanner() {
  const [upgradeOpen, setUpgradeOpen] = useState(false);

  return (
    <>
      <div className="bg-amber-50 border-b border-amber-200 px-4 py-2 text-sm text-amber-800 flex items-center justify-between">
        <span>
          You're in demo mode — data will be deleted after 7 days.
        </span>
        <Button
          variant="outline"
          size="sm"
          className="border-amber-300 text-amber-800 hover:bg-amber-100"
          onClick={() => setUpgradeOpen(true)}
        >
          Create an account
        </Button>
      </div>
      <UpgradeDialog open={upgradeOpen} onOpenChange={setUpgradeOpen} />
    </>
  );
}
