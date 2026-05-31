import { z } from "zod";
import {
  ROLE_ADVERTISER,
  EXPIRY_UNITS,
} from "./constants";

export const loginSchema = z.object({
  email: z.email("Invalid email address"),
  password: z.string().min(1, "Password is required"),
});

const passwordMatchRefine = z.object({
  password: z.string(),
  confirmPassword: z.string(),
}).refine((data) => data.password === data.confirmPassword, {
  message: "Passwords don't match",
  path: ["confirmPassword"],
});

const baseRegisterSchema = z.object({
  name: z.string().min(2, "Name must be at least 2 characters"),
  email: z.email("Invalid email address"),
  password: z.string().min(6, "Password must be at least 6 characters"),
  confirmPassword: z.string().min(6, "Confirm password must be at least 6 characters"),
});

export const registerSchema = baseRegisterSchema.and(passwordMatchRefine);

export const advertiserRegisterSchema = baseRegisterSchema.extend({
  role: z.literal(ROLE_ADVERTISER),
}).and(passwordMatchRefine);

const expiresValue = z.union([z.number().min(1, "Minimum value is 1"), z.nan()]).optional();

const expiresCheck = (data: { expires_value?: number; expires_unit?: string }) => {
  if (data.expires_value && !isNaN(data.expires_value) && !data.expires_unit) {
    return false;
  }
  return true;
};

export const linkSchema = z.object({
  url: z.string()
    .min(1, "Please enter a URL")
    .refine((v) => {
      try {
        const withProto = /^https?:\/\//i.test(v) ? v : `https://${v}`;
        new URL(withProto);
        return withProto.includes(".");
      } catch {
        return false;
      }
    }, "Please enter a valid URL"),
  custom_slug: z.string().regex(/^[a-zA-Z0-9-]*$/, "Only letters, numbers, and dashes allowed").min(3).max(20).optional().or(z.literal("")),
  expires_value: expiresValue,
  expires_unit: z.enum(EXPIRY_UNITS).optional(),
}).refine(expiresCheck, { message: "Please select a unit", path: ["expires_unit"] });

export const editSchema = z.object({
  custom_slug: z.string().regex(/^[a-zA-Z0-9-]*$/, "Only letters, numbers, and dashes allowed").min(3).max(20).optional().or(z.literal("")),
  expires_value: expiresValue,
  expires_unit: z.enum(EXPIRY_UNITS).optional(),
}).refine(expiresCheck, { message: "Please select a unit", path: ["expires_unit"] });

export type EditForm = z.infer<typeof editSchema>;


export const createAdSchema = z.object({
  title: z.string().min(3, "Title must be at least 3 characters"),
  description: z.string().optional(),
  image_url: z.string().min(1, "Ad creative media file is required"),
  target_url: z.url("Must be a valid URL"),
  category: z.string().min(1, "Category is required"),
  total_budget: z.number().min(1, "Minimum budget is 1"),
  ad_type: z.string().min(1, "Ad format type is required"),
});

export const forgotPasswordSchema = z.object({
  email: z.email("Invalid email address"),
});

export const resetPasswordSchema = z.object({
  token: z.string().min(1, "Reset token is required"),
  password: z.string().min(6, "Password must be at least 6 characters"),
  confirmPassword: z.string().min(6, "Confirm password must be at least 6 characters"),
}).refine((data) => data.password === data.confirmPassword, {
  message: "Passwords don't match",
  path: ["confirmPassword"],
});

export const sendVerificationSchema = z.object({
  email: z.email("Invalid email address"),
});

export type CreateAdForm = z.infer<typeof createAdSchema>;
