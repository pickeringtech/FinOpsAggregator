import { format, subDays } from "date-fns"
import { Calendar } from "lucide-react"
import { Button } from "@/components/ui/button"

interface DateRange {
  from: Date
  to: Date
}

interface DateRangePickerProps {
  value: DateRange
  onChange: (range: DateRange) => void
}

export function DateRangePicker({ value, onChange }: DateRangePickerProps) {
  const presets = [
    { label: "Last 7 days", days: 7 },
    { label: "Last 14 days", days: 14 },
    { label: "Last 30 days", days: 30 },
    { label: "Last 90 days", days: 90 },
  ]

  return (
    <div className="flex items-center gap-2">
      <Button variant="outline" className="justify-start text-left font-normal">
        <Calendar className="mr-2 h-4 w-4" />
        {format(value.from, "MMM dd, yyyy")} - {format(value.to, "MMM dd, yyyy")}
      </Button>
      <div className="flex gap-1">
        {presets.map((preset) => (
          <Button
            key={preset.days}
            variant="ghost"
            size="sm"
            onClick={() => {
              const to = new Date()
              const from = subDays(to, preset.days)
              onChange({ from, to })
            }}
          >
            {preset.label}
          </Button>
        ))}
      </div>
    </div>
  )
}

