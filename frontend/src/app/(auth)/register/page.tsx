"use client";

import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { z } from "zod";
import { useRegister } from "@/hooks/useAuth";
import { registerSchema } from "@/lib/validators";
import Link from "next/link";
import { useState } from "react";
import { Loader2, Link2, ArrowRight, Eye, EyeOff } from "lucide-react";
import { toast } from "sonner";

type RegisterForm = z.infer<typeof registerSchema>;

export default function RegisterPage() {
  const [showPassword, setShowPassword] = useState(false);
  const [showConfirmPassword, setShowConfirmPassword] = useState(false);
  const registerMutation = useRegister();

  const {
    register,
    handleSubmit,
    formState: { errors },
  } = useForm<RegisterForm>({
    resolver: zodResolver(registerSchema),
  });

  const onSubmit = (data: RegisterForm) => {
    registerMutation.mutate(data, {
      onError: (error) => {
        toast.error(error.response?.data?.message || "Registration failed. Please try again.");
      }
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

          <h2 className="text-4xl font-black tracking-tight text-white mb-2">CREATE ACCOUNT.</h2>
          <p className="text-sm text-white/40 mb-10 font-mono-dm uppercase tracking-wider">{"// Join the speed of redirects"}</p>

          <form onSubmit={handleSubmit(onSubmit)} className="space-y-6">
            <div>
              <label className="block text-xs font-bold text-[#6EE7B7] mb-2 uppercase tracking-widest font-mono-dm" htmlFor="name">
                Full Name
              </label>
              <input
                id="name"
                type="text"
                placeholder="John Doe"
                className="block w-full appearance-none rounded-xl border border-white/10 bg-white/[0.03] px-4 py-3 text-white placeholder-white/20 focus:border-[#6EE7B7]/50 focus:outline-none transition-all sm:text-sm"
                {...register("name")}
              />
              {errors.name && <p className="mt-1 text-xs text-red-400 font-medium">{errors.name.message}</p>}
            </div>

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
            </div>

            <div>
              <label className="block text-xs font-bold text-[#6EE7B7] mb-2 uppercase tracking-widest font-mono-dm" htmlFor="confirmPassword">
                Confirm Password
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
              disabled={registerMutation.isPending}
              className="btn-primary cursor-pointer flex w-full justify-center items-center gap-2 rounded-xl px-4 py-3.5 text-sm font-bold uppercase tracking-wider transition-all disabled:opacity-50 disabled:cursor-not-allowed"
            >
              {registerMutation.isPending ? <Loader2 className="animate-spin h-5 w-5" /> : (
                <>Sign up <ArrowRight size={18} /></>
              )}
            </button>
          </form>
        </div>

        <div className="bg-white/[0.02] px-8 py-5 border-t border-white/[0.05]">
          <p className="text-center text-sm text-white/40">
            Already have an account?{" "}
            <Link href="/login" className="font-bold text-[#6EE7B7] hover:text-[#A7F3D0] transition-colors">
              Sign in
            </Link>
          </p>
        </div>
      </div>
    </div>
  );
}
