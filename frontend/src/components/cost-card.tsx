import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { formatCurrency } from "@/lib/utils"
import { TrendingUp, TrendingDown, Minus } from "lucide-react"

interface CostCardProps {
  title: string
  amount: string
  currency?: string
  subtitle?: string
  trend?: {
    value: number
    direction: "up" | "down" | "neutral"
  }
  icon?: React.ReactNode
  showCurrency?: boolean
}

export function CostCard({
  title,
  amount,
  currency,
  subtitle,
  trend,
  icon,
  showCurrency = true,
}: CostCardProps) {
  const TrendIcon = trend
    ? trend.direction === "up"
      ? TrendingUp
      : trend.direction === "down"
      ? TrendingDown
      : Minus
    : null

  return (
    <Card>
      <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
        <CardTitle className="text-sm font-medium">{title}</CardTitle>
        {icon && <div className="text-muted-foreground">{icon}</div>}
      </CardHeader>
      <CardContent>
        <div className="text-2xl font-bold">
          {showCurrency ? formatCurrency(amount, currency) : amount}
        </div>
        {subtitle && (
          <p className="text-xs text-muted-foreground mt-1">{subtitle}</p>
        )}
        {trend && TrendIcon && (
          <div className="flex items-center gap-1 mt-2">
            <TrendIcon
              className={`h-4 w-4 ${
                trend.direction === "up"
                  ? "text-red-500"
                  : trend.direction === "down"
                  ? "text-green-500"
                  : "text-muted-foreground"
              }`}
            />
            <span
              className={`text-xs font-medium ${
                trend.direction === "up"
                  ? "text-red-500"
                  : trend.direction === "down"
                  ? "text-green-500"
                  : "text-muted-foreground"
              }`}
            >
              {Math.abs(trend.value)}%
            </span>
          </div>
        )}
      </CardContent>
    </Card>
  )
}

