import Link from "next/link"
import { useRouter } from "next/router"
import { LayoutDashboard, FolderTree, Server } from "lucide-react"
import { cn } from "@/lib/utils"
import { DateRangePicker } from "@/components/date-range-picker"
import { useDateRange } from "@/context/date-range-context"

export function Header() {
  const router = useRouter()
  const { dateRange, setDateRange } = useDateRange()

  const navItems = [
    { href: "/", label: "Dashboard", icon: LayoutDashboard },
    { href: "/products", label: "Products", icon: FolderTree },
    { href: "/platform", label: "Platform & Shared", icon: Server },
  ]

  return (
    <header className="sticky top-0 z-50 w-full border-b bg-background/95 backdrop-blur supports-[backdrop-filter]:bg-background/60">
      <div className="container flex h-16 items-center justify-between gap-4">
        <div className="flex items-center gap-8">
          <Link href="/" className="flex items-center space-x-2 hover:opacity-80 transition-opacity">
            <div className="flex h-8 w-8 items-center justify-center rounded-lg bg-primary text-primary-foreground">
              <span className="text-lg font-bold">F</span>
            </div>
            <span className="hidden text-xl font-bold sm:inline">FinOps Aggregator</span>
          </Link>
          <nav className="flex items-center space-x-6 text-sm font-medium">
            {navItems.map((item) => {
              const Icon = item.icon
              const isActive = router.pathname === item.href
              return (
                <Link
                  key={item.href}
                  href={item.href}
                  className={cn(
                    "flex items-center gap-2 transition-colors hover:text-foreground/80",
                    isActive ? "text-foreground" : "text-foreground/60"
                  )}
                >
                  <Icon className="h-4 w-4" />
                  <span className="hidden sm:inline">{item.label}</span>
                </Link>
              )
            })}
          </nav>
        </div>
        <div className="flex items-center">
          <DateRangePicker
            value={dateRange}
            onChange={(range) => {
              if (!range.from || !range.to) return
              setDateRange({ from: range.from, to: range.to })
            }}
          />
        </div>
      </div>
    </header>
  )
}

