'use client';

import Image from 'next/image';
import Link from 'next/link';
import { MdDashboard } from 'react-icons/md';
import { AiOutlineHome } from 'react-icons/ai';
import { BsPeople } from 'react-icons/bs';
import { FiMail } from 'react-icons/fi';
import { MdOutlineKeyboardArrowLeft } from 'react-icons/md';
import { useState } from 'react';

function SideBar() {
    const sidebarItems = [
        {
            name: "Sites",
            href: "/sites",
            icon: AiOutlineHome,
        },
        {
            name: "Catalogs",
            href: "/catalogs",
            icon: BsPeople,
        },
        {
            name: "Targets",
            href: "/targets",
            icon: FiMail,
        },
        {
            name: "Dashboard",
            href: "/dashboard",
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
                <div className='sidebar_top'>
                    <Image src='/next.svg' width={80} height={80} className='sidebar_logo' alt='sidebar logo'/>
                    <p className='sidebar_logo_name'>Symphony</p>
                </div>                
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