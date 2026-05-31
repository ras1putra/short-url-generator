"use client";

import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { z } from "zod";
import { useForgotPassword } from "@/hooks/useAuth";
import { forgotPasswordSchema } from "@/lib/validators";
import Link from "next/link";
import { ROUTE_LOGIN } from "@/lib/constants";
import { useState, useRef, useCallback } from "react";
import { Loader2, Link2, ArrowLeft, MailCheck, RefreshCw } from "lucide-react";
import { AxiosError } from "axios";
import { ApiErrorResponse } from "@/types/api";
import { toast } from "sonner";

type ForgotPasswordForm = z.infer<typeof forgotPasswordSchema>;

const formatCooldown = (seconds: number) => {
  const m = Math.floor(seconds / 60);
  const s = seconds % 60;
  if (m > 0) return `${m}m ${s}s`;
  return `${s}s`;
};

export default function ForgotPasswordPage() {
  const [sent, setSent] = useState(false);
  const [email, setEmail] = useState("");
  const [cooldown, setCooldown] = useState(0);
  const timerRef = useRef<ReturnType<typeof setInterval>>(null);
  const forgotMutation = useForgotPassword();

  const {
    register,
    handleSubmit,
    formState: { errors },
  } = useForm<ForgotPasswordForm>({
    resolver: zodResolver(forgotPasswordSchema),
  });

  const startCooldown = (seconds: number) => {
    setCooldown(seconds);
    timerRef.current = setInterval(() => {
      setCooldown((prev) => {
        if (prev <= 1) {
          if (timerRef.current) clearInterval(timerRef.current);
          return 0;
        }
        return prev - 1;
      });
    }, 1000);
  };

  const onSubmit = (data: ForgotPasswordForm) => {
    setEmail(data.email);
    forgotMutation.mutate(data, {
      onSuccess: () => {
        setSent(true);
      },
      onError: (error: AxiosError<ApiErrorResponse>) => {
        const msg = error.response?.data?.message || "";
        const match = msg.match(/(\d+)/);
        if (error.response?.status === 429 && match) {
          startCooldown(parseInt(match[1]));
        }
      },
    });
  };

  const handleResend = () => {
    if (cooldown > 0) {
      toast.error(`Please wait ${formatCooldown(cooldown)} before requesting a new email`);
      return;
    }
    forgotMutation.mutate({ email }, {
      onSuccess: () => {
        setSent(true);
      },
      onError: (error: AxiosError<ApiErrorResponse>) => {
        const msg = error.response?.data?.message || "";
        const match = msg.match(/(\d+)/);
        if (error.response?.status === 429 && match) {
          startCooldown(parseInt(match[1]));
        }
      },
    });
  };

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

          {sent ? (
            <div className="text-center fade-in">
              <div className="inline-flex items-center justify-center w-16 h-16 rounded-full bg-[#6EE7B7]/10 mb-6">
                <MailCheck size={32} className="text-[#6EE7B7]" />
              </div>
              <h2 className="text-4xl font-black tracking-tight text-white mb-2">CHECK YOUR INBOX.</h2>
              <p className="text-sm text-white/40 mb-6 font-mono-dm uppercase tracking-wider">
                {"// If that email is registered, a password reset link has been sent."}
              </p>
              <button
                type="button"
                onClick={handleResend}
                disabled={cooldown > 0 || forgotMutation.isPending}
                className="btn-primary cursor-pointer flex w-full justify-center items-center gap-2 rounded-xl px-4 py-3.5 text-sm font-bold uppercase tracking-wider transition-all disabled:opacity-50 disabled:cursor-not-allowed mb-3"
              >
                {forgotMutation.isPending ? (
                  <Loader2 size={18} className="animate-spin" />
                ) : cooldown > 0 ? (
                  <>{`Resend in ${formatCooldown(cooldown)}`}</>
                ) : (
                  <><RefreshCw size={18} /> Resend email</>
                )}
              </button>
              <Link
                href={ROUTE_LOGIN}
                className="cursor-pointer flex w-full items-center justify-center gap-3 rounded-xl border border-white/[0.12] bg-white/[0.04] px-4 py-3.5 text-sm font-bold text-white uppercase tracking-wider hover:bg-white/[0.08] hover:border-white/20 hover:-translate-y-0.5 transition-all"
              >
                <ArrowLeft size={18} /> Back to sign in
              </Link>
            </div>
          ) : (
            <>
              <h2 className="text-4xl font-black tracking-tight text-white mb-2">FORGOT PASSWORD?</h2>
              <p className="text-sm text-white/40 mb-10 font-mono-dm uppercase tracking-wider">
                {"// Enter your email to receive a reset link"}
              </p>

              <form onSubmit={handleSubmit(onSubmit)} className="space-y-6">
                <div>
                  <label className="block text-xs font-bold text-[#6EE7B7] mb-2 uppercase tracking-widest font-mono-dm" htmlFor="email">
                    Email Address
                  </label>
                  <input
                    id="email"
                    type="email"
                    placeholder="you@example.com"
                    className="block w-full appearance-none rounded-xl border border-white/10 bg-white/[0.03] px-4 py-3 text-white placeholder-white/20 focus:border-[#6EE7B7]/50 focus:outline-none transition-all sm:text-sm"
                    {...register("email")}
                  />
                  {errors.email && <p className="mt-1 text-xs text-red-400 font-medium">{errors.email.message}</p>}
                </div>

                <button
                  type="submit"
                  disabled={forgotMutation.isPending}
                  className="btn-primary cursor-pointer flex w-full justify-center items-center gap-2 rounded-xl px-4 py-3.5 text-sm font-bold uppercase tracking-wider transition-all disabled:opacity-50 disabled:cursor-not-allowed"
                >
                  {forgotMutation.isPending ? <Loader2 className="animate-spin h-5 w-5" /> : "Send reset link"}
                </button>
              </form>
            </>
          )}
        </div>

        <div className="bg-white/[0.02] px-8 py-5 border-t border-white/[0.05]">
          <p className="text-center text-sm text-white/40">
            Remember your password?{" "}
            <Link href={ROUTE_LOGIN} className="font-bold text-[#6EE7B7] hover:text-[#A7F3D0] transition-colors">
              Sign in
            </Link>
          </p>
        </div>
      </div>
    </div>
  );
}
