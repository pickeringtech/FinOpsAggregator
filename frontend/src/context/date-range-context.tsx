import { createContext, useContext, useState, ReactNode } from "react"
import { subDays } from "date-fns"

export interface DateRange {
  from: Date
  to: Date
}

interface DateRangeContextValue {
  dateRange: DateRange
  setDateRange: (range: DateRange) => void
}

const DateRangeContext = createContext<DateRangeContextValue | undefined>(undefined)

interface DateRangeProviderProps {
  children: ReactNode
}

export function DateRangeProvider({ children }: DateRangeProviderProps) {
  const [dateRange, setDateRange] = useState<DateRange>(() => ({
    from: subDays(new Date(), 30),
    to: new Date(),
  }))

  return (
    <DateRangeContext.Provider value={{ dateRange, setDateRange }}>
      {children}
    </DateRangeContext.Provider>
  )
}

export function useDateRange(): DateRangeContextValue {
  const ctx = useContext(DateRangeContext)
  if (!ctx) {
    throw new Error("useDateRange must be used within a DateRangeProvider")
  }
  return ctx
}

