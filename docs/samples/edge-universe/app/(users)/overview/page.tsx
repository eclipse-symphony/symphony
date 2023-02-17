'use client';

import React from "react";
import Toolbar from '@mui/material/Toolbar';
import Button from '@mui/material/Button';
import Menu from '@mui/material/Menu';
import MenuItem from '@mui/material/MenuItem';
import Map from './map';
function Overview() {
    const [clicks, setClicks] = React.useState<google.maps.LatLng[]>([]);      
    const [anchorEl, setAnchorEl] = React.useState<null | HTMLElement>(null);
    const open = Boolean(anchorEl);
    const handleClick = (event: React.MouseEvent<HTMLButtonElement>) => {
        setAnchorEl(event.currentTarget);
    };
    const handleClose = () => {
        setAnchorEl(null);
    };
      
    return <div className="map-container">
        <Toolbar>
            <Button
                id="basic-button"
                aria-controls={open ? 'basic-menu' : undefined}
                aria-haspopup="true"
                aria-expanded={open ? 'true' : undefined}
                onClick={handleClick}>
                Layers
            </Button>
            <Menu
                id="basic-menu"
                anchorEl={anchorEl}
                open={open}
                onClose={handleClose}
                MenuListProps={{
                'aria-labelledby': 'basic-button',
                }}>
                <MenuItem onClick={handleClose}>Profile</MenuItem>
                <MenuItem onClick={handleClose}>My account</MenuItem>
                <MenuItem onClick={handleClose}>Logout</MenuItem>
            </Menu>
        </Toolbar>        
       <Map markers={[
            {id:'1', lat:36.17, lng: -115.14},
            {id:'2', lat:40.73, lng: -73.94},
            {id:'3', lat:25.76, lng: -80.20},
            {id:'4', lat:47.61, lng: -122.33},
            {id:'4', lat:36.17, lng: -115.14},
            {id:'6', lat:40.73, lng: -73.94},
            {id:'7', lat:25.76, lng: -80.20},
            {id:'8', lat:47.61, lng: -122.33}
       ]} />
    </div>;
}

export default Overview;