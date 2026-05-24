import { useMutation, useQueryClient } from "@tanstack/react-query";
import { api } from "@/lib/api";
import { loginSchema, registerSchema, advertiserRegisterSchema } from "@/lib/validators";
import { z } from "zod";
import { AxiosError } from "axios";
import { ApiErrorResponse } from "@/types/api";
import { toast } from "sonner";
import { useUserStore } from "@/store/useUserStore";
import { useRouter } from "next/navigation";

import { ROLE_ADVERTISER, ROLE_ADMIN, API_AUTH_LOGIN, API_AUTH_REGISTER, API_AUTH_LOGOUT, ROUTE_CAMPAIGNS, ROUTE_ADMIN_DASHBOARD, ROUTE_LINKS, ROUTE_LOGIN } from "@/lib/constants";

type LoginForm = z.infer<typeof loginSchema>;

type RegisterForm = z.infer<typeof registerSchema>;
type AdvertiserRegisterForm = z.infer<typeof advertiserRegisterSchema>;

type AuthResponse = { data: { access_token: string; refresh_token: string; user: { id: string; email: string; name: string; role: string; created_at: string } } };

export function useLogin() {
  const router = useRouter();
  const setUser = useUserStore((state) => state.setUser);
  
  return useMutation<AuthResponse, AxiosError<ApiErrorResponse>, LoginForm>({
    mutationFn: async (data: LoginForm) => {
      const response = await api.post<AuthResponse>(API_AUTH_LOGIN, data);
      return response.data;
    },
    onSuccess: (res) => {
      const { user } = res.data;
      setUser({
        id: user.id,
        email: user.email,
        name: user.name,
        role: user.role,
        created_at: user.created_at,
      });
      toast.success("Welcome back!");
      if (user.role === ROLE_ADVERTISER) {
        router.push(ROUTE_CAMPAIGNS);
      } else if (user.role === ROLE_ADMIN) {
        router.push(ROUTE_ADMIN_DASHBOARD);
      } else {
        router.push(ROUTE_LINKS);
      }
    },
  });
}

export function useRegister() {
  const router = useRouter();
  return useMutation<void, AxiosError<ApiErrorResponse>, RegisterForm>({
    mutationFn: async (data: RegisterForm) => {
      const response = await api.post(API_AUTH_REGISTER, data);
      return response.data;
    },
    onSuccess: () => {
      toast.success("Account created! Please sign in.");
      router.push(ROUTE_LOGIN);
    },
  });
}

export function useRegisterAdvertiser() {
  const router = useRouter();
  return useMutation<void, AxiosError<ApiErrorResponse>, AdvertiserRegisterForm>({
    mutationFn: async (data: AdvertiserRegisterForm) => {
      const response = await api.post(API_AUTH_REGISTER, data);
      return response.data;
    },
    onSuccess: () => {
      toast.success("Advertiser account created! Sign in to manage campaigns.");
      router.push(ROUTE_LOGIN);
    },
  });
}

export function useLogout() {
  const router = useRouter();
  const clearUser = useUserStore((state) => state.clearUser);
  const queryClient = useQueryClient();
  
  return useMutation({
    mutationFn: async () => {
      await api.post(API_AUTH_LOGOUT);
    },
    onSuccess: () => {
      toast.success("Signed out successfully");
    },
    onSettled: () => {
      clearUser();
      queryClient.clear();
      router.push(ROUTE_LOGIN);
    },
  });
}
