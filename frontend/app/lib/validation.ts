import { z } from 'zod';

/**
 * NeoBank Form Validation Schemas
 * Using Zod for type-safe form validation
 */

// Common validators
const email = z.string().email('Invalid email address');
const password = z.string()
  .min(8, 'Password must be at least 8 characters')
  .regex(/[A-Z]/, 'Password must contain at least one uppercase letter')
  .regex(/[a-z]/, 'Password must contain at least one lowercase letter')
  .regex(/[0-9]/, 'Password must contain at least one number');

const uuid = z.string().uuid('Invalid ID format');

const amount = z.string()
  .regex(/^[0-9]+(\.[0-9]{1,2})?$/, 'Invalid amount format')
  .refine((val) => parseFloat(val) > 0, 'Amount must be greater than 0');

const currency = z.enum(['USD', 'EUR', 'GBP', 'CAD'], {
  errorMap: () => ({ message: 'Invalid currency' }),
});

// Auth Schemas
export const loginSchema = z.object({
  email,
  password: z.string().min(1, 'Password is required'),
});

export const registerSchema = z.object({
  email,
  password,
  confirmPassword: z.string(),
  firstName: z.string().min(1, 'First name is required').max(50),
  lastName: z.string().min(1, 'Last name is required').max(50),
}).refine((data) => data.password === data.confirmPassword, {
  message: "Passwords don't match",
  path: ['confirmPassword'],
});

export const forgotPasswordSchema = z.object({
  email,
});

export const resetPasswordSchema = z.object({
  password,
  confirmPassword: z.string(),
  token: z.string().min(1),
}).refine((data) => data.password === data.confirmPassword, {
  message: "Passwords don't match",
  path: ['confirmPassword'],
});

// Account Schemas
export const createAccountSchema = z.object({
  accountType: z.enum(['CHECKING', 'SAVINGS', 'INVESTMENT'], {
    errorMap: () => ({ message: 'Select an account type' }),
  }),
  currency,
});

// Transfer Schemas
export const transferSchema = z.object({
  fromAccountId: uuid,
  toAccountId: uuid,
  amount,
  currency,
  description: z.string().max(500).optional(),
}).refine((data) => data.fromAccountId !== data.toAccountId, {
  message: 'Cannot transfer to the same account',
  path: ['toAccountId'],
});

// Card Schemas
export const issueCardSchema = z.object({
  accountId: uuid,
});

// Product Schemas
export const createProductSchema = z.object({
  name: z.string().min(1, 'Name is required').max(100),
  description: z.string().max(500).optional(),
  type: z.enum(['CHECKING', 'SAVINGS', 'INVESTMENT', 'CREDIT']),
  active: z.boolean().default(true),
});

// Profile Schemas
export const updateProfileSchema = z.object({
  firstName: z.string().min(1).max(50).optional(),
  lastName: z.string().min(1).max(50).optional(),
  email: email.optional(),
});

export const changePasswordSchema = z.object({
  currentPassword: z.string().min(1, 'Current password is required'),
  newPassword: password,
  confirmPassword: z.string(),
}).refine((data) => data.newPassword !== data.currentPassword, {
  message: 'New password must be different from current password',
  path: ['newPassword'],
}).refine((data) => data.newPassword === data.confirmPassword, {
  message: "Passwords don't match",
  path: ['confirmPassword'],
});

// Type exports
export type LoginInput = z.infer<typeof loginSchema>;
export type RegisterInput = z.infer<typeof registerSchema>;
export type ForgotPasswordInput = z.infer<typeof forgotPasswordSchema>;
export type ResetPasswordInput = z.infer<typeof resetPasswordSchema>;
export type CreateAccountInput = z.infer<typeof createAccountSchema>;
export type TransferInput = z.infer<typeof transferSchema>;
export type IssueCardInput = z.infer<typeof issueCardSchema>;
export type CreateProductInput = z.infer<typeof createProductSchema>;
export type UpdateProfileInput = z.infer<typeof updateProfileSchema>;
export type ChangePasswordInput = z.infer<typeof changePasswordSchema>;

// Validation helper
export function validateForm<T>(schema: z.ZodSchema<T>, data: unknown): { success: true; data: T } | { success: false; errors: Record<string, string> } {
  const result = schema.safeParse(data);
  
  if (result.success) {
    return { success: true, data: result.data };
  }
  
  const errors: Record<string, string> = {};
  result.error.errors.forEach((err) => {
    const path = err.path.join('.');
    if (!errors[path]) {
      errors[path] = err.message;
    }
  });
  
  return { success: false, errors };
}
