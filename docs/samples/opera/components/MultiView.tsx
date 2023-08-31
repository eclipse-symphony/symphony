'use client';

import {Navbar, NavbarBrand, NavbarContent, NavbarItem, Link, Button, DropdownItem, DropdownTrigger, Dropdown, DropdownMenu} from "@nextui-org/react";
import {FiMenu} from 'react-icons/fi';
import { FiPlus } from 'react-icons/fi';
import {Tabs, Tab} from "@nextui-org/react";
import { PiCards } from 'react-icons/pi';
import { PiTable } from 'react-icons/pi';
import { PiMapTrifold } from 'react-icons/pi';
import { useState } from 'react';
import CampaignCardList from "./campaigns/CampaignCardList";
import SiteCardList from "./sites/SiteCardList";
import SiteMap from "./sites/SiteMap";
import AssetList from "./assets/AssetList";
import GraphTable from "./graph/GraphTable";

interface MenuInfo {
    name: string;
    href: string;    
}

interface ColumnSet {
    name: string;
    data: any[];
}

interface Params {
    type: string;
    menuItems: MenuInfo[];
    views: string[];
    items: any[];
    refItems?: any[];
    columns: ColumnSet[];
}

interface MultiViewProps {
    params: Params;
}

function MultiView(props: MultiViewProps) {
    const { params } = props;
    const [selected, setSelected] = useState("");
    const [selectedColumn, setSelectedColumn] = useState("");
    function handleSelectionChange(key: any) {
        setSelected(key.toString());
    }
    function handleColumnSelectionChange(key: any) {
        setSelectedColumn(key.toString());
    }
    return (
        <div>
            <Navbar isBordered className="top_navbar">
                <NavbarContent justify="start">
                    <Dropdown>
                        <NavbarItem>
                            <DropdownTrigger>
                                <Button disableRipple className="p-0 bg-transparent data-[hover=true]:bg-transparent text-2xl" radius="sm" variant="light">                          
                                    <FiMenu />      
                                </Button>
                            </DropdownTrigger>
                        </NavbarItem>
                        <DropdownMenu aria-label="View features" className="w-[340px]" itemClasses={{base: "gap-4",}}>
                            {params.menuItems.map((item: MenuInfo) => (
                                <DropdownItem key={item.name}>
                                    <Link href={item.href} className="flex gap-2 items-center">
                                        <span><FiPlus/></span>
                                        <span>{item.name}</span>
                                    </Link>
                                </DropdownItem>
                            ))}                            
                        </DropdownMenu>
                    </Dropdown>
                </NavbarContent>    
                <NavbarContent justify="start">
                    <Tabs aria-label="Options" color="primary" variant="bordered" onSelectionChange={handleSelectionChange}>
                        {params.views.map((view: string) => (
                            <Tab key={view}
                                title={
                                    <div className="flex items-center space-x-2">
                                        {view === 'cards' && <PiCards />}  
                                        {view === 'table' && <PiTable />}
                                        {view === 'map' && <PiMapTrifold />}
                                        <span>{view}</span>
                                        </div>}
                                    />  
                        ))}                    
                    </Tabs>
                </NavbarContent>
                {selected === 'table' && params.columns && params.columns.length > 0 && (
                    <NavbarContent justify="start" id="columnSets">
                        <Tabs aria-label="Options" color="primary" variant="bordered" onSelectionChange={handleColumnSelectionChange}>
                            {params.columns.map((column: ColumnSet) => (
                                <Tab key={column.name}
                                    title={
                                        <div className="flex items-center space-x-2">
                                            {column.name === 'configs' && <PiCards />}  
                                            {column.name === 'solutions' && <PiTable />}
                                            {column.name === 'instances' && <PiMapTrifold />}
                                            {column.name === 'targets' && <PiMapTrifold />}
                                            <span>{column.name}</span>
                                            </div>}
                                        />  
                            ))}                    
                        </Tabs>
                    </NavbarContent>
                )}
            </Navbar>
            <div className='view_container'>
                <Tabs isDisabled aria-label="Options" selectedKey={`c${selected}`}>
                    {params.views.map((view: string) => (
                        <Tab key={`c${view}`}
                            title={
                                <div className="flex items-center space-x-2">
                                    {view === 'cards' && <PiCards />}  
                                    {view === 'table' && <PiTable />}
                                    {view === 'map' && <PiMapTrifold />}
                                    <span>{view}</span>
                                    </div>}>
                            {view === 'cards' && params.type === 'campaigns' && <CampaignCardList campaigns={params.items} activations={params.refItems} />}
                            {view === 'cards' && params.type === 'sites' && <SiteCardList sites={params.items} />}
                            {view === 'map' && params.type === 'sites' && <SiteMap sites={params.items} />}
                            {view === "cards" && params.type === "assets" && <AssetList catalogs={params.items} />}
                            {view === "table" && params.type === "assets" && <GraphTable catalogs={params.items} columns = {params.columns.find((column: ColumnSet) => column.name === selectedColumn)?.data} />}
                        </Tab>
                    ))}
                </Tabs>
            </div>
        </div>
    );
}

export default MultiView;