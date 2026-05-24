// Date: 2026-05-25
// Author: XinYang Li

import * as React from "react";
import { cva, type VariantProps } from "class-variance-authority";

import { cn } from "@/lib/utils";

const buttonVariants = cva(
  "inline-flex items-center justify-center rounded-full border text-sm font-semibold transition duration-200 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-pine/30 disabled:pointer-events-none disabled:opacity-50",
  {
    variants: {
      variant: {
        primary: "border-pine bg-pine px-5 py-3 text-paper shadow-panel hover:-translate-y-0.5 hover:bg-[#164939]",
        secondary: "border-line bg-paper px-5 py-3 text-ink hover:border-pine/40 hover:bg-mist",
        ghost: "border-transparent bg-transparent px-4 py-2 text-ink hover:bg-mist",
      },
      size: {
        default: "h-11",
        sm: "h-9 text-xs",
        icon: "h-12 w-12",
      },
    },
    defaultVariants: {
      variant: "primary",
      size: "default",
    },
  },
);

export interface ButtonProps
  extends React.ButtonHTMLAttributes<HTMLButtonElement>,
    VariantProps<typeof buttonVariants> {}

/**
 * Renders a reusable button with shared AgentHub motion and visual states.
 * @param props Standard button props plus size and variant options.
 * @returns The styled button element.
 */
export function Button({ className, variant, size, ...props }: ButtonProps): JSX.Element {
  return <button className={cn(buttonVariants({ variant, size }), className)} {...props} />;
}
