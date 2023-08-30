'use client';
import TreeView from '@mui/lab/TreeView'
import TreeItem from '@mui/lab/TreeItem';
import { CatalogState } from '@/app/types';
import { AiOutlineMinusSquare, AiOutlinePlusSquare } from 'react-icons/ai';
import { SiKubernetes } from 'react-icons/si';
import { TbBuildingCommunity } from 'react-icons/tb';
import { FaDatabase } from 'react-icons/fa';
import { MdHub } from 'react-icons/md';
import { FaSitemap } from 'react-icons/fa';
import React, { useState, useEffect, useRef } from 'react';
import {Table, TableHeader, TableColumn, TableBody, TableRow, TableCell, Chip} from "@nextui-org/react";

interface GraphTableProps {
    catalogs: CatalogState[];
    columns: any[];
}

function BuildForest(catalogs: CatalogState[]) {
    const forest: any = [];
    catalogs.forEach((catalog: CatalogState) => {
        if (catalog.spec.parentName==="" || catalog.spec.parentName===undefined) {
            forest.push(BuildTree(catalogs, catalog));
        }
    });
    return forest;
}
function BuildTree(catalogs: CatalogState[], catalog: CatalogState) {
    const nodes: any = [];
    catalogs.forEach((cat: CatalogState) => {
        if (cat.spec.parentName===catalog.spec.name) {
            nodes.push(BuildTree(catalogs, cat));
        }
    });        
    return BuildTreeNode(catalog, nodes);  
}
function BuildTreeNode(catalog: CatalogState, children: any) {
    return <TreeItem nodeId={catalog.spec.name} label={BuildTreeNodeLabel(catalog)}>{children}</TreeItem>    
}
function BuildTreeNodeLabel(catalog: CatalogState) {
    return <div className='treenode'>
        {(catalog.spec.parentName === '' || catalog.spec.parentName === undefined) && <TbBuildingCommunity />}
        {catalog.spec.objectRef?.kind === 'arc' && <SiKubernetes />}
        {catalog.spec.objectRef?.kind === 'adr' && <FaDatabase />}
        {catalog.spec.objectRef?.kind === 'iot-hub' && <MdHub />}
        {catalog.spec.objectRef?.kind === 'site' && <FaSitemap />}
        {catalog.spec.properties.name}
    </div>
}

function GraphTable(props: GraphTableProps) {
    const [visibleNodes, setVisibleNodes] = useState<string[]>([]);
    const { catalogs } = props;
    const treeViewRef = useRef<typeof TreeView>(null);
    const columns = [
        { field: 'name', headerName: 'A', width: 200 },
        { field: 'name', headerName: 'B', width: 200 },
        { field: 'name', headerName: 'C', width: 200 },
    ]    
    
    const updateVisibleNodes = () => {
        const visibleNodes: string[] = [];
        const treeview = document.querySelector('#tree');
        treeview?.querySelectorAll('li').forEach((node) => {
            const nodeId = node.getAttribute('id');
            visibleNodes.push(nodeId ?? '');
        });
        setVisibleNodes(visibleNodes);
    }

    useEffect(() => {
        const observer = new MutationObserver(() => {
          updateVisibleNodes();
        });
        const treeview = document.querySelector('#tree');
        observer.observe(treeview, { childList: true, subtree: true });
      }, []);

    useEffect(() => {
        const tableBody = document.querySelector('#tree-table tbody');
        Array.from(tableBody?.childNodes ?? []).forEach((node) => {
            tableBody?.removeChild(node);
        });
        visibleNodes.forEach((nodeId: string) => {
            const catalog = catalogs.find((catalog: CatalogState) => "tree-" + catalog.spec.name === nodeId);
            if (catalog) {        
                const row = document.createElement('tr');
                const cell1 = document.createElement('td');
                const cell2 = document.createElement('td');
                cell1.innerText = catalog.spec.properties.name;
                cell2.innerText = catalog.spec.properties.description;
                row.appendChild(cell1);
                row.appendChild(cell2);
                tableBody?.appendChild(row);
            }
        });
      }, [visibleNodes, catalogs]);
    useEffect(() => {
        updateVisibleNodes();
    }, []);
    const findNodeById = (node: any, id: string): any => {
        if (node.type === TreeItem && node.props.nodeId === id) {
          return node;
        }
        if (node.props && node.props.children) {
          for (let i = 0; i < node.props.children.length; i++) {
            const child = node.props.children[i];
            const result = findNodeById(child, id);
            if (result) {
              return result;
            }
          }
        }
        return null;
      };
    const handleToggle = (event: React.SyntheticEvent, nodeIds: string[]) => {
       updateVisibleNodes();
    };

    const treeNodes = BuildForest(catalogs);
    return (
        <div className='graph_container'>
            <div className='tree_container'>
            <TreeView defaultCollapseIcon={<AiOutlineMinusSquare/>} defaultExpandIcon={<AiOutlinePlusSquare />}
                ref={treeViewRef}
                onNodeToggle={handleToggle} id="tree">
                {treeNodes}
            </TreeView>   
            </div> 
            <Table removeWrapper id="tree-table">
                <TableHeader>
                    <TableColumn>PROPERTY</TableColumn>
                    <TableColumn>VALUE</TableColumn>
                </TableHeader>
                <TableBody>
                </TableBody>
            </Table>
        </div>    
    );
}
export default GraphTable;