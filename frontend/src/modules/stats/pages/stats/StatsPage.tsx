import { statsPageStyles as s } from "./StatsPage.styles";
import { RoutineHeatmaps } from "../../components";

const MOCK_STATS = [
  { value: "0",  label: "Total Logs" },
  { value: "0",  label: "Best Streak" },
  { value: "0",  label: "XP Earned" },
  { value: "0",  label: "Quests Done" },
] as const;

export function StatsPage() {
  return (
    <div className={s.root}>
      <div className={s.header}>
        <h1 className={s.title}>Statistics</h1>
        <p className={s.subtitle}>Your progress at a glance</p>
      </div>

      <div className={s.grid}>
        {MOCK_STATS.map((stat) => (
          <div key={stat.label} className={s.statCard}>
            <span className={s.statValue}>{stat.value}</span>
            <span className={s.statLabel}>{stat.label}</span>
          </div>
        ))}
      </div>

      <section className={s.section}>
        <h2 className={s.sectionTitle}>Contribution Heatmaps</h2>
        <RoutineHeatmaps />
      </section>
    </div>
  );
}
