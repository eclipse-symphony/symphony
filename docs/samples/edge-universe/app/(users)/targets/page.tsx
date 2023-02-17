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
import Divider from '@mui/material/Divider';

interface Data {
    connections: State[];
    site: string;
    solutions: State[];
    name: string;
    stack: Artifact[];
    type: string;    
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
): Data {
    return {
        name,
        stack,
        type,
        connections,
        solutions,
        site,        
    };
}

const rows = [
    createData('Lenovo SE350', [            
        createArtifact('My Policy', 'policy', '1.0', "1.0", 'gatekeeper'),
        createArtifact('Microsoft Defender', 'agent', '1.381.2387.0', "1.381.2387.0", 'intune'),
        createArtifact('Symphony', 'agent', '0.41.68', '0.41.68', 'arc'),
        createArtifact('ADU', 'agent', '1.0.0', '1.0.0','adu'),
        createArtifact('Docker', 'runtime', '1.13.1', '1.13.0','intune'),
        createArtifact('Tensorflow', 'runtime', '2.11.0', '2.11.0', 'adu'),
        createArtifact('My Config', 'config', '1.3', '1.3', 'osconfig'),
        createArtifact('Windows IoT Core', 'os', '1809', '1809', 'intune'),
    ], 'server', [
        createState('front A', 'connected', 'camera'),
        createState('front B', 'connected', 'camera'),
        createState('main switch', 'static', 'router'),
    ], [
        createState('Queue Monitor', 'connected', 'server')
    ], 'Las Vegas'),  
    createData('ASE Pro2', [                
        createArtifact('My Policy', 'policy', '1.3', "1.3", 'gatekeeper'),
        createArtifact('Microsoft Defender', 'agent', '1.381.2387.0', "1.381.2387.0", 'intune'),
        createArtifact('Symphony', 'agent', '0.41.68', "0.41.68", 'arc'),
        createArtifact('ADU', 'agent', '1.0.0', "1.0.0", 'adu'),
        createArtifact('Docker', 'runtime', '1.13.1', "1.13.1", 'intune'),
        createArtifact('Windows IoT Core', 'os', '1809', '1809', 'intune'),
    ], 'server', [
        createState('station 1', 'connected', 'camera'),
        createState('station 2', 'connected', 'camera'),
        createState('station 3', 'connected', 'camera'),
        createState('main switch', 'static', 'router'),
    ], [
        createState('Worker Safety', 'connected', 'server')
    ], 'New York'),
    createData('Win Server B', [                
      createArtifact('Microsoft Defender', 'agent', '1.381.2387.0', "1.381.2387.0", 'intune'),
        createArtifact('Symphony', 'agent', '0.41.68', "0.41.68", 'arc'),
        createArtifact('ADU', 'agent', '1.0.0', "1.0.0", 'adu'),
        createArtifact('Docker', 'runtime', '1.13.1', "1.13.1", 'intune'),
        createArtifact('Windows IoT Core', 'os', '1809', '1809', 'intune'),
    ], 'server', [
        createState('loading', 'disconnected', 'camera'),
        createState('flipper', 'connected', 'robot'),
        createState('sorter', 'connected', 'robot'),
    ], [
        createState('IMS', 'disconnected', 'server')
    ], 'Miami'),
    createData('Win Server A', [                
      createArtifact('Microsoft Defender', 'agent', '1.381.2387.0', "1.381.2387.0", 'intune'),
        createArtifact('Symphony', 'agent', '0.41.68', "0.41.68", 'arc'),
        createArtifact('ADU', 'agent', '1.0.0', "0.9", 'adu'),
        createArtifact('Docker', 'runtime', '1.13.1', "1.13.1", 'intune'),
        createArtifact('Windows IoT Core', 'os', '1809', '1809', 'intune'),
    ], 'server', [
        createState('twister', 'connected', 'robot'),    
    ], [], 'Miami'),
    createData('AKS 1', [
        createArtifact('Cloud Policy', 'policy', '1.0', "1.0", 'azurepolicy'),
        createArtifact('Arc', 'agent', '1.13', "1.13", 'arc'),
        createArtifact('Flux', 'agent', '1.25.4', "1.25.3", 'arc'),
    ], 'server', [
        createState('flipper', 'connected', 'robot'),
        createState('twister', 'connected', 'robot'),    
    ], [], 'Las Vegas'),  
];

export default function Targets() {
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
      <Card>
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
          </Stack>
          <Divider />
          <h1>Connected Devices</h1>
          {row.connections.map((c)=>(
            <Chip label={c.name} size="small" color={c.status==='connected'? 'success': (
                                c.status === 'static'? 'info':
                                'error')} icon={c.annotation === 'camera' ? <PhotoCameraIcon/>:(
                                    c.annotation === 'robot'? <PrecisionManufacturingIcon/>:
                                    <RouterIcon/>)} className="m-1"/>                                

                        ))}
          <Divider />
          <h1>Solutions</h1>
          {row.solutions.map((c)=>(
            <Chip label={c.name} size="small" color={c.status==='connected'? 'success': (
                                c.status === 'static'? 'info':
                                'error')} className="m-1"/>))}
        </CardContent>
      </Card>);})}
    </Box>
  </div>);
}