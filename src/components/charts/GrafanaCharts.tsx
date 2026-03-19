import type { ReactNode } from 'react'
import ReactECharts from 'echarts-for-react'

type Datum = Record<string, number | string>

interface PanelProps {
  title: string
  subtitle?: string
  children: ReactNode
}

interface Series {
  key: string
  label: string
  color: string
  strokeWidth?: number
}

interface TimeSeriesProps {
  title: string
  subtitle?: string
  data: Datum[]
  xKey: string
  lines: Series[]
  yDomain?: [number, number] | ['auto', 'auto']
}

interface AreaSeriesProps {
  title: string
  subtitle?: string
  data: Datum[]
  xKey: string
  areas: Series[]
}

interface BarSeriesProps {
  title: string
  subtitle?: string
  data: Datum[]
  xKey: string
  bars: Series[]
}

function ChartPanel({ title, subtitle, children }: PanelProps) {
  return (
    <section className="border-t border-border-muted pt-5">
      <p className="industrial-label">{title}</p>
      {subtitle ? <p className="mt-1 text-xs text-text-secondary">{subtitle}</p> : null}
      <div className="mt-4 h-56">{children}</div>
    </section>
  )
}

function baseOption(data: Datum[], xKey: string) {
  return {
    animationDuration: 120,
    grid: { top: 20, left: 40, right: 20, bottom: 36 },
    tooltip: {
      trigger: 'axis',
      borderWidth: 1,
      borderColor: 'rgba(255,255,255,0.15)',
      backgroundColor: '#0A0A0B',
      textStyle: { color: 'rgba(255,255,255,0.9)', fontSize: 11 },
    },
    legend: {
      top: 0,
      textStyle: { color: 'rgba(255,255,255,0.55)', fontSize: 11 },
    },
    xAxis: {
      type: 'category',
      data: data.map((row) => String(row[xKey] ?? '')),
      boundaryGap: false,
      axisLine: { lineStyle: { color: 'rgba(255,255,255,0.25)' } },
      axisLabel: { color: 'rgba(255,255,255,0.45)', fontSize: 11 },
      splitLine: { show: false },
    },
    yAxis: {
      type: 'value',
      axisLine: { show: false },
      axisLabel: { color: 'rgba(255,255,255,0.45)', fontSize: 11 },
      splitLine: { lineStyle: { color: 'rgba(255,255,255,0.1)' } },
    },
  }
}

export function GrafanaTimeSeries({
  title,
  subtitle,
  data,
  xKey,
  lines,
  yDomain = ['auto', 'auto'],
}: TimeSeriesProps) {
  const option = {
    ...baseOption(data, xKey),
    yAxis: {
      ...(baseOption(data, xKey).yAxis as object),
      min: yDomain[0] === 'auto' ? undefined : yDomain[0],
      max: yDomain[1] === 'auto' ? undefined : yDomain[1],
    },
    series: lines.map((line) => ({
      name: line.label,
      type: 'line',
      data: data.map((row) => Number(row[line.key] ?? 0)),
      showSymbol: false,
      smooth: true,
      lineStyle: { color: line.color, width: line.strokeWidth ?? 2 },
      itemStyle: { color: line.color },
    })),
  }

  return (
    <ChartPanel title={title} subtitle={subtitle}>
      <ReactECharts option={option} style={{ width: '100%', height: '100%' }} />
    </ChartPanel>
  )
}

export function GrafanaAreaSeries({ title, subtitle, data, xKey, areas }: AreaSeriesProps) {
  const option = {
    ...baseOption(data, xKey),
    series: areas.map((area) => ({
      name: area.label,
      type: 'line',
      data: data.map((row) => Number(row[area.key] ?? 0)),
      showSymbol: false,
      smooth: true,
      lineStyle: { color: area.color, width: 2 },
      areaStyle: { color: area.color, opacity: 0.18 },
      itemStyle: { color: area.color },
    })),
  }

  return (
    <ChartPanel title={title} subtitle={subtitle}>
      <ReactECharts option={option} style={{ width: '100%', height: '100%' }} />
    </ChartPanel>
  )
}

export function GrafanaBarSeries({ title, subtitle, data, xKey, bars }: BarSeriesProps) {
  const base = baseOption(data, xKey)
  const option = {
    ...base,
    xAxis: {
      ...(base.xAxis as object),
      boundaryGap: true,
    },
    series: bars.map((bar) => ({
      name: bar.label,
      type: 'bar',
      barMaxWidth: 18,
      data: data.map((row) => Number(row[bar.key] ?? 0)),
      itemStyle: { color: bar.color },
    })),
  }

  return (
    <ChartPanel title={title} subtitle={subtitle}>
      <ReactECharts option={option} style={{ width: '100%', height: '100%' }} />
    </ChartPanel>
  )
}
