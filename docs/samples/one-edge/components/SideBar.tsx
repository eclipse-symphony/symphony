'use client';

import Image from 'next/image';
import Link from 'next/link';
import { MdDashboard } from 'react-icons/md';
import { AiOutlineHome } from 'react-icons/ai';
import { FiMail } from 'react-icons/fi';
import { MdOutlineKeyboardArrowLeft } from 'react-icons/md';
import { useState } from 'react';
import { HiOfficeBuilding} from 'react-icons/hi';
import {GrCatalog} from 'react-icons/gr';
function SideBar() {
    const sidebarItems = [
        {
            name: "Home",
            href: "/coming-soon",
            icon: AiOutlineHome,
        },
        {
            name: "Sites",
            href: "/sites",
            icon: HiOfficeBuilding,
        },
        {
            name: "Catalogs",
            href: "/catalogs",
            icon: GrCatalog,
        },
        {
            name: "Solutions",
            href: "/coming-soon",
            icon: FiMail,
        },
        {
            name: "Instances",
            href: "/coming-soon",
            icon: FiMail,
        },
        {
            name: "Targets",
            href: "/coming-soon",
            icon: FiMail,
        },
        {
            name: "AI Models",
            href: "/coming-soon",
            icon: FiMail,
        },
        {
            name: "AI Skills",
            href: "/coming-soon",
            icon: FiMail,
        },
        {
            name: "AI Skill Packages",
            href: "/coming-soon",
            icon: FiMail,
        },
        {
            name: "Dashboard",
            href: "/coming-soon",
            icon: MdDashboard,
        }
    ];
    const [isSidebarCollapse, setSidebarCollapse] = useState(false);
    const toggleSidebar = () => setSidebarCollapse(!isSidebarCollapse);
    return (
       <div className='sidebar_wrapper'>
            <button className='sidebar_toggle' onClick={()=> toggleSidebar()}>
                <MdOutlineKeyboardArrowLeft />
            </button>
            <aside className='sidebar' data-collapse={isSidebarCollapse}>                           
                <ul className='sidebar_list'>
                    {sidebarItems.map(({name, href, icon: Icon}) => (
                        <li className='sidebar_item' key={name}>
                            <Link href={href} className='sidebar_link'>
                                <span className='sidebar_icon'><Icon /></span>
                                <span className='sidebar_name'>{name}</span>
                            </Link>
                        </li>
                    ))}                    
                </ul>
            </aside>    
        </div>        
    );
}

export default SideBar;