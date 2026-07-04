import type { Config } from "tailwindcss";

const config: Config = {
  content: [
    "./src/pages/**/*.{js,ts,jsx,tsx,mdx}",
    "./src/components/**/*.{js,ts,jsx,tsx,mdx}",
    "./src/app/**/*.{js,ts,jsx,tsx,mdx}",
    "./src/modules/**/*.{js,ts,jsx,tsx,mdx}",
    "./src/shared/**/*.{js,ts,jsx,tsx,mdx}",
  ],
  theme: {
    extend: {
      colors: {
        background: "var(--background)",
        foreground: "var(--foreground)",
        emerald: {
          300: "#6EE7B7",
          400: "#34D399",
          600: "#059669",
          900: "#064E3B",
          DEFAULT: "#10B981", // Primary Accent
        },
        amber: {
          DEFAULT: "#F59E0B", // Streak Color
        },
        red: {
          DEFAULT: "#EF4444", // Punishment Color
        },
        purple: {
          DEFAULT: "#8B5CF6", // Gamification/XP Color
        }
      },
      fontFamily: {
        sans: ["var(--font-inter)", "sans-serif"],
        heading: ["var(--font-outfit)", "sans-serif"],
      }
    },
  },
  plugins: [],
};
export default config;
