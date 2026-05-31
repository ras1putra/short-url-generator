"use client";

import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { z } from "zod";
import { useLogin, useSendVerification } from "@/hooks/useAuth";
import { loginSchema } from "@/lib/validators";
import Link from "next/link";
import { ROUTE_REGISTER, ROUTE_FORGOT_PASSWORD, API_AUTH_GOOGLE_LOGIN } from "@/lib/constants";
import { useState, useRef } from "react";
import { Loader2, Link2, ArrowRight, Eye, EyeOff, MailCheck } from "lucide-react";
import { toast } from "sonner";
import { AxiosError } from "axios";
import { ApiErrorResponse } from "@/types/api";

type LoginForm = z.infer<typeof loginSchema>;

const formatCooldown = (seconds: number) => {
  const m = Math.floor(seconds / 60);
  const s = seconds % 60;
  if (m > 0) return `${m}m ${s}s`;
  return `${s}s`;
};

export default function LoginPage() {
  const [showPassword, setShowPassword] = useState(false);
  const [unverifiedEmail, setUnverifiedEmail] = useState<string | null>(null);
  const [cooldown, setCooldown] = useState(0);
  const timerRef = useRef<ReturnType<typeof setInterval>>(null);
  const loginMutation = useLogin();
  const sendVerificationMutation = useSendVerification();

  const {
    register,
    handleSubmit,
    formState: { errors },
  } = useForm<LoginForm>({
    resolver: zodResolver(loginSchema),
  });

  const onSubmit = (data: LoginForm) => {
    setUnverifiedEmail(null);
    loginMutation.mutate(data, {
      onError: (error) => {
        if (error.response?.data?.code === "EMAIL_NOT_VERIFIED") {
          setUnverifiedEmail(data.email);
        } else {
          toast.error(error.response?.data?.message || "Login failed. Please try again.");
        }
      }
    });
  };

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

  const handleResendVerification = () => {
    if (!unverifiedEmail) return;
    if (cooldown > 0) {
      toast.error(`Please wait ${formatCooldown(cooldown)} before requesting a new verification email`);
      return;
    }
    sendVerificationMutation.mutate({ email: unverifiedEmail }, {
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

          <h2 className="text-4xl font-black tracking-tight text-white mb-2">WELCOME BACK.</h2>
          <p className="text-sm text-white/40 mb-10 font-mono-dm uppercase tracking-wider">{"// Sign in to manage your links"}</p>

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

            <div>
              <label className="block text-xs font-bold text-[#6EE7B7] mb-2 uppercase tracking-widest font-mono-dm" htmlFor="password">
                Password
              </label>
              <div className="relative">
                <input
                  id="password"
                  type={showPassword ? "text" : "password"}
                  placeholder="••••••••"
                  className="block w-full appearance-none rounded-xl border border-white/10 bg-white/[0.03] px-4 py-3 text-white placeholder-white/20 focus:border-[#6EE7B7]/50 focus:outline-none transition-all sm:text-sm"
                  {...register("password")}
                />
                <button
                  type="button"
                  onClick={() => setShowPassword(!showPassword)}
                  className="absolute right-3 top-1/2 -translate-y-1/2 text-white/30 hover:text-white/60 transition-colors cursor-pointer"
                >
                  {showPassword ? <EyeOff size={18} /> : <Eye size={18} />}
                </button>
              </div>
              {errors.password && <p className="mt-1 text-xs text-red-400 font-medium">{errors.password.message}</p>}
              <div className="mt-2 text-right">
                <Link href={ROUTE_FORGOT_PASSWORD} className="text-xs text-white/40 hover:text-[#6EE7B7] transition-colors">
                  Forgot password?
                </Link>
              </div>
            </div>

            {unverifiedEmail && (
              <div className="rounded-2xl bg-white/[0.06] border border-white/[0.12] p-5 fade-in">
                <p className="text-xs font-mono-dm text-white/50 uppercase tracking-wider mb-3">{"// Email not verified"}</p>
                <p className="text-sm text-white/70 mb-4 leading-relaxed">Please verify your email before signing in. Check your inbox for the verification link.</p>
                <button
                  type="button"
                  onClick={handleResendVerification}
                  disabled={cooldown > 0 || sendVerificationMutation.isPending}
                  className="cursor-pointer inline-flex items-center gap-2 text-sm font-bold text-[#6EE7B7] hover:text-[#A7F3D0] transition-colors disabled:opacity-50"
                >
                  {sendVerificationMutation.isPending ? <Loader2 size={16} className="animate-spin" /> : <MailCheck size={16} />}
                  {sendVerificationMutation.isPending ? "Sending..." : cooldown > 0 ? `Resend in ${formatCooldown(cooldown)}` : "Resend verification email"}
                </button>
              </div>
            )}

            <button
              type="submit"
              disabled={loginMutation.isPending}
              className="btn-primary cursor-pointer flex w-full justify-center items-center gap-2 rounded-xl px-4 py-3.5 text-sm font-bold uppercase tracking-wider transition-all disabled:opacity-50 disabled:cursor-not-allowed"
            >
              {loginMutation.isPending ? <Loader2 className="animate-spin h-5 w-5" /> : (
                <>Sign in <ArrowRight size={18} /></>
              )}
            </button>
          </form>

          <div className="relative my-6 flex items-center">
            <div className="flex-1 border-t border-white/[0.06]" />
            <span className="mx-4 text-[10px] font-mono-dm uppercase tracking-[0.2em] text-white/20">or</span>
            <div className="flex-1 border-t border-white/[0.06]" />
          </div>

          <a
            href={API_AUTH_GOOGLE_LOGIN}
            className="cursor-pointer flex w-full items-center justify-center gap-3 rounded-xl border border-white/[0.12] bg-white/[0.04] px-4 py-3.5 text-sm font-bold text-white uppercase tracking-wider hover:bg-white/[0.08] hover:border-white/20 hover:-translate-y-0.5 transition-all"
          >
            <svg width="20" height="20" viewBox="0 0 24 24" className="shrink-0">
              <path d="M22.56 12.25c0-.78-.07-1.53-.2-2.25H12v4.26h5.92a5.06 5.06 0 0 1-2.2 3.32v2.77h3.57c2.08-1.92 3.28-4.74 3.28-8.1z" fill="#4285F4" />
              <path d="M12 23c2.97 0 5.46-.98 7.28-2.66l-3.57-2.77c-.98.66-2.23 1.06-3.71 1.06-2.86 0-5.29-1.93-6.16-4.53H2.18v2.84C3.99 20.53 7.7 23 12 23z" fill="#34A853" />
              <path d="M5.84 14.09c-.22-.66-.35-1.36-.35-2.09s.13-1.43.35-2.09V7.07H2.18C1.43 8.55 1 10.22 1 12s.43 3.45 1.18 4.93l2.85-2.22.81-.62z" fill="#FBBC05" />
              <path d="M12 5.38c1.62 0 3.06.56 4.21 1.64l3.15-3.15C17.45 2.09 14.97 1 12 1 7.7 1 3.99 3.47 2.18 7.07l3.66 2.84c.87-2.6 3.3-4.53 6.16-4.53z" fill="#EA4335" />
            </svg>
            Sign in with Google
          </a>
        </div>

        <div className="bg-white/[0.02] px-8 py-5 border-t border-white/[0.05]">
          <p className="text-center text-sm text-white/40">
            Don&apos;t have an account?{" "}
            <Link href={ROUTE_REGISTER} className="font-bold text-[#6EE7B7] hover:text-[#A7F3D0] transition-colors">
              Sign up
            </Link>
          </p>
        </div>
      </div>
    </div>
  );
}
