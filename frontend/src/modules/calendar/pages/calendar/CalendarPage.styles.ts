export const calendarPageStyles = {
  root: "flex flex-col gap-6 py-4",
  header: "flex flex-col gap-1",
  title: "font-heading text-2xl font-bold text-zinc-100",
  subtitle: "text-sm text-zinc-500",
  card: "glass-card p-5 flex flex-col gap-3",
  cardTitle: "text-base font-semibold text-zinc-200",
  cardBody: "text-sm text-zinc-500",
  grid: "grid grid-cols-7 gap-1",
  cell: "aspect-square rounded-sm bg-zinc-800/60",
  cellActive: "aspect-square rounded-sm bg-emerald/60",
} as const;
