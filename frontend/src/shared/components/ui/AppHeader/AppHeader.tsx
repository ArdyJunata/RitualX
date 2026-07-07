import { appHeaderStyles as s } from "./AppHeader.styles";

export interface IAppHeaderProps {
  displayName: string;
  level: number;
  xp: number;
  xpToNextLevel: number;
  streakCount: number;
}

export function AppHeader({
  displayName,
  level,
  xp,
  xpToNextLevel,
  streakCount,
}: IAppHeaderProps) {
  const initials = displayName.slice(0, 2).toUpperCase();
  const xpPercent = Math.min(100, Math.round((xp / xpToNextLevel) * 100));

  return (
    <header className={s.root}>
      <div className={s.inner}>
        {/* Top row: avatar + name | streak */}
        <div className={s.topRow}>
          <div className={s.leftGroup}>
            <div className={s.avatar} aria-hidden="true">
              {initials}
            </div>
            <span className={s.appName}>RitualX</span>
          </div>
          <div className={s.streakGroup} aria-label={`${streakCount} day streak`}>
            <span className={s.streakIcon}>🔥</span>
            <span className={s.streakCount}>{streakCount}</span>
          </div>
        </div>

        {/* XP row: level | bar | xp text */}
        <div className={s.xpRow}>
          <span className={s.levelBadge} aria-label={`Level ${level}`}>
            Lv.{level}
          </span>
          <div
            className={s.xpTrack}
            role="progressbar"
            aria-valuenow={xp}
            aria-valuemin={0}
            aria-valuemax={xpToNextLevel}
            aria-label="XP progress"
          >
            <div
              className={s.xpFill}
              style={{ width: `${xpPercent}%` }}
            />
          </div>
          <span className={s.xpText}>{xp}/{xpToNextLevel} XP</span>
        </div>
      </div>
    </header>
  );
}
