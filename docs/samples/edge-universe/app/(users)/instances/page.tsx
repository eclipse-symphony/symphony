'use client';

import * as React from 'react';
import Box from '@mui/material/Box';
import DeleteIcon from '@mui/icons-material/Delete';
import PhotoCameraIcon from '@mui/icons-material/PhotoCamera';
import PrecisionManufacturingIcon from '@mui/icons-material/PrecisionManufacturing';
import RouterIcon from '@mui/icons-material/Router';
import BoltIcon from '@mui/icons-material/Bolt';
import VideoLabelIcon from '@mui/icons-material/VideoLabel';
import PrintIcon from '@mui/icons-material/Print';
import PetsIcon from '@mui/icons-material/Pets';
import Chip from '@mui/material/Chip';
import DnsIcon from '@mui/icons-material/Dns';
import Toolbar from '@mui/material/Toolbar';
import Button from '@mui/material/Button';
import Menu from '@mui/material/Menu';
import MenuItem from '@mui/material/MenuItem';
import CloudUploadIcon from '@mui/icons-material/CloudUpload';
import ListItemIcon from '@mui/material/ListItemIcon';
import UploadFileIcon from '@mui/icons-material/UploadFile';
import Stack from '@mui/material/Stack';
import FaceIcon from '@mui/icons-material/Face';
import WindowIcon from '@mui/icons-material/Window';
import SettingsApplicationsIcon from '@mui/icons-material/SettingsApplications';
import PolicyIcon from '@mui/icons-material/Policy';
import Avatar from '@mui/material/Avatar';
import Card from '@mui/material/Card';
import CardHeader from '@mui/material/CardHeader';
import CardContent from '@mui/material/CardContent';
import IconTabs from '@elements/IconTabs';
import GraphVis from '@elements/GraphVis';
import { Chart as ChartJS, ArcElement, Tooltip, Legend } from "chart.js";
import { Pie } from 'react-chartjs-2';

ChartJS.register(ArcElement, Tooltip, Legend);

interface Data {
    connections: State[];
    site: string;
    solutions: State[];
    name: string;
    stack: Artifact[];
    type: string;    
    graph: {};
    pie: {};    
}
interface Artifact {
    version: string,
    status: string
}
interface State {
    name: string,
    status: string,
    annotation: string
}
function createArtifact(
    version: string,
    status: string
): Artifact {
    return {
        version,
        status
    };
}
function createState(
    name: string,
    status: string,
    annotation: string
): State {
    return {
        name,
        status,
        annotation
    };
}
function createData(
    name: string,
    stack: Artifact[],
    type: string,
    connections: State[],
    solutions: State[],
    site: string,    
    graph: {},
    pie: {}
): Data {
    return {
        name,
        stack,
        type,
        connections,
        solutions,
        site,        
        graph,
        pie
    };
}
const rows = [
    createData('Queue Monitor', [            
        createArtifact('1.0', 'ok'),
        createArtifact('1.0', 'ok'),
        createArtifact('1.0', 'ok'),
        createArtifact('1.0', 'ok'),
        createArtifact('1.0', 'ok'),
        createArtifact('1.0', 'ok'),
        createArtifact('1.0', 'ok'),
        createArtifact('1.0', 'ok'),
        createArtifact('1.0', 'ok'),
        createArtifact('1.0', ''),
    ], 'server', [], [], 'targets: 9/10', {nodes: [
        { id: 1, label: 'Servicing Policy' },
        { id: 2, label: 'Gatekeeper' },
        { id: 3, label: 'People Counter' },
        { id: 4, label: 'Camera Operator' },
        { id: 5, label: 'Nvdia Driver' },
    ],
    edges: [
        { from: 1, to: 2 },
        { from: 3, to: 5 },
        { from: 3, to: 4 },
    ],}, {
        datasets: [{
          data: [1, 9],
          backgroundColor: ['#FF6384', '#36A2EB'],
          hoverBackgroundColor: ['#FF6384', '#36A2EB']
        }]
      }),  
    createData('Worker Safety', [                
        createArtifact('1.0', 'ok'),
        createArtifact('1.0', 'ok'),
        createArtifact('1.0', 'ok'),
        createArtifact('1.0', 'ok'),
        createArtifact('1.0', 'ok'),
        createArtifact('2.0', 'ok'),
        createArtifact('1.0', 'ok'),
        createArtifact('1.0', 'ok'),
        createArtifact('1.0', 'ok'),
        createArtifact('1.0', 'ok'),
        createArtifact('1.0', 'ok'),
        createArtifact('1.0', 'ok'),
        createArtifact('1.0', 'ok'),
        createArtifact('1.0', 'ok'),
        createArtifact('1.0', 'ok'),
        createArtifact('1.0', 'ok'),
        createArtifact('1.0', 'ok'),
        createArtifact('1.0', 'ok'),
        createArtifact('1.0', 'ok'),
        createArtifact('1.0', 'ok'),
        createArtifact('1.0', 'ok'),
        createArtifact('1.0', 'ok'),
        createArtifact('1.0', 'ok'),
        createArtifact('1.0', 'ok'),
        createArtifact('1.0', 'ok'),
        createArtifact('1.0', 'ok'),
        createArtifact('1.0', 'ok'),
        createArtifact('1.0', ''),
        createArtifact('1.0', ''),
        createArtifact('1.0', '')
    ], 'server', [ ], [], 'targets: 27/30, canary: 1/27', {nodes: [
        { id: 1, label: 'Workplace Policy' },
        { id: 2, label: 'Gatekeeper' },
        { id: 3, label: 'People Detection' },
        { id: 4, label: 'Camera Operator' },
        { id: 5, label: 'Nvdia Driver' },
    ],
    edges: [
        { from: 1, to: 2 },
        { from: 3, to: 5 },
        { from: 3, to: 4 },
    ],}, {
        datasets: [{
          data: [1, 26, 3],
          backgroundColor: ['#FF6384', '#36A2EB', '#FFCE56'],
          hoverBackgroundColor: ['#FF6384', '#36A2EB', '#FFCE56']
        }]
      })
];

export default function Instances() {
  const [anchorEl, setAnchorEl] = React.useState<null | HTMLElement>(null);
  const open = Boolean(anchorEl);
  const handleMenuClick = (event: React.MouseEvent<HTMLButtonElement>) => {
    setAnchorEl(event.currentTarget);
  };
  const handleClose = () => {
    setAnchorEl(null);
  };
  return ( 
  <div>
    <Toolbar>
      <Button
        id="basic-button"
        aria-controls={open ? 'basic-menu' : undefined}
        aria-haspopup="true"
        aria-expanded={open ? 'true' : undefined}
        onClick={handleMenuClick}>
        ACTIONS
      </Button>
      <Menu id="basic-menu" anchorEl={anchorEl} open={open} onClose={handleClose} MenuListProps={{'aria-labelledby': 'basic-button',}}>
        <MenuItem onClick={handleClose}>
          <ListItemIcon>
            <CloudUploadIcon fontSize="small" />
          </ListItemIcon>
          Onboard
        </MenuItem>
        <MenuItem onClick={handleClose}>
          <ListItemIcon>
            <UploadFileIcon fontSize="small" />
          </ListItemIcon>
          Bulk Onboard
        </MenuItem>
        <MenuItem onClick={handleClose}>
          <ListItemIcon>
            <DeleteIcon fontSize="small" />
          </ListItemIcon>
            Decommission
        </MenuItem>
      </Menu>
    </Toolbar>    
    <Box sx={{ display: 'flex', flexWrap: 'wrap', '& > :not(style)': {m: 1,},}}>
        <Stack className="w-full">
    {rows.map((row,index)=>{
      return (
      <Card key={'c'+row.name} className="w-full">
        <CardHeader
          avatar={
            <Avatar  aria-label="recipe">
                <DnsIcon />         
            </Avatar>}              
          title={row.name}
          subheader={row.site}/>
        <CardContent>
                <Stack className="flex items-start rounded-md bg-gray-100" direction="row">
                <Box style={{width:100, height:100}}>
                    <Pie key={'pp'+row.name} data={row.pie}  />  
                </Box>
                <Box sx={{ display: 'flex', flexWrap: 'wrap', '& > :not(style)': {m: 1,},}}>
                { row.stack.map((c)=>(
                    <div>
                    <Box sx={{ bgcolor: c.status === 'ok'? (c.version === '2.0'? 'orange': 'green') : 'red', height: 30, width: 30 }}/>                    
                    </div>
                ))}
                </Box>
            </Stack>
        </CardContent>
      </Card>);})}
      </Stack>
    </Box>
  </div>);
}