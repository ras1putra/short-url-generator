"use client";

import { ReactNode } from "react";

interface DialogProps {
  open: boolean;
  onClose: () => void;
  title: string;
  message: string;
  confirmLabel: string;
  icon?: ReactNode;
  loading?: boolean;
  onConfirm: () => void;
}

export default function Dialog({ open, onClose, title, message, confirmLabel, icon, loading, onConfirm }: DialogProps) {
  if (!open) return null;

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center p-4">
      <div className="absolute inset-0 bg-black/60 backdrop-blur-sm" onClick={onClose} />
      <div className="relative w-full max-w-md rounded-2xl bg-[#141414] border border-white/[0.08] shadow-2xl p-4 sm:p-6">
        {icon && (
          <div className="flex items-center gap-3 mb-4">
            <div className="h-10 w-10 rounded-xl bg-red-500/10 flex items-center justify-center">
              {icon}
            </div>
            <h3 className="text-lg font-bold text-white">{title}</h3>
          </div>
        )}
        {!icon && <h3 className="text-lg font-bold text-white mb-4">{title}</h3>}
        <p className="text-sm text-white/60 leading-relaxed mb-4 sm:mb-6">{message}</p>
        <div className="flex gap-3 justify-end">
          <button
            onClick={onClose}
            disabled={loading}
            className="px-4 py-2 rounded-xl text-sm font-medium text-white/60 hover:text-white bg-white/[0.06] hover:bg-white/[0.1] transition-colors cursor-pointer disabled:opacity-50"
          >
            Cancel
          </button>
          <button
            onClick={onConfirm}
            disabled={loading}
            className="px-4 py-2 rounded-xl text-sm font-bold text-white bg-red-500/80 hover:bg-red-500 transition-colors cursor-pointer disabled:opacity-50 flex items-center gap-2"
          >
            {loading && (
              <svg className="animate-spin h-4 w-4"viewBox="0 0 24 24"fill="none">
                <circle className="opacity-25"cx="12"cy="12"r="10"stroke="currentColor"strokeWidth="4" />
                <path className="opacity-75"fill="currentColor"d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z" />
              </svg>
            )}
            {confirmLabel}
          </button>
        </div>
      </div>
    </div>
  );
}
