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
        <Card>
            <CardHeader className="flex gap-3 justify-between">
               {solution.id}     
               <Tabs color="primary" radius="full" selectedKey={activeView} onSelectionChange={updateActiveView} size='sm'>
                    <Tab key="properties" title="properties" />
                    <Tab key="topology" title="topology" />
                    <Tab key="json" title="json" />
               </Tabs>
            </CardHeader>
            <Divider/>
            <CardBody>    
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
            <Divider/>
            <CardFooter  className="flex gap-3 justify-between">                                
            </CardFooter>
        </Card>
    );
}

export default SolutionCard;