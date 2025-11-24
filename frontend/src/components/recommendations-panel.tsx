import { AlertTriangle, TrendingDown, Info } from "lucide-react"
import { CostRecommendation, RecommendationSeverity } from "@/types/api"
import { formatCurrency } from "@/lib/utils"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Badge } from "@/components/ui/badge"
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert"

interface RecommendationsPanelProps {
  recommendations: CostRecommendation[]
  totalSavings: string
  currency: string
  highCount: number
  mediumCount: number
  lowCount: number
}

function getSeverityColor(severity: RecommendationSeverity): string {
  switch (severity) {
    case "high":
      return "destructive"
    case "medium":
      return "default"
    case "low":
      return "secondary"
    default:
      return "outline"
  }
}

function getSeverityIcon(severity: RecommendationSeverity) {
  switch (severity) {
    case "high":
      return <AlertTriangle className="h-5 w-5 text-destructive" />
    case "medium":
      return <TrendingDown className="h-5 w-5 text-orange-500" />
    case "low":
      return <Info className="h-5 w-5 text-blue-500" />
    default:
      return <Info className="h-5 w-5" />
  }
}

export function RecommendationsPanel({
  recommendations,
  totalSavings,
  currency,
  highCount,
  mediumCount,
  lowCount,
}: RecommendationsPanelProps) {
  if (recommendations.length === 0) {
    return (
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <TrendingDown className="h-5 w-5" />
            Cost Optimization Recommendations
          </CardTitle>
        </CardHeader>
        <CardContent>
          <Alert>
            <Info className="h-4 w-4" />
            <AlertTitle>No Recommendations</AlertTitle>
            <AlertDescription>
              All services are running efficiently. No cost optimization opportunities detected at this time.
            </AlertDescription>
          </Alert>
        </CardContent>
      </Card>
    )
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center gap-2">
          <TrendingDown className="h-5 w-5" />
          Cost Optimization Recommendations
        </CardTitle>
        <div className="flex items-center gap-4 mt-2">
          <div className="text-sm">
            <span className="font-semibold">Potential Savings: </span>
            <span className="text-lg font-bold text-green-600">
              {formatCurrency(totalSavings, currency)}
            </span>
          </div>
          <div className="flex items-center gap-2 text-sm text-muted-foreground">
            {highCount > 0 && (
              <Badge variant="destructive" className="text-xs">
                {highCount} High
              </Badge>
            )}
            {mediumCount > 0 && (
              <Badge variant="default" className="text-xs">
                {mediumCount} Medium
              </Badge>
            )}
            {lowCount > 0 && (
              <Badge variant="secondary" className="text-xs">
                {lowCount} Low
              </Badge>
            )}
          </div>
        </div>
      </CardHeader>
      <CardContent>
        <div className="space-y-4">
          {recommendations.map((rec) => (
            <Alert key={rec.id} className="relative">
              <div className="flex items-start gap-3">
                <div className="mt-0.5">{getSeverityIcon(rec.severity)}</div>
                <div className="flex-1 space-y-2">
                  <div className="flex items-start justify-between gap-4">
                    <div>
                      <AlertTitle className="flex items-center gap-2">
                        {rec.title}
                        <Badge variant={getSeverityColor(rec.severity) as any} className="text-xs">
                          {rec.severity}
                        </Badge>
                        <Badge variant="outline" className="text-xs">
                          {rec.node_type}
                        </Badge>
                      </AlertTitle>
                      <AlertDescription className="mt-2">
                        {rec.description}
                      </AlertDescription>
                    </div>
                    <div className="text-right shrink-0">
                      <div className="text-sm text-muted-foreground">Potential Savings</div>
                      <div className="text-xl font-bold text-green-600">
                        {formatCurrency(rec.potential_savings, rec.currency)}
                      </div>
                      <div className="text-xs text-muted-foreground mt-1">
                        Current: {formatCurrency(rec.current_cost, rec.currency)}
                      </div>
                    </div>
                  </div>

                  <div className="grid grid-cols-3 gap-4 pt-2 border-t">
                    <div>
                      <div className="text-xs text-muted-foreground">Metric</div>
                      <div className="text-sm font-medium">{rec.metric.replace(/_/g, " ")}</div>
                    </div>
                    <div>
                      <div className="text-xs text-muted-foreground">Utilization</div>
                      <div className="text-sm font-medium">
                        {parseFloat(rec.utilization_percent).toFixed(1)}%
                      </div>
                    </div>
                    <div>
                      <div className="text-xs text-muted-foreground">Analysis Period</div>
                      <div className="text-sm font-medium">{rec.analysis_period}</div>
                    </div>
                  </div>

                  <div className="grid grid-cols-3 gap-4 text-xs">
                    <div>
                      <span className="text-muted-foreground">Peak: </span>
                      <span className="font-medium">{parseFloat(rec.peak_value).toFixed(2)}</span>
                    </div>
                    <div>
                      <span className="text-muted-foreground">Average: </span>
                      <span className="font-medium">{parseFloat(rec.average_value).toFixed(2)}</span>
                    </div>
                    <div>
                      <span className="text-muted-foreground">Current: </span>
                      <span className="font-medium">{parseFloat(rec.current_value).toFixed(2)}</span>
                    </div>
                  </div>

                  <div className="pt-2 border-t">
                    <div className="text-xs text-muted-foreground mb-1">Recommended Action:</div>
                    <div className="text-sm font-medium text-blue-600">{rec.recommended_action}</div>
                  </div>
                </div>
              </div>
            </Alert>
          ))}
        </div>
      </CardContent>
    </Card>
  )
}

