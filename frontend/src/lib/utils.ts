import { type ClassValue, clsx } from "clsx"
import { twMerge } from "tailwind-merge"

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs))
}

export function formatCurrency(amount: string | number | undefined, currency?: string): string {
  if (amount === undefined || amount === null || amount === "") {
    return "$0.00"
  }

  const numAmount = typeof amount === "string" ? parseFloat(amount) : amount

  if (isNaN(numAmount)) {
    console.warn("formatCurrency received NaN value:", amount)
    return "$0.00"
  }

  const currencyCode = currency && currency.trim() !== "" ? currency : "USD"

  return new Intl.NumberFormat("en-US", {
    style: "currency",
    currency: currencyCode,
    minimumFractionDigits: 2,
    maximumFractionDigits: 2,
  }).format(numAmount)
}

export function formatNumber(value: string | number): string {
  const numValue = typeof value === "string" ? parseFloat(value) : value
  return new Intl.NumberFormat("en-US", {
    minimumFractionDigits: 2,
    maximumFractionDigits: 2,
  }).format(numValue)
}

export function formatDate(date: Date | string): string {
  const dateObj = typeof date === "string" ? new Date(date) : date
  return dateObj.toLocaleDateString("en-US", {
    year: "numeric",
    month: "short",
    day: "numeric",
  })
}

