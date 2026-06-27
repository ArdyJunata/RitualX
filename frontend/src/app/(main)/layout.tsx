import QueryProvider from "@/components/providers/QueryProvider";

export default function MainLayout({ children }: { children: React.ReactNode }) {
  return (
    <QueryProvider>
      <div className="min-h-screen pb-20">
        {children}
        {/* Bottom Nav will go here later */}
      </div>
    </QueryProvider>
  );
}
