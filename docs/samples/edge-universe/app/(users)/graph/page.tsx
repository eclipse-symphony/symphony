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
import GraphTabs from '@elements/GraphTabs';
import GraphVis from '@elements/GraphVis';
interface Data {
    name: string;
    graph: {}    
}
function createData(
    name: string,
    graph: {},
): Data {
    return {
        name,
        graph
    };
}

const rows = [
    createData('Queue Monitor', {nodes: [
        { id: 1, label: 'Lenovo SE350' , shape: "box", color: "#97C2FC" },
        { id: 2, label: 'ASE Pro2', shape: "box", color: "#97C2FC" },
        { id: 3, label: 'front A' , shape: "diamond", color: "#FB7E81"},
        { id: 4, label: 'front B' , shape: "diamond", color: "#FB7E81"},
        { id: 5, label: 'station 1' , shape: "diamond", color: "#FB7E81"},
        { id: 6, label: 'station 2' , shape: "diamond", color: "#FB7E81"},
        { id: 7, label: 'station 3' , shape: "diamond", color: "#FB7E81"},
        { id: 8, label: 'AKS 1', shape: "box", color: "#97C2FC" },
        {id: 9, label: 'node 1', shape: "circle", color: "#FFFF00"},
        {id: 10, label: 'node 2', shape: "circle", color: "#FFFF00"},
        {id: 11, label: 'node 3', shape: "circle", color: "#FFFF00"},
        {id: 12, label: 'node 4', shape: "circle", color: "#FFFF00"},
        {id: 13, label: 'node 5', shape: "circle", color: "#FFFF00"},
        {id: 14, label: 'node 6', shape: "circle", color: "#FFFF00"},
        {id: 15, label: 'node 7', shape: "circle", color: "#FFFF00"},
        {id: 16, label: 'node 8', shape: "circle", color: "#FFFF00"},
        {id: 17, label: 'node 9', shape: "circle", color: "#FFFF00"},
        {id: 18, label: 'node 10', shape: "circle", color: "#FFFF00"},
        {id: 19, label: 'IoT Edge a', shape: "star",  color: "#C2FABC"},
        {id: 20, label: 'IoT Edge b', shape: "star",  color: "#C2FABC"},
        {id: 21, label: 'IoT Edge c', shape: "star",  color: "#C2FABC"},
    ],
    edges: [
        { from: 3, to: 1 },
        { from: 4, to: 1 },
        { from: 5, to: 2 },
        { from: 6, to: 2 },
        { from: 7, to: 2 },
        {from: 1, to: 2, dashes: true },
        {from: 2, to: 1, dashes: true },
        { from: 9, to: 8, color: "rgb(20,24,200)" },
        { from: 10, to: 8, color: "rgb(20,24,200)" },
        { from: 11, to: 8, color: "rgb(20,24,200)" },
        { from: 12, to: 8, color: "rgb(20,24,200)" },
        { from: 13, to: 8, color: "rgb(20,24,200)" },
        { from: 14, to: 8, color: "rgb(20,24,200)" },
        { from: 15, to: 8, color: "rgb(20,24,200)" },
        { from: 16, to: 8, color: "rgb(20,24,200)" },
        { from: 17, to: 8, color: "rgb(20,24,200)" },
        { from: 18, to: 8, color: "rgb(20,24,200)" },
        {from: 20, to: 21, dashes: true },
        {from: 21, to: 20, dashes: true },
    ],}),  
];

export default function Graph() {
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
     
            <GraphTabs  content1={<GraphVis key={'g1'+row.name} data={row.graph} />}          
                content2 = {<GraphVis key={'g2'+row.name} data={row.graph} />}
             />)})}
    </Box>
  </div>);
}