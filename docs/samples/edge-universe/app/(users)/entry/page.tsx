"use client";
import React, { Suspense } from "react";
import { useState, useEffect } from 'react';
import Table from '@mui/material/Table';
import TableBody from '@mui/material/TableBody';
import TableCell, { tableCellClasses } from '@mui/material/TableCell';
import TableHead from '@mui/material/TableHead';
import TableRow from '@mui/material/TableRow';
import TextField from '@mui/material/TextField';
import Paper from '@mui/material/Paper';
import Button from '@mui/material/Button';
import TableContainer from '@mui/material/TableContainer';
import { styled } from '@mui/material/styles';
import Typography from '@mui/material/Typography';
import Fab from '@mui/material/Fab';
import NavigationIcon from '@mui/icons-material/Save';

const StyledTableCell = styled(TableCell)(({ theme }) => ({
    [`&.${tableCellClasses.head}`]: {
      backgroundColor: theme.palette.common.black,
      color: theme.palette.common.white,
    },
    [`&.${tableCellClasses.body}`]: {
      fontSize: 14,
    },
  }));
  
  const StyledTableRow = styled(TableRow)(({ theme }) => ({
    '&:nth-of-type(odd)': {
      backgroundColor: theme.palette.action.hover,
    },
    // hide last border
    '&:last-child td, &:last-child th': {
      border: 0,
    },
  }));

function classNames(...classes: string[]) {
    return classes.filter(Boolean).join(' ')
}
function Entry() {
    const [perspectives, setPerspectives] = useState([]);
    const [perspectiveValues, setPerspectiveValues] = useState({});

    const handleCopyClick = (name, value) => {
        setPerspectiveValues({...perspectiveValues, [name]: value});
    };

    const handleSaveClick = () => {
        const textFields = document.querySelectorAll("input");        
        let data = {};
        textFields.forEach((textField)=> {
            if (textField.value) {
                data[textField.name] = textField.value;
            }
        })
        alert(JSON.stringify(data));
    };

    useEffect(()=> {
      const callAPI = async () => {
        const ret = new Map();
        const response = await fetch('/api/perspectives');
        const data = await response.json();        
        await Promise.all(data.map(async (p)=>{
          const r = await fetch('/api/perspective?name=' + p + '&twin=bob');            
          ret.set(p, await r.json());
        }));        
        return ret;
      }
      callAPI().then((data)=>{
        const arr = Array.from(data.entries());
        const values = arr.map(([_,value])=>value);        
        setPerspectives(values);        
      });
    },[]);
    
    return (        
        <div className="w-full h-screen">                        
            {perspectives.map((perspective, idx) => (
                <Paper className="pt-6 pb-6" elevation={12}>
                    <TableContainer key={idx}>
                        <Typography variant="h5">{perspective.name}</Typography>
                        <Table size="small" aria-label="simple table">
                            <TableHead>
                                <TableRow>
                                    <StyledTableCell>Property</StyledTableCell>
                                    <StyledTableCell>Current</StyledTableCell>
                                    <StyledTableCell></StyledTableCell>
                                    <StyledTableCell>New</StyledTableCell>
                                </TableRow>
                            </TableHead>
                            <TableBody>
                            {Object.entries(perspective.properties)
                                .map(([name,value]) => {                            
                                return <StyledTableRow key={name}  sx={{ '&:last-child td, &:last-child th': { border: 0 } }}>
                                    <TableCell component="th" scope="row">{name}</TableCell>
                                    <TableCell>{value}</TableCell>
                                    <TableCell><Button variant="contained" color="primary" onClick={() => handleCopyClick(name, value)}>Copy</Button></TableCell>
                                    <TableCell><TextField name={name} value={perspectiveValues[name] || ''} variant="outlined" onChange={(event)=>{
                                        handleCopyClick(event.target.name, event.target.value);
                                    }}/></TableCell>
                                </StyledTableRow>;
                                })}
                            </TableBody>
                        </Table>                    
                    </TableContainer>
                </Paper>
            ))}       
             <Fab variant="extended" style={{position:'fixed', bottom:16, right:16, background: '#00FF00'}} onClick={()=>handleSaveClick()}>
                <NavigationIcon sx={{ mr: 1 }} />
                Save
            </Fab>                
        </div>        
    );
}

export default Entry;