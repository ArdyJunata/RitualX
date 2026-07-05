export const bottomNavStyles = {
  container:
    "fixed bottom-6 left-1/2 -translate-x-1/2 z-50 flex items-center gap-6 px-6 py-3 rounded-full bg-zinc-900/70 backdrop-blur-xl border border-white/10 shadow-lg shadow-black/40",
  tab: "flex flex-col items-center gap-0.5 transition-colors duration-200",
  icon: "w-6 h-6",
  iconActive: "text-emerald",
  iconInactive: "text-zinc-500",
  dot: "w-1.5 h-1.5 rounded-full",
  dotActive: "bg-emerald",
  dotInactive: "bg-transparent",
} as const;
