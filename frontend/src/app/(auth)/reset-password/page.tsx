"use client";

import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { z } from "zod";
import { useResetPassword } from "@/hooks/useAuth";
import { resetPasswordSchema } from "@/lib/validators";
import Link from "next/link";
import { ROUTE_LOGIN } from "@/lib/constants";
import { useSearchParams } from "next/navigation";
import { Suspense, useState } from "react";
import { Loader2, Link2, Eye, EyeOff } from "lucide-react";

type ResetPasswordForm = z.infer<typeof resetPasswordSchema>;

function ResetPasswordFormInner() {
  const searchParams = useSearchParams();
  const token = searchParams.get("token") || "";
  const [showPassword, setShowPassword] = useState(false);
  const [showConfirmPassword, setShowConfirmPassword] = useState(false);
  const resetMutation = useResetPassword();

  const {
    register,
    handleSubmit,
    formState: { errors },
  } = useForm<ResetPasswordForm>({
    resolver: zodResolver(resetPasswordSchema),
    defaultValues: { token },
  });

  const onSubmit = (data: ResetPasswordForm) => {
    resetMutation.mutate({ token: data.token, password: data.password });
  };

  if (!token) {
    return (
      <div className="text-center">
        <h2 className="text-3xl font-black tracking-tight text-white mb-2">INVALID LINK.</h2>
        <p className="text-sm text-white/40 mb-8 font-mono-dm uppercase tracking-wider">
          {"// This reset link is invalid or expired"}
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
    <>
      <h2 className="text-4xl font-black tracking-tight text-white mb-2">RESET PASSWORD.</h2>
      <p className="text-sm text-white/40 mb-10 font-mono-dm uppercase tracking-wider">
        {"// Enter your new password"}
      </p>

      <form onSubmit={handleSubmit(onSubmit)} className="space-y-6">
        <input type="hidden" {...register("token")} />

        <div>
          <label className="block text-xs font-bold text-[#6EE7B7] mb-2 uppercase tracking-widest font-mono-dm" htmlFor="password">
            New Password
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
        </div>

        <div>
          <label className="block text-xs font-bold text-[#6EE7B7] mb-2 uppercase tracking-widest font-mono-dm" htmlFor="confirmPassword">
            Confirm New Password
          </label>
          <div className="relative">
            <input
              id="confirmPassword"
              type={showConfirmPassword ? "text" : "password"}
              placeholder="••••••••"
              className="block w-full appearance-none rounded-xl border border-white/10 bg-white/[0.03] px-4 py-3 text-white placeholder-white/20 focus:border-[#6EE7B7]/50 focus:outline-none transition-all sm:text-sm"
              {...register("confirmPassword")}
            />
            <button
              type="button"
              onClick={() => setShowConfirmPassword(!showConfirmPassword)}
              className="absolute right-3 top-1/2 -translate-y-1/2 text-white/30 hover:text-white/60 transition-colors cursor-pointer"
            >
              {showConfirmPassword ? <EyeOff size={18} /> : <Eye size={18} />}
            </button>
          </div>
          {errors.confirmPassword && <p className="mt-1 text-xs text-red-400 font-medium">{errors.confirmPassword.message}</p>}
        </div>

        <button
          type="submit"
          disabled={resetMutation.isPending}
          className="btn-primary cursor-pointer flex w-full justify-center items-center gap-2 rounded-xl px-4 py-3.5 text-sm font-bold uppercase tracking-wider transition-all disabled:opacity-50 disabled:cursor-not-allowed"
        >
          {resetMutation.isPending ? <Loader2 className="animate-spin h-5 w-5" /> : "Reset Password"}
        </button>
      </form>
    </>
  );
}

export default function ResetPasswordPage() {
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
            <ResetPasswordFormInner />
          </Suspense>
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
