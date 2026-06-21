import { useEffect, useState } from 'react'
import { Box, Paper, Typography, Chip, CircularProgress, Divider, List, ListItem, ListItemText } from '@mui/material'
import SettingsIcon from '@mui/icons-material/Settings'
import { httpClient } from '../../utils/httpClient'
import { useNotification } from '@gofreego/tsutils'
import { PageHeader } from '../../components/PageHeader'

interface Permission {
  key: string
  description?: string
}

interface ListPermissionsResponse {
  permissions: Permission[]
}

export function SettingsPage() {
  const [permissions, setPermissions] = useState<Permission[]>([])
  const [loading, setLoading] = useState(false)
  const notify = useNotification()

  useEffect(() => {
    setLoading(true)
    httpClient.get<ListPermissionsResponse>('/api/v1/permissions')
      .then(res => setPermissions(res.data?.permissions || []))
      .catch(() => notify.error('Failed to load permissions'))
      .finally(() => setLoading(false))
  }, [])

  return (
    <Box sx={{ p: 3 }}>
      <PageHeader
        title="Settings"
        infoTitle="About Settings"
        infoDescription="View system permissions and configuration. Use the Projects page to manage project-specific settings like name, timezone, and API keys."
      />

      <Paper sx={{ p: 3, mb: 3 }}>
        <Box display="flex" alignItems="center" gap={1} mb={2}>
          <SettingsIcon color="primary" />
          <Typography variant="h6" fontWeight={600}>System Permissions</Typography>
        </Box>
        <Divider sx={{ mb: 2 }} />

        {loading && <Box display="flex" justifyContent="center" py={4}><CircularProgress /></Box>}

        {!loading && permissions.length === 0 && (
          <Typography color="text.secondary">No permissions found.</Typography>
        )}

        {!loading && permissions.length > 0 && (
          <List dense>
            {permissions.map((perm) => (
              <ListItem key={perm.key} divider>
                <ListItemText
                  primary={
                    <Box display="flex" alignItems="center" gap={1}>
                      <Chip label={perm.key} size="small" variant="outlined" sx={{ fontFamily: 'monospace' }} />
                    </Box>
                  }
                  secondary={perm.description}
                />
              </ListItem>
            ))}
          </List>
        )}
      </Paper>

      <Paper sx={{ p: 3 }}>
        <Typography variant="h6" fontWeight={600} gutterBottom>About OpenClick</Typography>
        <Divider sx={{ mb: 2 }} />
        <Typography variant="body2" color="text.secondary">
          OpenClick is an open-source, self-hostable product analytics platform. It provides event tracking, session replay, funnel analysis, feature flags, and cohort analytics with a low memory footprint.
        </Typography>
      </Paper>
    </Box>
  )
}
