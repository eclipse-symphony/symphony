/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

'use client';

import {Card, CardHeader, CardBody, CardFooter, Divider} from '@nextui-org/react';
import {InstanceState} from '../../app/types';
import {useState} from 'react';
import ReactFlow from 'reactflow';
import 'reactflow/dist/style.css';
import {LuFileJson2} from 'react-icons/lu';
import InstanceSpecCard from "../InstanceSpecCard";
import {Tabs, Tab} from "@nextui-org/react";
import { FcOk } from "react-icons/fc";
import { FcHighPriority } from "react-icons/fc";

interface InstanceCardProps {
    instance: InstanceState;    
}

const getSolutionTopology = async() => {
    return  { nodes: [
             { id: '1', position: { x: 0, y: 0 }, data: { label: '1' } },
             { id: '2', position: { x: 0, y: 100 }, data: { label: '2' } },
           ],
        edges: [{ id: 'e1-2', source: '1', target: '2' }]
    };
}

function InstanceCard(props: InstanceCardProps) {

    const { instance } = props;
    const [activeView, setActiveView] = useState('properties');
    const [nodes, setNodes] = useState<any[]>([]);
    const [edges, setEdges] = useState<any[]>([]);

    // const initialNodes = [
    //     { id: '1', position: { x: 0, y: 0 }, data: { label: '1' } },
    //     { id: '2', position: { x: 0, y: 100 }, data: { label: '2' } },
    //   ];
    // const initialEdges = [{ id: 'e1-2', source: '1', target: '2' }];
    
    // get json from campaign with new lines
    const json = JSON.stringify(instance, null, 2);    

    const updateActiveView = (key: any) => {
        if (key.toString() == 'topology') {
            getSolutionTopology().then((topology) => {                
                setNodes(topology.nodes);
                setEdges(topology.edges);
            });
        }
        setActiveView(key.toString());
    }

    const [isHovered, setIsHovered] = useState(false);

    return (
        <Card radius='none' shadow='lg' className='card'
            onMouseEnter={() => setIsHovered(true)}
            onMouseLeave={() => setIsHovered(false)}>
            <CardHeader className="flex gap-3 justify-between">
               <div className="card_title">{instance.metadata.name}</div>
               {isHovered && (
               <Tabs color="primary" radius="full" selectedKey={activeView} onSelectionChange={updateActiveView} size='sm'>
                    <Tab key="properties" title="properties" />
                    <Tab key="topology" title="topology" />
                    <Tab key="json" title="json" />
               </Tabs>)}
            </CardHeader>
            <Divider/>
            <CardBody>    
                {activeView == 'properties' && (
                    <InstanceSpecCard instance={instance.spec} />
                )}
                {activeView == 'topology' && (                    
                    <div style={{ width: '600px', height: '400px' }}>
                    <ReactFlow nodes={nodes} edges={edges} />
                  </div>
                )}
                {activeView == 'json' && (
                    <div className="w-[600px] h-[400px]"><pre>{json}</pre></div>                
                )}                
            </CardBody>
            <Divider/>
            <CardFooter  className="flex gap-3 justify-between">    
                <div className="flex gap-2">
                    {instance.status.properties && instance.status.properties.status === 'Succeeded' && (
                        <span className="flex gap-2">
                        <FcOk className="text-[#AAAAF9] text-xl"/> OK
                        </span>
                    )}    
                    {instance.status.properties && instance.status.properties.status === 'Reconciling' && (
                        <span className="flex gap-2">
                        <FcOk className="text-[#AAAAF9] text-xl"/> Reconciling
                        </span>
                    )}    
                    {instance.status.properties && instance.status.properties.status != 'Succeeded' && instance.status.properties.status != 'Reconciling' && (
                        <span className="flex gap-2">
                        <FcHighPriority className="text-[#AAAAF9] text-xl"/> {instance.status.properties.status}
                        </span>
                    )}                          
                 </div>
            </CardFooter>
        </Card>
    );
}

export default InstanceCard;