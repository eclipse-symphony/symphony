import React from 'react'
import { makeStyles, createStyles } from '@mui/styles'

import List from '@mui/material/List'
import ListItem from '@mui/material/ListItem'
import ListItemIcon from '@mui/material/ListItemIcon'
import ListItemText from '@mui/material/ListItemText'
import Divider from '@mui/material/Divider'
import Collapse from '@mui/material/Collapse'

import IconGlobal from '@mui/icons-material/Public'
import IconExpandLess from '@mui/icons-material/ExpandLess'
import IconExpandMore from '@mui/icons-material/ExpandMore'
import IconDashboard from '@mui/icons-material/Dashboard'
import IconShoppingCart from '@mui/icons-material/ShoppingCart'
import IconPeople from '@mui/icons-material/People'
import IconBarChart from '@mui/icons-material/BarChart'
import IconLibraryBooks from '@mui/icons-material/LibraryBooks'
import VrpanoIcon from '@mui/icons-material/Vrpano';
import Link from 'next/link'
import HubIcon from '@mui/icons-material/Hub';
import DevicesOtherIcon from '@mui/icons-material/DevicesOther';
import DnsIcon from '@mui/icons-material/Dns';
import AppsIcon from '@mui/icons-material/Apps';
import BrowserUpdatedIcon from '@mui/icons-material/BrowserUpdated';
import SettingsSuggestIcon from '@mui/icons-material/SettingsSuggest';
import TuneIcon from '@mui/icons-material/Tune';
import PolicyIcon from '@mui/icons-material/Policy';
import CleaningServicesIcon from '@mui/icons-material/CleaningServices';
import AccountTreeIcon from '@mui/icons-material/AccountTree';
import PsychologyIcon from '@mui/icons-material/Psychology';
import HdrWeakIcon from '@mui/icons-material/HdrWeak';
import ImagesearchRollerIcon from '@mui/icons-material/ImagesearchRoller';
import ImageAspectRatioIcon from '@mui/icons-material/ImageAspectRatio';
import SettingsIcon from '@mui/icons-material/Settings';

const NavBar: React.FC = () => {
  const classes = useStyles();

  const [templatesOpen, setTemplatesOpen] = React.useState(false);
  const [catalogsOpen, setCatalogsOpen] = React.useState(false);
  const [perspectivesOpen, setPerspectivesOpen] = React.useState(false);
  const [open, setOpen] = React.useState(false);

  function handleClick() {
    setOpen(!open)
  }

  return (
    <List component="nav" className={`${classes.appMenu} mt-20`} disablePadding>      
      <ListItem className={classes.menuItem} onClick={()=>setPerspectivesOpen(!perspectivesOpen)}>
        <ListItemIcon className={classes.menuItemIcon}>
          <VrpanoIcon />
        </ListItemIcon>
        <ListItemText primary="Overview" />
        {perspectivesOpen ? <IconExpandLess /> : <IconExpandMore />}
      </ListItem>      
      <Collapse in={perspectivesOpen} timeout="auto" unmountOnExit>
        <Divider />
        <List component="div" disablePadding>
          <ListItem  className={classes.menuItem}>
            <ListItemText inset primary="Dashboard" />
            <ListItemIcon className={classes.menuItemIcon}>
              <IconDashboard />
            </ListItemIcon>                       
          </ListItem>
          <Link href="/overview">
            <ListItem className={classes.menuItem}>     
            <ListItemText inset primary="Map" />
            <ListItemIcon className={classes.menuItemIcon}>
              <IconGlobal />
            </ListItemIcon>                       
            </ListItem>
          </Link>
          <Link href="/graph">
            <ListItem  className={classes.menuItem}>
              <ListItemText inset primary="Graph" />
              <ListItemIcon className={classes.menuItemIcon}>
                <HubIcon />
              </ListItemIcon>                       
            </ListItem>         
          </Link>
        </List>
      </Collapse>
      <Link href="/devices">
        <ListItem className={classes.menuItem}>
          <ListItemIcon className={classes.menuItemIcon}>
            <DevicesOtherIcon />
          </ListItemIcon>
          <ListItemText primary="Devices" />
        </ListItem>
      </Link>
      <Link href="/targets">
      <ListItem button className={classes.menuItem}>
        <ListItemIcon className={classes.menuItemIcon}>
          <DnsIcon />
        </ListItemIcon>
        <ListItemText primary="Targets" />
      </ListItem>
      </Link>
      <Link href="/solutions">
      <ListItem className={classes.menuItem}>
        <ListItemIcon className={classes.menuItemIcon}>
          <AppsIcon />
        </ListItemIcon>
        <ListItemText primary="Solutions" />
      </ListItem>
      </Link>
      <Link href="/instances">
        <ListItem className={classes.menuItem}>
          <ListItemIcon className={classes.menuItemIcon}>
            <SettingsSuggestIcon />
          </ListItemIcon>
          <ListItemText primary="Instances" />        
        </ListItem>
      </Link>
      <ListItem  className={classes.menuItem} onClick={()=>setCatalogsOpen(!catalogsOpen)}>
        <ListItemIcon className={classes.menuItemIcon}>
          <IconBarChart />
        </ListItemIcon>
        <ListItemText primary="Artifacts" />
        {catalogsOpen ? <IconExpandLess /> : <IconExpandMore />}
      </ListItem>
      <Collapse in={catalogsOpen} timeout="auto" unmountOnExit>
        <Divider />
        <List component="div" disablePadding>
          <Link href="/configs">
            <ListItem  className={classes.menuItem}>
              <ListItemText inset primary="Configs" />
              <ListItemIcon className={classes.menuItemIcon}>
                <TuneIcon />
              </ListItemIcon>
            </ListItem>
          </Link>
          <ListItem  className={classes.menuItem}>
            <ListItemText inset primary="Policies" />
            <ListItemIcon className={classes.menuItemIcon}>
              <PolicyIcon />
            </ListItemIcon>
          </ListItem>
          <Divider/>
          <ListItem  className={classes.menuItem}>
            <ListItemText inset primary="AI Models" />
            <ListItemIcon className={classes.menuItemIcon}>
              <PsychologyIcon />
            </ListItemIcon>
          </ListItem>
          <ListItem  className={classes.menuItem}>
            <ListItemText inset primary="AI Skills" />
            <ListItemIcon className={classes.menuItemIcon}>
              <AccountTreeIcon />
            </ListItemIcon>
          </ListItem>
          <ListItem  className={classes.menuItem}>
            <ListItemText inset primary="AI Skill Nodes" />
            <ListItemIcon className={classes.menuItemIcon}>
              <HdrWeakIcon />
            </ListItemIcon>
          </ListItem>
          <Divider/>
          <ListItem  className={classes.menuItem}>
            <ListItemText inset primary="Updates" />
            <ListItemIcon className={classes.menuItemIcon}>
              <BrowserUpdatedIcon />
            </ListItemIcon>
          </ListItem>
          <ListItem  className={classes.menuItem}>
            <ListItemText inset primary="Images" />        
            <ListItemIcon className={classes.menuItemIcon}>
              <ImageAspectRatioIcon />
            </ListItemIcon>    
          </ListItem>          
        </List>
      </Collapse>
      <ListItem  className={classes.menuItem} onClick={()=>setTemplatesOpen(!templatesOpen)}>
        <ListItemIcon className={classes.menuItemIcon}>
          <CleaningServicesIcon />
        </ListItemIcon>
        <ListItemText primary="Templates" />
        {templatesOpen ? <IconExpandLess /> : <IconExpandMore />}
      </ListItem>
      <Collapse in={templatesOpen} timeout="auto" unmountOnExit>
        <Divider />
        <List component="div" disablePadding>
          <Link href="/templates">
            <ListItem  className={classes.menuItem}>
              <ListItemText inset primary="Site" />
              <ListItemIcon className={classes.menuItemIcon}>
                <ImagesearchRollerIcon />
              </ListItemIcon>    
            </ListItem>
          </Link>
          <ListItem  className={classes.menuItem}>
            <ListItemText inset primary="Device" />
            <ListItemIcon className={classes.menuItemIcon}>
              <DevicesOtherIcon />
            </ListItemIcon>    
          </ListItem>
          <ListItem  className={classes.menuItem}>
            <ListItemText inset primary="Target" />
            <ListItemIcon className={classes.menuItemIcon}>
              <DnsIcon />
            </ListItemIcon>    
          </ListItem>          
        </List>        
      </Collapse>
      <Divider />
      <ListItem  className={classes.menuItem}>
        <ListItemIcon className={classes.menuItemIcon}>
          <SettingsIcon />
        </ListItemIcon>
        <ListItemText primary="Settings" />        
      </ListItem>      
    </List>
  )
}

const drawerWidth = 240

const useStyles = makeStyles(theme =>
  createStyles({
    appMenu: {
      
    },
    navList: {
      width: drawerWidth,
    },
    menuItem: {
      width: drawerWidth,
    },
    menuItemIcon: {
      color: '#97c05c',
    },
  }),
)

export default NavBar
