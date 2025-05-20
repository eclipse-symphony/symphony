/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

'use client';

import {Card, CardHeader, CardBody, CardFooter, Divider} from '@nextui-org/react';
import {TargetState} from '../../app/types';
import {useState} from 'react';
import ReactFlow from 'reactflow';
import 'reactflow/dist/style.css';
import {LuFileJson2} from 'react-icons/lu';
import TargetSpecCard from "../TargetSpecCard";
import {Tabs, Tab} from "@nextui-org/react";
import { FcOk } from "react-icons/fc";
import { FcHighPriority } from "react-icons/fc";
import CoverImage from '../CoverImage';

interface TargetCardProps {
    target: TargetState;    
}
function TargetCard(props: TargetCardProps) {

    const { target } = props;
    const [activeView, setActiveView] = useState('properties');
    const [nodes, setNodes] = useState([]);
    const [edges, setEdges] = useState([]);

    const json = JSON.stringify(target, null, 2);    

    const updateActiveView = (key: any) => {
        setActiveView(key.toString());
    }

    const [isHovered, setIsHovered] = useState(false);

    return (
        <Card radius='none' shadow='lg' className='card'
            onMouseEnter={() => setIsHovered(true)}
            onMouseLeave={() => setIsHovered(false)}>
            <CardHeader className="absolute z-10 top-0 flex-col !items-start bg-black/10">                
               <div className="card_title">{target.metadata.name}</div>
               {isHovered && (
                <Tabs color="secondary" radius="full" selectedKey={activeView} onSelectionChange={updateActiveView} size='sm' className='absolute right-5 top-5'>
                        <Tab key="properties" title="properties" />
                        <Tab key="json" title="json" />
                </Tabs>)}
            </CardHeader>
            {target.spec.properties?.['image.url'] && (
                <CoverImage src={target.spec.properties['image.url']} />
            )}
            <CardBody className='absolute top-[80px] h-full bg-white/70'>    
                {activeView == 'properties' && (
                    <TargetSpecCard target={target.spec} />
                )}
                {activeView == 'json' && (
                    <div className="relative w-[600px] h-[400px]"><pre>{json}</pre></div>                
                )}                
            </CardBody>
            <CardFooter  className="absolute bg-black/30 bottom-0 border-t-1 border-zinc-100/50 z-10 justify-between">    
                <div className="flex gap-2">
                    {target.status.properties && target.status.properties.status === 'Succeeded' && (
                        <span className="flex gap-2">
                        <FcOk className="text-[#AAAAF9] text-xl"/> OK
                        </span>
                    )}    
                    {target.status.properties && target.status.properties.status === 'Reconciling' && (
                        <span className="flex gap-2">
                        <FcOk className="text-[#AAAAF9] text-xl"/> Reconciling
                        </span>
                    )}    
                    {target.status.properties && target.status.properties.status != 'Succeeded' && target.status.properties.status != 'Reconciling' && (
                        <span className="flex gap-2">
                        <FcHighPriority className="text-[#AAAAF9] text-xl"/> {target.status.properties.status}
                        </span>
                    )}                          
                 </div>
            </CardFooter>
        </Card>
    );
}

export default TargetCard;