// Date: 2026-05-25
// Author: XinYang Li

import { clsx, type ClassValue } from "clsx";
import { twMerge } from "tailwind-merge";

/**
 * Merges conditional class names into one Tailwind-safe string.
 * @param inputs Class values that may contain arrays, objects, or strings.
 * @returns The merged class name string.
 */
export function cn(...inputs: ClassValue[]): string {
  return twMerge(clsx(inputs));
}
