"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import { Home, CalendarDays, BarChart2, Sword, User } from "lucide-react";
import { bottomNavStyles as s } from "./BottomNav.styles";

const NAV_TABS = [
  { label: "Home",     href: "/",         Icon: Home         },
  { label: "Calendar", href: "/calendar", Icon: CalendarDays },
  { label: "Stats",    href: "/stats",    Icon: BarChart2    },
  { label: "Quest",    href: "/quest",    Icon: Sword        },
  { label: "Profile",  href: "/profile",  Icon: User         },
] as const;

function isActive(pathname: string, href: string): boolean {
  if (href === "/") return pathname === "/";
  return pathname.startsWith(href);
}

export function BottomNav() {
  const pathname = usePathname();

  return (
    <nav aria-label="Main navigation" className={s.container}>
      {NAV_TABS.map(({ label, href, Icon }) => {
        const active = isActive(pathname, href);
        return (
          <Link
            key={href}
            href={href}
            aria-label={label}
            title={label}
            aria-current={active ? "page" : undefined}
            className={s.tab}
          >
            <Icon className={`${s.icon} ${active ? s.iconActive : s.iconInactive}`} />
            <span className={`${s.dot} ${active ? s.dotActive : s.dotInactive}`} />
          </Link>
        );
      })}
    </nav>
  );
}
