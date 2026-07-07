export const profilePageStyles = {
  root: "flex flex-col gap-6 py-4",
  avatarSection: "flex flex-col items-center gap-3 pt-4",
  avatar:
    "w-20 h-20 rounded-full bg-emerald-600 flex items-center justify-center text-white text-3xl font-bold",
  displayName: "font-heading text-xl font-bold text-zinc-100",
  userTitle: "text-sm text-zinc-500",
  statsRow: "flex gap-4 justify-center",
  statItem: "flex flex-col items-center gap-0.5",
  statValue: "font-heading text-lg font-bold text-zinc-100",
  statLabel: "text-xs text-zinc-500",
  card: "glass-card p-5 flex flex-col gap-3",
  cardTitle: "text-base font-semibold text-zinc-200",
  menuItem:
    "flex items-center justify-between py-2 border-b border-white/5 last:border-0 text-sm text-zinc-300",
  menuArrow: "text-zinc-600",
} as const;
