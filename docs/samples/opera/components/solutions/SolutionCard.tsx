/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 * SPDX-License-Identifier: MIT
 */

'use client';

import {Card, CardHeader, CardBody, CardFooter, Divider} from '@nextui-org/react';
import {SolutionState} from '../../app/types';
import {useState} from 'react';
import ReactFlow from 'reactflow';
import 'reactflow/dist/style.css';
import {LuFileJson2} from 'react-icons/lu';
import SolutionSpecCard from "../SolutionSpecCard";
import {Tabs, Tab} from "@nextui-org/react";

interface SolutionCardProps {
    solution: SolutionState;    
}

const getSolutionTopology = async() => {
    return  { nodes: [
             { id: '1', position: { x: 0, y: 0 }, data: { label: '1' } },
             { id: '2', position: { x: 0, y: 100 }, data: { label: '2' } },
           ],
        edges: [{ id: 'e1-2', source: '1', target: '2' }]
    };
}

function SolutionCard(props: SolutionCardProps) {

    const { solution } = props;
    const [activeView, setActiveView] = useState('properties');
    const [nodes, setNodes] = useState<any[]>([]);
    const [edges, setEdges] = useState<any[]>([]);

    const [isHovered, setIsHovered] = useState(false);

    // const initialNodes = [
    //     { id: '1', position: { x: 0, y: 0 }, data: { label: '1' } },
    //     { id: '2', position: { x: 0, y: 100 }, data: { label: '2' } },
    //   ];
    // const initialEdges = [{ id: 'e1-2', source: '1', target: '2' }];
    
    // get json from campaign with new lines
    const json = JSON.stringify(solution, null, 2);    

    const updateActiveView = (key: any) => {
        if (key.toString() == 'topology') {
            getSolutionTopology().then((topology) => {                
                setNodes(topology.nodes);
                setEdges(topology.edges);
            });
        }
        setActiveView(key.toString());
    }

    return (
        <Card radius='none' shadow='lg' className='card'
            onMouseEnter={() => setIsHovered(true)}
            onMouseLeave={() => setIsHovered(false)}>
            <CardHeader className="absolute z-10 top-0 flex-col !items-start bg-black/10">
                <div className="card_title">
                    {solution.metadata.name.replace(`${solution.metadata.labels.rootResource}-v-`, `${solution.metadata.labels.rootResource}: `)}
                </div>
               {isHovered && (
                <Tabs color="primary" radius="full" selectedKey={activeView} onSelectionChange={updateActiveView} size='sm'  className='absolute right-5 top-5'>
                        <Tab key="properties" title="properties" />
                        <Tab key="topology" title="topology" />
                        <Tab key="json" title="json" />
                </Tabs>
               )}
            </CardHeader>
            <CardBody className='absolute top-[80px] h-full bg-white/70'>    
                {activeView == 'properties' && (
                    <SolutionSpecCard solution={solution.spec} />
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
            {/* <CardFooter  className="absolute bg-black/30 bottom-0 border-t-1 border-zinc-100/50 z-10 justify-between">                                
            </CardFooter> */}
        </Card>
    );
}

export default SolutionCard;