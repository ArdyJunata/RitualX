import { profilePageStyles as s } from "./ProfilePage.styles";

const MENU_ITEMS = [
  "Edit Profile",
  "Notification Settings",
  "Privacy",
  "Linked Partners",
  "Sign Out",
] as const;

export function ProfilePage() {
  return (
    <div className={s.root}>
      <div className={s.avatarSection}>
        <div className={s.avatar}>YO</div>
        <p className={s.displayName}>You</p>
        <p className={s.userTitle}>Novice · Level 1</p>
        <div className={s.statsRow}>
          <div className={s.statItem}>
            <span className={s.statValue}>0</span>
            <span className={s.statLabel}>Routines</span>
          </div>
          <div className={s.statItem}>
            <span className={s.statValue}>0</span>
            <span className={s.statLabel}>Streak</span>
          </div>
          <div className={s.statItem}>
            <span className={s.statValue}>0</span>
            <span className={s.statLabel}>XP</span>
          </div>
        </div>
      </div>

      <div className={s.card}>
        <p className={s.cardTitle}>Settings</p>
        {MENU_ITEMS.map((item) => (
          <div key={item} className={s.menuItem}>
            <span>{item}</span>
            <span className={s.menuArrow}>›</span>
          </div>
        ))}
      </div>
    </div>
  );
}
