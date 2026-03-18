import type { ReactNode } from 'react'
import {
  Area,
  AreaChart,
  Bar,
  BarChart,
  CartesianGrid,
  Legend,
  Line,
  LineChart,
  ResponsiveContainer,
  Tooltip,
  XAxis,
  YAxis,
} from 'recharts'

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

function tooltipStyle() {
  return {
    backgroundColor: '#0a0a0b',
    border: '1px solid rgba(255,255,255,0.15)',
    color: 'rgba(255,255,255,0.95)',
    borderRadius: 0,
  }
}

const gridStroke = 'rgba(255,255,255,0.1)'
const axisStroke = 'rgba(255,255,255,0.35)'
const tickColor = 'rgba(255,255,255,0.55)'

export function GrafanaTimeSeries({
  title,
  subtitle,
  data,
  xKey,
  lines,
  yDomain = ['auto', 'auto'],
}: TimeSeriesProps) {
  return (
    <ChartPanel title={title} subtitle={subtitle}>
      <ResponsiveContainer width="100%" height="100%">
        <LineChart data={data}>
          <CartesianGrid stroke={gridStroke} strokeDasharray="2 5" vertical={false} />
          <XAxis dataKey={xKey} stroke={axisStroke} tick={{ fontSize: 11, fill: tickColor }} />
          <YAxis stroke={axisStroke} tick={{ fontSize: 11, fill: tickColor }} domain={yDomain} />
          <Tooltip
            contentStyle={tooltipStyle()}
            labelStyle={{ color: 'rgba(255,255,255,0.7)' }}
            itemStyle={{ color: 'rgba(255,255,255,0.95)' }}
          />
          <Legend
            wrapperStyle={{
              color: 'rgba(255,255,255,0.65)',
              fontSize: 11,
              letterSpacing: '0.04em',
            }}
          />
          {lines.map((line) => (
            <Line
              key={line.key}
              type="monotone"
              dataKey={line.key}
              name={line.label}
              stroke={line.color}
              dot={false}
              strokeWidth={line.strokeWidth ?? 2}
              isAnimationActive={false}
              activeDot={{ r: 3, fill: line.color, stroke: 'transparent' }}
            />
          ))}
        </LineChart>
      </ResponsiveContainer>
    </ChartPanel>
  )
}

export function GrafanaAreaSeries({ title, subtitle, data, xKey, areas }: AreaSeriesProps) {
  return (
    <ChartPanel title={title} subtitle={subtitle}>
      <ResponsiveContainer width="100%" height="100%">
        <AreaChart data={data}>
          <CartesianGrid stroke={gridStroke} strokeDasharray="2 5" vertical={false} />
          <XAxis dataKey={xKey} stroke={axisStroke} tick={{ fontSize: 11, fill: tickColor }} />
          <YAxis stroke={axisStroke} tick={{ fontSize: 11, fill: tickColor }} />
          <Tooltip
            contentStyle={tooltipStyle()}
            labelStyle={{ color: 'rgba(255,255,255,0.7)' }}
            itemStyle={{ color: 'rgba(255,255,255,0.95)' }}
          />
          <Legend
            wrapperStyle={{
              color: 'rgba(255,255,255,0.65)',
              fontSize: 11,
              letterSpacing: '0.04em',
            }}
          />
          {areas.map((area) => (
            <Area
              key={area.key}
              type="monotone"
              dataKey={area.key}
              name={area.label}
              stroke={area.color}
              fill={area.color}
              fillOpacity={0.18}
              strokeWidth={2}
              isAnimationActive={false}
            />
          ))}
        </AreaChart>
      </ResponsiveContainer>
    </ChartPanel>
  )
}

export function GrafanaBarSeries({ title, subtitle, data, xKey, bars }: BarSeriesProps) {
  return (
    <ChartPanel title={title} subtitle={subtitle}>
      <ResponsiveContainer width="100%" height="100%">
        <BarChart data={data}>
          <CartesianGrid stroke={gridStroke} strokeDasharray="2 5" vertical={false} />
          <XAxis dataKey={xKey} stroke={axisStroke} tick={{ fontSize: 11, fill: tickColor }} />
          <YAxis stroke={axisStroke} tick={{ fontSize: 11, fill: tickColor }} />
          <Tooltip
            contentStyle={tooltipStyle()}
            labelStyle={{ color: 'rgba(255,255,255,0.7)' }}
            itemStyle={{ color: 'rgba(255,255,255,0.95)' }}
          />
          <Legend
            wrapperStyle={{
              color: 'rgba(255,255,255,0.65)',
              fontSize: 11,
              letterSpacing: '0.04em',
            }}
          />
          {bars.map((bar) => (
            <Bar
              key={bar.key}
              dataKey={bar.key}
              name={bar.label}
              fill={bar.color}
              isAnimationActive={false}
              maxBarSize={18}
            />
          ))}
        </BarChart>
      </ResponsiveContainer>
    </ChartPanel>
  )
}
