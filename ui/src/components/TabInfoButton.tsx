import { useState } from 'react'
import {
  Box, Button, Dialog, DialogActions, DialogContent, DialogTitle, Typography
} from '@mui/material'
import InfoIcon from '@mui/icons-material/Info'

export interface TabInfo {
  title: string
  meaning: string
  howToUse: string
  example: string
}

export function TabInfoButton({ info }: { info: TabInfo }) {
  const [open, setOpen] = useState(false)

  return (
    <>
      <Box
        component="span"
        onClick={(e) => { e.stopPropagation(); setOpen(true) }}
        sx={{
          display: 'inline-flex',
          p: 0.25,
          ml: 0.5,
          color: 'text.secondary',
          cursor: 'pointer',
          borderRadius: '50%',
          '&:hover': { color: 'primary.main', bgcolor: 'action.hover' },
        }}
      >
        <InfoIcon sx={{ fontSize: 16 }} />
      </Box>

      <Dialog open={open} onClose={() => setOpen(false)} maxWidth="sm" fullWidth PaperProps={{ sx: { borderRadius: '12px', p: 1 } }}>
        <DialogTitle sx={{ fontWeight: 700, pb: 1, display: 'flex', alignItems: 'center', gap: 1 }}>
          <InfoIcon color="primary" />
          {info.title}
        </DialogTitle>
        <DialogContent dividers sx={{ py: 2 }}>
          <Box mb={2.5}>
            <Typography variant="subtitle2" fontWeight={600} color="primary" gutterBottom>What is it?</Typography>
            <Typography variant="body2" sx={{ lineHeight: 1.6 }}>{info.meaning}</Typography>
          </Box>
          <Box mb={2.5}>
            <Typography variant="subtitle2" fontWeight={600} color="primary" gutterBottom>How to Use</Typography>
            <Typography variant="body2" sx={{ lineHeight: 1.6 }}>{info.howToUse}</Typography>
          </Box>
          <Box sx={{ p: 2, bgcolor: 'action.hover', borderRadius: '8px', border: '1px solid', borderColor: 'divider' }}>
            <Typography variant="subtitle2" fontWeight={600} gutterBottom>Example Scenario</Typography>
            <Typography variant="body2" color="text.secondary" sx={{ lineHeight: 1.6 }}>{info.example}</Typography>
          </Box>
        </DialogContent>
        <DialogActions sx={{ px: 3, py: 1.5 }}>
          <Button onClick={() => setOpen(false)} variant="contained" color="primary">Got it</Button>
        </DialogActions>
      </Dialog>
    </>
  )
}
