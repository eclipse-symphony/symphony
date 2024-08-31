import React, { useCallback, useEffect } from 'react';
import {
  ReactFlow,
  MiniMap,
  Controls,
  Background,
  useNodesState,
  useEdgesState,
  addEdge,
} from '@xyflow/react';
 
import '@xyflow/react/dist/style.css';
import { useGlobalState } from '../GlobalStateProvider'; // Import the global state hook

// const initialNodes = [
//   { id: '1', position: { x: 0, y: 0 }, data: { label: '1' } },
//   { id: '2', position: { x: 0, y: 100 }, data: { label: '2' } },
// ];
// const initialEdges = [{ id: 'e1-2', source: '1', target: '2' }];
 
export default function SystemDiagram() {
    const { objects } = useGlobalState(); // Access objects from global state

   
    const [nodes, setNodes, onNodesChange] = useNodesState([]);
    const [edges, setEdges, onEdgesChange] = useEdgesState([]);
 
    useEffect(() => {
         // Separate objects into their respective types
        const solutions = objects.filter((obj) => obj.type === 'solution');
        const instances = objects.filter((obj) => obj.type === 'instance');
        const targets = objects.filter((obj) => obj.type === 'target');
        
        const newNodes = [
            ...solutions.map((sol, index) => ({
              id: `solution-${index}`,
              position: { x: index * 200, y: 0 }, // Position solutions at the top layer
              data: { label: sol.name },
            })),
            ...instances.map((inst, index) => ({
              id: `instance-${index}`,
              position: { x: index * 200, y: 150 }, // Position instances in the middle layer
              data: { label: inst.name },
            })),
            ...targets.map((tgt, index) => ({
              id: `target-${index}`,
              position: { x: index * 200, y: 300 }, // Position targets at the bottom layer
              data: { label: tgt.name },
            })),
          ];
        if (newNodes.length === 0) {
            return;
        }

            // Generate edges for each instance to its corresponding solution
    const solutionEdges = instances.map((inst, index) => ({
        id: `edge-solution-instance-${index}`,
        source: `solution-${index}`, // Connect instance to corresponding solution
        target: `instance-${index}`,
      }));
  
      // Generate edges for each instance to all targets
      const targetEdges = instances.flatMap((inst, instanceIndex) =>
        targets.map((tgt, targetIndex) => ({
          id: `edge-instance-target-${instanceIndex}-${targetIndex}`,
          source: `instance-${instanceIndex}`, // Connect each target to each instance
          target: `target-${targetIndex}`,
        }))
      );
  
      const newEdges = [...solutionEdges, ...targetEdges];

        setNodes(newNodes);
        setEdges(newEdges);
    }, [objects]);
    // const onConnect = useCallback(
    //     (params) => setEdges((eds) => addEdge(params, eds)),
    //     [setEdges],
    // );
 
  return (
    <div style={{ width: '100%', height: '100%' }}>
      <ReactFlow
        nodes={nodes}
        edges={edges}
        onNodesChange={onNodesChange}
        onEdgesChange={onEdgesChange}
        // onConnect={onConnect}
      >
        <Controls />
        <MiniMap />
        <Background variant="dots" gap={12} size={1} />
      </ReactFlow>
    </div>
  );
}