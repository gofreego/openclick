import { useEffect, useState } from 'react'
import {
  Box,
  Paper,
  Typography,
  Chip,
  CircularProgress,
  Divider,
  List,
  ListItem,
  ListItemText,
  Accordion,
  AccordionSummary,
  AccordionDetails
} from '@mui/material'
import SettingsIcon from '@mui/icons-material/Settings'
import ExpandMoreIcon from '@mui/icons-material/ExpandMore'
import ShieldIcon from '@mui/icons-material/Shield'
import PeopleIcon from '@mui/icons-material/People'
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

const ROLES_INFO = [
  {
    role: 'owner',
    title: 'Owner',
    description: 'The creator of the project. Has absolute control over the project, including project deletion, settings modification, and member management. Bypasses standard permission restrictions.',
    color: 'error' as const,
  },
  {
    role: 'admin',
    title: 'Admin',
    description: 'Administrative role. Can manage project members (add/remove), configure project settings, and modify all project resources.',
    color: 'primary' as const,
  },
  {
    role: 'member',
    title: 'Member',
    description: 'Standard collaborator. Can create, edit, and view all project resources (analytics, dashboards, feature flags, events, persons) but cannot manage team members or overall project configurations.',
    color: 'info' as const,
  },
  {
    role: 'viewer',
    title: 'Viewer',
    description: 'Read-only viewer. Can inspect dashboards, run analytics, and view event logs, but cannot create, modify, or delete any project resources or settings.',
    color: 'default' as const,
  },
]

export function SettingsPage() {
  const [permissions, setPermissions] = useState<Permission[]>([])
  const [loading, setLoading] = useState(false)
  const [permissionsExpanded, setPermissionsExpanded] = useState(false)
  const [rolesExpanded, setRolesExpanded] = useState(false)
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

      {/* System Permissions Accordion */}
      <Accordion
        expanded={permissionsExpanded}
        onChange={() => setPermissionsExpanded(!permissionsExpanded)}
        sx={{
          mb: 3,
          borderRadius: '8px !important',
          boxShadow: 'var(--shadow)',
          border: '1px solid var(--border)',
          background: 'var(--bg)',
          '&:before': { display: 'none' },
          overflow: 'hidden',
        }}
      >
        <AccordionSummary
          expandIcon={<ExpandMoreIcon />}
          sx={{
            px: 3,
            py: 1,
            '& .MuiAccordionSummary-content': {
              display: 'flex',
              alignItems: 'center',
              gap: 1.5,
            },
          }}
        >
          <ShieldIcon color="primary" />
          <Typography variant="h6" fontWeight={600}>System Permissions</Typography>
        </AccordionSummary>
        <Divider />
        <AccordionDetails sx={{ p: 3 }}>
          {loading && (
            <Box display="flex" justifyContent="center" py={4}>
              <CircularProgress />
            </Box>
          )}

          {!loading && permissions.length === 0 && (
            <Typography color="text.secondary">No permissions found.</Typography>
          )}

          {!loading && permissions.length > 0 && (
            <List dense disablePadding>
              {permissions.map((perm) => (
                <ListItem key={perm.key} divider sx={{ px: 0, py: 1.5 }}>
                  <ListItemText
                    primary={
                      <Box display="flex" alignItems="center" gap={1} mb={0.5}>
                        <Chip
                          label={perm.key}
                          size="small"
                          // variant="outlined"
                          color="default"
                          sx={{ fontFamily: 'monospace', fontWeight: 600 }}
                        />
                      </Box>
                    }
                    secondary={
                      <Typography variant="body2" color="text.primary">
                        {perm.description}
                      </Typography>
                    }
                  />
                </ListItem>
              ))}
            </List>
          )}
        </AccordionDetails>
      </Accordion>

      {/* Project Roles Accordion */}
      <Accordion
        expanded={rolesExpanded}
        onChange={() => setRolesExpanded(!rolesExpanded)}
        sx={{
          mb: 3,
          borderRadius: '8px !important',
          boxShadow: 'var(--shadow)',
          border: '1px solid var(--border)',
          background: 'var(--bg)',
          '&:before': { display: 'none' },
          overflow: 'hidden',
        }}
      >
        <AccordionSummary
          expandIcon={<ExpandMoreIcon />}
          sx={{
            px: 3,
            py: 1,
            '& .MuiAccordionSummary-content': {
              display: 'flex',
              alignItems: 'center',
              gap: 1.5,
            },
          }}
        >
          <PeopleIcon color="primary" />
          <Typography variant="h6" fontWeight={600}>Project Roles</Typography>
        </AccordionSummary>
        <Divider />
        <AccordionDetails sx={{ p: 3 }}>
          <List dense disablePadding>
            {ROLES_INFO.map((role) => (
              <ListItem key={role.role} divider sx={{ px: 0, py: 1.5 }}>
                <ListItemText
                  primary={
                    <Box display="flex" alignItems="center" gap={1} mb={0.5}>
                      <Chip
                        label={role.title}
                        size="small"
                        color={role.color}
                        sx={{ fontWeight: 600, minWidth: 70 }}
                      />
                      <Typography variant="body2" color="text.secondary" sx={{ fontFamily: 'monospace' }}>
                        ({role.role})
                      </Typography>
                    </Box>
                  }
                  secondary={
                    <Typography variant="body2" color="text.primary">
                      {role.description}
                    </Typography>
                  }
                />
              </ListItem>
            ))}
          </List>
        </AccordionDetails>
      </Accordion>

      {/* About OpenClick */}
      <Paper sx={{ p: 3, borderRadius: '8px', border: '1px solid var(--border)', boxShadow: 'var(--shadow)', background: 'var(--bg)' }}>
        <Box display="flex" alignItems="center" gap={1.5} mb={2}>
          <SettingsIcon color="primary" />
          <Typography variant="h6" fontWeight={600}>About OpenClick</Typography>
        </Box>
        <Divider sx={{ mb: 2 }} />
        <Typography variant="body2" color="text.secondary">
          OpenClick is an open-source, self-hostable product analytics platform. It provides event tracking, session replay, funnel analysis, feature flags, and cohort analytics with a low memory footprint.
        </Typography>
      </Paper>
    </Box>
  )
}

