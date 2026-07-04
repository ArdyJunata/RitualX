export default function AuthLayout({ children }: { children: React.ReactNode }) {
  return (
    <div className="min-h-screen bg-zinc-950 flex items-center justify-center py-12">
      <div className="animate-auth-fade w-full">
        {children}
      </div>
    </div>
  )
}
