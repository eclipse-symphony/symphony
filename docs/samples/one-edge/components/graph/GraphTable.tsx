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
interface GraphTableProps {
    catalogs: CatalogState[];
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
    return <TreeItem nodeId={catalog.spec.name} label={<div className='treenode'>
            {(catalog.spec.parentName === '' || catalog.spec.parentName === undefined) && <TbBuildingCommunity />}
            {catalog.spec.objectRef?.kind === 'arc' && <SiKubernetes />}
            {catalog.spec.objectRef?.kind === 'adr' && <FaDatabase />}
            {catalog.spec.objectRef?.kind === 'iot-hub' && <MdHub />}
            {catalog.spec.objectRef?.kind === 'site' && <FaSitemap />}
            {catalog.spec.properties.name}
        </div>}>{children}</TreeItem>    
}
function GraphTable(props: GraphTableProps) {
    const { catalogs } = props;

    return (
        <TreeView defaultCollapseIcon={<AiOutlineMinusSquare/>} defaultExpandIcon={<AiOutlinePlusSquare />}>
            {BuildForest(catalogs)}
        </TreeView>        
    );
}
export default GraphTable;