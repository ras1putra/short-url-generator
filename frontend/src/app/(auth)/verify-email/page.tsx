"use client";

import { useVerifyEmail } from "@/hooks/useAuth";
import Link from "next/link";
import { ROUTE_LOGIN } from "@/lib/constants";
import { useSearchParams, useRouter } from "next/navigation";
import { Suspense, useEffect } from "react";
import { Loader2, Link2, CheckCircle, XCircle } from "lucide-react";
import axios from "axios";
import type { ApiErrorResponse } from "@/types/api";
import { toast } from "sonner";

function VerifyEmailInner() {
  const router = useRouter();
  const searchParams = useSearchParams();
  const token = searchParams.get("token") || "";

  const { isPending, isError, error, isSuccess } = useVerifyEmail(token);

  useEffect(() => {
    if (isSuccess) {
      toast.success("Email verified! You can now sign in.");
      router.push(ROUTE_LOGIN);
    }
  }, [isSuccess, router]);

  if (!token) {
    return (
      <div className="text-center fade-in">
        <div className="inline-flex items-center justify-center w-16 h-16 rounded-full bg-red-500/10 mb-6">
          <XCircle size={32} className="text-red-400" />
        </div>
        <h2 className="text-3xl font-black tracking-tight text-white mb-2">INVALID LINK.</h2>
        <p className="text-sm text-white/40 mb-8 font-mono-dm uppercase tracking-wider">
          {"// No verification token found"}
        </p>
        <Link
          href={ROUTE_LOGIN}
          className="text-sm font-bold text-[#6EE7B7] hover:text-[#A7F3D0] transition-colors"
        >
          Back to sign in
        </Link>
      </div>
    );
  }

  if (isPending) {
    return (
      <div className="text-center fade-in">
        <Loader2 className="animate-spin h-8 w-8 mx-auto text-white/40 mb-6" />
        <h2 className="text-3xl font-black tracking-tight text-white mb-2">VERIFYING...</h2>
        <p className="text-sm text-white/40 font-mono-dm uppercase tracking-wider">
          {"// Please wait"}
        </p>
      </div>
    );
  }

  if (isError) {
    const message = axios.isAxiosError(error)
      ? (error.response?.data as ApiErrorResponse)?.message
      : null;

    return (
      <div className="text-center fade-in">
        <div className="inline-flex items-center justify-center w-16 h-16 rounded-full bg-red-500/10 mb-6">
          <XCircle size={32} className="text-red-400" />
        </div>
        <h2 className="text-3xl font-black tracking-tight text-white mb-2">VERIFICATION FAILED.</h2>
        <p className="text-sm text-white/40 mb-8 font-mono-dm uppercase tracking-wider">
          {message || "The link may be invalid or expired."}
        </p>
        <Link
          href={ROUTE_LOGIN}
          className="text-sm font-bold text-[#6EE7B7] hover:text-[#A7F3D0] transition-colors"
        >
          Back to sign in
        </Link>
      </div>
    );
  }

  return (
    <div className="text-center fade-in">
      <div className="inline-flex items-center justify-center w-16 h-16 rounded-full bg-[#6EE7B7]/10 mb-6">
        <CheckCircle size={32} className="text-[#6EE7B7]" />
      </div>
      <h2 className="text-3xl font-black tracking-tight text-white mb-2">EMAIL VERIFIED!</h2>
      <p className="text-sm text-white/40 mb-8 font-mono-dm uppercase tracking-wider">
        {"// You can now sign in to your account"}
      </p>
      <Link
        href={ROUTE_LOGIN}
        className="btn-primary inline-flex items-center gap-2 rounded-xl px-6 py-3.5 text-sm font-bold uppercase tracking-wider transition-all"
      >
        Sign in
      </Link>
    </div>
  );
}

export default function VerifyEmailPage() {
  return (
    <div className="flex min-h-screen items-center justify-center bg-[#0A0A0A] p-4 font-syne grain-overlay">
      <div className="w-full max-w-md overflow-hidden rounded-2xl bg-white/[0.02] shadow-2xl border border-white/[0.08] backdrop-blur-xl glow-green fade-in">
        <div className="p-8">
          <div className="flex items-center gap-2 mb-10">
            <div className="h-8 w-8 rounded-lg bg-[#6EE7B7] flex items-center justify-center">
              <Link2 size={16} className="text-[#0A0A0A] stroke-[2.5]" />
            </div>
            <span className="text-xl font-bold text-white tracking-tight">go-short</span>
          </div>

          <Suspense fallback={<Loader2 className="animate-spin h-8 w-8 mx-auto text-white/40" />}>
            <VerifyEmailInner />
          </Suspense>
        </div>
      </div>
    </div>
  );
}
