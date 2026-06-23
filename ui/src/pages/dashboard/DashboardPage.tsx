import { useState } from 'react'
import { Typography, Box, Paper, Tabs, Tab } from '@mui/material'
import { TabInfoButton } from '../../components/TabInfoButton'
import { useCurrentProject } from '../../contexts/ProjectContext'
import { PageHeader } from '../../components/PageHeader'
import { TrendsTab } from './TrendsTab'
import { FunnelTab } from './FunnelTab'
import { RetentionTab } from './RetentionTab'
import { PathsTab } from './PathsTab'
import { DashboardsTab } from './DashboardsTab'
import { TAB_INFO } from './tabInfo'

const TAB_KEYS = ['dashboards', 'trends', 'funnel', 'retention', 'paths'] as const

export function DashboardPage() {
  const selectedProjectId = useCurrentProject()
  const [activeTab, setActiveTab] = useState(0)

  return (
    <Box sx={{ p: 3 }}>
      <PageHeader
        title="Analytics Dashboard"
        infoTitle="About Analytics"
        infoDescription="Run Trends, Funnel, Retention, and Paths queries against your event data. Create and manage dashboards to organize your key metrics."
      />

      {selectedProjectId ? (
        <>
          <Tabs value={activeTab} onChange={(_, v) => setActiveTab(v)} sx={{ mb: 3 }}>
            {TAB_KEYS.map((key) => {
              const label = key.charAt(0).toUpperCase() + key.slice(1)
              return (
                <Tab
                  key={key}
                  label={
                    <Box display="flex" alignItems="center" gap={0.5}>
                      {label}
                      <TabInfoButton info={TAB_INFO[key]} />
                    </Box>
                  }
                />
              )
            })}
          </Tabs>
          {activeTab === 0 && <DashboardsTab projectId={selectedProjectId} />}
          {activeTab === 1 && <TrendsTab projectId={selectedProjectId} />}
          {activeTab === 2 && <FunnelTab projectId={selectedProjectId} />}
          {activeTab === 3 && <RetentionTab projectId={selectedProjectId} />}
          {activeTab === 4 && <PathsTab projectId={selectedProjectId} />}
        </>
      ) : (
        <Paper sx={{ p: 3, textAlign: 'center', mt: 2 }}>
          <Typography variant="body1">Please select a project to view analytics.</Typography>
        </Paper>
      )}
    </Box>
  )
}
