import { Typography, Box, Paper } from '@mui/material'

export function DashboardPage() {
  return (
    <Box sx={{ p: 3 }}>
      <Typography variant="h4" gutterBottom>
        Analytics Dashboard
      </Typography>
      <Paper sx={{ p: 3, mt: 2 }}>
        <Typography variant="body1">
          Trends, Funnels, and Retention charts will appear here.
        </Typography>
      </Paper>
    </Box>
  )
}
