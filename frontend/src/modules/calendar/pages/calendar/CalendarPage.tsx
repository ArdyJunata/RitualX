import { calendarPageStyles as s } from "./CalendarPage.styles";

// 35 cells = 5 weeks × 7 days — mock heatmap preview
const MOCK_CELLS = Array.from({ length: 35 }, (_, i) => ({
  id: i,
  // Create a realistic-looking pattern
  active: [2, 3, 5, 7, 9, 10, 12, 14, 16, 17, 19, 21, 23, 24, 26, 28, 30, 31, 33].includes(i),
}));

export function CalendarPage() {
  return (
    <div className={s.root}>
      <div className={s.header}>
        <h1 className={s.title}>Calendar</h1>
        <p className={s.subtitle}>Contribution heatmap — coming soon</p>
      </div>

      <div className={s.card}>
        <p className={s.cardTitle}>Activity Heatmap Preview</p>
        <div className={s.grid}>
          {MOCK_CELLS.map((cell) => (
            <div
              key={cell.id}
              className={cell.active ? s.cellActive : s.cell}
              aria-hidden="true"
            />
          ))}
        </div>
        <p className={s.cardBody}>Full heatmap calendar — coming soon</p>
      </div>
    </div>
  );
}
