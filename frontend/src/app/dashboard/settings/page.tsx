"use client";

import { Settings } from "lucide-react";

export default function SettingsPage() {
  return (
    <div className="mb-8">
      <div className="flex items-center gap-3 mb-2">
        <div className="h-8 w-8 rounded-lg bg-[#22D3EE]/10 flex items-center justify-center">
          <Settings size={16} className="text-[#22D3EE]" />
        </div>
        <h1 className="text-3xl font-black tracking-tight text-white">Settings</h1>
      </div>
      <p className="mt-2 text-white/50 font-mono-dm text-sm">{"// Coming soon"}</p>
    </div>
  );
}
