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

    return (
        <Card>
            <CardHeader className="flex gap-3 justify-between">
               {target.metadata.name}     
               <Tabs color="primary" radius="full" selectedKey={activeView} onSelectionChange={updateActiveView} size='sm'>
                    <Tab key="properties" title="properties" />
                    <Tab key="json" title="json" />
               </Tabs>
            </CardHeader>
            <Divider/>
            <CardBody>    
                {activeView == 'properties' && (
                    <TargetSpecCard target={target.spec} />
                )}
                {activeView == 'json' && (
                    <div className="w-[600px] h-[400px]"><pre>{json}</pre></div>                
                )}                
            </CardBody>
            <Divider/>
            <CardFooter  className="flex gap-3 justify-between">    
                <div className="flex gap-2">
                    {target.status.properties && target.status.properties.status === 'Succeeded' && (
                        <span className="flex gap-2">
                        <FcOk className="text-[#AAAAF9] text-xl"/> OK
                        </span>
                    )}    
                    {target.status.properties && target.status.properties.status != 'Succeeded' && (
                        <span className="flex gap-2">
                        <FcHighPriority className="text-[#AAAAF9] text-xl"/> Failed
                        </span>
                    )}                          
                 </div>
            </CardFooter>
        </Card>
    );
}

export default TargetCard;