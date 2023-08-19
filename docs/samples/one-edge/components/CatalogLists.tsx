'use client';

import { Catalog, GroupInfo } from '../types';
import CatalogList from './CatalogList';
import {Accordion, AccordionItem} from "@nextui-org/react";

interface CalalogListsProps {
    groups: GroupInfo[];
}
function CatalogLists(props: CalalogListsProps) {
    const { groups } = props;
    return (
        <div className='sitelist'>
             <Accordion className='cards_view' defaultExpandedKeys={["config"]}>
                {groups.map((group: any) =>  
                    <AccordionItem className='cards_row' key={group.type} aria-label={group.title} title={group.title}>
                        <CatalogList catalogs={group.catalogs} />
                    </AccordionItem>)}
            </Accordion>
        </div>
    );
}
export default CatalogLists;