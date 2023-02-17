import Graph  from 'react-graph-vis';
import * as React from 'react';
import Box from '@mui/material/Box';
import Card from '@mui/material/Card';
import CardActions from '@mui/material/CardActions';
import CardContent from '@mui/material/CardContent';
import Button from '@mui/material/Button';
import Typography from '@mui/material/Typography';
import Chip from '@mui/material/Chip';
import Stack from '@mui/material/Stack';

const data1 = {
    nodes: [
        { id: 1, label: 'Node 1' },
        { id: 2, label: 'Node 2' },
    ],
    edges: [
        { from: 1, to: 2 },
    ],
};

const options = {
    layout: {
        hierarchical: false
    },
    edges: {
        color: "#000000"
    }
};

const bull = (
    <Box
      component="span"
      sx={{ display: 'inline-block', mx: '2px', transform: 'scale(0.8)' }}
    >
      •
    </Box>
  );

const events = {
    select: function(event) {
        var { nodes, edges } = event;
    }
};

function GraphVis({data}) {
    return <div className="flex h-screen">
            <Graph graph={data} options={options} events={events} />
            <div className="absolute right-0 top-50% transform -translate-y-50">
            <Card sx={{ minWidth: 275 }}>
      <CardContent>
        <Typography sx={{ fontSize: 14 }} color="text.secondary" gutterBottom>
          Infrasture Mapping
        </Typography>
        <Typography variant="h5" component="div">
          AKS 1
        </Typography>
        <Typography sx={{ mb: 1.5 }} color="text.secondary">
          name: my-aks-1
        </Typography>
        <Typography variant="body2">
          <Stack>
            <Chip label="Queue Monitor" color="success" className="m-2"/>
            <Chip label="Workspace Safety" color="success"  className="m-2" />
          </Stack>
        </Typography>
      </CardContent>
      <CardActions>
        <Button size="small">Edit</Button>
      </CardActions>
    </Card>
            </div>
        </div>;
}

export default GraphVis;