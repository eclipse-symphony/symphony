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
interface Data {
    connections: State[];
    site: string;
    solutions: State[];
    name: string;
    stack: Artifact[];
    type: string;    
    graph: {}    
}
interface Artifact {
    name: string,
    type: string,
    version: string,
    current: string,
    manager: string
}
interface State {
    name: string,
    status: string,
    annotation: string
}
function createArtifact(
    name: string,
    type: string,
    version: string,
    current: string,
    manager: string
): Artifact {
    return {
        name,
        type,
        version,
        current,
        manager
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
): Data {
    return {
        name,
        stack,
        type,
        connections,
        solutions,
        site,        
        graph
    };
}

const rows = [
    createData('Queue Monitor', [            
        createArtifact('Servicing Policy', 'policy', '1.0', "1.0", 'gatekeeper'),
        createArtifact('People Counter', 'runtime', '1.0.0', '1.0.0','container'),
        createArtifact('Gatekeeper', 'runtime', '1.13.1', '1.13.1','helm'),
        createArtifact('Camera Operator', 'runtime', '1.3.0', '1.3.0','container'),
        createArtifact('Nvidia Driver', 'runtime', '2.11.0', '2.11.0', 'container'),        
    ], 'server', [], [], 'instances: 1', {nodes: [
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
    ],}),  
    createData('Worker Safety', [                
        createArtifact('Workplace Policy', 'policy', '1.3', "1.3", 'gatekeeper'),
        createArtifact('People Detection', 'runtime', '1.0.0', '1.0.0','container'),
        createArtifact('Gatekeeper', 'runtime', '1.13.1', '1.13.1','helm'),
        createArtifact('Camera Operator', 'runtime', '1.3.0', '1.3.0','container'),
        createArtifact('Nvidia Driver', 'runtime', '2.11.0', '2.11.0', 'container'),   
    ], 'server', [ ], [], 'instances: 3', {nodes: [
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
    ],}),
    createData('Kiosk', [                
        createArtifact('Kiosk UWP', 'os', '1.3', "1.3", 'intune'),
        createArtifact('Order Manager', 'runtime', '1.0.0', '1.0.0','container'),        
    ], 'server', [ ], [], 'instances: 2', {nodes: [
        { id: 1, label: 'Kiosk UWP' },
        { id: 2, label: 'Order Manager' }
    ],
    edges: [        
    ],})
];

export default function Solutions() {
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
    {rows.map((row,index)=>{
      return (
      <Card key={'c'+row.name}>
        <CardHeader
          avatar={
            <Avatar  aria-label="recipe">
            {row.type ==='server'?
              <DnsIcon/>:(row.type === 'router'?
              <RouterIcon/>:(row.type === 'ups'?
              <BoltIcon/>:(row.type === 'tv'?
              <VideoLabelIcon/>:(row.type === 'printer'?
              <PrintIcon/>: (row.type === 'pets'?
              <PetsIcon/>: 
              <PrecisionManufacturingIcon/>)))))}
            </Avatar>}              
          title={row.name}
          subheader={row.site}/>
        <CardContent>
            <IconTabs  content1={
                <Stack className="flex items-start rounded-md bg-gray-100 ">
                { row.stack.map((c)=>(
                    <Stack direction="row">
                        <Chip label={c.name + ' (' + c.version+ ')' + (c.version === c.current? '': ' \u2190 ' + c.current)} size="small" color={c.version === c.current ? 'success': 'error'} icon={c.type === 'agent' ? <FaceIcon/>:(
                            c.type === 'os'? <WindowIcon/>:(
                            c.type === 'policy'? <PolicyIcon/>:
                            <SettingsApplicationsIcon/>))} className="m-1" variant="outlined"/>          
                            {c.manager === 'azurepolicy'?<Avatar alt="Arc" src="/images/Policy.svg" sx={{ width: 18, height: 22 }} className="m-1"/>:(
                            c.manager === 'gatekeeper'?<Avatar alt="Arc" src="/images/opa.png" sx={{ width: 18, height: 22 }} className="m-1"/>:(
                            c.manager === 'intune'?<Avatar alt="Arc" src="/images/intune.png" sx={{ width: 18, height: 22 }} className="m-1"/>:
                            <Avatar alt="Arc" src="/images/Azure-Arc.svg" sx={{ width: 26, height: 22 }} className="m-1"/>))}
                    </Stack>))}
                </Stack>}            
                content2 = {<GraphVis key={'g'+row.name} data={row.graph} />}
             />
          
        </CardContent>
      </Card>);})}
    </Box>
  </div>);
}