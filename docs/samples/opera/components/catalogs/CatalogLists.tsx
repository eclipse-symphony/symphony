'use client';

import { CatalogVersionState, GroupInfo } from '../../app/types';
import CatalogVersionList from './CatalogVersionList';
import {Accordion, AccordionItem} from "@nextui-org/react";

interface CalalogListsProps {
    groups: GroupInfo[];
}
function CatalogVersionLists(props: CalalogListsProps) {
    const { groups } = props;
    return (
        <div className='sitelist'>
             <Accordion className='cards_view' defaultExpandedKeys={["config"]}>
                {groups.map((group: any) =>  
                    <AccordionItem className='cards_row' key={group.type} aria-label={group.title} title={group.title}>
                        <CatalogVersionList catalogversions={group.catalogversions} />
                    </AccordionItem>)}
            </Accordion>
        </div>
    );
}
export default CatalogVersionLists;