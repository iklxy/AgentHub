// Date: 2026-05-25
// Author: XinYang Li

import type { Config } from "tailwindcss";

const config: Config = {
  content: ["./src/**/*.{ts,tsx}"],
  theme: {
    extend: {
      colors: {
        ink: "#153026",
        mist: "#f3f1e8",
        moss: "#cad8cb",
        pine: "#1d5745",
        ember: "#b85d3d",
        line: "#d5d1c7",
        paper: "#fcfaf5",
      },
      fontFamily: {
        display: ['"Baskervville"', '"Times New Roman"', "serif"],
        body: ['"Manrope"', '"Helvetica Neue"', "sans-serif"],
      },
      boxShadow: {
        panel: "0 18px 40px rgba(25, 55, 42, 0.08)",
      },
      borderRadius: {
        xl2: "1.5rem",
      },
      keyframes: {
        rise: {
          "0%": { opacity: "0", transform: "translateY(14px)" },
          "100%": { opacity: "1", transform: "translateY(0)" },
        },
        pulseLine: {
          "0%, 100%": { opacity: "0.35" },
          "50%": { opacity: "0.9" },
        },
      },
      animation: {
        rise: "rise 420ms ease-out both",
        pulseLine: "pulseLine 2.2s ease-in-out infinite",
      },
    },
  },
  plugins: [],
};

export default config;
