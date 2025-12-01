"use client"

import { useMemo, useState } from "react"
import { addDays, format, subDays } from "date-fns"
import { Calendar as CalendarIcon, ChevronDown } from "lucide-react"
import { DateRange } from "react-day-picker"

import { Button } from "@/components/ui/button"
import { Popover, PopoverContent, PopoverTrigger } from "@/components/ui/popover"
import { Calendar } from "@/components/ui/calendar"

const PRESETS: { label: string; getRange: () => DateRange }[] = [
  {
    label: "Last 7 days",
    getRange: () => {
      const to = new Date()
      const from = subDays(to, 7)
      return { from, to }
    },
  },
  {
    label: "Last 30 days",
    getRange: () => {
      const to = new Date()
      const from = subDays(to, 30)
      return { from, to }
    },
  },
  {
    label: "Last 60 days",
    getRange: () => {
      const to = new Date()
      const from = subDays(to, 60)
      return { from, to }
    },
  },
  {
    label: "Last 90 days",
    getRange: () => {
      const to = new Date()
      const from = subDays(to, 90)
      return { from, to }
    },
  },
  {
    label: "Last 6 months",
    getRange: () => {
      const to = new Date()
      const from = subDays(to, 182)
      return { from, to }
    },
  },
  {
    label: "Last 12 months",
    getRange: () => {
      const to = new Date()
      const from = subDays(to, 365)
      return { from, to }
    },
  },
]

export interface DateRangePickerProps {
  value: DateRange
  onChange: (range: DateRange) => void
}

export function DateRangePicker({ value, onChange }: DateRangePickerProps) {
  const [showCustom, setShowCustom] = useState(false)

  const label = useMemo(() => {
    if (!value?.from || !value?.to) return "Select a date range"
    return `${format(value.from, "MMM dd, yyyy")} - ${format(value.to, "MMM dd, yyyy")}`
  }, [value])

  return (
    <Popover>
      <PopoverTrigger asChild>
        <Button
          variant="outline"
          className="h-9 gap-2 rounded-md border-dashed px-3 text-sm font-normal text-muted-foreground hover:border-primary/60 hover:text-foreground"
        >
          <CalendarIcon className="h-4 w-4" />
          <span className="hidden sm:inline-flex">{label}</span>
          <span className="inline-flex sm:hidden">Date range</span>
          <ChevronDown className="h-3 w-3 opacity-60" />
        </Button>
      </PopoverTrigger>
      <PopoverContent className="w-auto p-3" align="end">
        <div className="flex flex-col gap-3 sm:flex-row">
          <div className="space-y-2">
            <p className="text-xs font-medium text-muted-foreground">Quick ranges</p>
            <div className="grid grid-cols-2 gap-1">
              {PRESETS.map((preset) => (
                <Button
                  key={preset.label}
                  variant="outline"
                  size="sm"
                  className="h-7 justify-start px-2 text-xs"
                  onClick={() => {
                    const range = preset.getRange()
                    onChange({ from: range.from!, to: range.to! })
                    setShowCustom(false)
                  }}
                >
                  {preset.label}
                </Button>
              ))}
            </div>
            <Button
              variant={showCustom ? "default" : "outline"}
              size="sm"
              className="mt-2 h-7 justify-start px-2 text-xs"
              onClick={() => setShowCustom((prev) => !prev)}
            >
              Custom range…
            </Button>
          </div>
          {showCustom && (
            <div className="flex-1">
              <p className="mb-1 text-xs font-medium text-muted-foreground">Custom range</p>
              <Calendar
                mode="range"
                defaultMonth={value?.from || value?.to || new Date()}
                selected={value}
                onSelect={(range) => {
                  if (!range?.from || !range?.to) return
                  const normalized: DateRange = {
                    from: range.from,
                    to: addDays(range.to, 0),
                  }
                  onChange(normalized)
                }}
                numberOfMonths={2}
              />
            </div>
          )}
        </div>
      </PopoverContent>
    </Popover>
  )
}

