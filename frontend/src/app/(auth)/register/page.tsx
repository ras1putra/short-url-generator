"use client";

import RegisterForm from "@/components/auth/RegisterForm";
import { ROUTE_LOGIN } from "@/lib/constants";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { toast } from "sonner";

export default function RegisterPage() {
  const router = useRouter();

  return (
    <RegisterForm
      title="CREATE ACCOUNT."
      subtitle="// Join the speed of redirects"
      onSuccess={async () => {
        toast.success("Account created! Check your email to verify.");
        router.push(ROUTE_LOGIN);
      }}
      footer={
        <>
          <p className="text-center text-sm text-white/40">
            Already have an account?{" "}
            <Link href={ROUTE_LOGIN} className="font-bold text-[#6EE7B7] hover:text-[#A7F3D0] transition-colors">
              Sign in
            </Link>
          </p>
          <p className="text-center text-sm text-white/40 mt-2">
            Want an advertiser account?{" "}
            <Link href="/register/advertiser"className="font-bold text-white/60 hover:text-white transition-colors">
              Sign up here
            </Link>
          </p>
        </>
      }
    />
  );
}
