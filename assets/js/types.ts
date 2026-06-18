export type ToastVariant = "info" | "success" | "error";
export interface ToastOptions {
  message: string;
  variant?: ToastVariant;
  timeout?: number;
}
