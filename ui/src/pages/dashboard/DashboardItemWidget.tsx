import { useState, useEffect } from 'react'
import { Box, Typography, CircularProgress } from '@mui/material'
import {
  LineChart, Line, BarChart, Bar, Cell, XAxis, YAxis, CartesianGrid,
  Tooltip as RechartsTooltip, Legend, ResponsiveContainer
} from 'recharts'
import { analyticsService } from '../../services/analyticsService'
import type { DashboardItem } from '../../services/dashboardService'
import { COLORS } from './tabInfo'

interface Props {
  projectId: string
  item: DashboardItem
}

export function DashboardItemWidget({ projectId, item }: Props) {
  const [loading, setLoading] = useState(true)
  const [data, setData] = useState<any>(null)
  const [error, setError] = useState(false)

  useEffect(() => {
    const q = item.query || {}
    setLoading(true)
    setError(false)

    const run = async () => {
      try {
        let result: any
        switch (item.type) {
          case 'trends':
            result = await analyticsService.queryTrends(projectId, q as any)
            break
          case 'funnel':
            result = await analyticsService.queryFunnel(projectId, q as any)
            break
          case 'retention':
            result = await analyticsService.queryRetention(projectId, q as any)
            break
          case 'paths':
            result = await analyticsService.queryPaths(projectId, q as any)
            break
          default:
            result = null
        }
        setData(result)
      } catch {
        setError(true)
      } finally {
        setLoading(false)
      }
    }
    run()
  }, [projectId, item])

  if (loading) {
    return (
      <Box display="flex" justifyContent="center" alignItems="center" height={180}>
        <CircularProgress size={24} />
      </Box>
    )
  }

  if (error || !data) {
    return (
      <Box display="flex" justifyContent="center" alignItems="center" height={180}>
        <Typography variant="body2" color="error">Failed to load data</Typography>
      </Box>
    )
  }

  return <WidgetChart type={item.type} data={data} />
}

function WidgetChart({ type, data }: { type: string; data: any }) {
  if (type === 'trends') {
    const results = data.results || []
    if (!results.length) return <NoData />
    const chartData = results[0].days.map((day: string, i: number) => ({
      day,
      ...Object.fromEntries(results.map((s: any) => [s.label || s.breakdownValue || 'Count', Number(s.data[i] || 0)]))
    }))
    return (
      <ResponsiveContainer width="100%" height={200}>
        <LineChart data={chartData} margin={{ top: 5, right: 10, left: 0, bottom: 5 }}>
          <CartesianGrid strokeDasharray="3 3" />
          <XAxis dataKey="day" tick={{ fontSize: 10 }} />
          <YAxis tick={{ fontSize: 10 }} />
          <RechartsTooltip />
          <Legend wrapperStyle={{ fontSize: 11 }} />
          {results.map((s: any, i: number) => (
            <Line key={i} type="monotone" dataKey={s.label || s.breakdownValue || 'Count'}
              stroke={COLORS[i % COLORS.length]} strokeWidth={2} dot={false} />
          ))}
        </LineChart>
      </ResponsiveContainer>
    )
  }

  if (type === 'funnel') {
    const chartData = (data.result || []).map((r: any) => ({
      name: r.name || r.actionId,
      count: Number(r.count),
      conversionRate: Math.round(r.conversionRate * 100) / 100,
    }))
    if (!chartData.length) return <NoData />
    return (
      <ResponsiveContainer width="100%" height={200}>
        <BarChart data={chartData} margin={{ top: 5, right: 10, left: 0, bottom: 5 }}>
          <CartesianGrid strokeDasharray="3 3" />
          <XAxis dataKey="name" tick={{ fontSize: 10 }} />
          <YAxis tick={{ fontSize: 10 }} />
          <RechartsTooltip />
          <Legend wrapperStyle={{ fontSize: 11 }} />
          <Bar dataKey="count" name="Count" radius={[4, 4, 0, 0]}>
            {chartData.map((_: any, i: number) => <Cell key={i} fill={COLORS[i % COLORS.length]} />)}
          </Bar>
        </BarChart>
      </ResponsiveContainer>
    )
  }

  if (type === 'retention') {
    const results: any[] = data.result || []
    if (!results.length) return <NoData />
    const maxPeriods = Math.max(...results.map((r: any) => r.values.length))
    return (
      <Box sx={{ overflow: 'auto', fontSize: '0.72rem' }}>
        <Box component="table" sx={{ borderCollapse: 'collapse', width: '100%' }}>
          <Box component="tbody">
            {results.slice(0, 5).map((cohort: any, ri: number) => (
              <Box key={ri} component="tr">
                <Box component="td" sx={{ p: '2px 4px', whiteSpace: 'nowrap', borderBottom: '1px solid', borderColor: 'divider', maxWidth: 80, overflow: 'hidden', textOverflow: 'ellipsis' }}>
                  {cohort.label || cohort.date}
                </Box>
                {cohort.values.slice(0, 7).map((v: any, vi: number) => {
                  const pct = Math.round(v.percentage)
                  const opacity = 0.15 + (pct / 100) * 0.85
                  return (
                    <Box key={vi} component="td" sx={{
                      p: '2px 4px', textAlign: 'center',
                      borderBottom: '1px solid', borderColor: 'divider',
                      backgroundColor: `rgba(99, 102, 241, ${opacity})`,
                      color: opacity > 0.5 ? 'white' : 'inherit',
                      minWidth: 36,
                    }}>
                      {pct}%
                    </Box>
                  )
                })}
                {Array.from({ length: Math.min(7, maxPeriods) - cohort.values.slice(0, 7).length }, (_, i) => (
                  <Box key={`e${i}`} component="td" sx={{ p: '2px 4px', borderBottom: '1px solid', borderColor: 'divider', bgcolor: 'action.hover', minWidth: 36 }} />
                ))}
              </Box>
            ))}
          </Box>
        </Box>
      </Box>
    )
  }

  if (type === 'paths') {
    const links = [...(data.links || [])].sort((a: any, b: any) => Number(b.value) - Number(a.value)).slice(0, 6)
    if (!links.length) return <NoData />
    return (
      <Box>
        {links.map((l: any, i: number) => (
          <Box key={i} display="flex" alignItems="center" gap={1} py={0.5} sx={{ borderBottom: '1px solid', borderColor: 'divider' }}>
            <Typography variant="caption" noWrap sx={{ flex: 1, minWidth: 0 }}>{l.source}</Typography>
            <Typography variant="caption" color="text.secondary" sx={{ flexShrink: 0 }}>→</Typography>
            <Typography variant="caption" noWrap sx={{ flex: 1, minWidth: 0 }}>{l.target}</Typography>
            <Typography variant="caption" fontWeight={700} sx={{ flexShrink: 0, color: 'primary.main' }}>{l.value}</Typography>
          </Box>
        ))}
      </Box>
    )
  }

  return <NoData />
}

function NoData() {
  return (
    <Box display="flex" justifyContent="center" alignItems="center" height={120}>
      <Typography variant="body2" color="text.secondary">No data available</Typography>
    </Box>
  )
}
