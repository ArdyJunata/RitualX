export const appHeaderStyles = {
  root: "sticky top-0 z-40 w-full bg-zinc-950/95 backdrop-blur-md border-b border-white/5",
  inner: "max-w-[430px] mx-auto px-4 py-2 flex flex-col gap-1.5",
  topRow: "flex items-center justify-between",
  // Left side: avatar + app name
  leftGroup: "flex items-center gap-2.5",
  avatar:
    "w-9 h-9 rounded-full bg-emerald-600 flex items-center justify-center text-white text-sm font-bold shrink-0",
  appName: "font-heading text-base font-semibold text-zinc-100 tracking-wide",
  // Right side: streak
  streakGroup: "flex items-center gap-1.5",
  streakIcon: "text-base leading-none select-none",
  streakCount: "text-sm font-semibold text-amber",
  // XP row
  xpRow: "flex items-center gap-2",
  levelBadge:
    "text-xs font-semibold text-purple px-1.5 py-0.5 rounded-full border border-purple/30 shrink-0",
  xpTrack: "flex-1 h-1 rounded-full bg-zinc-800 overflow-hidden",
  xpFill: "h-full rounded-full bg-emerald transition-[width] duration-500 ease-out",
  xpText: "text-xs text-zinc-500 shrink-0 tabular-nums",
} as const;
