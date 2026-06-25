import { useState, useEffect, useCallback } from 'react'
import { Box, Typography, CircularProgress, Button } from '@mui/material'
import { useTheme, alpha } from '@mui/material/styles'
import {
  LineChart, Line, BarChart, Bar, Cell, XAxis, YAxis, CartesianGrid,
  Tooltip as RechartsTooltip, Legend, LabelList, ResponsiveContainer
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
  const [retryCount, setRetryCount] = useState(0)

  // Stable serialized deps to avoid re-firing when parent passes a new object reference
  const itemType = item.type
  const itemQuery = JSON.stringify(item.query || {})

  const fetchData = useCallback(async () => {
    const q = item.query || {}
    setLoading(true)
    setError(false)

    try {
      let result: any
      switch (itemType) {
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
  // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [projectId, itemType, itemQuery, retryCount])

  useEffect(() => {
    fetchData()
  }, [fetchData])

  if (loading) {
    return (
      <Box display="flex" justifyContent="center" alignItems="center" height={180}>
        <CircularProgress size={24} />
      </Box>
    )
  }

  if (error || !data) {
    return (
      <Box display="flex" flexDirection="column" justifyContent="center" alignItems="center" gap={1} height={180}>
        <Typography variant="body2" color="error">Failed to load data</Typography>
        <Button size="small" variant="outlined" onClick={() => setRetryCount(c => c + 1)}>
          Retry
        </Button>
      </Box>
    )
  }

  return <WidgetChart type={item.type} data={data} />
}

function WidgetChart({ type, data }: { type: string; data: any }) {
  const theme = useTheme()

  const tooltipStyle = {
    backgroundColor: theme.palette.background.paper,
    border: `1px solid ${theme.palette.divider}`,
    color: theme.palette.text.primary,
  }
  const axisProps = {
    axisLine: { stroke: theme.palette.divider },
    tickLine: { stroke: theme.palette.divider },
  }

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
          <CartesianGrid strokeDasharray="3 3" stroke={theme.palette.divider} />
          <XAxis dataKey="day" tick={{ fontSize: 10, fill: theme.palette.text.secondary }} {...axisProps} />
          <YAxis tick={{ fontSize: 10, fill: theme.palette.text.secondary }} {...axisProps} />
          <RechartsTooltip contentStyle={tooltipStyle} />
          <Legend wrapperStyle={{ fontSize: 11, color: theme.palette.text.secondary }} />
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
        <BarChart data={chartData} margin={{ top: 24, right: 10, left: 0, bottom: 5 }}>
          <CartesianGrid strokeDasharray="3 3" stroke={theme.palette.divider} />
          <XAxis dataKey="name" tick={{ fontSize: 10, fill: theme.palette.text.primary }} {...axisProps} />
          <YAxis tick={{ fontSize: 10, fill: theme.palette.text.secondary }} {...axisProps} />
          <RechartsTooltip contentStyle={tooltipStyle}
            formatter={(value: any, name: any) => name === 'count' ? [value, 'Users'] : [`${value}%`, 'Conversion Rate']} />
          <Bar dataKey="count" name="count" radius={[4, 4, 0, 0]}>
            {chartData.map((_: any, i: number) => <Cell key={i} fill={COLORS[i % COLORS.length]} />)}
            <LabelList dataKey="conversionRate" position="top" formatter={(v: any) => `${v}%`}
              style={{ fontSize: 11, fontWeight: 700, fill: theme.palette.text.primary }} />
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
                      backgroundColor: alpha(theme.palette.primary.main, opacity),
                      color: opacity > 0.5 ? theme.palette.primary.contrastText : theme.palette.text.primary,
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
