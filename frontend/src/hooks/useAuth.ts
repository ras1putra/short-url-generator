import { useMutation, useQueryClient } from "@tanstack/react-query";
import { api } from "@/lib/api";
import { loginSchema, registerSchema } from "@/lib/validators";
import { z } from "zod";
import { AxiosError } from "axios";
import { ApiErrorResponse } from "@/types/api";
import { toast } from "sonner";
import { useUserStore } from "@/store/useUserStore";
import { useRouter } from "next/navigation";

type LoginForm = z.infer<typeof loginSchema>;
type RegisterForm = z.infer<typeof registerSchema>;

type AuthResponse = { data: { id: number; email: string; name: string; created_at: string } };

export function useLogin() {
  const router = useRouter();
  const setUser = useUserStore((state) => state.setUser);
  
  return useMutation<AuthResponse, AxiosError<ApiErrorResponse>, LoginForm>({
    mutationFn: async (data: LoginForm) => {
      const response = await api.post<AuthResponse>("/api/auth/login", data);
      return response.data;
    },
    onSuccess: (res) => {
      const user = res.data;
      setUser({
        id: String(user.id),
        email: user.email,
        name: user.name,
        created_at: user.created_at,
      });
      toast.success("Welcome back!");
      router.push("/dashboard");
    },
  });
}

export function useRegister() {
  const router = useRouter();
  return useMutation<void, AxiosError<ApiErrorResponse>, RegisterForm>({
    mutationFn: async (data: RegisterForm) => {
      const response = await api.post("/api/auth/register", data);
      return response.data;
    },
    onSuccess: () => {
      toast.success("Account created! Please sign in.");
      router.push("/login");
    },
  });
}

export function useLogout() {
  const router = useRouter();
  const clearUser = useUserStore((state) => state.clearUser);
  const queryClient = useQueryClient();
  
  return useMutation({
    mutationFn: async () => {
      await api.post("/api/auth/logout");
    },
    onSuccess: () => {
      toast.success("Signed out successfully");
    },
    onSettled: () => {
      clearUser();
      queryClient.clear();
      router.push("/login");
    },
  });
}
