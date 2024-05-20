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
import {Table, TableHeader, TableColumn, TableBody} from "@nextui-org/react";
import TableContainer from '@mui/material/TableContainer';

interface GraphTableProps {
    catalogs: CatalogState[];
    columns: any[] | undefined;
}

function BuildForest(catalogs: Record<string, CatalogState[]>) {
    const forest: any = [];
    console.log(catalogs);
    for (const [_, cats] of Object.entries(catalogs)) {
        cats.forEach((catalog: CatalogState) => {
            if (catalog.spec.parentName==="" || catalog.spec.parentName===undefined) {
                forest.push(BuildTree(cats, catalog));
            }
        });    
    }

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
    const { catalogs, columns } = props;
    const treeViewRef = useRef<HTMLDivElement>(null);
    
    const updateVisibleNodes = () => {
        const visibleNodes: string[] = [];
        const treeview = document.querySelector('#tree');
        treeview?.querySelectorAll('li').forEach((node) => {
            const nodeId = node.getAttribute('id');
            visibleNodes.push(nodeId ?? '');
        });
        setVisibleNodes(visibleNodes);
    }

    const mergedCatalogs: CatalogState[] = [];
   
    const mergedColumns: any[] = [];
    if (columns) {
        for (const [_, cols] of Object.entries(columns)) {
            mergedColumns.push(cols);
        }
    }
    //for (const [_, cats] of Object.entries(catalogs)) {
    //    mergedCatalogs.push(...cats);
    //}
    mergedCatalogs.push(...catalogs);

    useEffect(() => {
        const observer = new MutationObserver(() => {
          updateVisibleNodes();
        });
        const treeview = document.querySelector('#tree');
        if (treeview) {
            observer.observe(treeview, { childList: true, subtree: true });
        }
      }, [mergedCatalogs, mergedColumns]);

    useEffect(() => {
        const tableBody = document.querySelector('#tree-table tbody');
        Array.from(tableBody?.childNodes ?? []).forEach((node) => {
            tableBody?.removeChild(node);
        });
        visibleNodes.forEach((nodeId: string) => {
            const catalog = mergedCatalogs.find((catalog: CatalogState) => "tree-" + catalog.spec.name === nodeId);
            if (catalog) {        
                const row = document.createElement('tr');
                mergedColumns?.forEach((column: any) => {
                    const cell = document.createElement('td');
                    cell.innerText = "...";
                    column.forEach((col: any) => {
                        if (col.spec.metadata?.asset == catalog.spec.name) {
                            cell.innerText = col.spec.name;                            
                        }                        
                    });
                    row.appendChild(cell);
                });
                tableBody?.appendChild(row);
            }
        });
      }, [visibleNodes, mergedCatalogs, mergedColumns]);
    useEffect(() => {
        updateVisibleNodes();
    }, []);    
    const handleToggle = (event: React.SyntheticEvent, nodeIds: string[]) => {
       updateVisibleNodes();
    };

    const treeNodes = BuildForest({"FIX-ME": catalogs});
    if (mergedColumns?.length) {
    return (
        <div className='graph_container'>
            <div className='tree_container'>
            <TreeView defaultCollapseIcon={<AiOutlineMinusSquare/>} defaultExpandIcon={<AiOutlinePlusSquare />}
                ref={treeViewRef}
                onNodeToggle={handleToggle} id="tree">
                {treeNodes}
            </TreeView>   
            </div> 
            <TableContainer>
                <Table aria-label="sticky table" removeWrapper id="tree-table">
                    <TableHeader>
                        {Array.isArray(mergedColumns) ? (
                            mergedColumns.map((column) => (                                
                                <TableColumn>
                                    {column[0].spec.name}
                                </TableColumn>
                            ))                            
                        ):(
                            <TableColumn>ABC</TableColumn>
                        )}
                    </TableHeader>
                    <TableBody>
                        <span></span>
                    </TableBody>
                </Table>
            </TableContainer>
        </div>);
    } else {
        return null;
    }
}
export default GraphTable;