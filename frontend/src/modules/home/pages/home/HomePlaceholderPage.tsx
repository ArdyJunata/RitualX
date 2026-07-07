import { homePlaceholderStyles as s } from "./HomePlaceholderPage.styles";

export function HomePlaceholderPage() {
  return (
    <div className={s.root}>
      <div className={s.greeting}>
        <h1 className={s.title}>Good morning 👋</h1>
        <p className={s.subtitle}>Your daily rituals await</p>
      </div>

      <div className={s.card}>
        <p className={s.cardTitle}>Today&apos;s Goal</p>
        <p className={s.cardBody}>Track your routines to build your streak</p>
        <div className={s.pillRow}>
          <span className={s.pill}>🏃 Morning Run</span>
          <span className={s.pill}>📚 Reading</span>
          <span className={s.pill}>💧 Hydration</span>
        </div>
      </div>

      <div className={s.card}>
        <p className={s.cardTitle}>Heatmap</p>
        <p className={s.cardBody}>GitHub-style contribution calendar — coming soon</p>
      </div>
    </div>
  );
}
