import { Typography, Box, Paper } from '@mui/material'
import { PageHeader } from '../../components/PageHeader'

export function DashboardPage() {
  return (
    <Box sx={{ p: 3 }}>
      <PageHeader 
        title="Analytics Dashboard" 
        infoTitle="About Dashboard"
        infoDescription="The Dashboard provides a high-level overview of your project's analytics. Use this page to visualize user activity, key metrics, and insights derived from events tracked in your application."
      />
      <Paper sx={{ p: 3, mt: 2 }}>
        <Typography variant="body1">
          Trends, Funnels, and Retention charts will appear here.
        </Typography>
      </Paper>
    </Box>
  )
}
