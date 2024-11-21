'use client';

import Image from 'next/image';
import Link from 'next/link';
import { MdDashboard } from 'react-icons/md';
import { AiOutlineHome } from 'react-icons/ai';
import { FiMail } from 'react-icons/fi';
import { MdOutlineKeyboardArrowLeft } from 'react-icons/md';
import { useState } from 'react';
import { GrCatalog } from 'react-icons/gr';
import { TbBuildingCommunity } from 'react-icons/tb';
import { HiOutlineTemplate } from 'react-icons/hi';
import { IoLogoAppleAr }   from 'react-icons/io5';
import { FiServer } from 'react-icons/fi';
import { GoWorkflow } from 'react-icons/go';
import { PiBrainLight } from 'react-icons/pi';
import { PiGraph } from 'react-icons/pi';
import { FiPackage } from 'react-icons/fi';
import { PiPathBold } from 'react-icons/pi';
import { HiOutlineCamera } from 'react-icons/hi';
import { GoCopilot } from 'react-icons/go';
function SideBar() {
    const sidebarItems = [
        {
            name: "Home",
            href: "/coming-soon",
            icon: AiOutlineHome,
        },
        {
            name: "Sites",
            href: "/assets?graph=site",
            icon: TbBuildingCommunity,
        },
        {
            name: "Symphony Control Planes",
            href: "/sites",
            icon: TbBuildingCommunity,
        },
        {
            name: "Catalogs",
            href: "/catalogs",
            icon: GrCatalog,
        },
        {
            name: "Solutions",
            href: "/solutions",
            icon: HiOutlineTemplate,
        },
        {
            name: "Instances",
            href: "/instances",
            icon: IoLogoAppleAr,
        },
        {
            name: "Targets",
            href: "/targets",
            icon: FiServer,
        },
        {
            name: "Devices",
            href: "/coming-soon",
            icon: HiOutlineCamera,
        },
        {
            name: "Campaigns",
            href: "/campaigns",
            icon: GoWorkflow,
        },
        {
            name: "AI Models",
            href: "/coming-soon",
            icon: PiBrainLight,
        },
        {
            name: "AI Skills",
            href: "/coming-soon",
            icon: PiGraph,
        },
        {
            name: "AI Skill Packages",
            href: "/coming-soon",
            icon: FiPackage,
        },
        {
            name: "Dashboard",
            href: "/coming-soon",
            icon: MdDashboard,
        },
        {
            name: "Trails",
            href: "/coming-soon",
            icon: PiPathBold,
        },
        {
            name: "Copilot",
            href: "/copilot",
            icon: GoCopilot,
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