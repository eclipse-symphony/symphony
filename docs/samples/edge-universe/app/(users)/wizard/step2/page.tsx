'use client';

import * as React from 'react';
import { alpha } from '@mui/material/styles';
import Box from '@mui/material/Box';
import Table from '@mui/material/Table';
import TableBody from '@mui/material/TableBody';
import TableCell from '@mui/material/TableCell';
import TableContainer from '@mui/material/TableContainer';
import TableHead from '@mui/material/TableHead';
import TablePagination from '@mui/material/TablePagination';
import TableRow from '@mui/material/TableRow';
import TableSortLabel from '@mui/material/TableSortLabel';
import Typography from '@mui/material/Typography';
import Paper from '@mui/material/Paper';
import Checkbox from '@mui/material/Checkbox';
import IconButton from '@mui/material/IconButton';
import Tooltip from '@mui/material/Tooltip';
import FormControlLabel from '@mui/material/FormControlLabel';
import Switch from '@mui/material/Switch';
import DeleteIcon from '@mui/icons-material/Delete';
import FilterListIcon from '@mui/icons-material/FilterList';
import { visuallyHidden } from '@mui/utils';
import PhotoCameraIcon from '@mui/icons-material/PhotoCamera';
import PrecisionManufacturingIcon from '@mui/icons-material/PrecisionManufacturing';
import RouterIcon from '@mui/icons-material/Router';
import BoltIcon from '@mui/icons-material/Bolt';
import VideoLabelIcon from '@mui/icons-material/VideoLabel';
import PrintIcon from '@mui/icons-material/Print';
import PetsIcon from '@mui/icons-material/Pets';
import Chip from '@mui/material/Chip';
import DnsIcon from '@mui/icons-material/Dns';
import CloudIcon from '@mui/icons-material/Cloud';
import Toolbar from '@mui/material/Toolbar';
import Button from '@mui/material/Button';
import Menu from '@mui/material/Menu';
import MenuItem from '@mui/material/MenuItem';
import CloudUploadIcon from '@mui/icons-material/CloudUpload';
import ListItemText from '@mui/material/ListItemText';
import ListItemIcon from '@mui/material/ListItemIcon';
import UploadFileIcon from '@mui/icons-material/UploadFile';
import AccountTreeIcon from '@mui/icons-material/AccountTree';
import GitHubIcon from '@mui/icons-material/GitHub';
import TextSnippetIcon from '@mui/icons-material/TextSnippet';
import StorageIcon from '@mui/icons-material/Storage';

interface Data {
    connections: State[];
    site: string;
    solutions: State[];
    name: string;
    registry: string;
    type: string;    
}

interface State {
    name: string,
    status: string,
    annotation: string
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
    registry: string,
    type: string,
    connections: State[],
    solutions: State[],
    site: string,    
): Data {
    return {
        name,
        registry,
        type,
        connections,
        solutions,
        site,        
    };
}

const rows = [
  createData('front A', 'Akri (ONVIF)', 'camera', [
    createState('Lenovo SE350', 'connected', 'server')], [
        createState('Queue Monitor', 'connected', 'server')
    ], 'Las Vegas'),
  createData('front B', 'Akri (ONVIF)', 'camera', [
    createState('Lenovo SE350', 'connected', 'server')], [
        createState('Queue Monitor', 'connected', 'server')
    ], 'Las Vegas'),
  createData('station 1', 'Akri (ONVIF)', 'camera', [
    createState('ASE Pro2', 'connected', 'server')], [
        createState('Worker Safety', 'connected', 'server')
    ], 'New York'),
  createData('station 2', 'Akri (ONVIF)', 'camera', [
    createState('ASE Pro2', 'connected', 'server')], [
        createState('Worker Safety', 'connected', 'server')
    ], 'New York'),
  createData('station 3', 'Akri (ONVIF)', 'camera', [
    createState('ASE Pro2', 'connected', 'server')], [
        createState('Worker Safety', 'connected', 'server')
    ], 'Seattle'),
  createData('camera switch', 'Symphony', 'router', [
    createState('front A', 'static', 'camera'),
    createState('front B', 'static', 'camera'),
    createState('station 1', 'static', 'camera'),
    createState('station 2', 'static', 'camera'),
    createState('station 3', 'static', 'camera'),
    ], [], 'Seattle'),
  createData('main switch', 'Symphony','router',  [
    createState('ASE Pro2', 'static', 'server'),
    createState('Leveno SE350', 'static', 'server')], [], 'Seattle'),
  createData('loading', 'Akri (OPC-UA)', 'camera', [
    createState('Win Server B', 'disconnected', 'server'),
  ], [
    createState('IMS', 'disconnected', 'server')
  ], 'Miami'),
  createData('flipper', 'DPS', 'robot', [
    createState('AKS 1', 'connected', 'k8s'),
    createState('Win Server B', 'disconnected', 'server'),
  ], [], 'Miami'),
  createData('twister', 'DPS', 'robot', [
    createState('AKS 1', 'connected', 'k8s'),
    createState('Win Server A', 'connected', 'server'),
  ], [], 'Miami'),
  createData('sealer', 'DPS', 'robot', [], [], 'Las Vegas'),
  createData('ups', 'Symphony', 'ups', [], [], 'New York'),
  createData('sorter', 'Akri (OPC-UA)', 'robot', [
    createState('Win Server B', 'disconnected', 'server'),
  ], [], 'Austin'),
  createData('tv 1', 'Custom (SQL Server)', 'tv', [], [], 'Seattle'),
  createData('tv 2', 'Custom (SQL Server)', 'tv', [], [], 'Seattle'),
  createData('printer', 'Custom (SAP)', 'printer', [], [], 'Las Vegas'),
  createData('snowball', 'Symphony', 'pets', [], [
    createState('Moral', 'static', 'server')
  ], 'New York'),
  createData('sorter5', 'Akri (OPC-UA)', 'robot', [], [], 'New York'),
];

function descendingComparator<T>(a: T, b: T, orderBy: keyof T) {
  if (b[orderBy] < a[orderBy]) {
    return -1;
  }
  if (b[orderBy] > a[orderBy]) {
    return 1;
  }
  return 0;
}

type Order = 'asc' | 'desc';

function getComparator<Key extends keyof any>(
  order: Order,
  orderBy: Key,
): (
  a: { [key in Key]: number | string },
  b: { [key in Key]: number | string },
) => number {
  return order === 'desc'
    ? (a, b) => descendingComparator(a, b, orderBy)
    : (a, b) => -descendingComparator(a, b, orderBy);
}

// Since 2020 all major browsers ensure sort stability with Array.prototype.sort().
// stableSort() brings sort stability to non-modern browsers (notably IE11). If you
// only support modern browsers you can replace stableSort(exampleArray, exampleComparator)
// with exampleArray.slice().sort(exampleComparator)
function stableSort<T>(array: readonly T[], comparator: (a: T, b: T) => number) {
  const stabilizedThis = array.map((el, index) => [el, index] as [T, number]);
  stabilizedThis.sort((a, b) => {
    const order = comparator(a[0], b[0]);
    if (order !== 0) {
      return order;
    }
    return a[1] - b[1];
  });
  return stabilizedThis.map((el) => el[0]);
}

interface HeadCell {
  disablePadding: boolean;
  id: keyof Data;
  label: string;
  numeric: boolean;
}

const headCells: readonly HeadCell[] = [    
    {
        id: 'type',
        numeric: false,
        disablePadding: true,
        label: '',
      },
  {
    id: 'name',
    numeric: false,
    disablePadding: true,
    label: 'Name',
  },
 
  {
    id: 'registry',
    numeric: false,
    disablePadding: false,
    label: 'Registry',
  },
  {
    id: 'connections',
    numeric: false,
    disablePadding: false,
    label: 'Connections',
  },
  {
    id: 'solutions',
    numeric: false,
    disablePadding: false,
    label: 'Capabilities',
  },
  {
    id: 'site',
    numeric: false,
    disablePadding: false,
    label: 'Site',
  },
];

interface EnhancedTableProps {
  numSelected: number;
  onRequestSort: (event: React.MouseEvent<unknown>, property: keyof Data) => void;
  onSelectAllClick: (event: React.ChangeEvent<HTMLInputElement>) => void;
  order: Order;
  orderBy: string;
  rowCount: number;
}

function EnhancedTableHead(props: EnhancedTableProps) {
  const { onSelectAllClick, order, orderBy, numSelected, rowCount, onRequestSort } =
    props;
  const createSortHandler =
    (property: keyof Data) => (event: React.MouseEvent<unknown>) => {
      onRequestSort(event, property);
    };

  return (
    <TableHead>
      <TableRow>
        <TableCell padding="checkbox">
          <Checkbox
            color="primary"
            indeterminate={numSelected > 0 && numSelected < rowCount}
            checked={rowCount > 0 && numSelected === rowCount}
            onChange={onSelectAllClick}
            inputProps={{
              'aria-label': 'select all desserts',
            }}
          />
        </TableCell>
        {headCells.map((headCell) => (
          <TableCell
            key={headCell.id}
            align={headCell.numeric ? 'right' : 'left'}
            padding={headCell.disablePadding ? 'none' : 'normal'}
            sortDirection={orderBy === headCell.id ? order : false}
          >
            <TableSortLabel
              active={orderBy === headCell.id}
              direction={orderBy === headCell.id ? order : 'asc'}
              onClick={createSortHandler(headCell.id)}
            >
              {headCell.label}
              {orderBy === headCell.id ? (
                <Box component="span" sx={visuallyHidden}>
                  {order === 'desc' ? 'sorted descending' : 'sorted ascending'}
                </Box>
              ) : null}
            </TableSortLabel>
          </TableCell>
        ))}
      </TableRow>
    </TableHead>
  );
}

interface EnhancedTableToolbarProps {
  numSelected: number;
}

function EnhancedTableToolbar(props: EnhancedTableToolbarProps) {
  const { numSelected } = props;

  return (
    <Toolbar
      sx={{
        pl: { sm: 2 },
        pr: { xs: 1, sm: 1 },
        ...(numSelected > 0 && {
          bgcolor: (theme) =>
            alpha(theme.palette.primary.main, theme.palette.action.activatedOpacity),
        }),
      }}
    >
      {numSelected > 0 ? (
        <Typography
          sx={{ flex: '1 1 100%' }}
          color="inherit"
          variant="subtitle1"
          component="div"
        >
          {numSelected} selected
        </Typography>
      ) : (
        <Typography
          sx={{ flex: '1 1 100%' }}
          variant="h6"
          id="tableTitle"
          component="div"
        >
          Devices
        </Typography>
      )}
      {numSelected > 0 ? (
        <Tooltip title="Delete">
          <IconButton>
            <DeleteIcon />
          </IconButton>
        </Tooltip>
      ) : (
        <Tooltip title="Filter list">
          <IconButton>
            <FilterListIcon />
          </IconButton>
        </Tooltip>
      )}
    </Toolbar>
  );
}

export default function Step2() {
    const [clicks, setClicks] = React.useState<google.maps.LatLng[]>([]);      
    const [anchorEl, setAnchorEl] = React.useState<null | HTMLElement>(null);
    const open = Boolean(anchorEl);
    const handleMenuClick = (event: React.MouseEvent<HTMLButtonElement>) => {
        setAnchorEl(event.currentTarget);
    };
    const handleClose = () => {
        setAnchorEl(null);
    };

  const [order, setOrder] = React.useState<Order>('asc');
  const [orderBy, setOrderBy] = React.useState<keyof Data>('name');
  const [selected, setSelected] = React.useState<readonly string[]>([]);
  const [page, setPage] = React.useState(0);
  const [dense, setDense] = React.useState(true);
  const [rowsPerPage, setRowsPerPage] = React.useState(15);

  const handleRequestSort = (
    event: React.MouseEvent<unknown>,
    property: keyof Data,
  ) => {
    const isAsc = orderBy === property && order === 'asc';
    setOrder(isAsc ? 'desc' : 'asc');
    setOrderBy(property);
  };

  const handleSelectAllClick = (event: React.ChangeEvent<HTMLInputElement>) => {
    if (event.target.checked) {
      const newSelected = rows.map((n) => n.name);
      setSelected(newSelected);
      return;
    }
    setSelected([]);
  };

  const handleClick = (event: React.MouseEvent<unknown>, name: string) => {
    const selectedIndex = selected.indexOf(name);
    let newSelected: readonly string[] = [];

    if (selectedIndex === -1) {
      newSelected = newSelected.concat(selected, name);
    } else if (selectedIndex === 0) {
      newSelected = newSelected.concat(selected.slice(1));
    } else if (selectedIndex === selected.length - 1) {
      newSelected = newSelected.concat(selected.slice(0, -1));
    } else if (selectedIndex > 0) {
      newSelected = newSelected.concat(
        selected.slice(0, selectedIndex),
        selected.slice(selectedIndex + 1),
      );
    }

    setSelected(newSelected);
  };

  const handleChangePage = (event: unknown, newPage: number) => {
    setPage(newPage);
  };

  const handleChangeRowsPerPage = (event: React.ChangeEvent<HTMLInputElement>) => {
    setRowsPerPage(parseInt(event.target.value, 10));
    setPage(0);
  };

  const handleChangeDense = (event: React.ChangeEvent<HTMLInputElement>) => {
    setDense(event.target.checked);
  };

  const isSelected = (name: string) => selected.indexOf(name) !== -1;

  // Avoid a layout jump when reaching the last page with empty rows.
  const emptyRows =
    page > 0 ? Math.max(0, (1 + page) * rowsPerPage - rows.length) : 0;

  return ( <div>
     <Toolbar>
            <Button
                id="basic-button"
                aria-controls={open ? 'basic-menu' : undefined}
                aria-haspopup="true"
                aria-expanded={open ? 'true' : undefined}
                onClick={handleMenuClick}>
                ONBOARD
            </Button>
            <Menu
                id="basic-menu"
                anchorEl={anchorEl}
                open={open}
                onClose={handleClose}
                MenuListProps={{
                'aria-labelledby': 'basic-button',
                }}>
                <MenuItem onClick={handleClose}>
                    <ListItemIcon>
                        <CloudUploadIcon fontSize="small" />
                    </ListItemIcon>
                    Arc Clusters...</MenuItem>
                <MenuItem onClick={handleClose}>
                    <ListItemIcon>
                        <UploadFileIcon fontSize="small" />
                    </ListItemIcon>
                    DPS...</MenuItem>
                <MenuItem onClick={handleClose}>
                    <ListItemIcon>
                        <AccountTreeIcon fontSize="small" />
                    </ListItemIcon>
                    Azure Digital Twin...</MenuItem>
                <MenuItem onClick={handleClose}>
                    <ListItemIcon>
                        <StorageIcon fontSize="small" />
                    </ListItemIcon>
                    Custom Repo...</MenuItem>
                <MenuItem onClick={handleClose}>
                    <ListItemIcon>
                        <TextSnippetIcon fontSize="small" />
                    </ListItemIcon>
                    CSV...</MenuItem>
                    <MenuItem onClick={handleClose}>
                    <ListItemIcon>
                        <GitHubIcon fontSize="small" />
                    </ListItemIcon>
                    Git...</MenuItem>
            </Menu>
        </Toolbar>       
    <Box sx={{ width: '100%' }}>
      <Paper sx={{ width: '100%', mb: 2 }}>
        <EnhancedTableToolbar numSelected={selected.length} />
        <TableContainer>
          <Table
            sx={{ minWidth: 750 }}
            aria-labelledby="tableTitle"
            size={dense ? 'small' : 'medium'}
          >
            <EnhancedTableHead
              numSelected={selected.length}
              order={order}
              orderBy={orderBy}
              onSelectAllClick={handleSelectAllClick}
              onRequestSort={handleRequestSort}
              rowCount={rows.length}
            />
            <TableBody>
              {stableSort(rows, getComparator(order, orderBy))
                .slice(page * rowsPerPage, page * rowsPerPage + rowsPerPage)
                .map((row, index) => {
                  const isItemSelected = isSelected(row.name);
                  const labelId = `enhanced-table-checkbox-${index}`;

                  return (
                    <TableRow
                      hover
                      onClick={(event) => handleClick(event, row.name)}
                      role="checkbox"
                      aria-checked={isItemSelected}
                      tabIndex={-1}
                      key={row.name}
                      selected={isItemSelected}
                    >                      
                      <TableCell padding="checkbox">                        
                        <Checkbox
                          color="primary"
                          checked={isItemSelected}
                          inputProps={{
                            'aria-labelledby': labelId,
                          }}
                        />
                      </TableCell>
                      <TableCell>                    
                            {row.type ==='camera'
                                ?<PhotoCameraIcon/>:(
                                    row.type === 'router'
                                    ?<RouterIcon/>:(
                                        row.type === 'ups'
                                    ?<BoltIcon/>:(
                                        row.type === 'tv'
                                    ?<VideoLabelIcon/>:(
                                        row.type === 'printer'
                                    ?<PrintIcon/>: (
                                        row.type === 'pets'
                                    ?<PetsIcon/>: <PrecisionManufacturingIcon/>)))))}
                        </TableCell>
                      <TableCell
                        component="th"
                        id={labelId}
                        scope="row"
                        padding="none"
                      >
                        {row.name}
                      </TableCell>
                      <TableCell>{row.registry}</TableCell>
                      <TableCell>{
                        row.connections.map((c)=>(
                            <Chip label={c.name} size="small" color={c.status==='connected'? 'success': (
                                c.status === 'static'? 'info':
                                'error')} icon={c.annotation === 'camera' ? <PhotoCameraIcon/>:(
                                    c.annotation === 'k8s'? <CloudIcon/>:
                                    <DnsIcon/>)}className="m-1"/>                                

                        ))
                      }</TableCell>
                      <TableCell>{
                        row.solutions.map((c)=>(
                            <Chip label={c.name} size="small" color={c.status==='connected'? 'success': (
                                c.status === 'static'? 'info':
                                'error')} className="m-1"/>                                
                        ))
                                }</TableCell>
                      <TableCell >{row.site}</TableCell>                      
                    </TableRow>
                  );
                })}
              {emptyRows > 0 && (
                <TableRow
                  style={{
                    height: (dense ? 33 : 53) * emptyRows,
                  }}
                >
                  <TableCell colSpan={6} />
                </TableRow>
              )}
            </TableBody>
          </Table>
        </TableContainer>
        <TablePagination
          rowsPerPageOptions={[25,15,5]}
          component="div"
          count={rows.length}
          rowsPerPage={rowsPerPage}
          page={page}
          onPageChange={handleChangePage}
          onRowsPerPageChange={handleChangeRowsPerPage}
        />
      </Paper>
      <FormControlLabel
        control={<Switch checked={dense} onChange={handleChangeDense} />}
        label="Dense padding"
      />
    </Box>
  </div>);
}