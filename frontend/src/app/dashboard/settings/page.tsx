"use client";

import { useState } from "react";
import { Settings, Megaphone, User, AlertTriangle, ArrowUp, ArrowDown } from "lucide-react";
import { useUserStore } from "@/store/useUserStore";
import { useUpgrade, useDowngrade } from "@/hooks/useAuth";
import { ROLE_USER, ROLE_ADVERTISER, ROUTE_CAMPAIGNS } from "@/lib/constants";
import Dialog from "@/components/ui/Dialog";
import Link from "next/link";

export default function SettingsPage() {
  const user = useUserStore((state) => state.user);
  const role = user?.role || ROLE_USER;
  const upgradeMutation = useUpgrade();
  const downgradeMutation = useDowngrade();
  const [showUpgrade, setShowUpgrade] = useState(false);
  const [showDowngrade, setShowDowngrade] = useState(false);

  return (
    <div className="mb-8 space-y-8">
      <div className="flex items-center gap-3 mb-2">
        <div className="h-8 w-8 rounded-lg bg-[#22D3EE]/10 flex items-center justify-center">
          <Settings size={16} className="text-[#22D3EE]" />
        </div>
        <h1 className="text-3xl font-black tracking-tight text-white">Settings</h1>
      </div>

      {/* Account Section */}
      <div className="rounded-2xl border border-white/[0.06] bg-white/[0.02] p-4 sm:p-6">
        <h2 className="text-lg font-bold text-white mb-4">Account</h2>

        <div className="flex items-center justify-between py-3 border-b border-white/[0.06]">
          <div>
            <p className="text-sm text-white/40 font-mono-dm uppercase tracking-wider">Email</p>
            <p className="text-sm text-white mt-1">{user?.email}</p>
          </div>
        </div>

        <div className="flex items-center justify-between py-3 border-b border-white/[0.06]">
          <div>
            <p className="text-sm text-white/40 font-mono-dm uppercase tracking-wider">Name</p>
            <p className="text-sm text-white mt-1">{user?.name}</p>
          </div>
        </div>

        <div className="flex items-center justify-between py-3">
          <div>
            <p className="text-sm text-white/40 font-mono-dm uppercase tracking-wider">Role</p>
            <div className="flex items-center gap-2 mt-1">
              {role === ROLE_ADVERTISER ? (
                <span className="inline-flex items-center gap-1.5 text-sm font-bold text-[#6EE7B7]">
                  <Megaphone size={14} /> Advertiser
                </span>
              ) : (
                <span className="inline-flex items-center gap-1.5 text-sm font-bold text-white/60">
                  <User size={14} /> Regular User
                </span>
              )}
            </div>
          </div>
        </div>
      </div>

      {/* Role Change Section */}
      <div className="rounded-2xl border border-white/[0.06] bg-white/[0.02] p-4 sm:p-6">
        <h2 className="text-lg font-bold text-white mb-2">Account Type</h2>
        <p className="text-sm text-white/40 mb-4 sm:mb-6">Change your account role and available features.</p>

        {role === ROLE_USER && (
          <div className="space-y-4">
            <div className="rounded-xl border border-[#6EE7B7]/20 bg-[#6EE7B7]/[0.03] p-4">
              <div className="flex items-start gap-3">
                <div className="h-8 w-8 rounded-lg bg-[#6EE7B7]/10 flex items-center justify-center shrink-0 mt-0.5">
                  <Megaphone size={16} className="text-[#6EE7B7]" />
                </div>
                <div>
                  <p className="text-sm font-bold text-white">Advertiser Account</p>
                  <ul className="mt-2 text-sm text-white/50 space-y-1 list-disc list-inside">
                    <li>Create and manage ad campaigns</li>
                    <li>Run banner, video, and interstitial ads</li>
                    <li>Access campaign analytics and stats</li>
                  </ul>
                </div>
              </div>
            </div>
            <button
              onClick={() => setShowUpgrade(true)}
              className="btn-primary cursor-pointer inline-flex items-center gap-2 rounded-xl px-4 sm:px-6 py-3 text-sm font-bold uppercase tracking-wider transition-all"
            >
              <ArrowUp size={16} />
              Upgrade to Advertiser
            </button>
          </div>
        )}

        {role === ROLE_ADVERTISER && (
          <div className="space-y-4">
            <div className="rounded-xl border border-red-500/20 bg-red-500/[0.03] p-4">
              <div className="flex items-start gap-3">
                <div className="h-8 w-8 rounded-lg bg-red-500/10 flex items-center justify-center shrink-0 mt-0.5">
                  <AlertTriangle size={16} className="text-red-400" />
                </div>
                <div>
                  <p className="text-sm font-bold text-white">Downgrade to Regular User</p>
                  <p className="mt-2 text-sm text-white/50">
                    Your ads will be paused and you will no longer be able to create or manage campaigns.
                    Your existing links, wallet balance, and analytics will be preserved.
                  </p>
                </div>
              </div>
            </div>
            <button
              onClick={() => setShowDowngrade(true)}
              className="cursor-pointer inline-flex items-center gap-2 rounded-xl px-4 sm:px-6 py-3 text-sm font-bold uppercase tracking-wider transition-all bg-red-500/10 text-red-400 hover:bg-red-500/20 border border-red-500/20"
            >
              <ArrowDown size={16} />
              Downgrade to Regular User
            </button>
            <div className="mt-4">
              <Link
                href={ROUTE_CAMPAIGNS}
                className="text-sm text-[#6EE7B7] hover:text-[#A7F3D0] transition-colors"
              >
                Manage your campaigns →
              </Link>
            </div>
          </div>
        )}
      </div>

      <Dialog
        open={showUpgrade}
        onClose={() => setShowUpgrade(false)}
        title="Upgrade to Advertiser?"
        message="You will be able to create ad campaigns, run ads on your links, and access campaign analytics. This action can be reversed later."
        confirmLabel="Upgrade"
        icon={<Megaphone size={20} className="text-white" />}
        loading={upgradeMutation.isPending}
        onConfirm={() => upgradeMutation.mutate(undefined, { onSuccess: () => setShowUpgrade(false) })}
      />

      <Dialog
        open={showDowngrade}
        onClose={() => setShowDowngrade(false)}
        title="Downgrade to Regular User?"
        message="All your active ad campaigns will be paused immediately. You will not be able to create or manage campaigns as a regular user. Your links, wallet balance, and other data will be preserved."
        confirmLabel="Downgrade & Pause Ads"
        icon={<AlertTriangle size={20} className="text-white" />}
        loading={downgradeMutation.isPending}
        onConfirm={() => downgradeMutation.mutate(undefined, { onSuccess: () => setShowDowngrade(false) })}
      />
    </div>
  );
}
