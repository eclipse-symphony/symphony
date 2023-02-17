import * as React from 'react';
import Tabs from '@mui/material/Tabs';
import Tab from '@mui/material/Tab';
import Typography from '@mui/material/Typography';
import Box from '@mui/material/Box';
import PhoneIcon from '@mui/icons-material/Phone';
import FavoriteIcon from '@mui/icons-material/Favorite';
import PersonPinIcon from '@mui/icons-material/PersonPin';
import TableRowsIcon from '@mui/icons-material/TableRows';
import SchemaIcon from '@mui/icons-material/Schema';
import HubIcon from '@mui/icons-material/Hub';
interface TabPanelProps {
  children?: React.ReactNode;
  index: number;
  value: number;
}

function TabPanel(props: TabPanelProps) {
  const { children, value, index, ...other } = props;

  return (
    <div
      role="tabpanel"
      hidden={value !== index}
      id={`simple-tabpanel-${index}`}
      aria-labelledby={`simple-tab-${index}`}
      {...other}
    >
      {value === index && (
        <Box sx={{ p: 3 }}>
          <Typography>{children}</Typography>
        </Box>
      )}
    </div>
  );
}

function a11yProps(index: number) {
  return {
    id: `simple-tab-${index}`,
    'aria-controls': `simple-tabpanel-${index}`,
  };
}

export default function IconTabs({content1, content2}) {
  const [value, setValue] = React.useState(0);

  const handleChange = (event: React.SyntheticEvent, newValue: number) => {
    setValue(newValue);
  };

  return (
    <Box sx={{ width: '100%' }}>      
      <TabPanel value={value} index={0}>
        {content1}
      </TabPanel>
      <TabPanel value={value} index={1}>
        <div style={{height:500, width:400}}>
            {content2}
        </div>
      </TabPanel>
      <TabPanel value={value} index={2}>
        Item Three
      </TabPanel>
      <Box sx={{ borderTop: 1, borderColor: 'divider' }}>
        <Tabs value={value} onChange={handleChange} aria-label="basic tabs example" >          
          <Tab icon={<TableRowsIcon />} {...a11yProps(0)} aria-label="phone" />
            <Tab icon={<HubIcon />} {...a11yProps(1)} aria-label="favorite" />            
        </Tabs>
      </Box>
    </Box>
  );
}