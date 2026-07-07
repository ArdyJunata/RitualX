import { questPageStyles as s } from "./QuestPage.styles";

const MOCK_QUESTS = [
  { icon: "⚡", name: "Log 3 routines today",  reward: "+50 XP",  progress: "0/3"  },
  { icon: "🔥", name: "Reach a 7-day streak",  reward: "+100 XP", progress: "0/7"  },
  { icon: "🏆", name: "Complete 10 total logs", reward: "+200 XP", progress: "0/10" },
] as const;

export function QuestPage() {
  return (
    <div className={s.root}>
      <div className={s.header}>
        <h1 className={s.title}>Quests</h1>
        <p className={s.subtitle}>Complete challenges to earn XP and rewards</p>
      </div>

      <div className={s.card}>
        <p className={s.cardTitle}>Active Quests</p>
        {MOCK_QUESTS.map((quest) => (
          <div key={quest.name} className={s.questItem}>
            <span className={s.questIcon}>{quest.icon}</span>
            <div className={s.questInfo}>
              <span className={s.questName}>{quest.name}</span>
              <span className={s.questReward}>{quest.reward}</span>
            </div>
            <span className={s.questProgress}>{quest.progress}</span>
          </div>
        ))}
      </div>
    </div>
  );
}
