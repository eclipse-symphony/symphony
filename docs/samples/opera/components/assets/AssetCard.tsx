'use client';

import {Card, CardHeader, CardBody, CardFooter, Divider, Link, Image} from '@nextui-org/react';
import { CatalogState } from '../../app/types';
import PropertyTable from '../PropertyTable';
import SolutionSpecCard from '../SolutionSpecCard';
import { FcSettings, FcTemplate } from 'react-icons/fc';
import { FaGithub } from 'react-icons/fa';

interface AssetCardProps {
    catalog: CatalogState;
    refCatalog?: CatalogState | null | undefined;
}
function AssetCard(props: AssetCardProps) {
    const { catalog, refCatalog } = props;
    return (
        <Card>
            <CardHeader className="flex gap-3">
                {catalog.spec.catalogType === 'config' && (
                    <FcSettings className="text-[#F6B519] text-3xl"/>
                )}
                {catalog.spec.catalogType === 'solution' && (
                    <FcTemplate className="text-[#F6B519] text-3xl"/>
                )}
                <div className="flex flex-col">
                <p className="text-md">{catalog.spec.name}</p>
                <p className="text-small text-default-500">{catalog.spec.catalogType}</p>
                </div>
            </CardHeader>
            <Divider/>
            <CardBody>
                {(catalog.spec.catalogType === 'config' && !catalog.spec.objectRef?.name) && (
                    <PropertyTable properties={catalog.spec.properties} refProperties={refCatalog?.spec.properties} />
                )}
                {(catalog.spec.catalogType === 'config' && catalog.spec.objectRef?.name) && (
                    <div style={{ whiteSpace: 'nowrap' , display: 'inline-flex', gap: '0.5rem', color: 'darkolivegreen'}}><FaGithub />{catalog.spec.objectRef.address}</div>                    
                )}
                {catalog.spec.catalogType === 'solution' && (
                    <SolutionSpecCard solution={catalog.spec.properties['spec']} />
                )}
            </CardBody>
            <Divider/>
            {catalog.spec.metadata?.['override'] && (
            
            <CardFooter>
                <span className='text-sm'>
                    <div>{`overrides: ${catalog.spec.metadata['override']}`}</div></span>
            </CardFooter>
            )}
        </Card>
    );
}
export default AssetCard;