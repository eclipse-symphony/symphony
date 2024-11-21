/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

'use client';

import './globals.css';
import Header from '@/components/Header';
import SideBar from '@/components/SideBar';
import AuthProvider from './context/AuthProvider';
import React from 'react';
import { GlobalStateProvider } from '@/components/GlobalStateProvider';

export default function RootLayout({
  children,
}: {
  children: React.ReactNode
}) {
  return (
    <html lang="en">      
      <body>
        <AuthProvider>
          <GlobalStateProvider>
            <Header/>        
            <div className='layout'>
              <SideBar />
              <div className='main_content'>
                {children}
              </div>
            </div>        
          </GlobalStateProvider>
        </AuthProvider>
      </body>
    </html>
  )
}
