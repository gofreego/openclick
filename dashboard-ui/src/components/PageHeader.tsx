import React, { useState } from 'react';
import { Typography, Box, IconButton, Dialog, DialogTitle, DialogContent, DialogContentText, DialogActions, Button } from '@mui/material';
import InfoOutlinedIcon from '@mui/icons-material/InfoOutlined';

interface PageHeaderProps {
  title: string;
  infoTitle: string;
  infoDescription: string | React.ReactNode;
  action?: React.ReactNode;
}

export function PageHeader({ title, infoTitle, infoDescription, action }: PageHeaderProps) {
  const [open, setOpen] = useState(false);

  return (
    <Box display="flex" justifyContent="space-between" alignItems="center" mb={3}>
      <Box display="flex" alignItems="center" gap={1}>
        <Typography variant="h4">{title}</Typography>
        <IconButton size="small" onClick={() => setOpen(true)} color="info">
          <InfoOutlinedIcon />
        </IconButton>
      </Box>
      {action && <Box>{action}</Box>}

      <Dialog open={open} onClose={() => setOpen(false)} maxWidth="sm" fullWidth>
        <DialogTitle>{infoTitle}</DialogTitle>
        <DialogContent dividers>
          <DialogContentText sx={{ whiteSpace: 'pre-line' }}>
            {infoDescription}
          </DialogContentText>
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setOpen(false)}>Close</Button>
        </DialogActions>
      </Dialog>
    </Box>
  );
}
