import type { ReactNode } from "react";
import { BottomNav } from "@/shared/components/ui/BottomNav";
import { AppHeader } from "@/shared/components/ui/AppHeader";

export default function MainLayout({ children }: { children: ReactNode }) {
  return (
    <div className="min-h-screen bg-background">
      <AppHeader
        displayName="You"
        level={1}
        xp={0}
        xpToNextLevel={100}
        streakCount={0}
      />
      <main className="max-w-[430px] mx-auto px-4 pt-4 pb-32">
        {children}
      </main>
      <BottomNav />
    </div>
  );
}

