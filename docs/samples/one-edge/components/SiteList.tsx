'use client';

import { Site } from '../app/types';
import SiteCard from './SiteCard';
import {Tabs, Tab} from "@nextui-org/react";
import { PiCards } from 'react-icons/pi';
import { PiTable } from 'react-icons/pi';
import { PiMapTrifold } from 'react-icons/pi';
interface SiteListProps {
    sites: Site[];
}
function SiteList(props: SiteListProps) {
    const { sites } = props;
    return (
        <Tabs aria-label="Options" color="primary" variant="bordered">
            <Tab key="cards"
                title={
                    <div className="flex items-center space-x-2">
                        <PiCards />
                        <span>Cards</span>
                        </div>}>
                <div className='sitelist'>
                    {sites.map((site: any) =>  <SiteCard site={site} />)}
                </div>
            </Tab>
            <Tab key="table"
                title={
                    <div className="flex items-center space-x-2">
                        <PiTable />
                        <span>Table</span>
                        </div>}>
                <div className='sitelist'>
                    Coming soon ...
                </div>
            </Tab>
            <Tab key="map"
                title={
                    <div className="flex items-center space-x-2">
                        <PiMapTrifold />
                        <span>Map</span>
                        </div>}>
                <div className='sitelist'>
                    Coming soon ...
                </div>
            </Tab>
        </Tabs>
    );
}
export default SiteList;