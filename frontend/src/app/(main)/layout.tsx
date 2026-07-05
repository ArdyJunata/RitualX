import type { ReactNode } from "react";
import { BottomNav } from "@/shared/components/ui/BottomNav";

export default function MainLayout({ children }: { children: ReactNode }) {
  return (
    <div className="min-h-screen pb-28">
      {children}
      <BottomNav />
    </div>
  );
}

