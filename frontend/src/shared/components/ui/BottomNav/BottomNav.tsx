"use client";

import { useState } from "react";
import Link from "next/link";
import { usePathname } from "next/navigation";
import { Home, CalendarDays, Sword, User, Plus } from "lucide-react";
import { bottomNavStyles as s } from "./BottomNav.styles";
import { CreateRoutineSheet } from "@/modules/routines";

const LEFT_TABS = [
  { label: "Home",     href: "/",         Icon: Home         },
  { label: "Calendar", href: "/calendar", Icon: CalendarDays },
] as const;

const RIGHT_TABS = [
  { label: "Quest",   href: "/quest",   Icon: Sword },
  { label: "Profile", href: "/profile", Icon: User  },
] as const;

function isActive(pathname: string, href: string): boolean {
  if (href === "/") return pathname === "/";
  return pathname.startsWith(href);
}

export function BottomNav() {
  const pathname = usePathname();
  const [sheetOpen, setSheetOpen] = useState(false);

  return (
    <>
      <nav aria-label="Main navigation" className={s.container}>
        {/* Left tabs */}
        {LEFT_TABS.map(({ label, href, Icon }) => {
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

        {/* FAB center button */}
        <button
          type="button"
          aria-label="Create routine"
          onClick={() => setSheetOpen(true)}
          className="flex items-center justify-center w-12 h-12 rounded-full bg-emerald-500 shadow-lg shadow-emerald-500/40 -mt-5 transition-transform duration-150 active:scale-95"
        >
          <Plus size={22} className="text-white" />
        </button>

        {/* Right tabs */}
        {RIGHT_TABS.map(({ label, href, Icon }) => {
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

      <CreateRoutineSheet
        isOpen={sheetOpen}
        onClose={() => {
          setSheetOpen(false)
          window.dispatchEvent(new CustomEvent('routineCreated'))
        }}
      />
    </>
  );
}
