"use client";
import '../styles/globals.css';
import Header from "./header";
import { ReactNode } from 'react';
import { SessionProvider } from 'next-auth/react';
import NavBar from './NavBar';

export default function RootLayout({ children }: { children: React.ReactNode })  {
  return (
    <html>
      <head>
        <title>Microsoft Edge 365 Universe PoC</title>
      </head>
      <body>
        <SessionProvider>
          <Header />
          <div className="flex">
            <NavBar />
            <div className="h-screen flex-1 mt-20">{children}</div>
          </div>
        </SessionProvider>
      </body>
    </html>
  )
}
