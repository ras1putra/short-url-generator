"use client";

import RegisterForm from "@/components/auth/RegisterForm";
import { API_AUTH_GOOGLE_LOGIN, ROUTE_LOGIN } from "@/lib/constants";
import { toast } from "sonner";
import { useRouter } from "next/navigation";
import Link from "next/link";

export default function AdvertiserRegisterPage() {
  const router = useRouter();

  // eslint-disable-next-line @typescript-eslint/no-unused-vars
  const onSuccess = async (email: string, password: string) => {
    toast.success("Advertiser account created! Check your email to verify before signing in.");
    router.push(ROUTE_LOGIN);
  };

  return (
    <RegisterForm
      title="ADVERTISER SIGNUP."
      subtitle="// Start your ad campaigns today"
      buttonLabel="Create advertiser account"
      googleLoginURL={`${API_AUTH_GOOGLE_LOGIN}?intent=advertiser`}
      onSuccess={onSuccess}
      footer={
        <>
          <p className="text-center text-sm text-white/40">
            Already have an account?{" "}
            <Link href={ROUTE_LOGIN} className="font-bold text-[#6EE7B7] hover:text-[#A7F3D0] transition-colors">
              Sign in
            </Link>
          </p>
          <p className="text-center text-sm text-white/40 mt-2">
            Want a regular account?{" "}
            <Link href="/register"className="font-bold text-white/60 hover:text-white transition-colors">
              Sign up here
            </Link>
          </p>
        </>
      }
    />
  );
}
