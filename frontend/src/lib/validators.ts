import { z } from "zod";

export const loginSchema = z.object({
  email: z.email("Invalid email address"),
  password: z.string().min(1, "Password is required"),
});

export const registerSchema = z.object({
  name: z.string().min(2, "Name must be at least 2 characters"),
  email: z.email("Invalid email address"),
  password: z.string().min(6, "Password must be at least 6 characters"),
  confirmPassword: z.string().min(6, "Confirm password must be at least 6 characters"),
}).refine((data) => data.password === data.confirmPassword, {
  message: "Passwords don't match",
  path: ["confirmPassword"],
});

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
  expires_value: z.number().min(1, "Minimum value is 1").optional().or(z.nan()),
  expires_unit: z.enum(["minutes", "hours", "days"]).optional(),
}).refine(
  (data) => {
    if (data.expires_value && !isNaN(data.expires_value) && !data.expires_unit) {
      return false;
    }
    return true;
  },
  { message: "Please select a unit", path: ["expires_unit"] }
);
