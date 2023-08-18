import './globals.css';
import type { Metadata } from 'next';
import Header from '@/components/Header';
import SideBar from '@/components/SideBar';
export const metadata: Metadata = {
  title: 'One Edge Universe Portal',
  description: 'A PoC of a unified portal experieence for One Edge',
}

export default function RootLayout({
  children,
}: {
  children: React.ReactNode
}) {
  return (
    <html lang="en">      
      <body>
        <Header/>        
        <div className='layout'>
          <SideBar/>
          {children}
        </div>        
      </body>
    </html>
  )
}
