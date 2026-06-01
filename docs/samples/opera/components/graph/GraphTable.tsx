'use client';
import TreeView from '@mui/lab/TreeView'
import TreeItem from '@mui/lab/TreeItem';
import { CatalogVersionState } from '@/app/types';
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
    catalogversions: CatalogVersionState[];
    columns: any[] | undefined;
}

function BuildForest(catalogversions: Record<string, CatalogVersionState[]>) {
    const forest: any = [];
    console.log(catalogversions);
    for (const [_, cats] of Object.entries(catalogversions)) {
        cats.forEach((catalogversion: CatalogVersionState) => {
            if (catalogversion.spec.parentName==="" || catalogversion.spec.parentName===undefined) {
                forest.push(BuildTree(cats, catalogversion));
            }
        });    
    }

    return forest;
}
function BuildTree(catalogversions: CatalogVersionState[], catalogversion: CatalogVersionState) {
    const nodes: any = [];
    catalogversions.forEach((cat: CatalogVersionState) => {
        if (cat.spec.parentName===catalogversion.spec.name) {
            nodes.push(BuildTree(catalogversions, cat));
        }
    });        
    return BuildTreeNode(catalogversion, nodes);  
}
function BuildTreeNode(catalogversion: CatalogVersionState, children: any) {
    return <TreeItem nodeId={catalogversion.spec.name} label={BuildTreeNodeLabel(catalogversion)}>{children}</TreeItem>    
}
function BuildTreeNodeLabel(catalogversion: CatalogVersionState) {
    return <div className='treenode'>
        {(catalogversion.spec.parentName === '' || catalogversion.spec.parentName === undefined) && <TbBuildingCommunity />}
        {catalogversion.spec.objectRef?.kind === 'arc' && <SiKubernetes />}
        {catalogversion.spec.objectRef?.kind === 'adr' && <FaDatabase />}
        {catalogversion.spec.objectRef?.kind === 'iot-hub' && <MdHub />}
        {catalogversion.spec.objectRef?.kind === 'site' && <FaSitemap />}
        {catalogversion.spec.properties.name}
    </div>
}

function GraphTable(props: GraphTableProps) {
    const [visibleNodes, setVisibleNodes] = useState<string[]>([]);
    const { catalogversions, columns } = props;
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

    const mergedCatalogVersions: CatalogVersionState[] = [];
   
    const mergedColumns: any[] = [];
    if (columns) {
        for (const [_, cols] of Object.entries(columns)) {
            mergedColumns.push(cols);
        }
    }
    //for (const [_, cats] of Object.entries(catalogversions)) {
    //    mergedCatalogVersions.push(...cats);
    //}
    mergedCatalogVersions.push(...catalogversions);

    useEffect(() => {
        const observer = new MutationObserver(() => {
          updateVisibleNodes();
        });
        const treeview = document.querySelector('#tree');
        if (treeview) {
            observer.observe(treeview, { childList: true, subtree: true });
        }
      }, [mergedCatalogVersions, mergedColumns]);

    useEffect(() => {
        const tableBody = document.querySelector('#tree-table tbody');
        Array.from(tableBody?.childNodes ?? []).forEach((node) => {
            tableBody?.removeChild(node);
        });
        visibleNodes.forEach((nodeId: string) => {
            const catalogversion = mergedCatalogVersions.find((catalogversion: CatalogVersionState) => "tree-" + catalogversion.spec.name === nodeId);
            if (catalogversion) {        
                const row = document.createElement('tr');
                mergedColumns?.forEach((column: any) => {
                    const cell = document.createElement('td');
                    cell.innerText = "...";
                    column.forEach((col: any) => {
                        if (col.spec.metadata?.asset == catalogversion.spec.name) {
                            cell.innerText = col.spec.name;                            
                        }                        
                    });
                    row.appendChild(cell);
                });
                tableBody?.appendChild(row);
            }
        });
      }, [visibleNodes, mergedCatalogVersions, mergedColumns]);
    useEffect(() => {
        updateVisibleNodes();
    }, []);    
    const handleToggle = (event: React.SyntheticEvent, nodeIds: string[]) => {
       updateVisibleNodes();
    };

    const treeNodes = BuildForest({"FIX-ME": catalogversions});
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