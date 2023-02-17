'use client';

import * as React from 'react';
import Box from '@mui/material/Box';
import DeleteIcon from '@mui/icons-material/Delete';
import Toolbar from '@mui/material/Toolbar';
import Button from '@mui/material/Button';
import Menu from '@mui/material/Menu';
import MenuItem from '@mui/material/MenuItem';
import CloudUploadIcon from '@mui/icons-material/CloudUpload';
import ListItemIcon from '@mui/material/ListItemIcon';
import UploadFileIcon from '@mui/icons-material/UploadFile';
import Card from '@mui/material/Card';
import CardHeader from '@mui/material/CardHeader';
import CardMedia from '@mui/material/CardMedia';
import CardContent from '@mui/material/CardContent';
import CardActions from '@mui/material/CardActions';
import Avatar from '@mui/material/Avatar';
import IconButton from '@mui/material/IconButton';
import FavoriteIcon from '@mui/icons-material/Favorite';
import CleaningServicesIcon from '@mui/icons-material/CleaningServices';
import StartIcon from '@mui/icons-material/Start';
import Link from 'next/link';

interface Data {
    id: string;
    name: string;
    desc: string;
}
function createData(
    id: string,
    name: string,
    desc: string
): Data {
    return {
        id,
        name,
        desc
    };
}

const rows = [
    createData('t1', 'Smart Agriculture', 'Improve crop yields, reduce waste, and increase sustainability using sensors, drones, and other technology.'),
    createData('t2', 'Manufacture Automation', 'Automate and streamline production processes with robots and computer vision.'),
    createData('t3', 'Smart Building', 'Enhance safety, comfort, and efficiency of buildings using automated lighting, climate control, and enhanced security.'),
    createData('t4', 'Worksite Safety', 'Improve worksite safety and reduce accidents using intelligent technologies, such as wearable devices and real-time monitoring systems.'),
    createData('t5', 'Personalized Healthcare', 'Improve patient outcomes and enhance personalized care using advanced technologies, such as machine learning and wearable devices.'),
];

export default function Templates() {
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
      <Card key={'c'+row.id} sx={{width: '20%'}}>
        <CardHeader
          avatar={
            <Avatar  aria-label="recipe">
              <CleaningServicesIcon/>              
            </Avatar>}              
          title={row.name}
          subheader=""/>
         <CardMedia
            component="img"            
            sx={{ height: 200, objectFit: 'cover'}}
            image={
                row.id === "t1" ?
                    "/images/drones.png" : (
                row.id === "t2" ?
                     "/images/robots.png":( 
                row.id === "t3" ?
                    "/images/building.png":(
                row.id === "t4" ?
                    "/images/cameras.png":
                     "/images/doctor.png"))
                    )
            }
            alt="Paella dish"
        / >
        <CardContent>
            <p>{row.desc}</p>
        </CardContent>
        <CardActions disableSpacing>
        <IconButton aria-label="add to favorites">
          <FavoriteIcon />
        </IconButton>
        <IconButton aria-label="share">
            <Link href="/wizard/step1">
            <StartIcon />
            </Link>
        </IconButton>       
      </CardActions>
      </Card>);})}
    </Box>
  </div>);
}