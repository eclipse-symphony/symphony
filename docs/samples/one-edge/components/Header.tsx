'use client';

import React, { useState } from 'react';
import Link from 'next/link';
import Image from 'next/image';
import { BsList, BsX, BsYoutube, BsPinterest} from 'react-icons/bs';

const styles={
    navLink: 'ml-10 uppercase border-b border-white hover:border-[#F6B519] text-xl'
}

function Header() {
    const [isMenuOpen, setIsMenuOpen] = useState(false);
    const toggleMenu = () => setIsMenuOpen(!isMenuOpen);
    return (
        <header>
            <nav className='w-full h-24 hadow-xl bg-black'>
                {/* View Menu */}
                <div className='flex items-center justify-between h-full px-4 w-full'>
                    <Link href='/'>
                        <Image src='https://static.boredpanda.com/blog/wp-content/uploads/2018/04/5acb63d83493f__700-png.jpg' alt='One Edge Logo' width={135} height={55} className='cursor-pointer'/>
                    </Link>
                    <div className='text-white hidden sm:flex'>
                        <ul className='hidden sm:flex'>
                            <li className={styles.navLink}>
                                <Link href='/about'>About</Link>
                            </li>
                            <li className={styles.navLink}>
                                <Link href='/contact'>Contact</Link>
                            </li>
                            <li className='flex items-center space-x-5 text-[#F6B519] ml-10'>
                                <h3 className='cursor-pointer border border-[#F6B519] px-4 py-1 rounded-full bg-[#F6B519] text-black hover:bg-black hover:text-[#F6B519] ease-in-out duration-300'>Sign In</h3>
                            </li>
                        </ul>
                    </div>
                    {/* Mobile Menu */}
                    <div onClick={() => toggleMenu()} className='sm:hidden cursor-pointer pl-24'>
                        <BsList className='h-8 w-8 text-[#F6B519]' />
                    </div>
                </div>
                <div className={isMenuOpen ? 'fixed z-40 top-0 left-0 w-[75%] sm:hidden h-screen bg-[#ecf0f3] p-10 ease-in-out duration-500': 'fixed left-[-100%] top-10 p-10 ease-in-out duration-500'}>
                    <div className='flex w-full items-center justify-end'>
                        <div className='cursor-pointer'>
                            <BsX className='h-8 w-8 text-[#F6B519]' />
                        </div>
                    </div>
                    {/* Mobile Menu Links */}
                    <div className='flex-col py-4'>
                        <ul>
                            <li onClick={() => setIsMenuOpen(false)} className='py-4 hover:underline hover:decoration-[#F6B519]'>
                                <Link href='/about'>About</Link>
                            </li>
                            <li onClick={() => setIsMenuOpen(false)} className='py-4 hover:underline hover:decoration-[#F6B519]'>
                                <Link href='/contact'>Contact</Link>
                            </li>
                            <li className='flex items-center py-4 text-[#F6B519]'>
                                <p className='cursor-pointter px-4 py-1 rounded-full bg-[#F6B519] text-black hover:bg-black hover:text-[#F6B519] eas-in-out duration-300'>
                                    Sign In
                                </p>
                            </li>
                        </ul>
                    </div>
                    {/* Links */}
                    <div className='flex flex-row justify-around pt-10 items-center'>
                        <Link href='https://www.facebook.com/'>
                            <BsYoutube size={30} className='cursor-pointer hover:text-[#F6B519] eas-in-out duration-3000' />
                        </Link>
                        <Link href='https://www.instagram.com/'>
                            <BsPinterest size={30} className='cursor-pointer hover:text-[#F6B519] eas-in-out duration-3000' />
                        </Link>
                    </div>
                    <Image src='https://static.boredpanda.com/blog/wp-content/uploads/2018/04/5acb63d83493f__700-png.jpg' alt='One Edge Logo' width={135} height={55} className='cursor-pointer pt-10 mx-auto'/>             
                </div>                   
            </nav>
        </header>
    );
}

export default Header;