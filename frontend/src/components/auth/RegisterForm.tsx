"use client";

import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { z } from "zod";
import { AxiosError } from "axios";
import { api } from "@/lib/api";
import { registerSchema } from "@/lib/validators";
import { API_AUTH_REGISTER, API_AUTH_LOGIN, API_AUTH_UPGRADE, API_AUTH_DOWNGRADE, API_AUTH_GOOGLE_LOGIN, ROUTE_LOGIN } from "@/lib/constants";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { useState } from "react";
import { Loader2, ArrowRight, Eye, EyeOff, ArrowUp, ArrowDown, Megaphone, AlertTriangle } from "lucide-react";
import { toast } from "sonner";
import { useUserStore } from "@/store/useUserStore";
import Dialog from "@/components/ui/Dialog";
import TurnstileWidget from "@/components/TurnstileWidget";

type RegisterForm = z.infer<typeof registerSchema>;

interface RegisterFormProps {
  title: string;
  subtitle: string;
  onSuccess: (email: string, password: string) => Promise<void>;
  buttonLabel?: string;
  footer: React.ReactNode;
  googleLoginURL?: string;
}

export default function RegisterForm({ title, subtitle, onSuccess, buttonLabel, footer, googleLoginURL }: RegisterFormProps) {
  const router = useRouter();
  const setUser = useUserStore((state) => state.setUser);
  const [showPassword, setShowPassword] = useState(false);
  const [showConfirmPassword, setShowConfirmPassword] = useState(false);
  const [pending, setPending] = useState(false);
  const [existingEmail, setExistingEmail] = useState<string | null>(null);
  const [existingRole, setExistingRole] = useState<"user" | "advertiser" | null>(null);
  const [savedPassword, setSavedPassword] = useState("");
  const [showUpgradeConfirm, setShowUpgradeConfirm] = useState(false);
  const [showDowngradeConfirm, setShowDowngradeConfirm] = useState(false);

  const {
    register,
    handleSubmit,
    formState: { errors },
  } = useForm<RegisterForm>({
    resolver: zodResolver(registerSchema),
  });

  const doLogin = async (email: string, password: string) => {
    const res = await api.post(API_AUTH_LOGIN, { email, password });
    setUser({
      id: res.data.data.user.id,
      email: res.data.data.user.email,
      name: res.data.data.user.name,
      role: res.data.data.user.role,
      email_verified: res.data.data.user.email_verified,
      created_at: res.data.data.user.created_at,
    });
  };

  const doUpgrade = async (email: string, password: string) => {
    await doLogin(email, password);
    const upgradeRes = await api.post(API_AUTH_UPGRADE);
    setUser({
      id: upgradeRes.data.data.user.id,
      email: upgradeRes.data.data.user.email,
      name: upgradeRes.data.data.user.name,
      role: upgradeRes.data.data.user.role,
      email_verified: upgradeRes.data.data.user.email_verified,
      created_at: upgradeRes.data.data.user.created_at,
    });
  };

  const doDowngrade = async (email: string, password: string) => {
    await doLogin(email, password);
    const downgradeRes = await api.post(API_AUTH_DOWNGRADE);
    setUser({
      id: downgradeRes.data.data.user.id,
      email: downgradeRes.data.data.user.email,
      name: downgradeRes.data.data.user.name,
      role: downgradeRes.data.data.user.role,
      email_verified: downgradeRes.data.data.user.email_verified,
      created_at: downgradeRes.data.data.user.created_at,
    });
  };

  const onSubmit = async (data: RegisterForm) => {
    setExistingEmail(null);
    setExistingRole(null);
    setPending(true);
    try {
      const token = window.turnstile?.getResponse();
      const payload = token ? { ...data, "cf-turnstile-response": token } : data;
      await api.post(API_AUTH_REGISTER, payload);
      await onSuccess(data.email, data.password);
    } catch (raw) {
      const err = raw instanceof AxiosError ? raw : null;
      const errData = err?.response?.data as
        { can_upgrade?: boolean; can_downgrade?: boolean; message?: string } | undefined;
      if (err?.response?.status === 409) {
        if (errData?.can_upgrade) {
          setExistingEmail(data.email);
          setExistingRole("user");
          setSavedPassword(data.password);
          return;
        }
        if (errData?.can_downgrade) {
          setExistingEmail(data.email);
          setExistingRole("advertiser");
          setSavedPassword(data.password);
          return;
        }
      }
      toast.error(errData?.message || "Something went wrong. Please try again.");
    } finally {
      setPending(false);
    }
  };

  const handleUpgrade = async () => {
    if (!existingEmail) return;
    setPending(true);
    try {
      await doUpgrade(existingEmail, savedPassword);
      toast.success("Account upgraded to advertiser!");
      router.push("/dashboard/campaigns");
    } catch {
      toast.error("Failed to upgrade. Try signing in first.");
      router.push(ROUTE_LOGIN);
    } finally {
      setPending(false);
    }
  };

  const handleDowngrade = async () => {
    if (!existingEmail) return;
    setPending(true);
    try {
      await doDowngrade(existingEmail, savedPassword);
      toast.success("Account downgraded to regular user. Your ads have been paused.");
      router.push("/dashboard/links");
    } catch {
      toast.error("Failed to downgrade. Try signing in first.");
      router.push(ROUTE_LOGIN);
    } finally {
      setPending(false);
    }
  };

  return (
    <div className="flex min-h-screen items-center justify-center bg-[#0A0A0A] p-4 font-syne grain-overlay">
      <div className="w-full max-w-md overflow-hidden rounded-2xl bg-white/[0.02] shadow-2xl border border-white/[0.08] backdrop-blur-xl glow-green fade-in">
        <div className="p-4 sm:p-8">
          <div className="flex items-center gap-2 mb-4 sm:mb-10">
            <div className="h-8 w-8 rounded-lg bg-[#6EE7B7] flex items-center justify-center">
              <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="#0A0A0A" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round">
                <path d="M10 13a5 5 0 0 0 7.54.54l3-3a5 5 0 0 0-7.07-7.07l-1.72 1.71" />
                <path d="M14 11a5 5 0 0 0-7.54-.54l-3 3a5 5 0 0 0 7.07 7.07l1.71-1.71" />
              </svg>
            </div>
            <span className="text-xl font-bold text-white tracking-tight">go-short</span>
          </div>

          <h2 className="text-3xl sm:text-4xl font-black tracking-tight text-white mb-2">{title}</h2>
          <p className="text-sm text-white/40 mb-4 sm:mb-10 font-mono-dm uppercase tracking-wider">{subtitle}</p>

          {existingEmail && existingRole === "user" && (
            <div className="fade-in rounded-2xl p-4 sm:p-6 bg-white/[0.06] border border-white/[0.12]">
              <div className="absolute inset-0 flex items-center justify-center pointer-events-none">
                <div className="w-48 h-24 rounded-full bg-[radial-gradient(ellipse,rgba(110,231,183,0.06)_0%,transparent_70%)]" />
              </div>
              <div className="relative">
                <div className="inline-flex items-center gap-2 mb-4 px-3 py-1.5 rounded-full text-xs font-mono-dm uppercase tracking-widest border border-[#6EE7B7]/30 text-[#6EE7B7] bg-[#6EE7B7]/5">
                  <span>Upgrade available</span>
                </div>
                <p className="text-xs font-mono-dm text-white/50 mb-3">{"// Account exists"}</p>
                <div className="flex items-center justify-between gap-3 p-3 rounded-xl mb-4 bg-[#6EE7B7]/10 border border-[#6EE7B7]/20">
                  <span className="font-bold font-mono-dm text-sm text-[#6EE7B7] truncate">
                    {existingEmail}
                  </span>
                </div>
                <p className="text-sm text-white/70 mb-5 leading-relaxed">
                  This email is already registered. Upgrade to advertiser and start running campaigns.
                </p>
                <button
                  onClick={() => setShowUpgradeConfirm(true)}
                  disabled={pending}
                  className="cursor-pointer w-full px-4 sm:px-6 py-3 rounded-xl text-sm font-bold tracking-tight inline-flex items-center justify-center gap-2 bg-[#22D3EE] text-[#0A0A0A] hover:bg-[#67E8F9] hover:-translate-y-0.5 hover:shadow-[0_8px_30px_rgba(34,211,238,0.3)] transition-all duration-200 disabled:opacity-50 disabled:cursor-not-allowed"
                >
                  {pending ? <Loader2 size={16} className="animate-spin" /> : <ArrowUp size={16} />}
                  {pending ? "Upgrading..." : "Upgrade to Advertiser"}
                </button>
                <div className="mt-4 pt-4 border-t border-white/[0.06] text-center">
                  <Link
                    href={ROUTE_LOGIN}
                    className="text-sm text-white/50 hover:text-white transition-colors"
                  >
                    Sign in instead
                  </Link>
                </div>
              </div>
            </div>
          )}

          {existingEmail && existingRole === "advertiser" && (
            <div className="fade-in rounded-2xl p-4 sm:p-6 bg-white/[0.06] border border-white/[0.12]">
              <div className="absolute inset-0 flex items-center justify-center pointer-events-none">
                <div className="w-48 h-24 rounded-full bg-[radial-gradient(ellipse,rgba(110,231,183,0.06)_0%,transparent_70%)]" />
              </div>
              <div className="relative">
                <div className="inline-flex items-center gap-2 mb-4 px-3 py-1.5 rounded-full text-xs font-mono-dm uppercase tracking-widest border border-yellow-500/30 text-yellow-500 bg-yellow-500/5">
                  <span>Already advertiser</span>
                </div>
                <p className="text-xs font-mono-dm text-white/50 mb-3">{"// Account exists"}</p>
                <div className="flex items-center justify-between gap-3 p-3 rounded-xl mb-4 bg-yellow-500/10 border border-yellow-500/20">
                  <span className="font-bold font-mono-dm text-sm text-yellow-500 truncate">
                    {existingEmail}
                  </span>
                </div>
                <p className="text-sm text-white/70 mb-5 leading-relaxed">
                  This email is already registered as an advertiser. Downgrade to a regular user account?
                </p>
                <button
                  onClick={() => setShowDowngradeConfirm(true)}
                  disabled={pending}
                  className="cursor-pointer w-full px-4 sm:px-6 py-3 rounded-xl text-sm font-bold tracking-tight inline-flex items-center justify-center gap-2 bg-white/10 text-white hover:bg-white/20 transition-all duration-200 disabled:opacity-50 disabled:cursor-not-allowed"
                >
                  {pending ? <Loader2 size={16} className="animate-spin" /> : <ArrowDown size={16} />}
                  {pending ? "Downgrading..." : "Downgrade to Regular User"}
                </button>
                <div className="mt-4 pt-4 border-t border-white/[0.06] text-center">
                  <Link
                    href={ROUTE_LOGIN}
                    className="text-sm text-white/50 hover:text-white transition-colors"
                  >
                    Sign in instead
                  </Link>
                </div>
              </div>
            </div>
          )}

          {!existingEmail && (
            <form onSubmit={handleSubmit(onSubmit)} className="space-y-4 sm:space-y-6">
              <div>
                <label className="block text-xs font-bold text-[#6EE7B7] mb-2 uppercase tracking-widest font-mono-dm"htmlFor="name">
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
                <label className="block text-xs font-bold text-[#6EE7B7] mb-2 uppercase tracking-widest font-mono-dm"htmlFor="email">
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
                <label className="block text-xs font-bold text-[#6EE7B7] mb-2 uppercase tracking-widest font-mono-dm"htmlFor="password">
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
                <label className="block text-xs font-bold text-[#6EE7B7] mb-2 uppercase tracking-widest font-mono-dm"htmlFor="confirmPassword">
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

              <TurnstileWidget />
              <button
                type="submit"
                disabled={pending}
                className="btn-primary cursor-pointer flex w-full justify-center items-center gap-2 rounded-xl px-4 py-3.5 text-sm font-bold uppercase tracking-wider transition-all disabled:opacity-50 disabled:cursor-not-allowed"
              >
                {pending ? <Loader2 className="animate-spin h-5 w-5" /> : (
                  <>{buttonLabel || "Sign up"} <ArrowRight size={18} /></>
                )}
              </button>
            </form>
          )}

          {!existingEmail && (
            <>
              <div className="relative my-4 sm:my-6 flex items-center">
                <div className="flex-1 border-t border-white/[0.06]" />
                <span className="mx-4 text-xs font-mono-dm uppercase tracking-[0.2em] text-white/20">or</span>
                <div className="flex-1 border-t border-white/[0.06]" />
              </div>

              <a
                href={googleLoginURL || API_AUTH_GOOGLE_LOGIN}
                className="cursor-pointer flex w-full items-center justify-center gap-3 rounded-xl border border-white/[0.12] bg-white/[0.04] px-4 py-3.5 text-sm font-bold text-white uppercase tracking-wider hover:bg-white/[0.08] hover:border-white/20 hover:-translate-y-0.5 transition-all"
              >
                <svg width="20"height="20"viewBox="0 0 24 24"className="shrink-0">
                  <path d="M22.56 12.25c0-.78-.07-1.53-.2-2.25H12v4.26h5.92a5.06 5.06 0 0 1-2.2 3.32v2.77h3.57c2.08-1.92 3.28-4.74 3.28-8.1z" fill="#4285F4" />
                  <path d="M12 23c2.97 0 5.46-.98 7.28-2.66l-3.57-2.77c-.98.66-2.23 1.06-3.71 1.06-2.86 0-5.29-1.93-6.16-4.53H2.18v2.84C3.99 20.53 7.7 23 12 23z" fill="#34A853" />
                  <path d="M5.84 14.09c-.22-.66-.35-1.36-.35-2.09s.13-1.43.35-2.09V7.07H2.18C1.43 8.55 1 10.22 1 12s.43 3.45 1.18 4.93l2.85-2.22.81-.62z" fill="#FBBC05" />
                  <path d="M12 5.38c1.62 0 3.06.56 4.21 1.64l3.15-3.15C17.45 2.09 14.97 1 12 1 7.7 1 3.99 3.47 2.18 7.07l3.66 2.84c.87-2.6 3.3-4.53 6.16-4.53z" fill="#EA4335" />
                </svg>
                Sign up with Google
              </a>
            </>
          )}
        </div>

        {footer && (
          <div className="bg-white/[0.02] px-4 sm:px-8 py-3 sm:py-5 border-t border-white/[0.05]">
            {footer}
          </div>
        )}
      </div>

      <Dialog
        open={showUpgradeConfirm}
        onClose={() => setShowUpgradeConfirm(false)}
        title="Upgrade to Advertiser?"
        message="You will be able to create ad campaigns, run ads on your links, and access campaign analytics. This action can be reversed later."
        confirmLabel="Upgrade"
        icon={<Megaphone size={20} className="text-white" />}
        loading={pending}
        onConfirm={() => {
          setShowUpgradeConfirm(false);
          handleUpgrade();
        }}
      />

      <Dialog
        open={showDowngradeConfirm}
        onClose={() => setShowDowngradeConfirm(false)}
        title="Downgrade to Regular User?"
        message="All your active ad campaigns will be paused immediately. You will not be able to create or manage campaigns as a regular user. Your links, wallet balance, and other data will be preserved."
        confirmLabel="Downgrade & Pause Ads"
        icon={<AlertTriangle size={20} className="text-white" />}
        loading={pending}
        onConfirm={() => {
          setShowDowngradeConfirm(false);
          handleDowngrade();
        }}
      />
    </div>
  );
}
